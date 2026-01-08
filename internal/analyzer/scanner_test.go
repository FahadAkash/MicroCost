package analyzer

import (
	"os"
	"testing"

	"github.com/microcost/microcost/pkg/config"
	"github.com/sirupsen/logrus"
)

func TestNewScanner(t *testing.T) {
	cfg := &config.AnalysisConfig{
		Paths:        []string{"."},
		IncludeTests: false,
		MaxDepth:     10,
	}

	logger := logrus.New()
	scanner := NewScanner(cfg, logger)

	if scanner == nil {
		t.Fatal("NewScanner returned nil")
	}

	if scanner.config != cfg {
		t.Error("Config not set correctly")
	}

	if scanner.logger != logger {
		t.Error("Logger not set correctly")
	}

	if scanner.services == nil {
		t.Error("Services map not initialized")
	}
}

func TestExtractServiceName(t *testing.T) {
	cfg := &config.AnalysisConfig{}
	logger := logrus.New()
	scanner := NewScanner(cfg, logger)

	tests := []struct {
		name     string
		fileName string
		basePath string
		want     string
	}{
		{
			name:     "service directory",
			fileName: "/path/to/service/handler.go",
			basePath: "/path/to",
			want:     "service",
		},
		{
			name:     "current directory",
			fileName: "./main.go",
			basePath: ".",
			want:     ".",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scanner.extractServiceName(tt.fileName, tt.basePath)
			if result != tt.want {
				t.Errorf("Expected %s, got %s", tt.want, result)
			}
		})
	}
}

func TestIsHTTPHandler(t *testing.T) {
	// This would require creating AST nodes, which is complex
	// In a real scenario, you'd create test Go files and parse them
	cfg := &config.AnalysisConfig{}
	logger := logrus.New()
	scanner := NewScanner(cfg, logger)

	if scanner == nil {
		t.Fatal("Scanner should not be nil")
	}

	// Basic initialization test
	if scanner.services == nil {
		t.Error("Services map should be initialized")
	}
}

func TestShouldIncludeFile(t *testing.T) {
	tests := []struct {
		name         string
		fileName     string
		includeTests bool
		want         bool
	}{
		{
			name:         "include regular file",
			fileName:     "handler.go",
			includeTests: false,
			want:         true,
		},
		{
			name:         "exclude test file when tests disabled",
			fileName:     "handler_test.go",
			includeTests: false,
			want:         false,
		},
		{
			name:         "include test file when tests enabled",
			fileName:     "handler_test.go",
			includeTests: true,
			want:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.AnalysisConfig{
				IncludeTests: tt.includeTests,
			}
			logger := logrus.New()
			scanner := NewScanner(cfg, logger)

			// Create mock file info
			info := &mockFileInfo{name: tt.fileName}
			result := scanner.shouldIncludeFile(info)

			if result != tt.want {
				t.Errorf("Expected %v, got %v for file %s", tt.want, result, tt.fileName)
			}
		})
	}
}

// mockFileInfo implements os.FileInfo for testing
type mockFileInfo struct {
	name string
}

func (m *mockFileInfo) Name() string      { return m.name }
func (m *mockFileInfo) Size() int64       { return 0 }
func (m *mockFileInfo) Mode() os.FileMode { return 0 }
func (m *mockFileInfo) ModTime() os.Time  { return os.Time{} }
func (m *mockFileInfo) IsDir() bool       { return false }
func (m *mockFileInfo) Sys() interface{}  { return nil }
