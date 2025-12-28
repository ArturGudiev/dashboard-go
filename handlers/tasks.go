package handlers

import (
	"log"
	"strconv"
	"time"

	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/ent/containerchild"
	"arturgudiev/dashboard/ent/schema"
	"arturgudiev/dashboard/ent/task"

	"github.com/gin-gonic/gin"
	"github.com/niemeyer/pretty"
)

// GetTaskByID handles GET /task/:id
// @Summary      Get task by ID
// @Description  Returns a task by its ID
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Task ID"
// @Success      200  {object}  TaskResponse
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /task/{id} [get]
func (h *Handler) GetTaskByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid task ID"})
		return
	}

	task, err := h.App.Client.Task.Get(c.Request.Context(), id)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(404, gin.H{"error": "Task not found"})
			return
		}
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	pretty.Print(task)

	// Convert to custom response type to ensure all fields are included
	response := TaskResponse{
		ID:               task.ID,
		Description:      task.Description,
		Tags:             task.Tags,
		Done:             task.Done,
		Notes:            task.Notes,
		Problems:         task.Problems,
		Questions:        task.Questions,
		Actions:          task.Actions,
		Definitions:      task.Definitions,
		KnowledgeBits:    task.KnowledgeBits,
		ParentContainers: task.ParentContainers,
		KnowledgeNodes:   task.KnowledgeNodes,
		DoneDateTime:     task.DoneDateTime,
	}

	c.JSON(200, response)
}

