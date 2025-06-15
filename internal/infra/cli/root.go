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
		cmd.Help()
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
	rootCmd.PersistentFlags().StringP("config", "c", "", "config file (default is $HOME/.memory-bank.yaml)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
}