package handlers

import (
	"log"
	"net/http"
	"strconv"

	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/ent/schema"
	"arturgudiev/dashboard/models"

	"github.com/gin-gonic/gin"
)

// GetKnowledgeNodeByID handles GET /knowledge-node/:id
// @Summary      Get knowledge node by ID
// @Description  Returns a knowledge node by its ID
// @Tags         knowledge-nodes
// @Produce      json
// @Param        id   path      int  true  "Knowledge node ID"
// @Success      200  {object}  models.KnowledgeNodeFull
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /knowledge-node/{id} [get]
func (h *Handler) GetKnowledgeNodeByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid knowledge node ID"})
		return
	}

	full, err := h.App.KnowledgeNodesService.GetKnowledgeNodeFull(c.Request.Context(), id)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Knowledge node not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, full)
}

// GetKnowledgeNodesByIDs handles POST /get-knowledge-nodes
// @Summary      Get knowledge nodes by IDs
// @Description  Returns multiple knowledge nodes by their IDs
// @Tags         knowledge-nodes
// @Accept       json
// @Produce      json
// @Param        request  body      IDsRequest  true  "List of knowledge node IDs"
// @Success      200      {array}   models.KnowledgeNodeFull
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /get-knowledge-nodes [post]
func (h *Handler) GetKnowledgeNodesByIDs(c *gin.Context) {
	var req IDsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fulls, err := h.App.KnowledgeNodesService.GetKnowledgeNodesFull(c.Request.Context(), req.IDs)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Knowledge nodes not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, fulls)
}

// NewKnowledgeNode handles POST /new-knowledge-node
// @Summary      Create new knowledge node
// @Description  Creates a new knowledge node with optional parent relationship
// @Tags         knowledge-nodes
// @Accept       json
// @Produce      json
// @Param        request  body      models.NewKnowledgeNodeRequest  true  "Knowledge node creation request"
// @Success      200      {object}  models.KnowledgeNodeFull
// @Failure      400      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /new-knowledge-node [post]
func (h *Handler) NewKnowledgeNode(c *gin.Context) {
	var req models.NewKnowledgeNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	full, err := h.App.KnowledgeNodesService.AddKnowledgeNode(ctx, req.KnowledgeNode, req.Parent)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Parent container not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, full)
}

// UpdateKnowledgeNode handles PUT /update-knowledge-node
// @Summary      Update knowledge node
// @Description  Updates an existing knowledge node by ID
// @Tags         knowledge-nodes
// @Accept       json
// @Produce      json
// @Param        request  body      models.KnowledgeNodePartial  true  "Knowledge node update request"
// @Success      200      {object}  models.KnowledgeNodeFull
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /update-knowledge-node [put]
func (h *Handler) UpdateKnowledgeNode(c *gin.Context) {
	var req models.KnowledgeNodePartial
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	full, err := h.App.KnowledgeNodesService.UpdateKnowledgeNode(c.Request.Context(), req)
	if err != nil {
		log.Printf("Error updating knowledge node: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, full)
}

// PatchKnowledgeNodeByID handles PATCH /knowledge-node/:id
// @Summary      Patch knowledge node by ID
// @Description  Partially updates a knowledge node's name (via description) and/or notes
// @Tags         knowledge-nodes
// @Accept       json
// @Produce      json
// @Param        id       path      int                       true  "Knowledge node ID"
// @Param        request  body      PatchContainerByIDRequest  true  "Fields to update"
// @Success      200      {object}  models.KnowledgeNodeFull
// @Failure      400      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /knowledge-node/{id} [patch]
func (h *Handler) PatchKnowledgeNodeByID(c *gin.Context) {
	id, ok := parsePatchID(c, "knowledge node")
	if !ok {
		return
	}
	req, ok := bindPatchContainerRequest(c)
	if !ok {
		return
	}

	partial := models.KnowledgeNodePartial{ID: id, Notes: req.Notes}
	if req.Description != nil {
		partial.Name = req.Description
	}

	full, err := h.App.KnowledgeNodesService.UpdateKnowledgeNode(c.Request.Context(), partial)
	if err != nil {
		writePatchError(c, id, "knowledge node", err)
		return
	}
	c.JSON(http.StatusOK, full)
}

// ListContainerFiles handles GET /files/container/:type/:id
// @Summary      List files for a container
// @Description  Lists immediate files/dirs in the container's related files folder (if present)
// @Tags         files
// @Produce      json
// @Param        type  path  string  true  "Container type"  example(knowledge-node)
// @Param        id    path  int     true  "Container ID"
// @Success      200   {array}  services.FileInfo
// @Failure      400   {object} map[string]string
// @Failure      500   {object} map[string]string
// @Router       /files/container/{type}/{id} [get]
func (h *Handler) ListContainerFiles(c *gin.Context) {
	containerType := schema.ContainerType(c.Param("type"))
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid container ID"})
		return
	}

	switch containerType {
	case schema.ContainerTypeTask, schema.ContainerTypeProblem, schema.ContainerTypeQuestion,
		schema.ContainerTypeAction, schema.ContainerTypeDefinition, schema.ContainerTypeKnowledgeBit,
		schema.ContainerTypeKnowledgeNode, schema.ContainerTypeStory, schema.ContainerTypeEpic:
		// ok
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported container type for files"})
		return
	}

	folder := h.App.ContainerService.GetFilesFolder(c.Request.Context(), containerType, id)
	if folder == nil {
		c.JSON(http.StatusOK, []any{})
		return
	}

	files, err := h.App.FilesService.ListFilesInAbsoluteDir(*folder)
	if err != nil {
		writeFilesError(c, err)
		return
	}
	c.JSON(http.StatusOK, files)
}
