package analyzer

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/microcost/microcost/pkg/config"
	"github.com/microcost/microcost/pkg/models"
	"github.com/sirupsen/logrus"
)

// Scanner scans Go source code to discover services and dependencies
type Scanner struct {
	config   *config.AnalysisConfig
	logger   *logrus.Logger
	services map[string]*models.Service
	fset     *token.FileSet
}

// NewScanner creates a new code scanner
func NewScanner(cfg *config.AnalysisConfig, logger *logrus.Logger) *Scanner {
	return &Scanner{
		config:   cfg,
		logger:   logger,
		services: make(map[string]*models.Service),
		fset:     token.NewFileSet(),
	}
}

// Scan scans the specified paths and returns discovered services
func (s *Scanner) Scan() (map[string]*models.Service, error) {
	s.logger.Info("Starting code scan...")

	for _, path := range s.config.Paths {
		if err := s.scanPath(path); err != nil {
			s.logger.WithError(err).Warnf("Error scanning path: %s", path)
			continue
		}
	}

	s.logger.Infof("Scan complete. Found %d services", len(s.services))
	return s.services, nil
}

// scanPath scans a single directory path
func (s *Scanner) scanPath(path string) error {
	s.logger.Debugf("Scanning path: %s", path)

	// Parse all Go files in the directory
	pkgs, err := parser.ParseDir(s.fset, path, s.shouldIncludeFile, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("error parsing directory: %w", err)
	}

	for pkgName, pkg := range pkgs {
		s.logger.Debugf("Analyzing package: %s", pkgName)
		s.analyzePackage(pkg, path)
	}

	return nil
}

// shouldIncludeFile determines if a file should be included in the scan
func (s *Scanner) shouldIncludeFile(info os.FileInfo) bool {
	// Skip test files unless configured to include them
	if !s.config.IncludeTests && strings.HasSuffix(info.Name(), "_test.go") {
		return false
	}
	return true
}

// analyzePackage analyzes a Go package to find services and handlers
func (s *Scanner) analyzePackage(pkg *ast.Package, basePath string) {
	for fileName, file := range pkg.Files {
		s.analyzeFile(file, fileName, basePath)
	}
}

// analyzeFile analyzes a single Go file
func (s *Scanner) analyzeFile(file *ast.File, fileName, basePath string) {
	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncDecl:
			s.analyzeFunction(node, fileName, basePath)
		case *ast.TypeSpec:
			s.analyzeTypeDecl(node, fileName, basePath)
		}
		return true
	})
}

// analyzeFunction analyzes a function declaration
func (s *Scanner) analyzeFunction(fn *ast.FuncDecl, fileName, basePath string) {
	if fn.Name == nil {
		return
	}

	funcName := fn.Name.Name

	// Check if this looks like an HTTP handler
	if s.isHTTPHandler(fn) {
		s.logger.Debugf("Found HTTP handler: %s in %s", funcName, fileName)
		s.registerEndpoint(funcName, "HTTP", fileName, basePath, fn)
	}

	// Check if this looks like a gRPC method
	if s.isGRPCMethod(fn) {
		s.logger.Debugf("Found gRPC method: %s in %s", funcName, fileName)
		s.registerEndpoint(funcName, "gRPC", fileName, basePath, fn)
	}
}

// analyzeTypeDecl analyzes type declarations to find services
func (s *Scanner) analyzeTypeDecl(typeSpec *ast.TypeSpec, fileName, basePath string) {
	if typeSpec.Name == nil {
		return
	}

	typeName := typeSpec.Name.Name

	// Check if this matches service patterns
	for _, pattern := range s.config.ServicePatterns {
		pattern = strings.ToLower(strings.ReplaceAll(pattern, "*", ""))
		if strings.Contains(strings.ToLower(typeName), pattern) {
			s.logger.Debugf("Found service type: %s in %s", typeName, fileName)
			s.registerService(typeName, fileName, basePath)
			break
		}
	}
}

// isHTTPHandler checks if a function is an HTTP handler
func (s *Scanner) isHTTPHandler(fn *ast.FuncDecl) bool {
	if fn.Type == nil || fn.Type.Params == nil {
		return false
	}

	// Check for http.ResponseWriter and *http.Request parameters
	for _, param := range fn.Type.Params.List {
		if selExpr, ok := param.Type.(*ast.SelectorExpr); ok {
			if ident, ok := selExpr.X.(*ast.Ident); ok {
				if ident.Name == "http" {
					return true
				}
			}
		}
	}

	return false
}

// isGRPCMethod checks if a function is a gRPC method
func (s *Scanner) isGRPCMethod(fn *ast.FuncDecl) bool {
	if fn.Type == nil || fn.Type.Params == nil {
		return false
	}

	// Check for context.Context parameter (common in gRPC)
	for _, param := range fn.Type.Params.List {
		if selExpr, ok := param.Type.(*ast.SelectorExpr); ok {
			if ident, ok := selExpr.X.(*ast.Ident); ok {
				if ident.Name == "context" {
					return true
				}
			}
		}
	}

	return false
}

// registerService registers a discovered service
func (s *Scanner) registerService(name, fileName, basePath string) {
	serviceName := s.extractServiceName(fileName, basePath)

	if _, exists := s.services[serviceName]; !exists {
		s.services[serviceName] = &models.Service{
			Name:         serviceName,
			Path:         basePath,
			Endpoints:    make([]*models.Endpoint, 0),
			Dependencies: make([]*models.Dependency, 0),
			Metadata:     map[string]string{"file": fileName},
		}
	}
}

// registerEndpoint registers a discovered endpoint
func (s *Scanner) registerEndpoint(funcName, endpointType, fileName, basePath string, fn *ast.FuncDecl) {
	serviceName := s.extractServiceName(fileName, basePath)

	// Ensure service exists
	if _, exists := s.services[serviceName]; !exists {
		s.registerService(serviceName, fileName, basePath)
	}

	service := s.services[serviceName]

	// Create endpoint
	endpoint := &models.Endpoint{
		Path:    "/" + strings.ToLower(funcName),
		Method:  "GET", // Default, can be refined with more analysis
		Service: service,
	}

	service.AddEndpoint(endpoint)
}

// extractServiceName extracts a service name from file path
func (s *Scanner) extractServiceName(fileName, basePath string) string {
	// Use directory name as service name
	dir := filepath.Dir(fileName)
	serviceName := filepath.Base(dir)

	if serviceName == "." || serviceName == "/" {
		serviceName = filepath.Base(basePath)
	}

	return serviceName
}

// GetServices returns all discovered services
func (s *Scanner) GetServices() map[string]*models.Service {
	return s.services
}
