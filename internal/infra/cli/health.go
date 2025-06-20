package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
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
- Configuration validation

Use --watch to continuously monitor health status.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get flags
		verbose, _ := cmd.Flags().GetBool("verbose")
		json, _ := cmd.Flags().GetBool("json")
		watch, _ := cmd.Flags().GetBool("watch")
		
		if watch {
			return runWatchMode(verbose, json)
		}
		
		return runHealthCheck(verbose, json)
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Alias for health command",
	Long:  `Alias for the health command - check system health and service connectivity.

Use --watch to continuously monitor health status.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get flags
		verbose, _ := cmd.Flags().GetBool("verbose")
		json, _ := cmd.Flags().GetBool("json")
		watch, _ := cmd.Flags().GetBool("watch")
		
		if watch {
			return runWatchMode(verbose, json)
		}
		
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

func runWatchMode(verbose bool, jsonOutput bool) error {
	// Set up signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	go func() {
		<-sigChan
		fmt.Println("\nReceived interrupt signal, stopping health monitoring...")
		cancel()
	}()
	
	// Load services once
	services, err := NewServiceContainerQuiet()
	if err != nil {
		return fmt.Errorf("failed to initialize services: %w", err)
	}
	
	fmt.Println("🔍 Memory Bank Health Monitor - Press Ctrl+C to stop")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	
	// Initial check
	if err := performHealthCheck(ctx, services, verbose, jsonOutput); err != nil {
		return err
	}
	
	// Watch mode loop
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Health monitoring stopped.")
			return nil
		case <-ticker.C:
			// Clear previous output (only in non-JSON mode)
			if !jsonOutput {
				fmt.Print("\033[H\033[2J") // Clear screen
				fmt.Println("🔍 Memory Bank Health Monitor - Press Ctrl+C to stop")
				fmt.Printf("Last updated: %s\n", time.Now().Format("15:04:05"))
				fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
			}
			
			if err := performHealthCheck(ctx, services, verbose, jsonOutput); err != nil {
				if !jsonOutput {
					fmt.Printf("Error performing health check: %v\n", err)
				}
				continue
			}
		}
	}
}

func performHealthCheck(ctx context.Context, services *ServiceContainer, verbose bool, jsonOutput bool) error {
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
	
	// Try health check with retry logic
	const maxRetries = 3
	var lastError error
	
	for retry := 0; retry < maxRetries; retry++ {
		if retry > 0 {
			fmt.Printf("  Retrying ChromaDB connection... (%d/%d)\n", retry+1, maxRetries)
			time.Sleep(time.Duration(retry) * 500 * time.Millisecond) // Progressive backoff
		}
		
		// Measure response time
		start := time.Now()
		err := chromaStore.HealthCheck(ctx)
		responseTime := time.Since(start)
		
		status.ResponseTime = responseTime
		lastError = err
		
		if err == nil {
			// Success!
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
			return status
		}
		
		// Continue retry loop for transient errors
		if retry < maxRetries-1 && isTransientError(err) {
			continue
		}
		
		// Final failure or non-transient error
		break
	}
	
	// Health check failed after retries
	status.Status = "unhealthy"
	status.Available = false
	status.Error = enhanceChromaDBError(lastError, services.Config.ChromaDB.BaseURL)
	status.Details = map[string]interface{}{
		"base_url":     services.Config.ChromaDB.BaseURL,
		"collection":   services.Config.ChromaDB.Collection,
		"tenant":       services.Config.ChromaDB.Tenant,
		"database":     services.Config.ChromaDB.Database,
		"fallback":     "mock vector store",
		"retry_count":  maxRetries,
		"setup_hints":  getChromaDBSetupHints(),
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

// isTransientError determines if an error is likely transient and worth retrying
func isTransientError(err error) bool {
	if err == nil {
		return false
	}
	
	errorStr := strings.ToLower(err.Error())
	
	// Common transient error patterns
	transientPatterns := []string{
		"connection refused",
		"timeout",
		"temporary failure",
		"network is unreachable",
		"connection reset",
		"no route to host",
	}
	
	for _, pattern := range transientPatterns {
		if strings.Contains(errorStr, pattern) {
			return true
		}
	}
	
	return false
}

// enhanceChromaDBError provides enhanced error messages with specific setup hints
func enhanceChromaDBError(err error, baseURL string) string {
	if err == nil {
		return ""
	}
	
	originalError := err.Error()
	errorStr := strings.ToLower(originalError)
	
	// Categorize error and provide specific hints
	if strings.Contains(errorStr, "connection refused") {
		return fmt.Sprintf("%s\n\nSetup Hint: ChromaDB is not running. Start it with:\n"+
			"  Option 1 (uvx): uvx --from 'chromadb[server]' chroma run --host localhost --port 8000 --path ./chromadb_data\n"+
			"  Option 2 (docker): docker run -p 8000:8000 -v ./chromadb_data:/chroma/chroma chromadb/chroma\n"+
			"  Then verify with: curl %s/api/v2/heartbeat", originalError, baseURL)
	}
	
	if strings.Contains(errorStr, "timeout") {
		return fmt.Sprintf("%s\n\nSetup Hint: ChromaDB appears to be slow or overloaded.\n"+
			"  - Check if ChromaDB is under heavy load\n"+
			"  - Increase timeout in configuration\n"+
			"  - Verify network connectivity to %s", originalError, baseURL)
	}
	
	if strings.Contains(errorStr, "404") || strings.Contains(errorStr, "not found") {
		return fmt.Sprintf("%s\n\nSetup Hint: ChromaDB endpoint not found.\n"+
			"  - Verify ChromaDB is running on the correct port\n"+
			"  - Check URL: %s\n"+
			"  - Ensure you're using ChromaDB v2 API", originalError, baseURL)
	}
	
	if strings.Contains(errorStr, "no such host") || strings.Contains(errorStr, "dns") {
		return fmt.Sprintf("%s\n\nSetup Hint: DNS resolution failed.\n"+
			"  - Check if hostname is correct: %s\n"+
			"  - Try using 127.0.0.1 instead of localhost\n"+
			"  - Verify network connectivity", originalError, baseURL)
	}
	
	// Generic connection error
	return fmt.Sprintf("%s\n\nSetup Hint: Unable to connect to ChromaDB.\n"+
		"  - Ensure ChromaDB is running on %s\n"+
		"  - Check firewall and network settings\n"+
		"  - Verify ChromaDB logs for additional information", originalError, baseURL)
}

// getChromaDBSetupHints returns setup hints for ChromaDB
func getChromaDBSetupHints() []string {
	return []string{
		"Start ChromaDB with uvx: uvx --from 'chromadb[server]' chroma run --host localhost --port 8000",
		"Start ChromaDB with Docker: docker run -p 8000:8000 chromadb/chroma",
		"Verify installation: curl http://localhost:8000/api/v2/heartbeat",
		"Check ChromaDB logs for detailed error information",
		"Memory Bank will fall back to mock vector store if ChromaDB is unavailable",
	}
}

// QuickHealthCheck performs a quick health check and displays any service issues
func QuickHealthCheck(ctx context.Context, services *ServiceContainer) {
	// Quick checks - only check ChromaDB as it's most likely to fail
	chromaConfig := vector.ChromaDBConfig{
		BaseURL:    services.Config.ChromaDB.BaseURL,
		Collection: services.Config.ChromaDB.Collection,
		Tenant:     services.Config.ChromaDB.Tenant,
		Database:   services.Config.ChromaDB.Database,
		Timeout:    time.Duration(services.Config.ChromaDB.Timeout) * time.Second,
	}
	chromaStore := vector.NewChromaDBVectorStore(chromaConfig, services.Logger)
	
	// Quick timeout for this check
	quickCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	
	if err := chromaStore.HealthCheck(quickCtx); err != nil {
		fmt.Printf("⚠️  Warning: ChromaDB is not available (%s)\n", err.Error())
		fmt.Printf("   💡 Falling back to mock vector store (semantic search disabled)\n")
		fmt.Printf("   💡 Start ChromaDB: uvx --from 'chromadb[server]' chroma run --host localhost --port 8000\n\n")
	}
}

func init() {
	// Add flags
	healthCmd.Flags().BoolP("json", "j", false, "output health status as JSON")
	healthCmd.Flags().BoolP("verbose", "v", false, "show detailed service information")
	healthCmd.Flags().BoolP("watch", "w", false, "watch mode - continuously monitor health status")
	
	statusCmd.Flags().BoolP("json", "j", false, "output health status as JSON")
	statusCmd.Flags().BoolP("verbose", "v", false, "show detailed service information")
	statusCmd.Flags().BoolP("watch", "w", false, "watch mode - continuously monitor health status")
	
	// Register commands
	rootCmd.AddCommand(healthCmd)
	rootCmd.AddCommand(statusCmd)
}