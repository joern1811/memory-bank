package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "memory-bank",
	Short: "Semantic memory management system for development knowledge",
	Long: `Memory Bank is a semantic memory management system for Claude Code using hexagonal architecture.
It provides intelligent storage and retrieval of development knowledge including decisions, patterns, 
error solutions, and session context.`,
	Run: func(cmd *cobra.Command, args []string) {
		// If no subcommand is provided, show help
		if err := cmd.Help(); err != nil {
			fmt.Printf("Error displaying help: %v\n", err)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// Global flags can be added here
	rootCmd.PersistentFlags().StringP("config", "c", "", "config file (default is $HOME/.config/memory-bank/config.yaml)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")

	// Add version flag to root command
	rootCmd.Flags().BoolP("version", "V", false, "show version")
	rootCmd.Run = func(cmd *cobra.Command, args []string) {
		if versionFlag, _ := cmd.Flags().GetBool("version"); versionFlag {
			versionCmd.Run(cmd, args)
			return
		}
		// If no subcommand is provided, show help
		if err := cmd.Help(); err != nil {
			fmt.Printf("Error displaying help: %v\n", err)
		}
	}
}
