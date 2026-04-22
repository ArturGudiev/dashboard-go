package handlers

import (
	"arturgudiev/dashboard/ent"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetRepetitiveTasks handles GET /repetitive-tasks/
// @Summary      Get repetitive tasks
// @Description  Returns all repetitive tasks
// @Param actual query boolean false "filter repetitive tasks by actual property"
// @Tags         repetitive-tasks
// @Accept       json
// @Produce      json
// @Success      200  {array}  ent.RepetitiveTask
// @Failure      500  {object}  map[string]string
// @Router       /repetitive-tasks/ [get]
func (h *Handler) GetRepetitiveTasks(c *gin.Context) {
	var query RepetitiveTasksQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	repetitiveTasks, err := h.App.RepetitiveTaskService.GetRepetitiveTasks(c.Request.Context(), query.Actual)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, repetitiveTasks)
}

// GetRepetitiveTaskById handles GET /repetitive-tasks/:id
// @Summary      Get repetitive task by ID
// @Description  Returns repetitive task by ID
// @Tags         repetitive-tasks
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Repetitive task ID"
// @Success      200  {object}  ent.RepetitiveTask
// @Failure      500  {object}  map[string]string
// @Router       /repetitive-tasks/{id} [get]
func (h *Handler) GetRepetitiveTaskById(c *gin.Context) {
	idParam := c.Param("id")
	repetitiveTaskID, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid repetitive task ID"})
		return
	}

	repetitiveTask, err := h.App.RepetitiveTaskService.GetRepetitiveTaskById(c.Request.Context(), repetitiveTaskID)
	if err != nil {
		c.JSON(404, gin.H{"error": "Repetitive task not found"})
		return
	}

	c.JSON(200, repetitiveTask)
}

// AddRepetitiveTaskExecution handles POST /repetitive-tasks/:id/executions
// @Summary      Record repetitive task execution
// @Description  Creates an execution record for the given repetitive task (timestamp is server time)
// @Tags         repetitive-tasks
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Repetitive task ID"
// @Success      200  {object}  ent.RepetitiveTaskExecution
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /repetitive-tasks/{id}/executions [post]
func (h *Handler) AddRepetitiveTaskExecution(c *gin.Context) {
	idParam := c.Param("id")
	repetitiveTaskID, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid repetitive task ID"})
		return
	}

	exec, err := h.App.RepetitiveTaskExecutionService.AddRepetitiveTaskExecution(c.Request.Context(), repetitiveTaskID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, exec)
}

// GetRepetitiveTaskExecutions handles GET /repetitive-tasks/:id/executions
// @Summary      List repetitive task executions
// @Description  Returns execution records for the given repetitive task (newest first)
// @Tags         repetitive-tasks
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Repetitive task ID"
// @Success      200  {array}   ent.RepetitiveTaskExecution
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /repetitive-tasks/{id}/executions [get]
func (h *Handler) GetRepetitiveTaskExecutions(c *gin.Context) {
	idParam := c.Param("id")
	repetitiveTaskID, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid repetitive task ID"})
		return
	}

	execs, err := h.App.RepetitiveTaskExecutionService.GetRepetitiveTaskExecutions(c.Request.Context(), repetitiveTaskID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, execs)
}

// NewRepetitiveTask handles POST /new-repetitive-task
// @Summary      Create new repetitive task
// @Description  Creates a new repetitive task with optional parent relationship
// @Tags         repetitive-tasks
// @Accept       json
// @Produce      json
// @Param        request  body      NewRepetitiveTaskRequest  true  "Repetitive task creation request"
// @Success      200  {object}  ent.RepetitiveTask
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /new-repetitive-task [post]
func (h *Handler) NewRepetitiveTask(c *gin.Context) {
	var req NewRepetitiveTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	newRepetitiveTask, err := h.App.RepetitiveTaskService.AddRepetitiveTask(ctx, req.RepetitiveTask, req.Parent)

	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(404, gin.H{"error": "Parent container not found"})
			return
		}
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, newRepetitiveTask)
}
