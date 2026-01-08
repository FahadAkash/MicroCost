package config

import (
	"os"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg == nil {
		t.Fatal("DefaultConfig returned nil")
	}

	// Test analysis defaults
	if len(cfg.Analysis.Paths) == 0 {
		t.Error("Default analysis paths not set")
	}

	if cfg.Analysis.MaxDepth != 10 {
		t.Errorf("Expected max depth 10, got %d", cfg.Analysis.MaxDepth)
	}

	// Test Prometheus defaults
	if cfg.Prometheus.URL != "http://localhost:9090" {
		t.Errorf("Expected Prometheus URL 'http://localhost:9090', got '%s'", cfg.Prometheus.URL)
	}

	if cfg.Prometheus.Timeout != 30*time.Second {
		t.Errorf("Expected timeout 30s, got %v", cfg.Prometheus.Timeout)
	}

	// Test cost model defaults
	if cfg.CostModel.Provider != "aws" {
		t.Errorf("Expected provider 'aws', got '%s'", cfg.CostModel.Provider)
	}

	if cfg.CostModel.CPUCostPerCoreHour <= 0 {
		t.Error("CPU cost should be positive")
	}

	// Test output defaults
	if cfg.Output.Format != "ascii" {
		t.Errorf("Expected format 'ascii', got '%s'", cfg.Output.Format)
	}

	if cfg.Output.TopN != 10 {
		t.Errorf("Expected TopN 10, got %d", cfg.Output.TopN)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		modify  func(*Config)
		wantErr bool
	}{
		{
			name:    "valid config",
			modify:  func(c *Config) {},
			wantErr: false,
		},
		{
			name: "empty analysis paths",
			modify: func(c *Config) {
				c.Analysis.Paths = []string{}
			},
			wantErr: true,
		},
		{
			name: "empty prometheus URL",
			modify: func(c *Config) {
				c.Prometheus.URL = ""
			},
			wantErr: true,
		},
		{
			name: "empty cost provider",
			modify: func(c *Config) {
				c.CostModel.Provider = ""
			},
			wantErr: true,
		},
		{
			name: "negative TopN",
			modify: func(c *Config) {
				c.Output.TopN = -1
			},
			wantErr: false, // Should auto-correct to 10
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			tt.modify(cfg)

			err := cfg.Validate()

			if tt.wantErr && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestValidateAutoCorrect(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Output.TopN = -1
	cfg.Analysis.MaxDepth = -5

	err := cfg.Validate()

	if err != nil {
		t.Errorf("Validation failed: %v", err)
	}

	if cfg.Output.TopN != 10 {
		t.Errorf("Expected TopN to be corrected to 10, got %d", cfg.Output.TopN)
	}

	if cfg.Analysis.MaxDepth != 10 {
		t.Errorf("Expected MaxDepth to be corrected to 10, got %d", cfg.Analysis.MaxDepth)
	}
}

func TestLoadNonExistentFile(t *testing.T) {
	cfg, err := Load("/nonexistent/config.yaml")

	// If explicit path is provided and it doesn't exist, it should return an error
	if err == nil {
		t.Error("Expected error for non-existent explicit config path")
	}

	if cfg == nil {
		t.Error("Config should not be nil even with error")
	}
}

func TestEnvironmentVariableOverride(t *testing.T) {
	// Set environment variables
	os.Setenv("AWS_ACCESS_KEY_ID", "test-key-123")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "test-secret-456")
	defer os.Unsetenv("AWS_ACCESS_KEY_ID")
	defer os.Unsetenv("AWS_SECRET_ACCESS_KEY")

	cfg, err := Load("")

	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.AWS.AccessKeyID != "test-key-123" {
		t.Errorf("Expected AWS key 'test-key-123', got '%s'", cfg.AWS.AccessKeyID)
	}

	if cfg.AWS.SecretAccessKey != "test-secret-456" {
		t.Errorf("Expected AWS secret 'test-secret-456', got '%s'", cfg.AWS.SecretAccessKey)
	}
}
