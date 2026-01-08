package analyzer

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"

	"github.com/microcost/microcost/pkg/models"
	"github.com/sirupsen/logrus"
)

// GRPCDetector detects gRPC client calls in Go code
type GRPCDetector struct {
	logger       *logrus.Logger
	dependencies []*models.Dependency
}

// NewGRPCDetector creates a new gRPC call detector
func NewGRPCDetector(logger *logrus.Logger) *GRPCDetector {
	return &GRPCDetector{
		logger:       logger,
		dependencies: make([]*models.Dependency, 0),
	}
}

// DetectInFile detects gRPC calls in a Go source file
func (d *GRPCDetector) DetectInFile(filePath, serviceName string) ([]*models.Dependency, error) {
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

// inspectNode inspects an AST node for gRPC calls
func (d *GRPCDetector) inspectNode(n ast.Node, fset *token.FileSet, fromService string) {
	callExpr, ok := n.(*ast.CallExpr)
	if !ok {
		return
	}

	// Check for gRPC client stub method calls
	if d.isGRPCCall(callExpr) {
		targetService, method := d.extractGRPCInfo(callExpr)

		if targetService != "" {
			pos := fset.Position(callExpr.Pos())

			dep := &models.Dependency{
				ID:          generateDependencyID(fromService, targetService, "/"+method),
				FromService: fromService,
				ToService:   targetService,
				ToEndpoint:  "/" + method,
				CallType:    "grpc",
				Weight:      1.0,
				DetectedAt:  pos.Filename,
				LineNumber:  pos.Line,
			}

			d.dependencies = append(d.dependencies, dep)
			d.logger.Debugf("Detected gRPC call: %s -> %s.%s", fromService, targetService, method)
		}
	}
}

// isGRPCCall checks if a call expression is a gRPC client call
func (d *GRPCDetector) isGRPCCall(call *ast.CallExpr) bool {
	selExpr, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	// Check if it's calling a method on a client
	// Common patterns: client.GetUser(), userClient.CreateUser(), etc.
	if ident, ok := selExpr.X.(*ast.Ident); ok {
		clientName := strings.ToLower(ident.Name)
		// Check if variable name contains "client" or "stub"
		if strings.Contains(clientName, "client") || strings.Contains(clientName, "stub") {
			return true
		}
	}

	return false
}

// extractGRPCInfo extracts service and method information from gRPC call
func (d *GRPCDetector) extractGRPCInfo(call *ast.CallExpr) (string, string) {
	selExpr, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return "", ""
	}

	method := selExpr.Sel.Name

	// Try to extract service name from client variable
	var serviceName string
	if ident, ok := selExpr.X.(*ast.Ident); ok {
		clientName := ident.Name
		// Remove common suffixes/prefixes
		serviceName = d.extractServiceFromClientName(clientName)
	}

	return serviceName, method
}

// extractServiceFromClientName extracts service name from client variable name
func (d *GRPCDetector) extractServiceFromClientName(clientName string) string {
	// Remove common prefixes and suffixes
	name := clientName
	name = strings.TrimSuffix(name, "Client")
	name = strings.TrimSuffix(name, "Stub")
	name = strings.TrimPrefix(name, "new")
	name = strings.TrimPrefix(name, "New")

	// Convert to lowercase
	return strings.ToLower(name)
}
