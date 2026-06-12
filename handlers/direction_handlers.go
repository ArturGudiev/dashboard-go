package handlers

import (
	"arturgudiev/dashboard/ent"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetDirections handles GET /directions
// @Summary      Get directions
// @Description  Returns all directions; use open=true to return only open directions
// @Tags         directions
// @Accept       json
// @Produce      json
// @Param        open  query  boolean  false  "Return only open directions"
// @Success      200   {array}   ent.Direction
// @Failure      400   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /directions [get]
func (h *Handler) GetDirections(c *gin.Context) {
	var query DirectionsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	directions, err := h.App.DirectionsService.GetDirections(c.Request.Context(), query.Open)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, directions)
}

// GetDirectionById handles GET /directions/:id
// @Summary      Get direction by ID
// @Description  Returns a direction by ID with child container IDs
// @Tags         directions
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Direction ID"
// @Success      200  {object}  models.DirectionFull
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /directions/{id} [get]
func (h *Handler) GetDirectionById(c *gin.Context) {
	id, ok := parsePatchID(c, "direction")
	if !ok {
		return
	}

	direction, err := h.App.DirectionsService.GetDirectionById(c.Request.Context(), id)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(404, gin.H{"error": "Direction not found"})
			return
		}
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, direction)
}

// NewDirection handles POST /directions
// @Summary      Create new direction
// @Description  Creates a new direction with optional parent relationship
// @Tags         directions
// @Accept       json
// @Produce      json
// @Param        request  body      NewDirectionRequest  true  "Direction creation request"
// @Success      200      {object}  models.DirectionFull
// @Failure      400      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /directions [post]
func (h *Handler) NewDirection(c *gin.Context) {
	var req NewDirectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	newDirection, err := h.App.DirectionsService.AddDirection(ctx, req.Direction, req.Parent)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(404, gin.H{"error": "Parent container not found"})
			return
		}
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, newDirection)
}

// AddDirectionSubmission handles POST /directions/:id/submissions
// @Summary      Add direction submission
// @Description  Records a text submission for a direction
// @Tags         directions
// @Accept       json
// @Produce      json
// @Param        id       path      int                           true  "Direction ID"
// @Param        request  body      AddDirectionSubmissionRequest  true  "Submission request"
// @Success      200      {object}  ent.DirectionSubmission
// @Failure      400      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /directions/{id}/submissions [post]
func (h *Handler) AddDirectionSubmission(c *gin.Context) {
	id, ok := parsePatchID(c, "direction")
	if !ok {
		return
	}

	var req AddDirectionSubmissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if req.Text == nil || *req.Text == "" {
		c.JSON(400, gin.H{"error": "text must be provided"})
		return
	}

	submission, err := h.App.DirectionsService.AddDirectionSubmission(c.Request.Context(), id, req.Text)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(404, gin.H{"error": "Direction not found"})
			return
		}
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, submission)
}

// GetDirectionSubmissions handles GET /directions/:id/submissions
// @Summary      List direction submissions
// @Description  Returns submission records for the given direction (newest first)
// @Tags         directions
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Direction ID"
// @Success      200  {array}   ent.DirectionSubmission
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /directions/{id}/submissions [get]
func (h *Handler) GetDirectionSubmissions(c *gin.Context) {
	idParam := c.Param("id")
	directionID, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid direction ID"})
		return
	}

	submissions, err := h.App.DirectionsService.GetDirectionSubmissions(c.Request.Context(), directionID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, submissions)
}

// GetDirectionStats handles GET /directions/:id/stats
// @Summary      Get direction stats
// @Description  Returns aggregated stats entries for a direction including submissions and descendant activity
// @Tags         directions
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Direction ID"
// @Success      200  {array}   models.DirectionStatsEntry
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /directions/{id}/stats [get]
func (h *Handler) GetDirectionStats(c *gin.Context) {
	id, ok := parsePatchID(c, "direction")
	if !ok {
		return
	}

	stats, err := h.App.DirectionsService.GetDirectionStats(c.Request.Context(), id)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(404, gin.H{"error": "Direction not found"})
			return
		}
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, stats)
}
