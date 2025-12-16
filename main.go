package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/ent/containerchild"
	"arturgudiev/dashboard/ent/task"
	"github.com/gin-gonic/gin"

	_ "github.com/lib/pq"
)

// getOpenDescendantTasks recursively gets all descendant tasks that are not done
func getOpenDescendantTasks(ctx context.Context, client *ent.Client, parentTask *ent.Task) []*ent.Task {
	var result []*ent.Task
	
	// Get all child relationships where this task is the parent
	childRelations, err := client.ContainerChild.Query().
		Where(
			containerchild.ParentTypeEQ(containerchild.ParentTypeTask),
			containerchild.ParentID(parentTask.ID),
			containerchild.ChildTypeEQ(containerchild.ChildTypeTask),
		).
		WithChild().
		All(ctx)
	
	if err != nil {
		log.Printf("Error querying children for task %d: %v", parentTask.ID, err)
		return result
	}

	// Process each child
	for _, relation := range childRelations {
		childTask := relation.Edges.Child
		if childTask == nil {
			continue
		}

		// Only include tasks that are not done
		if !childTask.Done {
			result = append(result, childTask)
		}

		// Recursively get descendants of this child
		descendants := getOpenDescendantTasks(ctx, client, childTask)
		result = append(result, descendants...)
	}

	return result
}

