package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/joern1811/memory-bank/internal/infra/embedding"
	"github.com/joern1811/memory-bank/internal/infra/vector"
	"github.com/spf13/cobra"
)

// HealthStatus represents the health status of a service
type HealthStatus struct {
	Service     string        `json:"service"`
	Status      string        `json:"status"`
	Available   bool          `json:"available"`
	ResponseTime time.Duration `json:"response_time"`
	Details     interface{}   `json:"details,omitempty"`
	Error       string        `json:"error,omitempty"`
}

// SystemHealth represents the overall system health
type SystemHealth struct {
	Overall     string         `json:"overall"`
	Timestamp   time.Time      `json:"timestamp"`
	Services    []HealthStatus `json:"services"`
	Configuration map[string]interface{} `json:"configuration"`
}

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Check system health and service connectivity",
	Long: `Check the health and connectivity status of all Memory Bank services including:
- Ollama embedding provider
- ChromaDB vector store
- Database connectivity
- Configuration validation`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get verbose flag
		verbose, _ := cmd.Flags().GetBool("verbose")
		json, _ := cmd.Flags().GetBool("json")
		
		return runHealthCheck(verbose, json)
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Alias for health command",
	Long:  `Alias for the health command - check system health and service connectivity.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get verbose flag
		verbose, _ := cmd.Flags().GetBool("verbose")
		json, _ := cmd.Flags().GetBool("json")
		
		return runHealthCheck(verbose, json)
	},
}

func runHealthCheck(verbose bool, jsonOutput bool) error {
	ctx := context.Background()
	
	// Load services for health checking (use quiet mode)
	services, err := NewServiceContainerQuiet()
	if err != nil {
		return fmt.Errorf("failed to initialize services: %w", err)
	}
	
	// Perform health checks
	systemHealth := checkSystemHealth(ctx, services)
	
	// Output results
	if jsonOutput {
		return outputHealthJSON(systemHealth)
	} else {
		return outputHealthText(systemHealth, verbose)
	}
}

func checkSystemHealth(ctx context.Context, services *ServiceContainer) *SystemHealth {
	health := &SystemHealth{
		Timestamp: time.Now(),
		Services:  make([]HealthStatus, 0),
		Configuration: map[string]interface{}{
			"database_path":      services.Config.Database.Path,
			"ollama_base_url":    services.Config.Ollama.BaseURL,
			"ollama_model":       services.Config.Ollama.Model,
			"chromadb_base_url":  services.Config.ChromaDB.BaseURL,
			"chromadb_collection": services.Config.ChromaDB.Collection,
			"chromadb_tenant":    services.Config.ChromaDB.Tenant,
			"chromadb_database":  services.Config.ChromaDB.Database,
		},
	}
	
	// Check Ollama health
	ollamaStatus := checkOllamaHealth(ctx, services)
	health.Services = append(health.Services, ollamaStatus)
	
	// Check ChromaDB health
	chromaStatus := checkChromaDBHealth(ctx, services)
	health.Services = append(health.Services, chromaStatus)
	
	// Check database health
	dbStatus := checkDatabaseHealth(ctx, services)
	health.Services = append(health.Services, dbStatus)
	
	// Determine overall status
	allHealthy := true
	for _, service := range health.Services {
		if !service.Available {
			allHealthy = false
			break
		}
	}
	
	if allHealthy {
		health.Overall = "healthy"
	} else {
		health.Overall = "degraded"
	}
	
	return health
}

func checkOllamaHealth(ctx context.Context, services *ServiceContainer) HealthStatus {
	status := HealthStatus{
		Service: "ollama",
		Status:  "unknown",
	}
	
	// Create Ollama provider for health checking
	ollamaConfig := embedding.OllamaConfig{
		BaseURL: services.Config.Ollama.BaseURL,
		Model:   services.Config.Ollama.Model,
	}
	ollamaProvider := embedding.NewOllamaProvider(ollamaConfig, services.Logger)
	
	// Measure response time
	start := time.Now()
	err := ollamaProvider.HealthCheck(ctx)
	responseTime := time.Since(start)
	
	status.ResponseTime = responseTime
	
	if err != nil {
		status.Status = "unhealthy"
		status.Available = false
		status.Error = err.Error()
		status.Details = map[string]interface{}{
			"base_url": services.Config.Ollama.BaseURL,
			"model":    services.Config.Ollama.Model,
			"fallback": "mock provider",
		}
	} else {
		status.Status = "healthy"
		status.Available = true
		status.Details = map[string]interface{}{
			"base_url":   services.Config.Ollama.BaseURL,
			"model":      services.Config.Ollama.Model,
			"dimensions": ollamaProvider.GetDimensions(),
		}
	}
	
	return status
}

func checkChromaDBHealth(ctx context.Context, services *ServiceContainer) HealthStatus {
	status := HealthStatus{
		Service: "chromadb",
		Status:  "unknown",
	}
	
	// Create ChromaDB store for health checking
	chromaConfig := vector.ChromaDBConfig{
		BaseURL:    services.Config.ChromaDB.BaseURL,
		Collection: services.Config.ChromaDB.Collection,
		Tenant:     services.Config.ChromaDB.Tenant,
		Database:   services.Config.ChromaDB.Database,
		Timeout:    time.Duration(services.Config.ChromaDB.Timeout) * time.Second,
	}
	chromaStore := vector.NewChromaDBVectorStore(chromaConfig, services.Logger)
	
	// Measure response time
	start := time.Now()
	err := chromaStore.HealthCheck(ctx)
	responseTime := time.Since(start)
	
	status.ResponseTime = responseTime
	
	if err != nil {
		status.Status = "unhealthy"
		status.Available = false
		status.Error = err.Error()
		status.Details = map[string]interface{}{
			"base_url":   services.Config.ChromaDB.BaseURL,
			"collection": services.Config.ChromaDB.Collection,
			"tenant":     services.Config.ChromaDB.Tenant,
			"database":   services.Config.ChromaDB.Database,
			"fallback":   "mock vector store",
		}
	} else {
		status.Status = "healthy"
		status.Available = true
		
		// Try to get additional details
		details := map[string]interface{}{
			"base_url":   services.Config.ChromaDB.BaseURL,
			"collection": services.Config.ChromaDB.Collection,
			"tenant":     services.Config.ChromaDB.Tenant,
			"database":   services.Config.ChromaDB.Database,
		}
		
		// Try to list collections to get more info
		if collections, err := chromaStore.ListCollections(ctx); err == nil {
			details["available_collections"] = collections
			details["collections_count"] = len(collections)
		}
		
		status.Details = details
	}
	
	return status
}

func checkDatabaseHealth(ctx context.Context, services *ServiceContainer) HealthStatus {
	status := HealthStatus{
		Service: "database",
		Status:  "unknown",
	}
	
	// We don't have direct access to the database from services, but we can
	// infer health by trying to perform a simple operation
	start := time.Now()
	
	// Try to list projects as a simple database health check
	_, err := services.ProjectService.ListProjects(ctx)
	responseTime := time.Since(start)
	
	status.ResponseTime = responseTime
	
	if err != nil {
		status.Status = "unhealthy"
		status.Available = false
		status.Error = err.Error()
		status.Details = map[string]interface{}{
			"path": services.Config.Database.Path,
			"type": "sqlite",
		}
	} else {
		status.Status = "healthy"
		status.Available = true
		status.Details = map[string]interface{}{
			"path": services.Config.Database.Path,
			"type": "sqlite",
		}
	}
	
	return status
}

func outputHealthJSON(health *SystemHealth) error {
	output, err := json.MarshalIndent(health, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to format health status as JSON: %w", err)
	}
	fmt.Println(string(output))
	return nil
}

func outputHealthText(health *SystemHealth, verbose bool) error {
	// Overall status
	fmt.Printf("System Health: %s\n", formatStatus(health.Overall))
	fmt.Printf("Checked at: %s\n\n", health.Timestamp.Format(time.RFC3339))
	
	// Service statuses
	fmt.Println("Services:")
	for _, service := range health.Services {
		fmt.Printf("  %s: %s", formatServiceName(service.Service), formatStatus(service.Status))
		
		if service.ResponseTime > 0 {
			fmt.Printf(" (%.2fms)", float64(service.ResponseTime.Nanoseconds())/1e6)
		}
		
		if !service.Available && service.Error != "" {
			fmt.Printf(" - %s", service.Error)
		}
		
		fmt.Println()
		
		// Show details in verbose mode
		if verbose && service.Details != nil {
			if details, ok := service.Details.(map[string]interface{}); ok {
				for key, value := range details {
					fmt.Printf("    %s: %v\n", key, value)
				}
			}
		}
	}
	
	// Configuration in verbose mode
	if verbose {
		fmt.Println("\nConfiguration:")
		for key, value := range health.Configuration {
			fmt.Printf("  %s: %v\n", key, value)
		}
	}
	
	return nil
}

func formatStatus(status string) string {
	switch status {
	case "healthy":
		return "✅ " + status
	case "unhealthy":
		return "❌ " + status
	case "degraded":
		return "⚠️  " + status
	default:
		return "❓ " + status
	}
}

func formatServiceName(service string) string {
	switch service {
	case "ollama":
		return "Ollama     "
	case "chromadb":
		return "ChromaDB   "
	case "database":
		return "Database   "
	default:
		return service + "     "
	}
}

func init() {
	// Add flags
	healthCmd.Flags().BoolP("json", "j", false, "output health status as JSON")
	statusCmd.Flags().BoolP("json", "j", false, "output health status as JSON")
	
	// Register commands
	rootCmd.AddCommand(healthCmd)
	rootCmd.AddCommand(statusCmd)
}