//go:build wireinject
// +build wireinject

package app

import (
	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/ent/migrate"
	"arturgudiev/dashboard/repositories"
	"arturgudiev/dashboard/services"
	"arturgudiev/dashboard/ws"
	"context"
	"log"
	"net/url"
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
		services.NewReportService,
		services.NewTaskService,
		services.NewProblemService,
		services.NewContainerService,
		services.NewCLIService,
		services.NewTasksRepository,
		services.NewProblemsRepository,
		services.NewQuestionService,
		services.NewStoriesService,
		services.NewEpicsService,
		services.NewKnowledgeNodesService,
		services.NewAliasesService,
		services.NewRepetitiveTaskService,
		services.NewRepetitiveTaskExecutionService,
		services.NewLongTasksService,
		services.NewLongTaskSubmissionsService,
		services.NewLongTaskProgressesService,
		services.NewDirectionsService,
		services.NewStatesService,
		repositories.NewStatesRepository,
		repositories.NewStateRequirementsRepository,
		repositories.NewStateRequirementChecksRepository,
		repositories.NewUsersRepository,
		repositories.NewRefreshTokensRepository,
		repositories.NewQuestionsRepository,
		repositories.NewStoriesRepository,
		repositories.NewEpicsRepository,
		repositories.NewKnowledgeNodesRepository,
		repositories.NewAliasesRepository,
		services.NewChildContainerRepository,
		repositories.NewLogMessagesRepository,
		repositories.NewVariablesStackRepository,
		repositories.NewContainerVariablesRepository,
		repositories.NewContainerChecksRepository,
		services.NewContainerChecksService,
		repositories.NewRepetitiveTasksRepository,
		repositories.NewRepetitiveTaskExecutionsRepository,
		repositories.NewLongTasksRepository,
		repositories.NewLongTaskSubmissionsRepository,
		repositories.NewDirectionsRepository,
		repositories.NewDirectionSubmissionsRepository,
		repositories.NewLongTasksProgressesSubmissionsRepository,
		repositories.NewLongTasksProgressesRepository,
		// App provider
		provideApp,
	)
	return nil, nil
}

