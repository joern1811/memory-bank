package cli

import (
	"fmt"
	"os"

	"github.com/joern1811/memory-bank/internal/infra/config"
	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration management",
	Long:  `Manage Memory Bank configuration including creating default config files and validating settings.`,
}

// configInitCmd creates a default configuration file
var configInitCmd = &cobra.Command{
	Use:   "init [path]",
	Short: "Create default configuration file",
	Long:  `Create a default configuration file at the specified path or default location.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var configPath string
		if len(args) > 0 {
			configPath = args[0]
		} else {
			configPath = config.GetDefaultConfigPath()
		}

		// Check if file already exists
		if _, err := os.Stat(configPath); err == nil {
			overwrite, _ := cmd.Flags().GetBool("force")
			if !overwrite {
				return fmt.Errorf("config file already exists at %s (use --force to overwrite)", configPath)
			}
		}

		// Create default config
		if err := config.SaveDefaultConfig(configPath); err != nil {
			return fmt.Errorf("failed to create config file: %w", err)
		}

		fmt.Printf("Default configuration created at: %s\n", configPath)
		return nil
	},
}

// configValidateCmd validates the current configuration
var configValidateCmd = &cobra.Command{
	Use:   "validate [path]",
	Short: "Validate configuration file",
	Long:  `Validate the configuration file for syntax and required values.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var configPath string
		if len(args) > 0 {
			configPath = args[0]
		}

		// Load configuration
		cfg, err := config.LoadConfig(configPath)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Validate configuration
		if err := cfg.ValidateConfig(); err != nil {
			return fmt.Errorf("configuration validation failed: %w", err)
		}

		fmt.Println("Configuration is valid")
		return nil
	},
}

// configShowCmd displays the current configuration
var configShowCmd = &cobra.Command{
	Use:   "show [path]",
	Short: "Show current configuration",
	Long:  `Display the current configuration values including defaults and environment variables.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var configPath string
		if len(args) > 0 {
			configPath = args[0]
		}

		// Load configuration
		cfg, err := config.LoadConfig(configPath)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Display configuration
		fmt.Println("Current Memory Bank Configuration:")
		fmt.Printf("\nDatabase:")
		fmt.Printf("\n  Path: %s", cfg.Database.Path)

		fmt.Printf("\n\nOllama:")
		fmt.Printf("\n  Base URL: %s", cfg.Ollama.BaseURL)
		fmt.Printf("\n  Model: %s", cfg.Ollama.Model)
		fmt.Printf("\n  Timeout: %d seconds", cfg.Ollama.Timeout)

		fmt.Printf("\n\nChromaDB:")
		fmt.Printf("\n  Base URL: %s", cfg.ChromaDB.BaseURL)
		fmt.Printf("\n  Collection: %s", cfg.ChromaDB.Collection)
		fmt.Printf("\n  Tenant: %s", cfg.ChromaDB.Tenant)
		fmt.Printf("\n  Database: %s", cfg.ChromaDB.Database)
		fmt.Printf("\n  Data Path: %s", cfg.ChromaDB.DataPath)
		fmt.Printf("\n  Auto Start: %t", cfg.ChromaDB.AutoStart)
		fmt.Printf("\n  Timeout: %d seconds", cfg.ChromaDB.Timeout)

		fmt.Printf("\n\nLogging:")
		fmt.Printf("\n  Level: %s", cfg.Logging.Level)
		fmt.Printf("\n  Format: %s", cfg.Logging.Format)
		fmt.Printf("\n")

		return nil
	},
}

// configPathCmd shows the default configuration file path
var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Show default configuration path",
	Long:  `Display the default configuration file path that will be used.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		defaultPath := config.GetDefaultConfigPath()
		fmt.Printf("Default config path: %s\n", defaultPath)

		// Check if file exists
		if _, err := os.Stat(defaultPath); err == nil {
			fmt.Println("Config file exists")
		} else {
			fmt.Println("Config file does not exist (run 'memory-bank config init' to create)")
		}

		return nil
	},
}

func init() {
	// Add config subcommands
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configValidateCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configPathCmd)

	// Add flags
	configInitCmd.Flags().BoolP("force", "f", false, "Overwrite existing configuration file")

	// Add config command to root
	rootCmd.AddCommand(configCmd)
}
