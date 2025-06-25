package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"
	"time"

	"github.com/joern1811/memory-bank/internal/domain"
	"github.com/spf13/cobra"
)

var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Manage projects",
	Long:  "Manage projects in the memory bank system.",
}

var projectListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all projects",
	Long:  "List all projects in the memory bank system.",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		// Initialize services
		services, err := GetServicesForCLI(cmd)
		if err != nil {
			return fmt.Errorf("failed to initialize services: %w", err)
		}

		// List projects
		projects, err := services.ProjectService.ListProjects(ctx)
		if err != nil {
			return fmt.Errorf("failed to list projects: %w", err)
		}

		// Display results
		fmt.Printf("\nProjects (%d found):\n", len(projects))
		if len(projects) == 0 {
			fmt.Println("No projects found. Use 'memory-bank init' to create a new project.")
			return nil
		}

		// Create table writer
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		if _, err := fmt.Fprintf(w, "ID\tName\tPath\tCreated\tDescription\n"); err != nil {
			return fmt.Errorf("failed to write table header: %w", err)
		}
		if _, err := fmt.Fprintf(w, "--\t----\t----\t-------\t-----------\n"); err != nil {
			return fmt.Errorf("failed to write table separator: %w", err)
		}

		for _, project := range projects {
			// Format creation time
			createdAt := project.CreatedAt.Format("2006-01-02 15:04")

			// Truncate description if too long
			description := project.Description
			if len(description) > 50 {
				description = description[:47] + "..."
			}

			if _, err := fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				project.ID,
				project.Name,
				project.Path,
				createdAt,
				description,
			); err != nil {
				return fmt.Errorf("failed to write project row: %w", err)
			}
		}

		if err := w.Flush(); err != nil {
			return fmt.Errorf("failed to flush table output: %w", err)
		}
		return nil
	},
}

var projectGetCmd = &cobra.Command{
	Use:   "get [project-id-or-path]",
	Short: "Get project information",
	Long:  "Get detailed information about a specific project by ID or path.",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		// Initialize services
		services, err := GetServicesForCLI(cmd)
		if err != nil {
			return fmt.Errorf("failed to initialize services: %w", err)
		}

		var project *domain.Project

		if len(args) > 0 {
			// Try to get by ID first, then by path
			projectID := domain.ProjectID(args[0])
			project, err = services.ProjectService.GetProject(ctx, projectID)
			if err != nil {
				// Try by path
				project, err = services.ProjectService.GetProjectByPath(ctx, args[0])
				if err != nil {
					return fmt.Errorf("project not found: %s", args[0])
				}
			}
		} else {
			// Use current directory
			currentDir, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current directory: %w", err)
			}

			project, err = services.ProjectService.GetProjectByPath(ctx, currentDir)
			if err != nil {
				return fmt.Errorf("no project found in current directory: %s", currentDir)
			}
		}

		// Display project information
		fmt.Printf("\nProject Information:\n")
		fmt.Printf("ID:          %s\n", project.ID)
		fmt.Printf("Name:        %s\n", project.Name)
		fmt.Printf("Path:        %s\n", project.Path)
		fmt.Printf("Description: %s\n", project.Description)
		fmt.Printf("Created:     %s\n", project.CreatedAt.Format(time.RFC3339))
		fmt.Printf("Updated:     %s\n", project.UpdatedAt.Format(time.RFC3339))

		// Get absolute path info
		absPath, err := filepath.Abs(project.Path)
		if err == nil {
			fmt.Printf("Absolute:    %s\n", absPath)
		}

		// Check if path exists
		if _, err := os.Stat(project.Path); os.IsNotExist(err) {
			fmt.Printf("Status:      ⚠️  Path does not exist\n")
		} else {
			fmt.Printf("Status:      ✅ Path exists\n")
		}

		return nil
	},
}

