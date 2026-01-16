package main

import (
	"log"
	"os"

	"arturgudiev/dashboard/app"
	_ "arturgudiev/dashboard/docs" // Swagger docs
	"arturgudiev/dashboard/handlers"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           Dashboard API
// @version         1.0
// @description     Task management dashboard API server
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /

// @schemes   http https

func main() {
	// Initialize app with all dependencies
	application, err := app.NewApp()
	if err != nil {
		log.Fatalf("Failed to initialize app: %v", err)
	}
	defer application.Close()

	// Check command-line arguments
	if len(os.Args) > 1 {
		command := os.Args[1]
		switch command {
		case "cli", "interactive":
			runCLI(application)
			return
		case "import":
			// Import tasks/epics/stories/knowledge-nodes from JSON file
			importType := "tasks"
			jsonPath := ""
			if len(os.Args) > 2 {
				importType = os.Args[2]
			}
			if len(os.Args) > 3 {
				jsonPath = os.Args[3]
			}
			var err error
			switch importType {
			case "tasks":
				err = importTasks(jsonPath)
			case "epics":
				err = importEpics(jsonPath)
			case "stories":
				err = importStories(jsonPath)
			case "knowledge-nodes", "knowledgenodes":
				err = importKnowledgeNodes(jsonPath)
			default:
				log.Fatalf("Unknown import type: %s. Use 'tasks', 'epics', 'stories', or 'knowledge-nodes'", importType)
			}
			if err != nil {
				log.Fatalf("Import failed: %v", err)
			}
			return
		}
	}

	// Setup Gin router
	router := gin.Default()

	// Swagger UI route
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Create handler instance
	h := handlers.NewHandler(application)

	router.GET("/", h.Root)
	router.GET("/tests", h.GetTests)

	// Task routes
	router.GET("/task/:id", h.GetTaskByID)
	router.POST("/get-tasks", h.GetTasksByIDs)
	router.PUT("/add-anonymous-task", h.AddAnonymousTask)
	router.GET("/done-tasks", h.GetDoneTasks)
	router.PUT("/finish-task/:id", h.FinishTask)
	router.PUT("/finish-tasks/", h.FinishTasks)
	router.PUT("/finish-tasks-by-ids/", h.FinishTasksByIDs)
	router.POST("/new-task", h.NewTask)
	router.PUT("/update-task", h.UpdateTask)

	// Problem routes
	router.GET("/problem/:id", h.GetProblemByID)
	router.POST("/get-problems", h.GetProblemsByIDs)
	router.PUT("/solve-problem/:id", h.SolveProblem)
	router.POST("/new-problem", h.NewProblem)
	router.PUT("/update-problem", h.UpdateProblem)

	// Question routes
	router.GET("/question/:id", h.GetQuestionByID)
	router.POST("/get-questions", h.GetQuestionsByIDs)
	router.POST("/new-question", h.NewQuestion)
	router.PUT("/update-question", h.UpdateQuestion)
	router.POST("/answer-question/:id", h.AnswerQuestion)

	// Stories routes
	router.GET("/story/:id", h.GetStoryByID)
	router.POST("/new-story", h.NewStory)
	router.POST("/get-stories", h.GetStoriesByIDs)
	router.PUT("/update-story", h.UpdateStory)

	// Epic routes
	router.GET("/epic/:id", h.GetEpicByID)
	router.POST("/get-epics", h.GetEpicsByIDs)
	router.POST("/new-epic", h.NewEpic)
	router.PUT("/update-epic", h.UpdateEpic)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(router.Run(":" + port))
}
