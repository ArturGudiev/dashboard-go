package handlers

import (
	"arturgudiev/dashboard/ent"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetLogMessageByID handles GET /log-messages/:id
// @Summary      Get log message by ID
// @Description  Returns a log message by its ID
// @Tags         log messages
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Log Message ID"
// @Success      200  {object}  ent.LogMessage
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /log-messages/{id} [get]
func (h *Handler) GetLogMessageByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid log message ID"})
		return
	}

	logMessage, err := h.App.LogMessagesRepository.GetLogMessage(c.Request.Context(), id)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, logMessage)
}

// GetTasksByIDs handles GET /log-messages/:containerType/:containerID
// @Summary      Get log messages by container type and ID
// @Description  Returns multiple log messages by their container type and ID
// @Tags         log messages
// @Accept       json
// @Produce      json
// @Param        containerType   path      string  true  "Container type"
// @Param        containerID   path      int  true  "Container ID"
// @Success      200      {array}   ent.LogMessage
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /container-log-messages/{containerType}/{containerID} [get]
// func (h *Handler) GetLogMessagesByContainer(c *gin.Context) {
// 	containerType := schema.ContainerType(c.Param("containerType"))
// 	switch containerType {
// 	case schema.ContainerTypeEpic,
// 		schema.ContainerTypeStory,
// 		schema.ContainerTypeTask,
// 		schema.ContainerTypeQuestion,
// 		schema.ContainerTypeProblem,
// 		schema.ContainerTypeKnowledgeNode,
// 		schema.ContainerTypeKnowledgeBit,
// 		schema.ContainerTypeDefinition,
// 		schema.ContainerTypeAction,
// 		schema.ContainerTypeScheduledTask,
// 		schema.ContainerTypeState:
// 		// valid
// 	default:
// 		c.JSON(400, gin.H{"error": "Invalid container type"})
// 		return
// 	}

// 	containerID, err := strconv.Atoi(c.Param("containerID"))
// 	if err != nil {
// 		c.JSON(400, gin.H{"error": "Invalid container ID"})
// 		return
// 	}
// 	logMessages, total, err := h.App.LogMessagesRepository.GetContainerLogMessages(c.Request.Context(), containerType, containerID)
// 	if err != nil {
// 		c.JSON(500, gin.H{"error": err.Error()})
// 		return
// 	}
// 	c.JSON(200, logMessages)
// }

// PaginatedResponse — универсальная структура для любого типа данных T
type PaginatedResponse[T any] struct {
	Items   []T `json:"items" binding:"required"`
	Total   int `json:"total" binding:"required"`
	Page    int `json:"page" binding:"required"`
	PerPage int `json:"perPage" binding:"required"`
}

// GetTasksByIDs handles GET /log-messages
// @Summary      Get log messages
// @Description  Returns multiple log messages
// @Param containerType query string false "Container type filter"
// @Param containerID query integer false "Container id filter"
// @Param perPage query integer false "Items per page"
// @Param page query integer false "Page number"
// @Param global query boolean false "Global log messages"
// @Tags         log messages
// @Accept       json
// @Produce      json
// @Success      200      {object}   PaginatedResponse[ent.LogMessage]
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /log-messages [get]
func (h *Handler) GetLogMessages(c *gin.Context) {
	var query logMessagesQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if query.PerPage == nil || *query.PerPage == 0 {
		v := 50
		query.PerPage = &v
	}
	if query.Page == nil {
		v := 0
		query.Page = &v
	}

	global := false
	if query.Global != nil {
		global = *query.Global
	}

	if query.ContainerType == nil && query.ContainerID == nil {
		logMessages, total, err := h.App.LogMessagesRepository.GetLogMessages(c.Request.Context(), *query.PerPage, *query.Page, global)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, PaginatedResponse[*ent.LogMessage]{
			Items:   logMessages,
			Total:   *total,
			Page:    *query.Page,
			PerPage: *query.PerPage,
		})
		return
	}
	if query.ContainerType != nil && query.ContainerID != nil {
		logMessages, total, err := h.App.LogMessagesRepository.GetContainerLogMessages(
			c.Request.Context(),
			*query.ContainerType,
			*query.ContainerID,
			*query.PerPage,
			*query.Page,
		)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, PaginatedResponse[*ent.LogMessage]{
			Items:   logMessages,
			Total:   *total,
			Page:    *query.Page,
			PerPage: *query.PerPage,
		})
		return
	}
	c.JSON(400, gin.H{"error": "Invalid container type or ID"})
}

