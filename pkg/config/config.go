package config

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	Analysis   AnalysisConfig   `mapstructure:"analysis"`
	Prometheus PrometheusConfig `mapstructure:"prometheus"`
	CostModel  CostModelConfig  `mapstructure:"cost_model"`
	AWS        AWSConfig        `mapstructure:"aws"`
	Output     OutputConfig     `mapstructure:"output"`
	Server     ServerConfig     `mapstructure:"server"`
	Logging    LoggingConfig    `mapstructure:"logging"`
}

// AnalysisConfig contains static analysis settings
type AnalysisConfig struct {
	Paths           []string `mapstructure:"paths"`
	Excludes        []string `mapstructure:"excludes"`
	IncludeTests    bool     `mapstructure:"include_tests"`
	FollowImports   bool     `mapstructure:"follow_imports"`
	MaxDepth        int      `mapstructure:"max_depth"`
	ServicePatterns []string `mapstructure:"service_patterns"`
}

// PrometheusConfig contains Prometheus connection settings
type PrometheusConfig struct {
	URL            string            `mapstructure:"url"`
	Timeout        time.Duration     `mapstructure:"timeout"`
	QueryInterval  time.Duration     `mapstructure:"query_interval"`
	LookbackWindow time.Duration     `mapstructure:"lookback_window"`
	CustomQueries  map[string]string `mapstructure:"custom_queries"`
}

// CostModelConfig contains cost calculation settings
type CostModelConfig struct {
	Provider            string  `mapstructure:"provider"`
	Region              string  `mapstructure:"region"`
	CPUCostPerCoreHour  float64 `mapstructure:"cpu_cost_per_core_hour"`
	MemoryCostPerGBHour float64 `mapstructure:"memory_cost_per_gb_hour"`
	NetworkCostPerGB    float64 `mapstructure:"network_cost_per_gb"`
	DiskCostPerGBHour   float64 `mapstructure:"disk_cost_per_gb_hour"`
	RequestCost         float64 `mapstructure:"request_cost"`
}

// AWSConfig contains AWS-specific settings
type AWSConfig struct {
	Region          string `mapstructure:"region"`
	AccessKeyID     string `mapstructure:"access_key_id"`
	SecretAccessKey string `mapstructure:"secret_access_key"`
	ProfileName     string `mapstructure:"profile_name"`
	UseCostExplorer bool   `mapstructure:"use_cost_explorer"`
}

// OutputConfig contains output formatting settings
type OutputConfig struct {
	Format         string `mapstructure:"format"` // ascii, json, yaml, html
	OutputPath     string `mapstructure:"output_path"`
	IncludeMetrics bool   `mapstructure:"include_metrics"`
	IncludeCosts   bool   `mapstructure:"include_costs"`
	TopN           int    `mapstructure:"top_n"` // top N costly endpoints
	ColorEnabled   bool   `mapstructure:"color_enabled"`
}

// ServerConfig contains web server settings
type ServerConfig struct {
	Port            int           `mapstructure:"port"`
	Host            string        `mapstructure:"host"`
	RefreshInterval time.Duration `mapstructure:"refresh_interval"`
	EnableCORS      bool          `mapstructure:"enable_cors"`
}

// LoggingConfig contains logging settings
type LoggingConfig struct {
	Level      string `mapstructure:"level"`  // debug, info, warn, error
	Format     string `mapstructure:"format"` // text, json
	OutputFile string `mapstructure:"output_file"`
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Analysis: AnalysisConfig{
			Paths:           []string{"./"},
			Excludes:        []string{"vendor", "node_modules", ".git"},
			IncludeTests:    false,
			FollowImports:   true,
			MaxDepth:        10,
			ServicePatterns: []string{"*service*", "*handler*", "*controller*"},
		},
		Prometheus: PrometheusConfig{
			URL:            "http://localhost:9090",
			Timeout:        30 * time.Second,
			QueryInterval:  1 * time.Minute,
			LookbackWindow: 1 * time.Hour,
			CustomQueries:  make(map[string]string),
		},
		CostModel: CostModelConfig{
			Provider:            "aws",
			Region:              "us-east-1",
			CPUCostPerCoreHour:  0.0416, // t3.medium equivalent
			MemoryCostPerGBHour: 0.0052,
			NetworkCostPerGB:    0.09,
			DiskCostPerGBHour:   0.10,
			RequestCost:         0.0000002,
		},
		AWS: AWSConfig{
			Region:          "us-east-1",
			ProfileName:     "default",
			UseCostExplorer: false,
		},
		Output: OutputConfig{
			Format:         "ascii",
			OutputPath:     "./output",
			IncludeMetrics: true,
			IncludeCosts:   true,
			TopN:           10,
			ColorEnabled:   true,
		},
		Server: ServerConfig{
			Port:            8080,
			Host:            "localhost",
			RefreshInterval: 5 * time.Minute,
			EnableCORS:      true,
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "text",
		},
	}
}

// Load loads configuration from file and environment variables
func Load(configPath string) (*Config, error) {
	cfg := DefaultConfig()

	v := viper.New()
	v.SetConfigType("yaml")

	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("config")
		v.AddConfigPath(".")
		v.AddConfigPath("$HOME/.microcost")
		v.AddConfigPath("/etc/microcost")
	}

	// Read from environment variables
	v.SetEnvPrefix("MICROCOST")
	v.AutomaticEnv()

	// Try to read config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return cfg, fmt.Errorf("error reading config file: %w", err)
		}
		// Config file not found; use defaults
	}

	// Unmarshal into config struct
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Override with environment variables for sensitive data
	if awsKey := os.Getenv("AWS_ACCESS_KEY_ID"); awsKey != "" {
		cfg.AWS.AccessKeyID = awsKey
	}
	if awsSecret := os.Getenv("AWS_SECRET_ACCESS_KEY"); awsSecret != "" {
		cfg.AWS.SecretAccessKey = awsSecret
	}

	return cfg, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if len(c.Analysis.Paths) == 0 {
		return fmt.Errorf("analysis paths cannot be empty")
	}

	if c.Prometheus.URL == "" {
		return fmt.Errorf("prometheus URL is required")
	}

	if c.CostModel.Provider == "" {
		return fmt.Errorf("cost model provider is required")
	}

	if c.Output.TopN < 1 {
		c.Output.TopN = 10
	}

	if c.Analysis.MaxDepth < 1 {
		c.Analysis.MaxDepth = 10
	}

	return nil
}

// Save saves the configuration to a file
func (c *Config) Save(path string) error {
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetConfigFile(path)

	// Marshal config to map
	cfg := map[string]interface{}{
		"analysis":   c.Analysis,
		"prometheus": c.Prometheus,
		"cost_model": c.CostModel,
		"aws":        c.AWS,
		"output":     c.Output,
		"server":     c.Server,
		"logging":    c.Logging,
	}

	for key, value := range cfg {
		v.Set(key, value)
	}

	return v.WriteConfig()
}
