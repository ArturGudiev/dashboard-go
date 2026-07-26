package handlers

import (
	"arturgudiev/dashboard/ent"
	_ "arturgudiev/dashboard/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

// AddContainerCheck handles POST /container-checks
// @Summary      Add check to container
// @Description  Creates a check entry for the given container
// @Tags         container checks
// @Accept       json
// @Produce      json
// @Param        request  body      AddContainerCheckRequest  true  "Check creation request"
// @Success      200      {object}  models.ContainerCheck
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /container-checks [post]
func (h *Handler) AddContainerCheck(c *gin.Context) {
	var req AddContainerCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if !validateContainerType(req.ContainerType) {
		c.JSON(400, gin.H{"error": "Invalid container type"})
		return
	}

	ctx := c.Request.Context()
	check, err := h.App.ContainerChecksService.AddCheck(
		ctx,
		req.Description,
		req.ContainerType,
		req.ContainerID,
	)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, check)
}

// PatchContainerCheck handles PATCH /container-checks/:id
// @Summary      Update container check
// @Description  Updates a container check's description
// @Tags         container checks
// @Accept       json
// @Produce      json
// @Param        id       path      int                         true  "Container check ID"
// @Param        request  body      PatchContainerCheckRequest  true  "Fields to update"
// @Success      200      {object}  models.ContainerCheck
// @Failure      400      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /container-checks/{id} [patch]
func (h *Handler) PatchContainerCheck(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid check ID"})
		return
	}

	var req PatchContainerCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	check, err := h.App.ContainerChecksService.UpdateCheck(ctx, id, req.Description)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(404, gin.H{"error": "Check not found"})
			return
		}
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, check)
}

// RemoveContainerCheck handles DELETE /container-checks/:id
// @Summary      Remove container check
// @Description  Deletes a container check by ID
// @Tags         container checks
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Container check ID"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /container-checks/{id} [delete]
func (h *Handler) RemoveContainerCheck(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid check ID"})
		return
	}

	ctx := c.Request.Context()
	err = h.App.ContainerChecksService.RemoveCheck(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(404, gin.H{"error": "Check not found"})
			return
		}
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"status": "ok"})
}
