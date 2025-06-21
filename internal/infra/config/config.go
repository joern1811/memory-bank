package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Config holds the application configuration
type Config struct {
	Database Database `mapstructure:"database" yaml:"database" json:"database"`
	Ollama   Ollama   `mapstructure:"ollama" yaml:"ollama" json:"ollama"`
	ChromaDB ChromaDB `mapstructure:"chromadb" yaml:"chromadb" json:"chromadb"`
	Logging  Logging  `mapstructure:"logging" yaml:"logging" json:"logging"`
}

// Database configuration
type Database struct {
	Path string `mapstructure:"path" yaml:"path" json:"path"`
}

// Ollama configuration
type Ollama struct {
	BaseURL string `mapstructure:"base_url" yaml:"base_url" json:"base_url"`
	Model   string `mapstructure:"model" yaml:"model" json:"model"`
	Timeout int    `mapstructure:"timeout" yaml:"timeout" json:"timeout"` // seconds
}

// ChromaDB configuration
type ChromaDB struct {
	BaseURL    string `mapstructure:"base_url" yaml:"base_url" json:"base_url"`
	Collection string `mapstructure:"collection" yaml:"collection" json:"collection"`
	Tenant     string `mapstructure:"tenant" yaml:"tenant" json:"tenant"`
	Database   string `mapstructure:"database" yaml:"database" json:"database"`
	Timeout    int    `mapstructure:"timeout" yaml:"timeout" json:"timeout"` // seconds
	DataPath   string `mapstructure:"data_path" yaml:"data_path" json:"data_path"`
	AutoStart  bool   `mapstructure:"auto_start" yaml:"auto_start" json:"auto_start"`
}

// Logging configuration
type Logging struct {
	Level  string `mapstructure:"level" yaml:"level" json:"level"`
	Format string `mapstructure:"format" yaml:"format" json:"format"` // "json" or "text"
}

// LoadConfig loads configuration from file and environment variables
func LoadConfig(configPath string) (*Config, error) {
	// Set defaults
	viper.SetDefault("database.path", "./memory_bank.db")
	viper.SetDefault("ollama.base_url", "http://localhost:11434")
	viper.SetDefault("ollama.model", "nomic-embed-text")
	viper.SetDefault("ollama.timeout", 30)
	viper.SetDefault("chromadb.base_url", "http://localhost:8000")
	viper.SetDefault("chromadb.collection", "memory_bank")
	viper.SetDefault("chromadb.tenant", "default_tenant")
	viper.SetDefault("chromadb.database", "default_database")
	viper.SetDefault("chromadb.timeout", 30)
	viper.SetDefault("chromadb.data_path", "./chromadb_data")
	viper.SetDefault("chromadb.auto_start", false)
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")

	// Configure viper
	viper.SetConfigType("yaml")
	viper.SetEnvPrefix("MEMORY_BANK")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Load config file if provided
	if configPath != "" {
		viper.SetConfigFile(configPath)
		if err := viper.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
		}
	} else {
		// Try to find config file in standard locations
		if err := findAndLoadConfig(); err != nil {
			// Config file not found is not an error - we'll use defaults and env vars
			if !isConfigNotFoundError(err) {
				return nil, fmt.Errorf("failed to load config: %w", err)
			}
		}
	}

	// Override with legacy environment variables for backward compatibility
	overrideWithLegacyEnvVars()

	// Unmarshal config
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

// findAndLoadConfig searches for config files in standard locations
func findAndLoadConfig() error {
	// Search paths in order of preference
	searchPaths := []string{
		".",                         // Current directory
		"$HOME/.config/memory-bank", // User config directory
		"/etc/memory-bank",          // System config directory
	}

	// Config file names to search for
	configNames := []string{
		"memory-bank",
		".memory-bank",
		"config",
	}

	for _, path := range searchPaths {
		expandedPath := os.ExpandEnv(path)
		viper.AddConfigPath(expandedPath)
	}

	for _, name := range configNames {
		viper.SetConfigName(name)
		if err := viper.ReadInConfig(); err == nil {
			return nil // Found and loaded config
		}
	}

	return fmt.Errorf("config file not found in search paths")
}

