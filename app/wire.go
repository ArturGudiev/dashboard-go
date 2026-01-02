//go:build wireinject
// +build wireinject

package app

import (
	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/services"
	"context"
	"log"
	"os"
	"strings"

	"github.com/google/wire"
	_ "github.com/lib/pq"
)

// InitializeApp creates App with all dependencies wired automatically
func InitializeApp() (*App, error) {
	wire.Build(
		// Provider for ent.Client (includes migration)
		provideEntClient,
		// Service providers
		services.NewTaskService,
		services.NewProblemService,
		services.NewContainerService,
		services.NewCLIService,
		// App provider
		provideApp,
	)
	return nil, nil
}

// provideEntClient creates and migrates the ent client
func provideEntClient() (*ent.Client, error) {
	// Database connection string
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost/dashboard?sslmode=disable"
	}

	// Create Ent client
	client, err := ent.Open("postgres", dbURL)
	if err != nil {
		return nil, err
	}

	// Run the auto migration tool
	ctx := context.Background()
	if err := client.Schema.Create(ctx); err != nil {
		// Check if the error is because table already exists, permission issue, or schema mismatch
		errMsg := strings.ToLower(err.Error())
		if strings.Contains(errMsg, "already exists") ||
			strings.Contains(errMsg, "permission denied") ||
			strings.Contains(errMsg, "unexpected attribute change") ||
			strings.Contains(errMsg, "expect identity") {
			log.Printf("Warning: Schema migration had issues: %v", err)
			log.Println("Continuing with existing schema...")
		} else {
			client.Close()
			return nil, err
		}
	} else {
		log.Println("Schema migration completed successfully")
	}

	return client, nil
}

// provideApp creates the App instance with all dependencies
func provideApp(
	client *ent.Client,
	taskService *services.TaskService,
	problemService *services.ProblemService,
	containerService *services.ContainerService,
	cliService *services.CLIService,
) *App {
	return &App{
		Client:           client,
		TaskService:      taskService,
		ProblemService:   problemService,
		ContainerService: containerService,
		CLIService:       cliService,
		ctx:              context.Background(), // Default context for CLI
	}
}
