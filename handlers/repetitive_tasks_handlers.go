package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetRepetitiveTasks handles GET /repetitive-tasks/
// @Summary      Get repetitive tasks
// @Description  Returns all repetitive tasks
// @Tags         repetitive-tasks
// @Accept       json
// @Produce      json
// @Success      200  {array}  ent.RepetitiveTask
// @Failure      500  {object}  map[string]string
// @Router       /repetitive-tasks/ [get]
func (h *Handler) GetRepetitiveTasks(c *gin.Context) {
	repetitiveTasks, err := h.App.RepetitiveTaskService.GetRepetitiveTasks(c.Request.Context())
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
