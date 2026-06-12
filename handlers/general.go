package handlers

import (
	"arturgudiev/dashboard/ent/schema"

	"github.com/gin-gonic/gin"
)

// Root handles GET /
// @Summary      Root endpoint
// @Description  Returns a welcome message
// @Tags         general
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]string
// @Router       / [get]
func (h *Handler) Root(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Dashboard server"})
}

// GetParentsPath handles POST /parents-path
// @Summary      Get parents path
// @Description  Returns the path of parent containers for a given container
// @Tags         general
// @Accept       json
// @Produce      json
// @Param        request  body      ParentsPathRequest  true  "Parents path request"
// @Success      200      {array}   string
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /parents-path [post]
func (h *Handler) GetParentsPath(c *gin.Context) {
	var req ParentsPathRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Convert string type to ContainerType
	containerType := schema.ContainerType(req.Type)

	// Validate the container type
	switch containerType {
	case schema.ContainerTypeEpic, schema.ContainerTypeStory, schema.ContainerTypeTask,
		schema.ContainerTypeQuestion, schema.ContainerTypeProblem, schema.ContainerTypeKnowledgeNode,
		schema.ContainerTypeKnowledgeBit, schema.ContainerTypeDefinition, schema.ContainerTypeAction,
		schema.ContainerTypeRepetitiveTask, schema.ContainerTypeLongTask, schema.ContainerTypeDirection, schema.ContainerTypeState:
		// Valid type
	default:
		c.JSON(400, gin.H{"error": "Invalid container type"})
		return
	}

	ctx := c.Request.Context()
	parentsPath := h.App.ContainerService.GetParentsPathDescriptions(ctx, containerType, req.ID)

	if parentsPath == nil {
		parentsPath = []string{}
	}

	c.JSON(200, parentsPath)
}
