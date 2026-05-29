package main

import (
	"log"
	"os"

	"arturgudiev/dashboard/app"
	_ "arturgudiev/dashboard/docs" // Swagger docs
	"arturgudiev/dashboard/handlers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
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
	// Load .env file (ignore error if file doesn't exist)
	_ = godotenv.Load()

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
			// Import tasks/epics/stories/knowledge-nodes/aliases from JSON file
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
			case "aliases":
				err = importAliases(jsonPath)
			case "aliases-files":
				err = importFileAliases(jsonPath)
			default:
				log.Fatalf("Unknown import type: %s. Use 'tasks', 'epics', 'stories', 'knowledge-nodes', or 'aliases'", importType)
			}
			if err != nil {
				log.Fatalf("Import failed: %v", err)
			}
			return
		case "parse-file-aliases":
			// Parse file aliases from aliases.json and create aliases_parsed_files.json
			inputPath := ""
			outputPath := ""
			if len(os.Args) > 2 {
				inputPath = os.Args[2]
			}
			if len(os.Args) > 3 {
				outputPath = os.Args[3]
			}
			if err := parseFileAliases(inputPath, outputPath); err != nil {
				log.Fatalf("Parse failed: %v", err)
			}
			return
		}
	}

	// Setup Gin router
	router := gin.Default()

	// Configure CORS middleware - matching NodeJS server behavior
	config := cors.DefaultConfig()

	// Use a function to allow any origin (like NodeJS server does)
	// This is required when credentials: true (Firefox requirement)
	config.AllowOriginFunc = func(origin string) bool {
		// Allow requests with no origin (like mobile apps, curl, Postman)
		if origin == "" {
			log.Println("CORS: No origin header, allowing request")
			return true
		}
		// Log the origin for debugging
		log.Printf("CORS: Allowing origin: %s", origin)
		// Allow all origins (return true for any origin)
		return true
	}

	// Allow credentials (cookies, authorization headers)
	config.AllowCredentials = true
	// Allow all methods
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD"}
	// Allow common headers
	config.AllowHeaders = []string{
		"Origin",
		"Content-Type",
		"Content-Length",
		"Accept-Encoding",
		"X-CSRF-Token",
		"Authorization",
		"Accept",
		"X-Requested-With",
		"X-HTTP-Method-Override",
		"Access-Control-Request-Method",
		"Access-Control-Request-Headers",
	}
	// Cache preflight requests for 24 hours
	config.MaxAge = 86400

	// Apply CORS middleware
	router.Use(cors.New(config))

	// Explicitly handle OPTIONS requests for better compatibility (like NodeJS server)
	router.OPTIONS("/*path", func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		log.Printf("OPTIONS preflight request: %s %s Origin: %s", c.Request.Method, c.Request.URL.Path, origin)
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH, HEAD")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Origin, X-HTTP-Method-Override")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400")
		c.Status(204)
	})

	// Swagger UI route
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Create handler instance
	h := handlers.NewHandler(application)

	router.GET("/", h.Root)
	router.GET("/tests", h.GetTests)
	router.POST("/parents-path", h.GetParentsPath)

	// Task routes
	router.GET("/task/:id", h.GetTaskByID)
	router.PATCH("/task/:id", h.PatchTaskByID)
	router.GET("/task-report/:id", h.GetTaskReport)
	router.POST("/get-tasks", h.GetTasksByIDs)
	router.PUT("/add-anonymous-task", h.AddAnonymousTask)
	router.GET("/done-tasks", h.GetDoneTasks)
	router.PUT("/finish-task/:id", h.FinishTask)
	router.PUT("/finish-tasks/", h.FinishTasks)
	router.PUT("/finish-tasks-by-ids/", h.FinishTasksByIDs)
	router.POST("/new-task", h.NewTask)
	router.PUT("/update-task", h.UpdateTask)
	router.POST("/change-tasks-order", h.ChangeTasksOrder)

	// Repetitive task routes
	router.GET("/repetitive-tasks", h.GetRepetitiveTasks)
	router.GET("/repetitive-tasks/:id", h.GetRepetitiveTaskById)
	router.PATCH("/repetitive-tasks/:id", h.PatchRepetitiveTaskByID)
	router.POST("/repetitive-tasks/:id/executions", h.AddRepetitiveTaskExecution)
	router.GET("/repetitive-tasks/:id/executions", h.GetRepetitiveTaskExecutions)
	router.POST("/new-repetitive-task", h.NewRepetitiveTask)

	// Problem routes
	router.GET("/problem/:id", h.GetProblemByID)
	router.PATCH("/problem/:id", h.PatchProblemByID)
	router.POST("/get-problems", h.GetProblemsByIDs)
	router.POST("/solve-problem/:id", h.SolveProblem)
	router.POST("/new-problem", h.NewProblem)
	router.PUT("/update-problem", h.UpdateProblem)

	// Question routes
	router.GET("/question/:id", h.GetQuestionByID)
	router.PATCH("/question/:id", h.PatchQuestionByID)
	router.POST("/get-questions", h.GetQuestionsByIDs)
	router.POST("/new-question", h.NewQuestion)
	router.PUT("/update-question", h.UpdateQuestion)
	router.POST("/answer-question/:id", h.AnswerQuestion)

	// Stories routes
	router.GET("/story/:id", h.GetStoryByID)
	router.PATCH("/story/:id", h.PatchStoryByID)
	router.POST("/new-story", h.NewStory)
	router.POST("/get-stories", h.GetStoriesByIDs)
	router.PUT("/update-story", h.UpdateStory)

	// Epic routes
	router.GET("/epic/:id", h.GetEpicByID)
	router.PATCH("/epic/:id", h.PatchEpicByID)
	router.POST("/get-epics", h.GetEpicsByIDs)
	router.POST("/new-epic", h.NewEpic)
	router.PUT("/update-epic", h.UpdateEpic)
	router.GET("/epics", h.GetAllOpenEpics)

	// Log messages routes
	router.GET("/log-messages/:id", h.GetLogMessageByID)
	router.GET("/log-messages", h.GetLogMessages)
	router.POST("/log-messages", h.NewLogMessage)

	// Container variables routes
	router.POST("/container-variables", h.AddContainerVariable)
	router.DELETE("/container-variables/:id", h.RemoveContainerVariable)

	// Aliase routes
	router.GET("/aliases/:alias", h.GetAliasByString)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Bind to all interfaces (0.0.0.0) to allow access from other machines
	address := "0.0.0.0:" + port
	log.Printf("Server starting on %s (accessible from network)", address)
	log.Fatal(router.Run(address))
}