// GetTasksByIDs handles POST /get-tasks
// @Summary      Get tasks by IDs
// @Description  Returns multiple tasks by their IDs
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Param        request  body      IDsRequest  true  "List of task IDs"
// @Success      200      {array}   TaskResponse
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /get-tasks [post]
func (h *Handler) GetTasksByIDs(c *gin.Context) {
	var req IDsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	tasks, err := h.App.Client.Task.Query().Where(task.IDIn(req.IDs...)).All(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// Convert to slice of TaskResponse to ensure all fields are included
	responses := make([]TaskResponse, len(tasks))
	for i, t := range tasks {
		responses[i] = TaskResponse{
			ID:               t.ID,
			Description:      t.Description,
			Tags:             t.Tags,
			Done:             t.Done,
			Notes:            t.Notes,
			Problems:         t.Problems,
			Questions:        t.Questions,
			Actions:          t.Actions,
			Definitions:      t.Definitions,
			KnowledgeBits:    t.KnowledgeBits,
			ParentContainers: t.ParentContainers,
			KnowledgeNodes:   t.KnowledgeNodes,
			DoneDateTime:     t.DoneDateTime,
		}
	}

	c.JSON(200, responses)
}

// AddAnonymousTask handles PUT /add-anonymous-task
// @Summary      Create anonymous task
// @Description  Creates a simple anonymous task with default values
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Success      200  {object}  TaskResponse
// @Failure      500  {object}  map[string]string
// @Router       /add-anonymous-task [put]
func (h *Handler) AddAnonymousTask(c *gin.Context) {
	newTask, err := h.App.Client.Task.Create().
		SetDescription("Simple task").
		SetDone(true).
		SetTags([]string{}).
		SetNotes("").
		SetDoneDateTime(time.Now()).
		Save(c.Request.Context())

	if err != nil {
		log.Printf("Error creating task: %v", err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Successfully created task with ID: %d", newTask.ID)

	// Convert to custom response type to ensure all fields are included
	response := TaskResponse{
		ID:               newTask.ID,
		Description:      newTask.Description,
		Tags:             newTask.Tags,
		Done:             newTask.Done,
		Notes:            newTask.Notes,
		Problems:         newTask.Problems,
		Questions:        newTask.Questions,
		Actions:          newTask.Actions,
		Definitions:      newTask.Definitions,
		KnowledgeBits:    newTask.KnowledgeBits,
		ParentContainers: newTask.ParentContainers,
		KnowledgeNodes:   newTask.KnowledgeNodes,
		DoneDateTime:     newTask.DoneDateTime,
	}

	c.JSON(200, response)
}

// GetDoneTasks handles GET /done-tasks
// @Summary      Get done tasks from today
// @Description  Returns all done tasks where doneDateTime is today
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Success      200  {array}   ent.Task
// @Failure      500  {object}  map[string]string
// @Router       /done-tasks [get]
func (h *Handler) GetDoneTasks(c *gin.Context) {
	ctx := c.Request.Context()

	// Get start of today (00:00:00)
	now := time.Now()
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	startOfTomorrow := startOfToday.AddDate(0, 0, 1)

	// Query tasks where Done is true and DoneDateTime is today
	tasks, err := h.App.Client.Task.Query().
		Where(
			task.DoneEQ(true),
			task.DoneDateTimeNotNil(),
			task.DoneDateTimeGTE(startOfToday),
			task.DoneDateTimeLT(startOfTomorrow),
		).
		All(ctx)

	if err != nil {
		log.Printf("Error querying done tasks: %v", err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, len(tasks))
}

// FinishTask handles PUT /finish-task/:id
// @Summary      Finish task recursively
// @Description  Marks a task and all its descendants as done
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Task ID"
// @Success      200  {object}  TaskResponse
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /finish-task/{id} [put]
func (h *Handler) FinishTask(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid task ID"})
		return
	}

	ctx := c.Request.Context()

	updatedTask, err := h.App.TaskService.FinishTaskById(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(404, gin.H{"error": "Task not found"})
			return
		}
		log.Printf("Error finishing task %d: %v", id, err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// Convert to custom response type to ensure all fields are included
	response := TaskResponse{
		ID:               updatedTask.ID,
		Description:      updatedTask.Description,
		Tags:             updatedTask.Tags,
		Done:             updatedTask.Done,
		Notes:            updatedTask.Notes,
		Problems:         updatedTask.Problems,
		Questions:        updatedTask.Questions,
		Actions:          updatedTask.Actions,
		Definitions:      updatedTask.Definitions,
		KnowledgeBits:    updatedTask.KnowledgeBits,
		ParentContainers: updatedTask.ParentContainers,
		KnowledgeNodes:   updatedTask.KnowledgeNodes,
		DoneDateTime:     updatedTask.DoneDateTime,
	}

	c.JSON(200, response)
}

// FinishTasks handles PUT /finish-tasks/
// @Summary      Finish multiple tasks recursively
// @Description  Marks multiple tasks and all their descendants as done
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Param        request  body      []TaskIDRequest  true  "List of task IDs"
// @Success      200      {object}  map[string]interface{}
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /finish-tasks/ [put]
func (h *Handler) FinishTasks(c *gin.Context) {
	var tasks []TaskIDRequest
	if err := c.ShouldBindJSON(&tasks); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()

	for _, t := range tasks {
		task, err := h.App.Client.Task.Get(ctx, t.ID)
		if err != nil {
			if ent.IsNotFound(err) {
				log.Printf("Task %d not found, skipping", t.ID)
				continue
			}
			log.Printf("Error getting task %d: %v", t.ID, err)
			continue
		}

		if err := h.App.TaskService.FinishTaskRecursively(ctx, task); err != nil {
			log.Printf("Error finishing task %d recursively: %v", t.ID, err)
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(200, gin.H{})
}

// FinishTasksByIDs handles PUT /finish-tasks-by-ids/
// @Summary      Finish tasks by IDs recursively
// @Description  Marks tasks by their IDs and all their descendants as done
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Param        request  body      []int  true  "List of task IDs"
// @Success      200      {object}  map[string]interface{}
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /finish-tasks-by-ids/ [put]
func (h *Handler) FinishTasksByIDs(c *gin.Context) {
	var ids []int
	if err := c.ShouldBindJSON(&ids); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()

	for _, id := range ids {
		task, err := h.App.Client.Task.Get(ctx, id)
		if err != nil {
			if ent.IsNotFound(err) {
				log.Printf("Task %d not found, skipping", id)
				continue
			}
			log.Printf("Error getting task %d: %v", id, err)
			continue
		}

		if err := h.App.TaskService.FinishTaskRecursively(ctx, task); err != nil {
			log.Printf("Error finishing task %d recursively: %v", id, err)
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(200, gin.H{})
}

// NewTask handles POST /new-task
// @Summary      Create new task
// @Description  Creates a new task with optional parent relationship
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Param        request  body      NewTaskRequest  true  "Task creation request"
// @Success      200      {object}  ent.Task
// @Failure      400      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /new-task [post]
func (h *Handler) NewTask(c *gin.Context) {
	var req NewTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()

	taskBuilder := h.App.Client.Task.Create().
		SetDescription(req.Task.Description).
		SetDone(req.Task.Done)

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

	if req.Parent != nil && req.Parent.Type == "task" {
		parentTask, err := h.App.Client.Task.Get(ctx, req.Parent.Obj.ID)
		if err != nil {
			if ent.IsNotFound(err) {
				c.JSON(404, gin.H{"error": "Parent task not found"})
				return
			}
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		exists, err := h.App.Client.ContainerChild.Query().
			Where(
				containerchild.ParentTypeEQ(schema.ContainerTypeTask),
				containerchild.ParentID(parentTask.ID),
				containerchild.ChildTypeEQ(schema.ContainerTypeTask),
				containerchild.ChildID(newTask.ID),
			).
			Exist(ctx)
		if err != nil {
			log.Printf("Error checking relationship: %v", err)
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		if !exists {
			childCount, err := h.App.Client.ContainerChild.Query().
				Where(
					containerchild.ParentTypeEQ(schema.ContainerTypeTask),
					containerchild.ParentID(parentTask.ID),
					containerchild.ChildTypeEQ(schema.ContainerTypeTask),
				).
				Count(ctx)
			if err != nil {
				log.Printf("Error counting children: %v", err)
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}

			parentCount, err := h.App.Client.ContainerChild.Query().
				Where(
					containerchild.ChildTypeEQ(schema.ContainerTypeTask),
					containerchild.ChildID(newTask.ID),
					containerchild.ParentTypeEQ(schema.ContainerTypeTask),
				).
				Count(ctx)
			if err != nil {
				log.Printf("Error counting parents: %v", err)
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}

			_, err = h.App.Client.ContainerChild.Create().
				SetParentType(schema.ContainerTypeTask).
				SetParentID(parentTask.ID).
				SetChildType(schema.ContainerTypeTask).
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

	c.JSON(200, newTask)
}

// UpdateTask handles PUT /update-task
// @Summary      Update task
// @Description  Updates an existing task by ID
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Param        request  body      UpdateTaskRequest  true  "Task update request"
// @Success      200      {object}  TaskResponse
// @Failure      400      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /update-task [put]
func (h *Handler) UpdateTask(c *gin.Context) {
	var req UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()

	_, err := h.App.Client.Task.Get(ctx, req.ID)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(404, gin.H{"error": "Task not found"})
			return
		}
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	taskBuilder := h.App.Client.Task.UpdateOneID(req.ID).
		SetDescription(req.Description).
		SetDone(req.Done)

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

	// Convert to custom response type to ensure all fields are included
	response := TaskResponse{
		ID:               updatedTask.ID,
		Description:      updatedTask.Description,
		Tags:             updatedTask.Tags,
		Done:             updatedTask.Done,
		Notes:            updatedTask.Notes,
		Problems:         updatedTask.Problems,
		Questions:        updatedTask.Questions,
		Actions:          updatedTask.Actions,
		Definitions:      updatedTask.Definitions,
		KnowledgeBits:    updatedTask.KnowledgeBits,
		ParentContainers: updatedTask.ParentContainers,
		KnowledgeNodes:   updatedTask.KnowledgeNodes,
		DoneDateTime:     updatedTask.DoneDateTime,
	}

	c.JSON(200, response)
}
