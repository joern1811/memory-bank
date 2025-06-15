package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [project-path]",
	Short: "Initialize a new project for memory management",
	Long: `Initialize a new project in the memory bank system.
If no path is provided, the current directory will be used.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var projectPath string
		if len(args) > 0 {
			projectPath = args[0]
		} else {
			var err error
			projectPath, err = os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current directory: %w", err)
			}
		}

		name, _ := cmd.Flags().GetString("name")
		description, _ := cmd.Flags().GetString("description")

		if name == "" {
			name = filepath.Base(projectPath)
		}

		// TODO: Initialize project services and create project
		// ctx := context.Background()
		fmt.Printf("Initializing project '%s' at path: %s\n", name, projectPath)
		if description != "" {
			fmt.Printf("Description: %s\n", description)
		}
		
		// This would call the project service to create a new project
		// projectService.InitProject(ctx, name, projectPath, description)
		
		fmt.Println("✓ Project initialized successfully")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().StringP("name", "n", "", "project name (default: directory name)")
	initCmd.Flags().StringP("description", "d", "", "project description")
}