// NewLogMessage handles POST /log-messages
// @Summary      Create new log message
// @Description  Creates a new log message with optional container relationship
// @Tags         log messages
// @Accept       json
// @Produce      json
// @Param        request  body      NewLogMessageRequest  true  "Log message creation request"
// @Success      200      {object}  ent.LogMessage
// @Failure      400      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /log-messages [post]
func (h *Handler) NewLogMessage(c *gin.Context) {
	var req NewLogMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	newLogMessage, err := h.App.LogMessagesRepository.AddLogMessage(ctx, req.Description, req.ContainerType, req.ContainerID)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, newLogMessage)
}

// // AddAnonymousTask handles PUT /add-anonymous-task
// // @Summary      Create anonymous task
// // @Description  Creates a simple anonymous task with default values
// // @Tags         tasks
// // @Accept       json
// // @Produce      json
// // @Success      200  {object}  ent.Task
// // @Failure      500  {object}  map[string]string
// // @Router       /add-anonymous-task [put]
// func (h *Handler) AddAnonymousTask(c *gin.Context) {
// 	newTask, err := h.App.TaskService.AddAnonymousTask(c.Request.Context())

// 	if err != nil {
// 		log.Printf("Error creating task: %v", err)
// 		c.JSON(500, gin.H{"error": err.Error()})
// 		return
// 	}

// 	c.JSON(200, newTask)
// }

// // GetDoneTasks handles GET /done-tasks
// // @Summary      Get done tasks from today
// // @Description  Returns all done tasks where doneDateTime is today
// // @Tags         tasks
// // @Accept       json
// // @Produce      json
// // @Success      200  {array}   ent.Task
// // @Failure      500  {object}  map[string]string
// // @Router       /done-tasks [get]
// func (h *Handler) GetDoneTasks(c *gin.Context) {
// 	tasksCount, err := h.App.TaskService.GetDoneTasksCount(c.Request.Context())

// 	if err != nil {
// 		log.Printf("Error querying done tasksCount: %v", err)
// 		c.JSON(500, gin.H{"error": err.Error()})
// 		return
// 	}

// 	response := DoneTasksResponse{
// 		DoneTasks: tasksCount,
// 	}
// 	c.JSON(200, response)
// }

// // FinishTask handles PUT /finish-task/:id
// // @Summary      Finish task recursively
// // @Description  Marks a task and all its descendants as done
// // @Tags         tasks
// // @Accept       json
// // @Produce      json
// // @Param        id   path      int  true  "Task ID"
// // @Success      200  {object}  TaskResponse
// // @Failure      400  {object}  map[string]string
// // @Failure      404  {object}  map[string]string
// // @Failure      500  {object}  map[string]string
// // @Router       /finish-task/{id} [put]
// func (h *Handler) FinishTask(c *gin.Context) {
// 	idParam := c.Param("id")
// 	id, err := strconv.Atoi(idParam)
// 	if err != nil {
// 		c.JSON(400, gin.H{"error": "Invalid task ID"})
// 		return
// 	}

// 	ctx := c.Request.Context()

// 	updatedTask, err := h.App.TaskService.FinishTaskById(ctx, id)
// 	if err != nil {
// 		if ent.IsNotFound(err) {
// 			c.JSON(404, gin.H{"error": "Task not found"})
// 			return
// 		}
// 		log.Printf("Error finishing task %d: %v", id, err)
// 		c.JSON(500, gin.H{"error": err.Error()})
// 		return
// 	}

// 	// Convert to custom response type to ensure all fields are included
// 	response := TaskResponse{
// 		ID:           updatedTask.ID,
// 		Description:  updatedTask.Description,
// 		Tags:         updatedTask.Tags,
// 		Done:         updatedTask.Done,
// 		Notes:        updatedTask.Notes,
// 		DoneDateTime: updatedTask.DoneDateTime,
// 	}

