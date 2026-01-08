package analyzer

import (
	"go/ast"
	"go/parser"
	"go/token"
	"regexp"
	"strings"

	"github.com/microcost/microcost/pkg/models"
	"github.com/sirupsen/logrus"
)

// HTTPDetector detects HTTP client calls in Go code
type HTTPDetector struct {
	logger       *logrus.Logger
	urlPatterns  []*regexp.Regexp
	dependencies []*models.Dependency
}

// NewHTTPDetector creates a new HTTP call detector
func NewHTTPDetector(logger *logrus.Logger) *HTTPDetector {
	// Patterns to extract service names from URLs
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`https?://([a-zA-Z0-9-]+)\.`),      // http://service.domain
		regexp.MustCompile(`https?://([a-zA-Z0-9-]+):[0-9]+`), // http://service:port
		regexp.MustCompile(`https?://([a-zA-Z0-9-]+)/`),       // http://service/
	}

	return &HTTPDetector{
		logger:       logger,
		urlPatterns:  patterns,
		dependencies: make([]*models.Dependency, 0),
	}
}

// DetectInFile detects HTTP calls in a Go source file
func (d *HTTPDetector) DetectInFile(filePath, serviceName string) ([]*models.Dependency, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	d.dependencies = make([]*models.Dependency, 0)

	ast.Inspect(node, func(n ast.Node) bool {
		d.inspectNode(n, fset, serviceName)
		return true
	})

	return d.dependencies, nil
}

// inspectNode inspects an AST node for HTTP calls
func (d *HTTPDetector) inspectNode(n ast.Node, fset *token.FileSet, fromService string) {
	callExpr, ok := n.(*ast.CallExpr)
	if !ok {
		return
	}

	// Check for http.Get, http.Post, http.Client.Do, etc.
	if d.isHTTPCall(callExpr) {
		url := d.extractURL(callExpr)
		if url != "" {
			targetService := d.extractServiceFromURL(url)
			endpoint := d.extractEndpointFromURL(url)

			pos := fset.Position(callExpr.Pos())

			dep := &models.Dependency{
				ID:          generateDependencyID(fromService, targetService, endpoint),
				FromService: fromService,
				ToService:   targetService,
				ToEndpoint:  endpoint,
				CallType:    "http",
				Weight:      1.0,
				DetectedAt:  pos.Filename,
				LineNumber:  pos.Line,
			}

			d.dependencies = append(d.dependencies, dep)
			d.logger.Debugf("Detected HTTP call: %s -> %s%s", fromService, targetService, endpoint)
		}
	}
}

// isHTTPCall checks if a call expression is an HTTP client call
func (d *HTTPDetector) isHTTPCall(call *ast.CallExpr) bool {
	switch fun := call.Fun.(type) {
	case *ast.SelectorExpr:
		// Check for http.Get, http.Post, etc.
		if ident, ok := fun.X.(*ast.Ident); ok {
			if ident.Name == "http" {
				method := fun.Sel.Name
				return method == "Get" || method == "Post" || method == "Put" ||
					method == "Delete" || method == "Head" || method == "Patch"
			}
		}

		// Check for client.Do()
		if fun.Sel.Name == "Do" || fun.Sel.Name == "Get" ||
			fun.Sel.Name == "Post" || fun.Sel.Name == "Put" {
			return true
		}
	}

	return false
}

// extractURL extracts the URL from an HTTP call
func (d *HTTPDetector) extractURL(call *ast.CallExpr) string {
	if len(call.Args) == 0 {
		return ""
	}

	// First argument is usually the URL
	arg := call.Args[0]

	// Handle string literals
	if lit, ok := arg.(*ast.BasicLit); ok {
		url := strings.Trim(lit.Value, `"`)
		return url
	}

	// Handle variables or constants (we can't resolve these at static analysis time)
	// In a production tool, you might use type information or constant evaluation

	return ""
}

// extractServiceFromURL extracts service name from URL
func (d *HTTPDetector) extractServiceFromURL(url string) string {
	for _, pattern := range d.urlPatterns {
		matches := pattern.FindStringSubmatch(url)
		if len(matches) > 1 {
			return matches[1]
		}
	}

	// Fallback: use the first part of the hostname
	parts := strings.Split(url, "/")
	if len(parts) > 2 {
		host := parts[2]
		if colonIdx := strings.Index(host, ":"); colonIdx != -1 {
			return host[:colonIdx]
		}
		if dotIdx := strings.Index(host, "."); dotIdx != -1 {
			return host[:dotIdx]
		}
		return host
	}

	return "unknown-service"
}

// extractEndpointFromURL extracts the endpoint path from URL
func (d *HTTPDetector) extractEndpointFromURL(url string) string {
	parts := strings.Split(url, "/")
	if len(parts) > 3 {
		return "/" + strings.Join(parts[3:], "/")
	}
	return "/"
}

// generateDependencyID generates a unique ID for a dependency
func generateDependencyID(from, to, endpoint string) string {
	return from + "->" + to + endpoint
}
