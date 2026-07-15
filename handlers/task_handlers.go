package handlers

import (
	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/models"
	"log"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// GetTaskByID handles GET /task/:id
// @Summary      Get task by ID
// @Description  Returns a task by its ID
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Task ID"
// @Success      200  {object}  models.TaskFull
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

	taskFull, err := h.App.TaskService.GetTaskFull(c.Request.Context(), id)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(404, gin.H{"error": "Task not found"})
			return
		}
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, taskFull)
}

// GetTaskReport handles GET /task-report/:id
// @Summary      Task completion report tree
// @Description  Returns a tree of done descendant tasks under the root task. The response body is JSON null when there are no done descendants.
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Root task ID"
// @Success      200  {object}  models.TaskReportTreeNode
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /task-report/{id} [get]
func (h *Handler) GetTaskReport(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid task ID"})
		return
	}

	report, err := h.App.TaskService.GetTaskReport(c.Request.Context(), id)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(404, gin.H{"error": "Task not found"})
			return
		}
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, report)
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

	tasks, err := h.App.TaskService.GetTasksFull(c.Request.Context(), req.IDs)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, tasks)
}

// AddAnonymousTask handles PUT /add-anonymous-task
// @Summary      Create anonymous task
// @Description  Creates a simple anonymous task with default values
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Success      200  {object}  ent.Task
// @Failure      500  {object}  map[string]string
// @Router       /add-anonymous-task [put]
func (h *Handler) AddAnonymousTask(c *gin.Context) {
	newTask, err := h.App.TaskService.AddAnonymousTask(c.Request.Context())

	if err != nil {
		log.Printf("Error creating task: %v", err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, newTask)
}

// GetDoneTasks handles GET /done-tasks
// @Summary      Get done tasks from today
// @Description  Returns all done tasks where doneDateTime is today
// @Param from query string false "From date-time filter (RFC3339)" format(date-time) example(2026-04-23T00:00:00+03:00)
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Success      200  {array}   ent.Task
// @Failure      500  {object}  map[string]string
// @Router       /done-tasks [get]
func (h *Handler) GetDoneTasks(c *gin.Context) {
	var from *time.Time
	if fromDateRaw := c.Query("from"); fromDateRaw != "" {
		parsed, parseErr := time.Parse(time.RFC3339, fromDateRaw)
		if parseErr != nil {
			// Also support UI format used by dashboard-ui: "15:04 02-01-2006"
			parsed, parseErr = time.ParseInLocation("15:04 02-01-2006", fromDateRaw, time.Local)
			if parseErr != nil {
				c.JSON(400, gin.H{"error": "from must be RFC3339 or 'HH:mm DD-MM-YYYY'"})
				return
			}
		}
		from = &parsed
	}

	tasksCount, err := h.App.TaskService.GetDoneTasksCount(c.Request.Context(), from)

	if err != nil {
		log.Printf("Error querying done tasksCount: %v", err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	response := DoneTasksResponse{
		DoneTasks: tasksCount,
	}
	c.JSON(200, response)
}

// GetOpenTasksByDueDate handles GET /tasks/by-due-date
// @Summary      Get open tasks by due date
// @Description  Returns open (not done) tasks whose due date falls on the specified calendar day
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Param        date   query     string  true  "Due date (YYYY-MM-DD)" example(2026-07-15)
// @Success      200    {array}   models.TaskFull
// @Failure      400    {object}  map[string]string
// @Failure      500    {object}  map[string]string
// @Security     AccessTokenCookie
// @Router       /tasks/by-due-date [get]
func (h *Handler) GetOpenTasksByDueDate(c *gin.Context) {
	dateRaw := c.Query("date")
	if dateRaw == "" {
		c.JSON(400, gin.H{"error": "date query parameter is required (YYYY-MM-DD)"})
		return
	}

	dayStart, err := time.ParseInLocation("2006-01-02", dateRaw, time.Local)
	if err != nil {
		c.JSON(400, gin.H{"error": "date must be YYYY-MM-DD"})
		return
	}

	tasks, err := h.App.TaskService.GetOpenTasksFullByDueDate(c.Request.Context(), dayStart)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, tasks)
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
		ID:           updatedTask.ID,
		Description:  updatedTask.Description,
		Tags:         updatedTask.Tags,
		Done:         updatedTask.Done,
		Notes:        updatedTask.Notes,
		DoneDateTime: updatedTask.DoneDateTime,
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
func (h *Handler) FinishTasks(c *gin.Context) { // TODO after Go: modify
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
		task, err := h.App.TasksRepository.GetTask(ctx, id)
		if err != nil {
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
	newTask, err := h.App.TaskService.AddTask(ctx, req.Task, req.Parent)

	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(404, gin.H{"error": "Parent container not found"})
			return
		}
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, newTask)
}

// NewHierarchicalTasks handles POST /new-hierarchical-tasks
// @Summary      Create hierarchical tasks
// @Description  Creates a tree of tasks under a parent container
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Param        request  body      handlers.NewHierarchicalTasksRequest  true  "Hierarchical tasks creation request"
// @Success      200      {array}   models.TaskFull
// @Failure      400      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Security     AccessTokenCookie
// @Router       /new-hierarchical-tasks [post]
func (h *Handler) NewHierarchicalTasks(c *gin.Context) {
	var req NewHierarchicalTasksRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if len(req.Nodes) == 0 {
		c.JSON(400, gin.H{"error": "at least one task node is required"})
		return
	}

	ctx := c.Request.Context()
	createdTasks, err := h.App.TaskService.AddHierarchicalTasks(ctx, req.Parent, req.Nodes)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(404, gin.H{"error": "Parent container not found"})
			return
		}
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, createdTasks)
}

