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
	"arturgudiev/dashboard/ent/epic"
	"arturgudiev/dashboard/ent/knowledgenode"
	"arturgudiev/dashboard/ent/problem"
	"arturgudiev/dashboard/ent/question"
	"arturgudiev/dashboard/ent/schema"
	"arturgudiev/dashboard/ent/story"
	"arturgudiev/dashboard/ent/task"

	_ "github.com/lib/pq"
)

// TaskJSON represents the structure of tasks in the JSON file
type TaskJSON struct {
	ID               int             `json:"_id"`
	Description      string          `json:"description"`
	Tags             []string        `json:"tags"`
	Done             FlexibleBool    `json:"done"`
	Notes            string          `json:"notes"`
	Tasks            []int           `json:"tasks"` // Child task IDs
	Problems         []int           `json:"problems"`
	Questions        []int           `json:"questions"`
	Actions          []int           `json:"actions"`
	Definitions      []int           `json:"definitions"`
	KnowledgeBits    []int           `json:"knowledgeBits"`
	KnowledgeNodes   []int           `json:"knowledgeNodes"`
	ParentContainers [][]interface{} `json:"parents"`
	DoneDateTime     *string         `json:"doneDateTime"`
}

// importTasks imports tasks from a JSON file
func importTasks(jsonPath string) error {
	// Database connection
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost/dashboard?sslmode=disable"
	}

	// Create Ent client
	client, err := ent.Open("postgres", dbURL)
	if err != nil {
		return fmt.Errorf("failed opening connection to postgres: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Use provided path or default
	if jsonPath == "" {
		jsonPath = `C:\Programming\NodeJS\dashboard\data\tasks.json`
	}
	log.Printf("Reading tasks from: %s", jsonPath)

	jsonData, err := os.ReadFile(jsonPath)
	if err != nil {
		return fmt.Errorf("failed to read JSON file: %v", err)
	}

	// Parse JSON
	var tasks []TaskJSON
	if err := json.Unmarshal(jsonData, &tasks); err != nil {
		return fmt.Errorf("failed to parse JSON: %v", err)
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
			SetDone(bool(taskJSON.Done)).
			SetNotes(taskJSON.Notes)

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

	// Create a map of tasks by ID for quick lookup of parent_order
	taskByID := make(map[int]*TaskJSON)
	for i := range tasks {
		taskByID[tasks[i].ID] = &tasks[i]
	}

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
		// The index in the array is the child_order
		for childOrder, childID := range taskJSON.Tasks {
			// Check if child task exists
			if !taskMap[childID] {
				log.Printf("Warning: Child task %d does not exist, skipping relationship", childID)
				continue
			}

			// Find parent_order: look for this parent in the child's parents array
			parentOrder := -1
			if childTask, exists := taskByID[childID]; exists {
				for idx, parentDesc := range childTask.ParentContainers {
					// parentDesc is [type, id] array
					if len(parentDesc) >= 2 {
						// Check if the ID matches (handle both string and float64 from JSON)
						var parentIDFromDesc int
						switch v := parentDesc[1].(type) {
						case float64:
							parentIDFromDesc = int(v)
						case int:
							parentIDFromDesc = v
						}

						if parentIDFromDesc == taskJSON.ID {
							parentOrder = idx
							break
						}
					}
				}
			}

			// If parent not found in child's parents array, default to 0
			if parentOrder == -1 {
				parentOrder = 0
			}

			// Check if relationship already exists
			exists, err := client.ContainerChild.Query().
				Where(
					containerchild.ParentTypeEQ(schema.ContainerTypeTask),
					containerchild.ParentID(taskJSON.ID),
					containerchild.ChildTypeEQ(schema.ContainerTypeTask),
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

			// Create relationship with both order fields
			_, err = client.ContainerChild.Create().
				SetParentType(schema.ContainerTypeTask).
				SetParentID(taskJSON.ID).
				SetChildType(schema.ContainerTypeTask).
				SetChildID(childID).
				SetChildOrder(childOrder).   // Position in parent's children array
				SetParentOrder(parentOrder). // Position in child's parents array
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

	return nil
}

// FlexibleBool can unmarshal both string and bool values
type FlexibleBool bool

func (fb *FlexibleBool) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*fb = FlexibleBool(s == "true" || s == "1" || s == "yes")
		return nil
	}
	var b bool
	if err := json.Unmarshal(data, &b); err != nil {
		return err
	}
	*fb = FlexibleBool(b)
	return nil
}

// EpicJSON represents the structure of epics in the JSON file
type EpicJSON struct {
	ID               int             `json:"_id"`
	Description      string          `json:"description"`
	Tags             []string        `json:"tags"`
	Closed           FlexibleBool    `json:"closed"`
	Notes            string          `json:"notes"`
	Epics            []int           `json:"epics"`   // Child epic IDs
	Stories          []int           `json:"stories"` // Child story IDs
	Tasks            []int           `json:"tasks"`   // Child task IDs
	ParentContainers [][]interface{} `json:"parents"`
	DoneDateTime     *string         `json:"doneDateTime"`
}

// importEpics imports epics from a JSON file
func importEpics(jsonPath string) error {
	// Database connection
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost/dashboard?sslmode=disable"
	}

	// Create Ent client
	client, err := ent.Open("postgres", dbURL)
	if err != nil {
		return fmt.Errorf("failed opening connection to postgres: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Use provided path or default
	if jsonPath == "" {
		jsonPath = `C:\Programming\NodeJS\dashboard\data\epics.json`
	}
	log.Printf("Reading epics from: %s", jsonPath)

	jsonData, err := os.ReadFile(jsonPath)
	if err != nil {
		return fmt.Errorf("failed to read JSON file: %v", err)
	}

	// Parse JSON
	var epics []EpicJSON
	if err := json.Unmarshal(jsonData, &epics); err != nil {
		return fmt.Errorf("failed to parse JSON: %v", err)
	}

	log.Printf("Found %d epics to import", len(epics))

	// Create a map to track which epics have been inserted
	epicMap := make(map[int]bool)

	// First pass: Insert all epics
	log.Println("Step 1: Inserting epics into database...")
	inserted := 0
	skipped := 0

	for _, epicJSON := range epics {
		// Check if epic already exists
		exists, err := client.Epic.Query().
			Where(epic.ID(epicJSON.ID)).
			Exist(ctx)
		if err != nil {
			log.Printf("Error checking if epic %d exists: %v", epicJSON.ID, err)
			continue
		}

		if exists {
			log.Printf("Epic %d already exists, skipping...", epicJSON.ID)
			skipped++
			epicMap[epicJSON.ID] = true
			continue
		}

		// Parse doneDateTime
		var doneDateTime *time.Time
		if epicJSON.DoneDateTime != nil && *epicJSON.DoneDateTime != "" && *epicJSON.DoneDateTime != "null" {
			parsed, err := time.Parse(time.RFC3339, *epicJSON.DoneDateTime)
			if err != nil {
				// Try alternative format
				parsed, err = time.Parse("2006-01-02T15:04:05.000Z", *epicJSON.DoneDateTime)
				if err != nil {
					log.Printf("Warning: Could not parse doneDateTime for epic %d: %v", epicJSON.ID, err)
				} else {
					doneDateTime = &parsed
				}
			} else {
				doneDateTime = &parsed
			}
		}

		// Create epic
		epicBuilder := client.Epic.Create().
			SetID(epicJSON.ID).
			SetDescription(epicJSON.Description).
			SetTags(epicJSON.Tags).
			SetClosed(bool(epicJSON.Closed)).
			SetNotes(epicJSON.Notes)

		if doneDateTime != nil {
			epicBuilder.SetDoneDateTime(*doneDateTime)
		}

		_, err = epicBuilder.Save(ctx)
		if err != nil {
			log.Printf("Error inserting epic %d: %v", epicJSON.ID, err)
			continue
		}

		epicMap[epicJSON.ID] = true
		inserted++

		if inserted%100 == 0 {
			log.Printf("Inserted %d epics...", inserted)
		}
	}

	log.Printf("Step 1 complete: Inserted %d epics, skipped %d existing epics", inserted, skipped)

	// Create a map of epics by ID for quick lookup of parent_order
	epicByID := make(map[int]*EpicJSON)
	for i := range epics {
		epicByID[epics[i].ID] = &epics[i]
	}

	// Second pass: Create parent-child relationships in container_children
	log.Println("Step 2: Creating parent-child relationships...")
	relationshipsCreated := 0
	relationshipsSkipped := 0

	// Helper function to create relationships for a child type
	createRelationships := func(parentID int, childIDs []int, childType schema.ContainerType, parentType schema.ContainerType) {
		for childOrder, childID := range childIDs {
			// Check if child exists in database
			var childExists bool
			switch childType {
			case schema.ContainerTypeEpic:
				exists, err := client.Epic.Query().Where(epic.ID(childID)).Exist(ctx)
				if err != nil {
					log.Printf("Error checking if epic %d exists: %v", childID, err)
					continue
				}
				childExists = exists
			case schema.ContainerTypeStory:
				exists, err := client.Story.Query().Where(story.ID(childID)).Exist(ctx)
				if err != nil {
					log.Printf("Error checking if story %d exists: %v", childID, err)
					continue
				}
				childExists = exists
			case schema.ContainerTypeTask:
				exists, err := client.Task.Query().Where(task.ID(childID)).Exist(ctx)
				if err != nil {
					log.Printf("Error checking if task %d exists: %v", childID, err)
					continue
				}
				childExists = exists
			default:
				log.Printf("Unknown child type: %v", childType)
				continue
			}

			if !childExists {
				log.Printf("Warning: Child %s %d does not exist, skipping relationship", childType, childID)
				continue
			}

			// Find parent_order: look for this parent in the child's parents array
			// Only works if child is same type (epic->epic), otherwise default to 0
			parentOrder := 0
			if childType == schema.ContainerTypeEpic {
				if childEpic, exists := epicByID[childID]; exists {
					for idx, parentDesc := range childEpic.ParentContainers {
						if len(parentDesc) >= 2 {
							var parentIDFromDesc int
							switch v := parentDesc[1].(type) {
							case float64:
								parentIDFromDesc = int(v)
							case int:
								parentIDFromDesc = v
							}

							if parentIDFromDesc == parentID {
								parentOrder = idx
								break
							}
						}
					}
				}
			}

			// Check if relationship already exists
			exists, err := client.ContainerChild.Query().
				Where(
					containerchild.ParentTypeEQ(parentType),
					containerchild.ParentID(parentID),
					containerchild.ChildTypeEQ(childType),
					containerchild.ChildID(childID),
				).
				Exist(ctx)
			if err != nil {
				log.Printf("Error checking relationship %d -> %d: %v", parentID, childID, err)
				continue
			}

			if exists {
				relationshipsSkipped++
				continue
			}

			// Create relationship with both order fields
			_, err = client.ContainerChild.Create().
				SetParentType(parentType).
				SetParentID(parentID).
				SetChildType(childType).
				SetChildID(childID).
				SetChildOrder(childOrder).
				SetParentOrder(parentOrder).
				Save(ctx)
			if err != nil {
				log.Printf("Error creating relationship %d -> %d: %v", parentID, childID, err)
				continue
			}

			relationshipsCreated++

			if relationshipsCreated%100 == 0 {
				log.Printf("Created %d relationships...", relationshipsCreated)
			}
		}
	}

	for _, epicJSON := range epics {
		// Only process if this epic exists in the database
		if !epicMap[epicJSON.ID] {
			continue
		}

		// Create relationships for child epics
		createRelationships(epicJSON.ID, epicJSON.Epics, schema.ContainerTypeEpic, schema.ContainerTypeEpic)

		// Create relationships for child stories
		createRelationships(epicJSON.ID, epicJSON.Stories, schema.ContainerTypeStory, schema.ContainerTypeEpic)

		// Create relationships for child tasks
		createRelationships(epicJSON.ID, epicJSON.Tasks, schema.ContainerTypeTask, schema.ContainerTypeEpic)
	}

	log.Printf("Step 2 complete: Created %d relationships, skipped %d existing relationships",
		relationshipsCreated, relationshipsSkipped)

	log.Println("Import completed successfully!")
	fmt.Printf("\nSummary:\n")
	fmt.Printf("  Epics inserted: %d\n", inserted)
	fmt.Printf("  Epics skipped (already exist): %d\n", skipped)
	fmt.Printf("  Relationships created: %d\n", relationshipsCreated)
	fmt.Printf("  Relationships skipped (already exist): %d\n", relationshipsSkipped)

	return nil
}

// StoryJSON represents the structure of stories in the JSON file
type StoryJSON struct {
	ID               int             `json:"_id"`
	Description      string          `json:"description"`
	Tags             []string        `json:"tags"`
	Closed           FlexibleBool    `json:"closed"`
	Notes            string          `json:"notes"`
	Stories          []int           `json:"stories"` // Child story IDs
	Tasks            []int           `json:"tasks"`   // Child task IDs
	ParentContainers [][]interface{} `json:"parents"`
	DoneDateTime     *string         `json:"doneDateTime"`
}

// importStories imports stories from a JSON file
func importStories(jsonPath string) error {
	// Database connection
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost/dashboard?sslmode=disable"
	}

	// Create Ent client
	client, err := ent.Open("postgres", dbURL)
	if err != nil {
		return fmt.Errorf("failed opening connection to postgres: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Use provided path or default
	if jsonPath == "" {
		jsonPath = `C:\Programming\NodeJS\dashboard\data\stories.json`
	}
	log.Printf("Reading stories from: %s", jsonPath)

	jsonData, err := os.ReadFile(jsonPath)
	if err != nil {
		return fmt.Errorf("failed to read JSON file: %v", err)
	}

	// Parse JSON
	var stories []StoryJSON
	if err := json.Unmarshal(jsonData, &stories); err != nil {
		return fmt.Errorf("failed to parse JSON: %v", err)
	}

	log.Printf("Found %d stories to import", len(stories))

	// Create a map to track which stories have been inserted
	storyMap := make(map[int]bool)

	// First pass: Insert all stories
	log.Println("Step 1: Inserting stories into database...")
	inserted := 0
	skipped := 0

	for _, storyJSON := range stories {
		// Check if story already exists
		exists, err := client.Story.Query().
			Where(story.ID(storyJSON.ID)).
			Exist(ctx)
		if err != nil {
			log.Printf("Error checking if story %d exists: %v", storyJSON.ID, err)
			continue
		}

		if exists {
			log.Printf("Story %d already exists, skipping...", storyJSON.ID)
			skipped++
			storyMap[storyJSON.ID] = true
			continue
		}

		// Parse doneDateTime
		var doneDateTime *time.Time
		if storyJSON.DoneDateTime != nil && *storyJSON.DoneDateTime != "" && *storyJSON.DoneDateTime != "null" {
			parsed, err := time.Parse(time.RFC3339, *storyJSON.DoneDateTime)
			if err != nil {
				// Try alternative format
				parsed, err = time.Parse("2006-01-02T15:04:05.000Z", *storyJSON.DoneDateTime)
				if err != nil {
					log.Printf("Warning: Could not parse doneDateTime for story %d: %v", storyJSON.ID, err)
				} else {
					doneDateTime = &parsed
				}
			} else {
				doneDateTime = &parsed
			}
		}

		// Create story
		storyBuilder := client.Story.Create().
			SetID(storyJSON.ID).
			SetDescription(storyJSON.Description).
			SetTags(storyJSON.Tags).
			SetClosed(bool(storyJSON.Closed)).
			SetNotes(storyJSON.Notes)

		if doneDateTime != nil {
			storyBuilder.SetDoneDateTime(*doneDateTime)
		}

		_, err = storyBuilder.Save(ctx)
		if err != nil {
			log.Printf("Error inserting story %d: %v", storyJSON.ID, err)
			continue
		}

		storyMap[storyJSON.ID] = true
		inserted++

		if inserted%100 == 0 {
			log.Printf("Inserted %d stories...", inserted)
		}
	}

	log.Printf("Step 1 complete: Inserted %d stories, skipped %d existing stories", inserted, skipped)

	// Create a map of stories by ID for quick lookup of parent_order
	storyByID := make(map[int]*StoryJSON)
	for i := range stories {
		storyByID[stories[i].ID] = &stories[i]
	}

	// Second pass: Create parent-child relationships in container_children
	log.Println("Step 2: Creating parent-child relationships...")
	relationshipsCreated := 0
	relationshipsSkipped := 0

	// Helper function to create relationships for a child type
	createRelationships := func(parentID int, childIDs []int, childType schema.ContainerType, parentType schema.ContainerType) {
		for childOrder, childID := range childIDs {
			// Check if child exists in database
			var childExists bool
			switch childType {
			case schema.ContainerTypeStory:
				exists, err := client.Story.Query().Where(story.ID(childID)).Exist(ctx)
				if err != nil {
					log.Printf("Error checking if story %d exists: %v", childID, err)
					continue
				}
				childExists = exists
			case schema.ContainerTypeTask:
				exists, err := client.Task.Query().Where(task.ID(childID)).Exist(ctx)
				if err != nil {
					log.Printf("Error checking if task %d exists: %v", childID, err)
					continue
				}
				childExists = exists
			default:
				log.Printf("Unknown child type: %v", childType)
				continue
			}

			if !childExists {
				log.Printf("Warning: Child %s %d does not exist, skipping relationship", childType, childID)
				continue
			}

			// Find parent_order: look for this parent in the child's parents array
			// Only works if child is same type (story->story), otherwise default to 0
			parentOrder := 0
			if childType == schema.ContainerTypeStory {
				if childStory, exists := storyByID[childID]; exists {
					for idx, parentDesc := range childStory.ParentContainers {
						if len(parentDesc) >= 2 {
							var parentIDFromDesc int
							switch v := parentDesc[1].(type) {
							case float64:
								parentIDFromDesc = int(v)
							case int:
								parentIDFromDesc = v
							}

							if parentIDFromDesc == parentID {
								parentOrder = idx
								break
							}
						}
					}
				}
			}

			// Check if relationship already exists
			exists, err := client.ContainerChild.Query().
				Where(
					containerchild.ParentTypeEQ(parentType),
					containerchild.ParentID(parentID),
					containerchild.ChildTypeEQ(childType),
					containerchild.ChildID(childID),
				).
				Exist(ctx)
			if err != nil {
				log.Printf("Error checking relationship %d -> %d: %v", parentID, childID, err)
				continue
			}

			if exists {
				relationshipsSkipped++
				continue
			}

			// Create relationship with both order fields
			_, err = client.ContainerChild.Create().
				SetParentType(parentType).
				SetParentID(parentID).
				SetChildType(childType).
				SetChildID(childID).
				SetChildOrder(childOrder).
				SetParentOrder(parentOrder).
				Save(ctx)
			if err != nil {
				log.Printf("Error creating relationship %d -> %d: %v", parentID, childID, err)
				continue
			}

			relationshipsCreated++

			if relationshipsCreated%100 == 0 {
				log.Printf("Created %d relationships...", relationshipsCreated)
			}
		}
	}

	for _, storyJSON := range stories {
		// Only process if this story exists in the database
		if !storyMap[storyJSON.ID] {
			continue
		}

		// Create relationships for child stories
		createRelationships(storyJSON.ID, storyJSON.Stories, schema.ContainerTypeStory, schema.ContainerTypeStory)

		// Create relationships for child tasks
		createRelationships(storyJSON.ID, storyJSON.Tasks, schema.ContainerTypeTask, schema.ContainerTypeStory)
	}

	log.Printf("Step 2 complete: Created %d relationships, skipped %d existing relationships",
		relationshipsCreated, relationshipsSkipped)

	log.Println("Import completed successfully!")
	fmt.Printf("\nSummary:\n")
	fmt.Printf("  Stories inserted: %d\n", inserted)
	fmt.Printf("  Stories skipped (already exist): %d\n", skipped)
	fmt.Printf("  Relationships created: %d\n", relationshipsCreated)
	fmt.Printf("  Relationships skipped (already exist): %d\n", relationshipsSkipped)

	return nil
}

// KnowledgeNodeJSON represents the structure of knowledge nodes in the JSON file
type KnowledgeNodeJSON struct {
	ID               int             `json:"_id"`
	Name             string          `json:"name"`
	Description      string          `json:"description"` // Not used in DB schema, but present in JSON
	Tags             []string        `json:"tags"`
	Notes            string          `json:"notes"`
	KnowledgeNodes   []int           `json:"knowledgeNodes"` // Child knowledge node IDs
	Tasks            []int           `json:"tasks"`          // Child task IDs
	Problems         []int           `json:"problems"`       // Child problem IDs
	Questions        []int           `json:"questions"`      // Child question IDs
	Actions          []int           `json:"actions"`        // Not in DB schema
	Definitions      []int           `json:"definitions"`    // Not in DB schema
	KnowledgeBits    []int           `json:"knowledgeBits"`  // Not in DB schema
	ParentContainers [][]interface{} `json:"parents"`
}

// importKnowledgeNodes imports knowledge nodes from a JSON file
func importKnowledgeNodes(jsonPath string) error {
	// Database connection
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost/dashboard?sslmode=disable"
	}

	// Create Ent client
	client, err := ent.Open("postgres", dbURL)
	if err != nil {
		return fmt.Errorf("failed opening connection to postgres: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Use provided path or default
	if jsonPath == "" {
		jsonPath = `C:\Programming\NodeJS\dashboard\data\knowledge-nodes.json`
	}
	log.Printf("Reading knowledge nodes from: %s", jsonPath)

	jsonData, err := os.ReadFile(jsonPath)
	if err != nil {
		return fmt.Errorf("failed to read JSON file: %v", err)
	}

	// Parse JSON
	var knowledgeNodes []KnowledgeNodeJSON
	if err := json.Unmarshal(jsonData, &knowledgeNodes); err != nil {
		return fmt.Errorf("failed to parse JSON: %v", err)
	}

	log.Printf("Found %d knowledge nodes to import", len(knowledgeNodes))

	// Create a map to track which knowledge nodes have been inserted
	knowledgeNodeMap := make(map[int]bool)

	// First pass: Insert all knowledge nodes
	log.Println("Step 1: Inserting knowledge nodes into database...")
	inserted := 0
	skipped := 0

	for _, knJSON := range knowledgeNodes {
		// Check if knowledge node already exists
		exists, err := client.KnowledgeNode.Query().
			Where(knowledgenode.ID(knJSON.ID)).
			Exist(ctx)
		if err != nil {
			log.Printf("Error checking if knowledge node %d exists: %v", knJSON.ID, err)
			continue
		}

		if exists {
			log.Printf("Knowledge node %d already exists, skipping...", knJSON.ID)
			skipped++
			knowledgeNodeMap[knJSON.ID] = true
			continue
		}

		// Create knowledge node
		_, err = client.KnowledgeNode.Create().
			SetID(knJSON.ID).
			SetName(knJSON.Name).
			SetTags(knJSON.Tags).
			SetNotes(knJSON.Notes).
			Save(ctx)
		if err != nil {
			log.Printf("Error inserting knowledge node %d: %v", knJSON.ID, err)
			continue
		}

		knowledgeNodeMap[knJSON.ID] = true
		inserted++

		if inserted%100 == 0 {
			log.Printf("Inserted %d knowledge nodes...", inserted)
		}
	}

	log.Printf("Step 1 complete: Inserted %d knowledge nodes, skipped %d existing knowledge nodes", inserted, skipped)

	// Create a map of knowledge nodes by ID for quick lookup of parent_order
	knowledgeNodeByID := make(map[int]*KnowledgeNodeJSON)
	for i := range knowledgeNodes {
		knowledgeNodeByID[knowledgeNodes[i].ID] = &knowledgeNodes[i]
	}

	// Second pass: Create parent-child relationships in container_children
	log.Println("Step 2: Creating parent-child relationships...")
	relationshipsCreated := 0
	relationshipsSkipped := 0

	// Helper function to create relationships for a child type
	createRelationships := func(parentID int, childIDs []int, childType schema.ContainerType, parentType schema.ContainerType) {
		for childOrder, childID := range childIDs {
			// Check if child exists in database
			var childExists bool
			switch childType {
			case schema.ContainerTypeKnowledgeNode:
				exists, err := client.KnowledgeNode.Query().Where(knowledgenode.ID(childID)).Exist(ctx)
				if err != nil {
					log.Printf("Error checking if knowledge node %d exists: %v", childID, err)
					continue
				}
				childExists = exists
			case schema.ContainerTypeTask:
				exists, err := client.Task.Query().Where(task.ID(childID)).Exist(ctx)
				if err != nil {
					log.Printf("Error checking if task %d exists: %v", childID, err)
					continue
				}
				childExists = exists
			case schema.ContainerTypeProblem:
				exists, err := client.Problem.Query().Where(problem.ID(childID)).Exist(ctx)
				if err != nil {
					log.Printf("Error checking if problem %d exists: %v", childID, err)
					continue
				}
				childExists = exists
			case schema.ContainerTypeQuestion:
				exists, err := client.Question.Query().Where(question.ID(childID)).Exist(ctx)
				if err != nil {
					log.Printf("Error checking if question %d exists: %v", childID, err)
					continue
				}
				childExists = exists
			default:
				log.Printf("Unknown child type: %v", childType)
				continue
			}

			if !childExists {
				log.Printf("Warning: Child %s %d does not exist, skipping relationship", childType, childID)
				continue
			}

			// Find parent_order: look for this parent in the child's parents array
			// Only works if child is same type (knowledge-node->knowledge-node), otherwise default to 0
			parentOrder := 0
			if childType == schema.ContainerTypeKnowledgeNode {
				if childKN, exists := knowledgeNodeByID[childID]; exists {
					for idx, parentDesc := range childKN.ParentContainers {
						if len(parentDesc) >= 2 {
							var parentIDFromDesc int
							switch v := parentDesc[1].(type) {
							case float64:
								parentIDFromDesc = int(v)
							case int:
								parentIDFromDesc = v
							}

							if parentIDFromDesc == parentID {
								parentOrder = idx
								break
							}
						}
					}
				}
			}

			// Check if relationship already exists
			exists, err := client.ContainerChild.Query().
				Where(
					containerchild.ParentTypeEQ(parentType),
					containerchild.ParentID(parentID),
					containerchild.ChildTypeEQ(childType),
					containerchild.ChildID(childID),
				).
				Exist(ctx)
			if err != nil {
				log.Printf("Error checking relationship %d -> %d: %v", parentID, childID, err)
				continue
			}

			if exists {
				relationshipsSkipped++
				continue
			}

			// Create relationship with both order fields
			_, err = client.ContainerChild.Create().
				SetParentType(parentType).
				SetParentID(parentID).
				SetChildType(childType).
				SetChildID(childID).
				SetChildOrder(childOrder).
				SetParentOrder(parentOrder).
				Save(ctx)
			if err != nil {
				log.Printf("Error creating relationship %d -> %d: %v", parentID, childID, err)
				continue
			}

			relationshipsCreated++

			if relationshipsCreated%100 == 0 {
				log.Printf("Created %d relationships...", relationshipsCreated)
			}
		}
	}

	for _, knJSON := range knowledgeNodes {
		// Only process if this knowledge node exists in the database
		if !knowledgeNodeMap[knJSON.ID] {
			continue
		}

		// Create relationships for child knowledge nodes
		createRelationships(knJSON.ID, knJSON.KnowledgeNodes, schema.ContainerTypeKnowledgeNode, schema.ContainerTypeKnowledgeNode)

		// Create relationships for child tasks
		createRelationships(knJSON.ID, knJSON.Tasks, schema.ContainerTypeTask, schema.ContainerTypeKnowledgeNode)

		// Create relationships for child problems
		createRelationships(knJSON.ID, knJSON.Problems, schema.ContainerTypeProblem, schema.ContainerTypeKnowledgeNode)

		// Create relationships for child questions
		createRelationships(knJSON.ID, knJSON.Questions, schema.ContainerTypeQuestion, schema.ContainerTypeKnowledgeNode)
	}

	log.Printf("Step 2 complete: Created %d relationships, skipped %d existing relationships",
		relationshipsCreated, relationshipsSkipped)

	log.Println("Import completed successfully!")
	fmt.Printf("\nSummary:\n")
	fmt.Printf("  Knowledge nodes inserted: %d\n", inserted)
	fmt.Printf("  Knowledge nodes skipped (already exist): %d\n", skipped)
	fmt.Printf("  Relationships created: %d\n", relationshipsCreated)
	fmt.Printf("  Relationships skipped (already exist): %d\n", relationshipsSkipped)

	return nil
}
