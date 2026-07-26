package handlers

import (
	"errors"
	"strconv"
	"strings"

	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/ent/schema"
	"arturgudiev/dashboard/services"

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

func (h *Handler) absoluteAliasFilePath(c *gin.Context, relPath string) (string, bool) {
	abs, err := h.App.FilesService.AbsolutePath(relPath)
	if err != nil {
		if errors.Is(err, services.ErrInvalidFilePath) {
			c.JSON(400, gin.H{"error": "Invalid file path"})
			return "", false
		}
		c.JSON(500, gin.H{"error": err.Error()})
		return "", false
	}
	return abs, true
}

// GetAliasesByFile handles GET /aliases/file/*filepath
// @Summary      List aliases for file
// @Description  Returns all aliases linked to the given logical relative file path under FILES_DIR
// @Tags         aliases
// @Accept       json
// @Produce      json
// @Param        filepath  path      string  true  "Relative file path"
// @Success      200       {array}   models.AliasModel
// @Failure      400       {object}  map[string]string
// @Failure      500       {object}  map[string]string
// @Router       /aliases/file/{filepath} [get]
func (h *Handler) GetAliasesByFile(c *gin.Context) {
	relPath := strings.TrimPrefix(c.Param("filepath"), "/")
	if _, ok := h.absoluteAliasFilePath(c, relPath); !ok {
		return
	}

	// Lookup by logical relative path so aliases stored with another machine's absolute
	// FILES_DIR (e.g. Windows) still match the current file page.
	aliases, err := h.App.AliasesService.GetAliasesByFilePath(c.Request.Context(), relPath)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, aliases)
}

// UpdateFileAliases handles PUT /aliases/file
// @Summary      Sync aliases for file
// @Description  Replaces file aliases: removes absent ones and adds new ones. filePath is the logical relative path under FILES_DIR.
// @Tags         aliases
// @Accept       json
// @Produce      json
// @Param        request  body      UpdateFileAliasesRequest  true  "Aliases sync request"
// @Success      200      {array}   models.AliasModel
// @Failure      400      {object}  map[string]string
// @Failure      409      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /aliases/file [put]
func (h *Handler) UpdateFileAliases(c *gin.Context) {
	var req UpdateFileAliasesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	absPath, ok := h.absoluteAliasFilePath(c, req.FilePath)
	if !ok {
		return
	}

	if req.Aliases == nil {
		req.Aliases = []string{}
	}

	// Lookup by logical relative path (cross-root match); store new aliases as local absolute.
	aliases, err := h.App.AliasesService.UpdateFileAliases(
		c.Request.Context(),
		req.FilePath,
		absPath,
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