// 	c.JSON(200, response)
// }

// // FinishTasks handles PUT /finish-tasks/
// // @Summary      Finish multiple tasks recursively
// // @Description  Marks multiple tasks and all their descendants as done
// // @Tags         tasks
// // @Accept       json
// // @Produce      json
// // @Param        request  body      []TaskIDRequest  true  "List of task IDs"
// // @Success      200      {object}  map[string]interface{}
// // @Failure      400      {object}  map[string]string
// // @Failure      500      {object}  map[string]string
// // @Router       /finish-tasks/ [put]
// func (h *Handler) FinishTasks(c *gin.Context) { // TODO after Go: modify
// 	var tasks []TaskIDRequest
// 	if err := c.ShouldBindJSON(&tasks); err != nil {
// 		c.JSON(400, gin.H{"error": err.Error()})
// 		return
// 	}

// 	ctx := c.Request.Context()

// 	for _, t := range tasks {
// 		task, err := h.App.Client.Task.Get(ctx, t.ID)
// 		if err != nil {
// 			if ent.IsNotFound(err) {
// 				log.Printf("Task %d not found, skipping", t.ID)
// 				continue
// 			}
// 			log.Printf("Error getting task %d: %v", t.ID, err)
// 			continue
// 		}

// 		if err := h.App.TaskService.FinishTaskRecursively(ctx, task); err != nil {
// 			log.Printf("Error finishing task %d recursively: %v", t.ID, err)
// 			c.JSON(500, gin.H{"error": err.Error()})
// 			return
// 		}
// 	}

// 	c.JSON(200, gin.H{})
// }

// // FinishTasksByIDs handles PUT /finish-tasks-by-ids/
// // @Summary      Finish tasks by IDs recursively
// // @Description  Marks tasks by their IDs and all their descendants as done
// // @Tags         tasks
// // @Accept       json
// // @Produce      json
// // @Param        request  body      []int  true  "List of task IDs"
// // @Success      200      {object}  map[string]interface{}
// // @Failure      400      {object}  map[string]string
// // @Failure      500      {object}  map[string]string
// // @Router       /finish-tasks-by-ids/ [put]
// func (h *Handler) FinishTasksByIDs(c *gin.Context) {
// 	var ids []int
// 	if err := c.ShouldBindJSON(&ids); err != nil {
// 		c.JSON(400, gin.H{"error": err.Error()})
// 		return
// 	}

// 	ctx := c.Request.Context()

// 	for _, id := range ids {
// 		task, err := h.App.TasksRepository.GetTask(ctx, id)
// 		if err != nil {
// 			log.Printf("Error getting task %d: %v", id, err)
// 			continue
// 		}

// 		if err := h.App.TaskService.FinishTaskRecursively(ctx, task); err != nil {
// 			log.Printf("Error finishing task %d recursively: %v", id, err)
// 			c.JSON(500, gin.H{"error": err.Error()})
// 			return
// 		}
// 	}

// 	c.JSON(200, gin.H{})
// }

// // UpdateTask handles PUT /update-task
// // @Summary      Update task
// // @Description  Updates an existing task by ID
// // @Tags         tasks
// // @Accept       json
// // @Produce      json
// // @Param        request  body      models.TaskPartial  true  "Task update request"
// // @Success      200      {object}  models.TaskFull
// // @Failure      400      {object}  map[string]string
// // @Failure      404      {object}  map[string]string
// // @Failure      500      {object}  map[string]string
// // @Router       /update-task [put]
// func (h *Handler) UpdateTask(c *gin.Context) {
// 	var req models.TaskPartial
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(400, gin.H{"error": err.Error()})
// 		return
// 	}

// 	ctx := c.Request.Context()

// 	taskFull, err := h.App.TaskService.UpdateTask(ctx, req)

// 	if err != nil {
// 		log.Printf("Error updating problem: %v", err)
// 		c.JSON(500, gin.H{"error": err.Error()})
// 		return
// 	}

// 	c.JSON(200, taskFull)
// }