// NewTask handles POST /change-tasks-order
// @Summary      Change tasks order
// @Description  Changes the order of tasks in a container
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Param        request  body      ChangeTasksOrderRequest  true  "Change tasks order request"
// @Success      200      {object}  map[string]interface{}
// @Failure      400      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /change-tasks-order [post]
func (h *Handler) ChangeTasksOrder(c *gin.Context) {
	var req ChangeTasksOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()

	_, err := h.App.ContainerService.ChangeTasksOrder(ctx, req.ContainerType, req.ContainerID, req.TasksInNewOrder)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	openTasksIDs, err := h.App.ContainerService.GetOpenSubtasksIDs(ctx, req.ContainerType, req.ContainerID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"subtasksIDs": openTasksIDs})
}

// PatchTaskByID handles PATCH /task/:id
// @Summary      Patch task by ID
// @Description  Partially updates a task's description and/or notes
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Param        id       path      int                   true  "Task ID"
// @Param        request  body      PatchTaskByIDRequest  true  "Fields to update"
// @Success      200      {object}  models.TaskFull
// @Failure      400      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /task/{id} [patch]
func (h *Handler) PatchTaskByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid task ID"})
		return
	}

	var req PatchTaskByIDRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if req.Description == nil && req.Notes == nil {
		c.JSON(400, gin.H{"error": "At least one of description or notes must be provided"})
		return
	}

	ctx := c.Request.Context()
	taskFull, err := h.App.TaskService.UpdateTask(ctx, models.TaskPartial{
		ID:          id,
		Description: req.Description,
		Notes:       req.Notes,
	})
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(404, gin.H{"error": "Task not found"})
			return
		}
		log.Printf("Error patching task %d: %v", id, err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, taskFull)
}

// UpdateTask handles PUT /update-task
// @Summary      Update task
// @Description  Updates an existing task by ID
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Param        request  body      models.TaskPartial  true  "Task update request"
// @Success      200      {object}  models.TaskFull
// @Failure      400      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /update-task [put]
func (h *Handler) UpdateTask(c *gin.Context) {
	var req models.TaskPartial
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()

	taskFull, err := h.App.TaskService.UpdateTask(ctx, req)

	if err != nil {
		log.Printf("Error updating problem: %v", err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, taskFull)
}
