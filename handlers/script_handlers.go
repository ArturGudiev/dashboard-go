package handlers

import (
	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/ent/schema"
	"arturgudiev/dashboard/models"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// ListScripts handles GET /scripts
// @Summary      List scripts
// @Description  Lists scripts filtered by name, scope (all|global|local), and optional container
// @Tags         scripts
// @Produce      json
// @Param        q              query     string  false  "Name filter"
// @Param        scope          query     string  false  "all, global, or local"
// @Param        containerType  query     string  false  "Container type for local/all scope"
// @Param        containerId    query     int     false  "Container ID for local/all scope"
// @Success      200  {array}   models.ScriptListItem
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /scripts [get]
func (h *Handler) ListScripts(c *gin.Context) {
	filter := models.ScriptListFilter{
		Query: c.Query("q"),
		Scope: c.Query("scope"),
	}
	if typeStr := c.Query("containerType"); typeStr != "" {
		filter.ContainerType = schema.ContainerType(typeStr)
		if !validateContainerType(filter.ContainerType) {
			c.JSON(400, gin.H{"error": "Invalid container type"})
			return
		}
	}
	if idStr := c.Query("containerId"); idStr != "" {
		id, err := strconv.Atoi(idStr)
		if err != nil || id <= 0 {
			c.JSON(400, gin.H{"error": "Invalid containerId"})
			return
		}
		filter.ContainerID = id
	}

	items, err := h.App.ScriptsService.List(c.Request.Context(), filter)
	if err != nil {
		if strings.Contains(err.Error(), "required") {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, items)
}

// GetScript handles GET /scripts/:id
// @Summary      Get script
// @Description  Returns a full script by ID
// @Tags         scripts
// @Produce      json
// @Param        id   path      int  true  "Script ID"
// @Success      200  {object}  models.ScriptFull
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /scripts/{id} [get]
func (h *Handler) GetScript(c *gin.Context) {
	id, ok := parsePositiveID(c, "id")
	if !ok {
		return
	}
	script, err := h.App.ScriptsService.Get(c.Request.Context(), id)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(404, gin.H{"error": "Script not found"})
			return
		}
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, script)
}

// CreateScript handles POST /scripts
// @Summary      Create script
// @Description  Creates a global script, or a local script when container is provided
// @Tags         scripts
// @Accept       json
// @Produce      json
// @Param        request  body      models.ScriptShort  true  "Script"
// @Success      200      {object}  models.ScriptFull
// @Failure      400      {object}  map[string]string
// @Failure      409      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /scripts [post]
func (h *Handler) CreateScript(c *gin.Context) {
	var req models.ScriptShort
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if req.Container != nil && !validateContainerType(req.Container.Type) {
		c.JSON(400, gin.H{"error": "Invalid container type"})
		return
	}
	script, err := h.App.ScriptsService.Create(c.Request.Context(), req)
	if err != nil {
		if ent.IsConstraintError(err) {
			c.JSON(409, gin.H{"error": "Script with this name already exists"})
			return
		}
		if isClientScriptError(err) {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, script)
}

// PatchScript handles PATCH /scripts/:id
// @Summary      Update script
// @Description  Partially updates a script
// @Tags         scripts
// @Accept       json
// @Produce      json
// @Param        id       path      int                  true  "Script ID"
// @Param        request  body      models.ScriptPartial  true  "Fields to update"
// @Success      200      {object}  models.ScriptFull
// @Failure      400      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Failure      409      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /scripts/{id} [patch]
func (h *Handler) PatchScript(c *gin.Context) {
	id, ok := parsePositiveID(c, "id")
	if !ok {
		return
	}
	var req models.ScriptPartial
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	script, err := h.App.ScriptsService.Update(c.Request.Context(), id, req)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(404, gin.H{"error": "Script not found"})
			return
		}
		if ent.IsConstraintError(err) {
			c.JSON(409, gin.H{"error": "Script with this name already exists"})
			return
		}
		if isClientScriptError(err) {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, script)
}

// DeleteScript handles DELETE /scripts/:id
// @Summary      Delete script
// @Description  Deletes a script by ID
// @Tags         scripts
// @Produce      json
// @Param        id   path      int  true  "Script ID"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /scripts/{id} [delete]
func (h *Handler) DeleteScript(c *gin.Context) {
	id, ok := parsePositiveID(c, "id")
	if !ok {
		return
	}
	if err := h.App.ScriptsService.Delete(c.Request.Context(), id); err != nil {
		if ent.IsNotFound(err) {
			c.JSON(404, gin.H{"error": "Script not found"})
			return
		}
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"status": "ok"})
}

// ValidateScript handles POST /scripts/validate
// @Summary      Validate script code
// @Description  Parses script code for syntax errors without executing host APIs
// @Tags         scripts
// @Accept       json
// @Produce      json
// @Param        request  body      models.ScriptValidateRequest  true  "Code to validate"
// @Success      200      {object}  models.ScriptValidateResponse
// @Failure      400      {object}  map[string]string
// @Router       /scripts/validate [post]
func (h *Handler) ValidateScript(c *gin.Context) {
	var req models.ScriptValidateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, h.App.ScriptsService.Validate(req.Code))
}

// RunScript handles POST /scripts/:id/run
// @Summary      Run script
// @Description  Executes a script against a container with optional params
// @Tags         scripts
// @Accept       json
// @Produce      json
// @Param        id       path      int                     true  "Script ID"
// @Param        request  body      models.ScriptRunRequest  true  "Run request"
// @Success      200      {object}  models.ScriptRunResponse
// @Failure      400      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /scripts/{id}/run [post]
func (h *Handler) RunScript(c *gin.Context) {
	id, ok := parsePositiveID(c, "id")
	if !ok {
		return
	}
	var req models.ScriptRunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if req.Container.ID <= 0 || req.Container.Type == "" {
		c.JSON(400, gin.H{"error": "container id and type are required"})
		return
	}
	if !validateContainerType(req.Container.Type) {
		c.JSON(400, gin.H{"error": "Invalid container type"})
		return
	}

	result, err := h.App.ScriptsService.Execute(c.Request.Context(), id, req.Container, req.Params)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(404, gin.H{"error": "Script not found"})
			return
		}
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	if result != nil && !result.OK {
		c.JSON(400, result)
		return
	}
	c.JSON(200, result)
}

func parsePositiveID(c *gin.Context, name string) (int, bool) {
	id, err := strconv.Atoi(c.Param(name))
	if err != nil || id <= 0 {
		c.JSON(400, gin.H{"error": "Invalid " + name})
		return 0, false
	}
	return id, true
}

func isClientScriptError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "must not be empty") ||
		strings.Contains(msg, "syntax error") ||
		strings.Contains(msg, "invalid type") ||
		strings.Contains(msg, "duplicate param") ||
		strings.Contains(msg, "missing required param") ||
		strings.Contains(msg, "must be ") ||
		strings.Contains(msg, "required for local")
}
