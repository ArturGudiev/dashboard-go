package app

import (
	"arturgudiev/dashboard/ent/schema"
	"context"
	"log"
	"os"
	"strings"

	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/services"

	_ "github.com/lib/pq"
)

// App holds all application dependencies
type App struct {
	Client           *ent.Client
	TaskService      *services.TaskService
	ProblemService   *services.ProblemService
	ContainerService *services.ContainerService
	ctx              context.Context // Default context for CLI operations
}

// NewApp creates a new App instance with all dependencies initialized
func NewApp() (*App, error) {
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

	// Initialize services
	taskService := services.NewTaskService(client)
	problemService := services.NewProblemService(client)
	containerService := services.NewContainerService(client)

	return &App{
		Client:           client,
		TaskService:      taskService,
		ProblemService:   problemService,
		ContainerService: containerService,
		ctx:              context.Background(), // Default context for CLI
	}, nil
}

// Close closes all resources
func (a *App) Close() error {
	return a.Client.Close()
}

// Wrapper methods that use the default context (for CLI convenience)

// GetChildSubtasks returns all child tasks for a given parent task ID using default context
func (a *App) GetChildSubtasks(parentID int) ([]*ent.Task, error) {
	return a.TaskService.GetChildSubtasks(a.ctx, parentID)
}

// FinishTaskRecursively finishes all open descendant tasks using default context
func (a *App) FinishTaskRecursively(task *ent.Task) error {
	return a.TaskService.FinishTaskRecursively(a.ctx, task)
}

// AddSubtask creates a new subtask for the given parent task using default context
func (a *App) AddSubtask(parentType schema.ContainerType, parentID int, description string) (*ent.Task, error) {
	return a.TaskService.AddSubtask(a.ctx, parentType, parentID, description)
}

// GetOpenDescendantTasks recursively gets all descendant tasks that are not done using default context
func (a *App) GetOpenDescendantTasks(parentTask *ent.Task) []*ent.Task {
	return a.TaskService.GetOpenDescendantTasks(a.ctx, parentTask)
}

// GetParent returns the parent task of the given task using default context
func (a *App) GetParent(task *ent.Task) *ent.Task {
	return a.TaskService.GetParent(a.ctx, task)
}
