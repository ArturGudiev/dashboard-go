package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/ent/containerchild"
	"arturgudiev/dashboard/ent/task"

	_ "github.com/lib/pq"
)

// TaskJSON represents the structure of tasks in the JSON file
type TaskJSON struct {
	ID           int       `json:"_id"`
	Description  string    `json:"description"`
	Tags         []string  `json:"tags"`
	Done         bool      `json:"done"`
	Notes        string    `json:"notes"`
	Tasks        []int     `json:"tasks"` // Child task IDs
	Problems     []int     `json:"problems"`
	Questions    []int     `json:"questions"`
	Actions      []int     `json:"actions"`
	Definitions  []int     `json:"definitions"`
	KnowledgeBits []int    `json:"knowledgeBits"`
	KnowledgeNodes []int   `json:"knowledgeNodes"`
	ParentContainers [][]interface{} `json:"parents"`
	DoneDateTime *string   `json:"doneDateTime"`
}

func main() {
	// Database connection
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

	ctx := context.Background()

	// Read JSON file
	jsonPath := `C:\Programming\NodeJS\dashboard\data\tasks.json`
	log.Printf("Reading tasks from: %s", jsonPath)
	
	jsonData, err := os.ReadFile(jsonPath)
	if err != nil {
		log.Fatalf("failed to read JSON file: %v", err)
	}

	// Parse JSON
	var tasks []TaskJSON
	if err := json.Unmarshal(jsonData, &tasks); err != nil {
		log.Fatalf("failed to parse JSON: %v", err)
	}

	log.Printf("Found %d tasks to import", len(tasks))

	// Create a map to track which tasks have been inserted
	// This helps us handle parent-child relationships
	taskMap := make(map[int]bool)
	
	// First pass: Insert all tasks
	log.Println("Step 1: Inserting tasks into database...")
	inserted := 0
	skipped := 0
	
	for _, taskJSON := range tasks {
		// Check if task already exists
		exists, err := client.Task.Query().
			Where(task.ID(taskJSON.ID)).
			Exist(ctx)
		if err != nil {
			log.Printf("Error checking if task %d exists: %v", taskJSON.ID, err)
			continue
		}
		
		if exists {
			log.Printf("Task %d already exists, skipping...", taskJSON.ID)
			skipped++
			taskMap[taskJSON.ID] = true
			continue
		}

		// Parse doneDateTime
		var doneDateTime *time.Time
		if taskJSON.DoneDateTime != nil && *taskJSON.DoneDateTime != "" && *taskJSON.DoneDateTime != "null" {
			parsed, err := time.Parse(time.RFC3339, *taskJSON.DoneDateTime)
			if err != nil {
				// Try alternative format
				parsed, err = time.Parse("2006-01-02T15:04:05.000Z", *taskJSON.DoneDateTime)
				if err != nil {
					log.Printf("Warning: Could not parse doneDateTime for task %d: %v", taskJSON.ID, err)
				} else {
					doneDateTime = &parsed
				}
			} else {
				doneDateTime = &parsed
			}
		}

		// Create task
		taskBuilder := client.Task.Create().
			SetID(taskJSON.ID).
			SetDescription(taskJSON.Description).
			SetTags(taskJSON.Tags).
			SetDone(taskJSON.Done).
			SetNotes(taskJSON.Notes).
			SetProblems(taskJSON.Problems).
			SetQuestions(taskJSON.Questions).
			SetActions(taskJSON.Actions).
			SetDefinitions(taskJSON.Definitions).
			SetKnowledgeBits(taskJSON.KnowledgeBits).
			SetKnowledgeNodes(taskJSON.KnowledgeNodes).
			SetParentContainers(taskJSON.ParentContainers)

		if doneDateTime != nil {
			taskBuilder.SetDoneDateTime(*doneDateTime)
		}

		_, err = taskBuilder.Save(ctx)
		if err != nil {
			log.Printf("Error inserting task %d: %v", taskJSON.ID, err)
			continue
		}

		taskMap[taskJSON.ID] = true
		inserted++
		
		if inserted%100 == 0 {
			log.Printf("Inserted %d tasks...", inserted)
		}
	}

	log.Printf("Step 1 complete: Inserted %d tasks, skipped %d existing tasks", inserted, skipped)

	// Second pass: Create parent-child relationships in container_children
	log.Println("Step 2: Creating parent-child relationships...")
	relationshipsCreated := 0
	relationshipsSkipped := 0

	for _, taskJSON := range tasks {
		// Only process if this task exists in the database
		if !taskMap[taskJSON.ID] {
			continue
		}

		// For each child task ID in the tasks array
		for _, childID := range taskJSON.Tasks {
			// Check if child task exists
			if !taskMap[childID] {
				log.Printf("Warning: Child task %d does not exist, skipping relationship", childID)
				continue
			}

			// Check if relationship already exists
			exists, err := client.ContainerChild.Query().
				Where(
					containerchild.ParentTypeEQ(containerchild.ParentTypeTask),
					containerchild.ParentID(taskJSON.ID),
					containerchild.ChildTypeEQ(containerchild.ChildTypeTask),
					containerchild.ChildID(childID),
				).
				Exist(ctx)
			if err != nil {
				log.Printf("Error checking relationship %d -> %d: %v", taskJSON.ID, childID, err)
				continue
			}

			if exists {
				relationshipsSkipped++
				continue
			}

			// Create relationship
			_, err = client.ContainerChild.Create().
				SetParentType(containerchild.ParentTypeTask).
				SetParentID(taskJSON.ID).
				SetChildType(containerchild.ChildTypeTask).
				SetChildID(childID).
				Save(ctx)
			if err != nil {
				log.Printf("Error creating relationship %d -> %d: %v", taskJSON.ID, childID, err)
				continue
			}

			relationshipsCreated++
			
			if relationshipsCreated%100 == 0 {
				log.Printf("Created %d relationships...", relationshipsCreated)
			}
		}
	}

	log.Printf("Step 2 complete: Created %d relationships, skipped %d existing relationships", 
		relationshipsCreated, relationshipsSkipped)
	
	log.Println("Import completed successfully!")
	fmt.Printf("\nSummary:\n")
	fmt.Printf("  Tasks inserted: %d\n", inserted)
	fmt.Printf("  Tasks skipped (already exist): %d\n", skipped)
	fmt.Printf("  Relationships created: %d\n", relationshipsCreated)
	fmt.Printf("  Relationships skipped (already exist): %d\n", relationshipsSkipped)
}