var projectDeleteCmd = &cobra.Command{
	Use:   "delete [project-id-or-path]",
	Short: "Delete a project",
	Long:  "Delete a project and all its associated memories by ID or path.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		// Initialize services
		services, err := GetServicesForCLI(cmd)
		if err != nil {
			return fmt.Errorf("failed to initialize services: %w", err)
		}

		// Get project first to show what will be deleted
		var project *domain.Project
		projectID := domain.ProjectID(args[0])
		project, err = services.ProjectService.GetProject(ctx, projectID)
		if err != nil {
			// Try by path
			project, err = services.ProjectService.GetProjectByPath(ctx, args[0])
			if err != nil {
				return fmt.Errorf("project not found: %s", args[0])
			}
		}

		// Show project info and ask for confirmation
		fmt.Printf("\nProject to be deleted:\n")
		fmt.Printf("ID:   %s\n", project.ID)
		fmt.Printf("Name: %s\n", project.Name)
		fmt.Printf("Path: %s\n", project.Path)
		fmt.Printf("\n⚠️  This will permanently delete the project and all its memories.\n")
		fmt.Printf("This action cannot be undone.\n\n")
		fmt.Printf("Are you sure? (y/N): ")

		var response string
		_, _ = fmt.Scanln(&response)

		if response != "y" && response != "Y" && response != "yes" && response != "YES" {
			fmt.Println("Deletion cancelled.")
			return nil
		}

		// Delete the project
		err = services.ProjectService.DeleteProject(ctx, project.ID)
		if err != nil {
			return fmt.Errorf("failed to delete project: %w", err)
		}

		fmt.Printf("✅ Project '%s' deleted successfully.\n", project.Name)
		return nil
	},
}

var projectUpdateCmd = &cobra.Command{
	Use:   "update [project-id-or-path]",
	Short: "Update project information",
	Long:  "Update project information such as name, description, or path.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		// Initialize services
		services, err := GetServicesForCLI(cmd)
		if err != nil {
			return fmt.Errorf("failed to initialize services: %w", err)
		}

		// Get project first
		var project *domain.Project
		projectID := domain.ProjectID(args[0])
		project, err = services.ProjectService.GetProject(ctx, projectID)
		if err != nil {
			// Try by path
			project, err = services.ProjectService.GetProjectByPath(ctx, args[0])
			if err != nil {
				return fmt.Errorf("project not found: %s", args[0])
			}
		}

		// Get flag values
		name, _ := cmd.Flags().GetString("name")
		description, _ := cmd.Flags().GetString("description")
		path, _ := cmd.Flags().GetString("path")

		// Check if any update flags were provided
		if name == "" && description == "" && path == "" {
			return fmt.Errorf("at least one field must be specified to update (--name, --description, or --path)")
		}

		// Update fields if provided
		if name != "" {
			project.Name = name
		}
		if description != "" {
			project.Description = description
		}
		if path != "" {
			// Validate path exists
			if _, err := os.Stat(path); os.IsNotExist(err) {
				fmt.Printf("⚠️  Warning: Path does not exist: %s\n", path)
				fmt.Printf("Continue anyway? (y/N): ")

				var response string
				_, _ = fmt.Scanln(&response)

				if response != "y" && response != "Y" && response != "yes" && response != "YES" {
					fmt.Println("Update cancelled.")
					return nil
				}
			}
			project.Path = path
		}

		// Update the project
		err = services.ProjectService.UpdateProject(ctx, project)
		if err != nil {
			return fmt.Errorf("failed to update project: %w", err)
		}

		fmt.Printf("✅ Project '%s' updated successfully.\n", project.Name)

		// Show updated information
		fmt.Printf("\nUpdated project information:\n")
		fmt.Printf("ID:          %s\n", project.ID)
		fmt.Printf("Name:        %s\n", project.Name)
		fmt.Printf("Path:        %s\n", project.Path)
		fmt.Printf("Description: %s\n", project.Description)

		return nil
	},
}

func init() {
	// Add flags to update command
	projectUpdateCmd.Flags().String("name", "", "New project name")
	projectUpdateCmd.Flags().String("description", "", "New project description")
	projectUpdateCmd.Flags().String("path", "", "New project path")

	// Add subcommands to project command
	projectCmd.AddCommand(projectListCmd)
	projectCmd.AddCommand(projectGetCmd)
	projectCmd.AddCommand(projectDeleteCmd)
	projectCmd.AddCommand(projectUpdateCmd)

	// Add project command to root
	rootCmd.AddCommand(projectCmd)
}
