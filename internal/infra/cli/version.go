package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// These will be set by goreleaser at build time
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of memory-bank",
	Long:  `Print the version number, commit hash, and build date of memory-bank`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("memory-bank %s\n", version)
		fmt.Printf("Commit: %s\n", commit)
		fmt.Printf("Built: %s\n", date)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