// provideEntClient creates and migrates the ent client
func provideEntClient() (*ent.Client, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbName := os.Getenv("DB_NAME")
		if dbName == "" {
			dbName = "dashboard"
		}
		dbPassword := os.Getenv("DB_PASSWORD")
		if dbPassword == "" {
			dbPassword = "postgres"
		}
		dbHost := os.Getenv("DB_HOST")
		if dbHost == "" {
			dbHost = "localhost"
		}
		dbUser := os.Getenv("DB_USER")
		if dbUser == "" {
			dbUser = "postgres"
		}
		u := &url.URL{
			Scheme:   "postgres",
			User:     url.UserPassword(dbUser, dbPassword),
			Host:     dbHost,
			Path:     "/" + url.PathEscape(dbName),
			RawQuery: "sslmode=disable",
		}
		dbURL = u.String()
		log.Printf("DATABASE_URL was not set; using constructed DB URL: %s", dbURL)
	}

	// Create Ent client
	client, err := ent.Open("postgres", dbURL)
	if err != nil {
		return nil, err
	}

	// Run the auto migration tool with DropColumn option
	ctx := context.Background()
	if err := client.Schema.Create(ctx, migrate.WithDropColumn(true)); err != nil {
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
	repetitiveTaskService *services.RepetitiveTaskService,
	repetitiveTaskExecutionService *services.RepetitiveTaskExecutionService,
	longTasksService *services.LongTasksService,
	longTaskSubmissionsService *services.LongTaskSubmissionsService,
	longTaskProgressesService *services.LongTaskProgressesService,
	directionsService *services.DirectionsService,
	problemService *services.ProblemService,
	containerService *services.ContainerService,
	cliService *services.CLIService,
	statesService *services.StatesService,
	tasksRepository *services.TasksRepository,
	problemsRepository *services.ProblemsRepository,
	questionsRepository *repositories.QuestionsRepository,
	questionsService *services.QuestionService,
	storiesRepository *repositories.StoriesRepository,
	storiesService *services.StoriesService,
	epicsRepository *repositories.EpicsRepository,
	epicsService *services.EpicsService,
	aliasesRepository *repositories.AliasesRepository,
	aliasesService *services.AliasesService,
	knowledgeNodesRepository *repositories.KnowledgeNodesRepository,
	knowledgeNodesService *services.KnowledgeNodesService,
	childContainerRepository *services.ChildContainerRepository,
	logMessagesRepository *repositories.LogMessagesRepository,
	variablesStackRepository *repositories.VariablesStackRepository,
	containerVariablesRepository *repositories.ContainerVariablesRepository,
	containerChecksRepository *repositories.ContainerChecksRepository,
	containerChecksService *services.ContainerChecksService,
	repetitiveTasksRepository *repositories.RepetitiveTasksRepository,
	repetitiveTaskExecutionsRepository *repositories.RepetitiveTaskExecutionsRepository,
	longTasksRepository *repositories.LongTasksRepository,
	longTaskSubmissionsRepository *repositories.LongTaskSubmissionsRepository,
	directionsRepository *repositories.DirectionsRepository,
	directionSubmissionsRepository *repositories.DirectionSubmissionsRepository,
	longTasksProgressesSubmissionsRepository *repositories.LongTasksProgressesSubmissionsRepository,
	longTasksProgressesRepository *repositories.LongTasksProgressesRepository,
	statesRepository *repositories.StatesRepository,
	stateRequirementsRepository *repositories.StateRequirementsRepository,
	stateRequirementChecksRepository *repositories.StateRequirementChecksRepository,
	usersRepository *repositories.UsersRepository,
	refreshTokensRepository *repositories.RefreshTokensRepository,
) *App {
	hub := ws.NewHub()
	go hub.Run()

	return &App{
		Client:                    client,
		TaskService:               taskService,
		RepetitiveTaskService:     repetitiveTaskService,
		RepetitiveTaskExecutionService: repetitiveTaskExecutionService,
		LongTasksService:         longTasksService,
		LongTaskSubmissionsService: longTaskSubmissionsService,
		LongTaskProgressesService: longTaskProgressesService,
		DirectionsService:         directionsService,
		StatesService:             statesService,
		StatesRepository:          statesRepository,
		StateRequirementsRepository: stateRequirementsRepository,
		StateRequirementChecksRepository: stateRequirementChecksRepository,
		ProblemService:            problemService,
		ContainerService:          containerService,
		CLIService:                cliService,
		TasksRepository:           tasksRepository,
		ProblemsRepository:        problemsRepository,
		QuestionsRepository:       questionsRepository,
		QuestionsService:          questionsService,
		StoriesRepository:         storiesRepository,
		StoriesService:            storiesService,
		EpicsRepository:           epicsRepository,
		EpicsService:              epicsService,
		KnowledgeNodesRepository:  knowledgeNodesRepository,
		KnowledgeNodesService:     knowledgeNodesService,
		AliasesRepository:         aliasesRepository,
		AliasesService:            aliasesService,
		ChildContainerRepository:  childContainerRepository,
		LogMessagesRepository:              logMessagesRepository,
		VariablesStackRepository:           variablesStackRepository,
		ContainerVariablesRepository:       containerVariablesRepository,
		ContainerChecksRepository:          containerChecksRepository,
		ContainerChecksService:             containerChecksService,
		RepetitiveTasksRepository:          repetitiveTasksRepository,
		RepetitiveTaskExecutionsRepository: repetitiveTaskExecutionsRepository,
		LongTasksRepository:                longTasksRepository,
		LongTaskSubmissionsRepository:     longTaskSubmissionsRepository,
		DirectionsRepository:              directionsRepository,
		DirectionSubmissionsRepository:    directionSubmissionsRepository,
		LongTasksProgressesSubmissionsRepository: longTasksProgressesSubmissionsRepository,
		LongTasksProgressesRepository: longTasksProgressesRepository,
		UsersRepository:                usersRepository,
		RefreshTokensRepository:        refreshTokensRepository,
		FilesService:                   services.NewFilesService(),
		Hub:                            hub,
		ctx:                       context.Background(), // Default context for CLI
	}
}
