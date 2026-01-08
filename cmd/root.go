package cmd

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	logger  *logrus.Logger
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "microcost",
	Short: "Microservices Dependency & Cost Mapper",
	Long: `A production-ready CLI tool that maps microservices architecture and calculates 
the true cost of each API endpoint by tracing dependencies across service boundaries.

Features:
  • Static code analysis to build dependency graphs
  • Runtime metrics collection from Prometheus
  • Cost attribution across service call chains
  • Interactive visualization and reporting`,
	Version: "1.0.0",
}

// Execute adds all child commands to the root command and sets flags appropriately
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig, initLogger)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./config.yaml)")
	rootCmd.PersistentFlags().String("log-level", "info", "log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().String("log-format", "text", "log format (text, json)")

	viper.BindPFlag("logging.level", rootCmd.PersistentFlags().Lookup("log-level"))
	viper.BindPFlag("logging.format", rootCmd.PersistentFlags().Lookup("log-format"))
}

// initConfig reads in config file and ENV variables
func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("config")
		viper.AddConfigPath(".")
		viper.AddConfigPath("$HOME/.microcost")
		viper.AddConfigPath("/etc/microcost")
	}

	viper.SetEnvPrefix("MICROCOST")
	viper.AutomaticEnv()

	// Read config file (not required)
	viper.ReadInConfig()
}

// initLogger initializes the logger
func initLogger() {
	logger = logrus.New()

	// Set log level
	level := viper.GetString("logging.level")
	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		logLevel = logrus.InfoLevel
	}
	logger.SetLevel(logLevel)

	// Set log format
	if viper.GetString("logging.format") == "json" {
		logger.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	}
}

// GetLogger returns the logger instance
func GetLogger() *logrus.Logger {
	if logger == nil {
		initLogger()
	}
	return logger
}
