package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetLongTaskProgressByID handles GET /long-task-progresses/:id
// @Summary      Get long task progress by ID
// @Description  Returns a long task progress by ID
// @Tags         long-task-progresses
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Long task progress ID"
// @Success      200  {object}  models.LongTaskProgressFull
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /long-task-progresses/{id} [get]
func (h *Handler) GetLongTaskProgressByID(c *gin.Context) {
	idParam := c.Param("id")
	longTaskProgressID, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid long task progress ID"})
		return
	}

	progress, err := h.App.LongTaskProgressesService.GetLongTaskProgressByID(c.Request.Context(), longTaskProgressID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, progress)
}

// AddLongTaskProgressSubmission handles POST /long-task-progresses/:id/submissions
// @Summary      Add long task progress submission
// @Description  Adds a new progress submission for the given long task progress
// @Tags         long-task-progresses
// @Accept       json
// @Produce      json
// @Param        id       path      int                         true  "Long task progress ID"
// @Param        request  body      AddLongTaskProgressSubmissionRequest  true  "Progress submission request"
// @Success      200      {object}  ent.LongTaskProgressSubmission
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /long-task-progresses/{id}/submissions [post]
func (h *Handler) AddLongTaskProgressSubmission(c *gin.Context) {
	idParam := c.Param("id")
	longTaskProgressID, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid long task progress ID"})
		return
	}
	var req AddLongTaskProgressSubmissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	submission, err := h.App.LongTaskProgressesService.AddLongTaskProgressSubmission(
		c.Request.Context(),
		longTaskProgressID,
		req.Comments,
		req.ProgressToAdd,
		req.ProgressToSet,
		req.ProgressRaw,
		req.ExecutionDate,
	)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, submission)
}

// GetLongTaskProgressSubmissions handles GET /long-task-progresses/:id/submissions
// @Summary      List long task progress submissions
// @Description  Returns submission records for the given long task progress
// @Tags         long-task-progresses
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Long task progress ID"
// @Success      200  {array}   models.LongTaskProgressSubmission
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /long-task-progresses/{id}/submissions [get]
func (h *Handler) GetLongTaskProgressSubmissions(c *gin.Context) {
	idParam := c.Param("id")
	longTaskProgressID, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid long task progress ID"})
		return
	}

	submissions, err := h.App.LongTaskProgressesService.GetLongTaskProgressSubmissions(
		c.Request.Context(),
		longTaskProgressID,
	)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, submissions)
}

// GetLongTaskProgresses handles GET /long-tasks/:id/progresses
// @Summary      List long task progresses
// @Description  Returns progress records for the given long task (newest first)
// @Tags         long-task-progresses
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Long task ID"
// @Success      200  {array}   ent.LongTaskProgress
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /long-tasks/{id}/progresses [get]
func (h *Handler) GetLongTaskProgresses(c *gin.Context) {
	idParam := c.Param("id")
	longTaskID, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid long task ID"})
		return
	}

	progresses, err := h.App.LongTaskProgressesService.GetLongTaskProgresses(c.Request.Context(), longTaskID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, progresses)
}

// AddLongTaskProgress handles POST /long-tasks/:id/progresses
// @Summary      Add long task progress
// @Description  Adds a new progress record for the given long task
// @Tags         long-task-progresses
// @Accept       json
// @Produce      json
// @Param        id       path      int                         true  "Long task ID"
// @Param        request  body      AddLongTaskProgressRequest  true  "Progress request"
// @Success      200      {object}  ent.LongTaskProgress
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /long-tasks/{id}/progresses [post]
func (h *Handler) AddLongTaskProgress(c *gin.Context) {
	idParam := c.Param("id")
	longTaskID, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid long task ID"})
		return
	}

	var req AddLongTaskProgressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	progress, err := h.App.LongTaskProgressesService.AddLongTaskProgress(
		c.Request.Context(),
		req.Name,
		longTaskID,
		req.Value,
		req.Total,
		req.Units,
	)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, progress)
}