// overrideWithLegacyEnvVars maintains backward compatibility with existing env vars
func overrideWithLegacyEnvVars() {
	// Legacy environment variables mapping
	legacyMappings := map[string]string{
		"MEMORY_BANK_DB_PATH":    "database.path",
		"OLLAMA_BASE_URL":        "ollama.base_url",
		"OLLAMA_MODEL":           "ollama.model",
		"CHROMADB_BASE_URL":      "chromadb.base_url",
		"CHROMADB_COLLECTION":    "chromadb.collection",
		"CHROMADB_DATA_PATH":     "chromadb.data_path",
		"CHROMADB_AUTO_START":    "chromadb.auto_start",
	}

	for envVar, configKey := range legacyMappings {
		if value := os.Getenv(envVar); value != "" {
			viper.Set(configKey, value)
		}
	}
}

// isConfigNotFoundError checks if the error is due to config file not found
func isConfigNotFoundError(err error) bool {
	return strings.Contains(err.Error(), "config file not found")
}

// SaveDefaultConfig creates a default configuration file
func SaveDefaultConfig(configPath string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Default configuration content
	defaultConfig := `# Memory Bank Configuration

database:
  path: "./memory_bank.db"

ollama:
  base_url: "http://localhost:11434"
  model: "nomic-embed-text"
  timeout: 30

chromadb:
  base_url: "http://localhost:8000"
  collection: "memory_bank"
  timeout: 30

logging:
  level: "info"    # debug, info, warn, error
  format: "json"   # json, text
`

	// Write configuration file
	if err := os.WriteFile(configPath, []byte(defaultConfig), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetDefaultConfigPath returns the default configuration file path
func GetDefaultConfigPath() string {
	// Try user config directory first
	if configDir, err := os.UserConfigDir(); err == nil {
		return filepath.Join(configDir, "memory-bank", "config.yaml")
	}

	// Fall back to home directory
	if homeDir, err := os.UserHomeDir(); err == nil {
		return filepath.Join(homeDir, ".memory-bank.yaml")
	}

	// Last resort: current directory
	return "./memory-bank.yaml"
}

// ValidateConfig performs basic validation on the configuration
func (c *Config) ValidateConfig() error {
	// Validate database path
	if c.Database.Path == "" {
		return fmt.Errorf("database path cannot be empty")
	}

	// Validate Ollama configuration
	if c.Ollama.BaseURL == "" {
		return fmt.Errorf("Ollama base URL cannot be empty")
	}
	if c.Ollama.Model == "" {
		return fmt.Errorf("Ollama model cannot be empty")
	}
	if c.Ollama.Timeout <= 0 {
		return fmt.Errorf("Ollama timeout must be positive")
	}

	// Validate ChromaDB configuration
	if c.ChromaDB.BaseURL == "" {
		return fmt.Errorf("ChromaDB base URL cannot be empty")
	}
	if c.ChromaDB.Collection == "" {
		return fmt.Errorf("ChromaDB collection cannot be empty")
	}
	if c.ChromaDB.Timeout <= 0 {
		return fmt.Errorf("ChromaDB timeout must be positive")
	}

	// Validate logging configuration
	validLogLevels := map[string]bool{
		"debug": true, "info": true, "warn": true, "error": true,
	}
	if !validLogLevels[c.Logging.Level] {
		return fmt.Errorf("invalid log level: %s (valid: debug, info, warn, error)", c.Logging.Level)
	}

	validLogFormats := map[string]bool{
		"json": true, "text": true,
	}
	if !validLogFormats[c.Logging.Format] {
		return fmt.Errorf("invalid log format: %s (valid: json, text)", c.Logging.Format)
	}

	return nil
}
