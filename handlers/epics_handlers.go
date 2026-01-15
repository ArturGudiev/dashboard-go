package handlers

import (
	"strconv"

	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/models"

	"github.com/gin-gonic/gin"
)

// GetEpicByID handles GET /epic/:id
// @Summary      Get epic by ID
// @Description  Returns a epic by its ID
// @Tags         epics
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Epic ID"
// @Success      200  {object}  models.EpicFull
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /epic/{id} [get]
func (h *Handler) GetEpicByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid epic ID"})
		return
	}

	epicFull, err := h.App.EpicsService.GetEpicFull(c.Request.Context(), id)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(404, gin.H{"error": "Epic not found"})
			return
		}
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, epicFull)
}

// GetEpicsByIDs handles POST /get-epics
// @Summary      Get epics by IDs
// @Description  Returns multiple epics by their IDs
// @Tags         epics
// @Accept       json
// @Produce      json
// @Param        request  body      IDsRequest  true  "List of epic IDs"
// @Success      200      {array}   models.EpicFull
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /get-epics [post]
func (h *Handler) GetEpicsByIDs(c *gin.Context) {
	var req IDsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	epicsFull, err := h.App.EpicsService.GetEpicsFull(c.Request.Context(), req.IDs)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(404, gin.H{"error": "Epics not found"})
			return
		}
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, epicsFull)
}

// NewEpic handles POST /new-epic
// @Summary      Create new epic
// @Description  Creates a new epic with optional parent relationship
// @Tags         epics
// @Accept       json
// @Produce      json
// @Param        request  body      models.NewEpicRequest  true  "Epic creation request"
// @Success      200      {object}  models.EpicFull
// @Failure      400      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /new-epic [post]
func (h *Handler) NewEpic(c *gin.Context) {
	var req models.NewEpicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	newEpic, err := h.App.EpicsService.AddEpic(ctx, req.Epic, req.Parent)

	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(404, gin.H{"error": "Parent container not found"})
			return
		}
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, newEpic)
}

// UpdateEpic handles PUT /update-epic
// @Summary      Update epic
// @Description  Updates an existing epic by ID
// @Tags         epics
// @Accept       json
// @Produce      json
// @Param        request  body      models.EpicPartial  true  "Epic update request"
// @Success      200      {object}  models.EpicFull
// @Failure      400      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /update-epic [put]
func (h *Handler) UpdateEpic(c *gin.Context) {
	var req models.EpicPartial
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()

	epicFull, err := h.App.EpicsService.UpdateEpic(ctx, req)

	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(404, gin.H{"error": "Epic not found"})
			return
		}
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, epicFull)
}
