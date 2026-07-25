package handlers

import (
	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/ent/schema"
	"strconv"

	"github.com/gin-gonic/gin"
)

func validateAliasContainerType(containerType schema.ContainerType) bool {
	switch containerType {
	case schema.ContainerTypeEpic, schema.ContainerTypeStory, schema.ContainerTypeTask,
		schema.ContainerTypeQuestion, schema.ContainerTypeProblem, schema.ContainerTypeKnowledgeNode,
		schema.ContainerTypeKnowledgeBit, schema.ContainerTypeDefinition, schema.ContainerTypeAction,
		schema.ContainerTypeRepetitiveTask, schema.ContainerTypeState:
		return true
	default:
		return false
	}
}

// GetAliasByString handles GET /aliases/:alias
// @Summary      Get alias by string
// @Description  Returns an alias by its string
// @Tags         aliases
// @Accept       json
// @Produce      json
// @Param        alias   path      string  true  "Alias"
// @Success      200      {object}  models.AliasModel
// @Failure      400      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /aliases/{alias} [get]
func (h *Handler) GetAliasByString(c *gin.Context) {
	aliasString := c.Param("alias")
	aliasModel, err := h.App.AliasesService.GetAlias(c.Request.Context(), aliasString)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(404, gin.H{"error": "Alias not found"})
			return
		}
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, aliasModel)
}

// GetAliasesByContainer handles GET /aliases/container/:type/:id
// @Summary      List aliases for container
// @Description  Returns all aliases linked to the given container type and id
// @Tags         aliases
// @Accept       json
// @Produce      json
// @Param        type  path      string  true  "Container type"
// @Param        id    path      int     true  "Container ID"
// @Success      200   {array}   models.AliasModel
// @Failure      400   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /aliases/container/{type}/{id} [get]
func (h *Handler) GetAliasesByContainer(c *gin.Context) {
	containerType := schema.ContainerType(c.Param("type"))
	if !validateAliasContainerType(containerType) {
		c.JSON(400, gin.H{"error": "Invalid container type"})
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(400, gin.H{"error": "Invalid container ID"})
		return
	}

	aliases, err := h.App.AliasesService.GetAliasesByTaskContainer(c.Request.Context(), containerType, id)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, aliases)
}

// UpdateContainerAliases handles PUT /aliases/container
// @Summary      Sync aliases for container
// @Description  Replaces container aliases: removes absent ones and adds new ones
// @Tags         aliases
// @Accept       json
// @Produce      json
// @Param        request  body      UpdateContainerAliasesRequest  true  "Aliases sync request"
// @Success      200      {array}   models.AliasModel
// @Failure      400      {object}  map[string]string
// @Failure      409      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /aliases/container [put]
func (h *Handler) UpdateContainerAliases(c *gin.Context) {
	var req UpdateContainerAliasesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if !validateAliasContainerType(req.ContainerType) {
		c.JSON(400, gin.H{"error": "Invalid container type"})
		return
	}

	if req.Aliases == nil {
		req.Aliases = []string{}
	}

	aliases, err := h.App.AliasesService.UpdateContainerAliases(
		c.Request.Context(),
		req.ContainerType,
		req.ContainerID,
		req.Aliases,
	)
	if err != nil {
		if ent.IsConstraintError(err) {
			c.JSON(409, gin.H{"error": "Alias already exists"})
			return
		}
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, aliases)
}