// finishTaskRecursively finishes all open descendant tasks of the given task
func finishTaskRecursively(ctx context.Context, client *ent.Client, task *ent.Task) error {
	// Recursively get all descendant tasks that are not done
	allTasksToFinish := getOpenDescendantTasks(ctx, client, task)

	// Mark all tasks as done with current timestamp
	now := time.Now()
	for _, taskToFinish := range allTasksToFinish {
		_, err := client.Task.UpdateOneID(taskToFinish.ID).
			SetDone(true).
			SetDoneDateTime(now).
			Save(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	// Check command-line arguments
	if len(os.Args) > 1 {
		command := os.Args[1]
		switch command {
		case "cli", "interactive":
			runCLI()
			return
		case "import":
			// Import tasks from JSON file
			jsonPath := ""
			if len(os.Args) > 2 {
				jsonPath = os.Args[2]
			}
			if err := importTasks(jsonPath); err != nil {
				log.Fatalf("Import failed: %v", err)
			}
			return
		}
	}

	// Database connection string
	// Default PostgreSQL connection: postgresql://postgres:postgres@localhost/dashboard
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost/dashboard?sslmode=disable"
	}

	// Create Ent client
	client, err := ent.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("failed opening connection to postgres: %v", err)
	}
	defer client.Close()

	// Run the auto migration tool
	// If tables already exist, this will handle it gracefully
	ctx := context.Background()
	if err := client.Schema.Create(ctx); err != nil {
		// Check if the error is because table already exists, permission issue, or schema mismatch
		// If table exists, we can continue - Ent will work with existing tables
		errMsg := strings.ToLower(err.Error())
		if strings.Contains(errMsg, "already exists") || 
		   strings.Contains(errMsg, "permission denied") ||
		   strings.Contains(errMsg, "unexpected attribute change") ||
		   strings.Contains(errMsg, "expect identity") {
			log.Printf("Warning: Schema migration had issues: %v", err)
			
			log.Println("Continuing with existing schema...")
		} else {
			log.Fatalf("failed creating schema resources: %v", err)
		}
	} else {
		log.Println("Schema migration completed successfully")
	}

	// Setup Gin router
	router := gin.Default()

	// GET /
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Dashboard server"})
	})

	// GET /tests
	router.GET("/tests", func(c *gin.Context) {
		// Query all tests
		tests, err := client.Test.Query().All(c.Request.Context())
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		// Convert to JSON response
		response := make([]TestResponse, len(tests))
		for i, t := range tests {
			response[i] = TestResponse{
				ID:   t.ID,
				Name: t.Name,
				Tags: t.Tags,
			}
		}

		c.JSON(200, response)
	})

	// GET /task/:id
	router.GET("/task/:id", func(c *gin.Context) {
		// Get ID from URL parameter
		idParam := c.Param("id")
		id, err := strconv.Atoi(idParam)
		if err != nil {
			c.JSON(400, gin.H{"error": "Invalid task ID"})
			return
		}

		// Get task by ID
		task, err := client.Task.Get(c.Request.Context(), id)
		if err != nil {
			if ent.IsNotFound(err) {
				c.JSON(404, gin.H{"error": "Task not found"})
				return
			}
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		// Return task as JSON
		c.JSON(200, task)
	})

	// POST /get-tasks
	router.POST("/get-tasks", func(c *gin.Context) {
		var req IDsRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		// Get tasks by IDs
		tasks, err := client.Task.Query().Where(task.IDIn(req.IDs...)).All(c.Request.Context())
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		// Return tasks as JSON
		c.JSON(200, tasks)
	})

	router.PUT("/add-anonymous-task", func(c *gin.Context) {
		// Create a simple task
		newTask, err := client.Task.Create().
			SetDescription("Simple task").
			SetDone(true).
			SetTags([]string{}).
			SetNotes("").
			Save(c.Request.Context())
		
		if err != nil {
			log.Printf("Error creating task: %v", err)
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		log.Printf("Successfully created task with ID: %d", newTask.ID)
		// Return the created task
		c.JSON(200, newTask)
	})


	
	// GET /task/:id
	router.GET("/done-tasks", func(c *gin.Context) {
		c.JSON(200, gin.H{"doneTasks": 777})
	})

	// PUT /finish-task/:id
	router.PUT("/finish-task/:id", func(c *gin.Context) {
		// Get ID from URL parameter
		idParam := c.Param("id")
		id, err := strconv.Atoi(idParam)
		if err != nil {
			c.JSON(400, gin.H{"error": "Invalid task ID"})
			return
		}

		ctx := c.Request.Context()

		// Get task by ID
		task, err := client.Task.Get(ctx, id)
		if err != nil {
			if ent.IsNotFound(err) {
				c.JSON(404, gin.H{"error": "Task not found"})
				return
			}
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		// Finish task recursively
		if err := finishTaskRecursively(ctx, client, task); err != nil {
			log.Printf("Error finishing task %d recursively: %v", id, err)
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		// Reload the updated task
		updatedTask, err := client.Task.Get(ctx, id)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		// Return the updated task
		c.JSON(200, updatedTask)
	})

	// PUT /finish-tasks/
	router.PUT("/finish-tasks/", func(c *gin.Context) {
		var tasks []TaskIDRequest
		if err := c.ShouldBindJSON(&tasks); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		ctx := c.Request.Context()

		for _, t := range tasks {
			// Get task by ID
			task, err := client.Task.Get(ctx, t.ID)
			if err != nil {
				if ent.IsNotFound(err) {
					log.Printf("Task %d not found, skipping", t.ID)
					continue
				}
				log.Printf("Error getting task %d: %v", t.ID, err)
				continue
			}

			// Finish task recursively
			if err := finishTaskRecursively(ctx, client, task); err != nil {
				log.Printf("Error finishing task %d recursively: %v", t.ID, err)
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
		}

		c.JSON(200, gin.H{})
	})

	// PUT /finish-tasks-by-ids/
	router.PUT("/finish-tasks-by-ids/", func(c *gin.Context) {
		var ids []int
		if err := c.ShouldBindJSON(&ids); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		ctx := c.Request.Context()

		for _, id := range ids {
			// Get task by ID
			task, err := client.Task.Get(ctx, id)
			if err != nil {
				if ent.IsNotFound(err) {
					log.Printf("Task %d not found, skipping", id)
					continue
				}
				log.Printf("Error getting task %d: %v", id, err)
				continue
			}

			// Finish task recursively
			if err := finishTaskRecursively(ctx, client, task); err != nil {
				log.Printf("Error finishing task %d recursively: %v", id, err)
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
		}

		c.JSON(200, gin.H{})
	})

	// POST /new-task
	router.POST("/new-task", func(c *gin.Context) {
		var req NewTaskRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		ctx := c.Request.Context()

		// Create the task - Ent will use schema defaults for unset fields
		taskBuilder := client.Task.Create().
			SetDescription(req.Task.Description).
			SetDone(req.Task.Done)

		// Only set fields that are explicitly provided (non-zero/non-nil)
		if req.Task.Tags != nil {
			taskBuilder = taskBuilder.SetTags(req.Task.Tags)
		}
		if req.Task.Notes != "" {
			taskBuilder = taskBuilder.SetNotes(req.Task.Notes)
		}
		if req.Task.Problems != nil {
			taskBuilder = taskBuilder.SetProblems(req.Task.Problems)
		}
		if req.Task.Questions != nil {
			taskBuilder = taskBuilder.SetQuestions(req.Task.Questions)
		}
		if req.Task.Actions != nil {
			taskBuilder = taskBuilder.SetActions(req.Task.Actions)
		}
		if req.Task.Definitions != nil {
			taskBuilder = taskBuilder.SetDefinitions(req.Task.Definitions)
		}
		if req.Task.KnowledgeBits != nil {
			taskBuilder = taskBuilder.SetKnowledgeBits(req.Task.KnowledgeBits)
		}
		if req.Task.KnowledgeNodes != nil {
			taskBuilder = taskBuilder.SetKnowledgeNodes(req.Task.KnowledgeNodes)
		}
		if req.Task.ParentContainers != nil {
			taskBuilder = taskBuilder.SetParentContainers(req.Task.ParentContainers)
		}
		if req.Task.DoneDateTime != nil {
			taskBuilder = taskBuilder.SetDoneDateTime(*req.Task.DoneDateTime)
		}

		newTask, err := taskBuilder.Save(ctx)
		if err != nil {
			log.Printf("Error creating task: %v", err)
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		// Create parent-child relationship if parent is provided
		if req.Parent != nil && req.Parent.Type == "task" {
			// Verify parent task exists
			parentTask, err := client.Task.Get(ctx, req.Parent.Obj.ID)
			if err != nil {
				if ent.IsNotFound(err) {
					c.JSON(404, gin.H{"error": "Parent task not found"})
					return
				}
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}

			// Check if relationship already exists
			exists, err := client.ContainerChild.Query().
				Where(
					containerchild.ParentTypeEQ(containerchild.ParentTypeTask),
					containerchild.ParentID(parentTask.ID),
					containerchild.ChildTypeEQ(containerchild.ChildTypeTask),
					containerchild.ChildID(newTask.ID),
				).
				Exist(ctx)
			if err != nil {
				log.Printf("Error checking relationship: %v", err)
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}

			if !exists {
				// Get the count of existing children to set child_order
				childCount, err := client.ContainerChild.Query().
					Where(
						containerchild.ParentTypeEQ(containerchild.ParentTypeTask),
						containerchild.ParentID(parentTask.ID),
						containerchild.ChildTypeEQ(containerchild.ChildTypeTask),
					).
					Count(ctx)
				if err != nil {
					log.Printf("Error counting children: %v", err)
					c.JSON(500, gin.H{"error": err.Error()})
					return
				}

				// Get the count of existing parents to set parent_order
				parentCount, err := client.ContainerChild.Query().
					Where(
						containerchild.ChildTypeEQ(containerchild.ChildTypeTask),
						containerchild.ChildID(newTask.ID),
						containerchild.ParentTypeEQ(containerchild.ParentTypeTask),
					).
					Count(ctx)
				if err != nil {
					log.Printf("Error counting parents: %v", err)
					c.JSON(500, gin.H{"error": err.Error()})
					return
				}

				// Create the relationship
				_, err = client.ContainerChild.Create().
					SetParentType(containerchild.ParentTypeTask).
					SetParentID(parentTask.ID).
					SetChildType(containerchild.ChildTypeTask).
					SetChildID(newTask.ID).
					SetChildOrder(childCount).
					SetParentOrder(parentCount).
					Save(ctx)
				if err != nil {
					log.Printf("Error creating relationship: %v", err)
					c.JSON(500, gin.H{"error": err.Error()})
					return
				}
			}
		}

		// Return the created task
		c.JSON(200, newTask)
	})

	// PUT /update-task
	router.PUT("/update-task", func(c *gin.Context) {
		var req UpdateTaskRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		ctx := c.Request.Context()

		// Verify task exists
		_, err := client.Task.Get(ctx, req.ID)
		if err != nil {
			if ent.IsNotFound(err) {
				c.JSON(404, gin.H{"error": "Task not found"})
				return
			}
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		// Update the task - build update query
		taskBuilder := client.Task.UpdateOneID(req.ID).
			SetDescription(req.Description).
			SetDone(req.Done)

		// Only set fields that are explicitly provided (non-zero/non-nil)
		if req.Tags != nil {
			taskBuilder = taskBuilder.SetTags(req.Tags)
		}
		if req.Notes != "" {
			taskBuilder = taskBuilder.SetNotes(req.Notes)
		}
		if req.Problems != nil {
			taskBuilder = taskBuilder.SetProblems(req.Problems)
		}
		if req.Questions != nil {
			taskBuilder = taskBuilder.SetQuestions(req.Questions)
		}
		if req.Actions != nil {
			taskBuilder = taskBuilder.SetActions(req.Actions)
		}
		if req.Definitions != nil {
			taskBuilder = taskBuilder.SetDefinitions(req.Definitions)
		}
		if req.KnowledgeBits != nil {
			taskBuilder = taskBuilder.SetKnowledgeBits(req.KnowledgeBits)
		}
		if req.KnowledgeNodes != nil {
			taskBuilder = taskBuilder.SetKnowledgeNodes(req.KnowledgeNodes)
		}
		if req.ParentContainers != nil {
			taskBuilder = taskBuilder.SetParentContainers(req.ParentContainers)
		}
		if req.DoneDateTime != nil {
			taskBuilder = taskBuilder.SetDoneDateTime(*req.DoneDateTime)
		}

		updatedTask, err := taskBuilder.Save(ctx)
		if err != nil {
			log.Printf("Error updating task: %v", err)
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		// Return the updated task
		c.JSON(200, updatedTask)
	})

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(router.Run(":" + port))
}

