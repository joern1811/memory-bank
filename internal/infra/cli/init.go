package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/joern1811/memory-bank/internal/ports"
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
		language, _ := cmd.Flags().GetString("language")
		framework, _ := cmd.Flags().GetString("framework")

		if name == "" {
			name = filepath.Base(projectPath)
		}

		// Get services
		services, err := GetServicesForCLI(cmd)
		if err != nil {
			return fmt.Errorf("failed to initialize services: %w", err)
		}

		// Create initialization request
		req := ports.InitializeProjectRequest{
			Name:        name,
			Description: description,
			Language:    language,
			Framework:   framework,
		}

		// Initialize project using the service
		ctx := context.Background()
		project, err := services.ProjectService.InitializeProject(ctx, projectPath, req)
		if err != nil {
			return fmt.Errorf("failed to initialize project: %w", err)
		}

		// Display results
		fmt.Printf("✓ Project initialized successfully\n")
		fmt.Printf("  ID: %s\n", project.ID)
		fmt.Printf("  Name: %s\n", project.Name)
		fmt.Printf("  Path: %s\n", project.Path)
		if project.Description != "" {
			fmt.Printf("  Description: %s\n", project.Description)
		}
		if project.Language != "" {
			fmt.Printf("  Language: %s\n", project.Language)
		}
		if project.Framework != "" {
			fmt.Printf("  Framework: %s\n", project.Framework)
		}
		fmt.Printf("  Created: %s\n", project.CreatedAt.Format("2006-01-02 15:04:05"))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().StringP("name", "n", "", "project name (default: directory name)")
	initCmd.Flags().StringP("description", "d", "", "project description")
	initCmd.Flags().StringP("language", "l", "", "programming language (auto-detected if not specified)")
	initCmd.Flags().StringP("framework", "f", "", "framework used (auto-detected if not specified)")
}