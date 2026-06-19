package handlers

import (
	"arturgudiev/dashboard/ent"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetLongTasks handles GET /long-tasks
// @Summary      Get long tasks
// @Description  Returns all long tasks; use open=true to return only open (not done) tasks
// @Tags         long-tasks
// @Accept       json
// @Produce      json
// @Param        open  query  boolean  false  "Return only open long tasks"
// @Success      200   {array}   []models.LongTaskFull
// @Failure      400   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /long-tasks [get]
func (h *Handler) GetLongTasks(c *gin.Context) {
	var query LongTasksQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	longTasks, err := h.App.LongTasksService.GetLongTasksFull(c.Request.Context(), query.Open)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, longTasks)
}

// GetLongTaskById handles GET /long-tasks/:id
// @Summary      Get long task by ID
// @Description  Returns a long task by ID
// @Tags         long-tasks
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Long task ID"
// @Success      200  {object}  models.LongTaskFull
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /long-tasks/{id} [get]
func (h *Handler) GetLongTaskById(c *gin.Context) {
	idParam := c.Param("id")
	longTaskID, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid long task ID"})
		return
	}

	longTask, err := h.App.LongTasksRepository.GetLongTask(c.Request.Context(), longTaskID)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(404, gin.H{"error": "Long task not found"})
			return
		}
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, longTask)
}

// AddLongTaskSubmission handles POST /long-tasks/:id/submissions
// @Summary      Add long task submission
// @Description  Records progress submission; increments progress_done when progressToAdd is set, or sets progress_done when progressToSet is set
// @Tags         long-tasks
// @Accept       json
// @Produce      json
// @Param        id       path      int                           true  "Long task ID"
// @Param        request  body      AddLongTaskSubmissionRequest  true  "Submission request"
// @Success      200      {object}  ent.LongTaskSubmission
// @Failure      400      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /long-tasks/{id}/submissions [post]
func (h *Handler) AddLongTaskSubmission(c *gin.Context) {
	idParam := c.Param("id")
	longTaskID, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid long task ID"})
		return
	}

	var req AddLongTaskSubmissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if req.ProgressToAdd == nil && req.ProgressToSet == nil && req.ProgressRaw == nil {
		c.JSON(400, gin.H{"error": "At least one of progressToAdd, progressToSet, or progressRaw must be provided"})
		return
	}

	submission, err := h.App.LongTaskSubmissionsService.AddLongTaskSubmission(
		c.Request.Context(),
		longTaskID,
		req.Comments,
		req.ProgressToAdd,
		req.ProgressToSet,
		req.ProgressRaw,
	)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(404, gin.H{"error": "Long task not found"})
			return
		}
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, submission)
}

// GetLongTaskSubmissions handles GET /long-tasks/:id/submissions
// @Summary      List long task submissions
// @Description  Returns submission records for the given long task (newest first)
// @Tags         long-tasks
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Long task ID"
// @Success      200  {array}   ent.LongTaskSubmission
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /long-tasks/{id}/submissions [get]
func (h *Handler) GetLongTaskSubmissions(c *gin.Context) {
	idParam := c.Param("id")
	longTaskID, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid long task ID"})
		return
	}

	submissions, err := h.App.LongTaskSubmissionsService.GetLongTaskSubmissions(c.Request.Context(), longTaskID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, submissions)
}

// NewLongTask handles POST /long-tasks
// @Summary      Create new long task
// @Description  Creates a new long task with optional parent relationship
// @Tags         long-tasks
// @Accept       json
// @Produce      json
// @Param        request  body      NewLongTaskRequest  true  "Long task creation request"
// @Success      200      {object}  ent.LongTask
// @Failure      400      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /long-tasks [post]
func (h *Handler) NewLongTask(c *gin.Context) {
	var req NewLongTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	newLongTask, err := h.App.LongTasksService.AddLongTask(ctx, req.LongTask, req.Parent)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(404, gin.H{"error": "Parent container not found"})
			return
		}
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, newLongTask)
}
