package handlers

import (
	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/ent/schema"
	"strconv"

	"github.com/gin-gonic/gin"
)

func validateContainerType(containerType schema.ContainerType) bool {
	switch containerType {
	case schema.ContainerTypeEpic, schema.ContainerTypeStory, schema.ContainerTypeTask,
		schema.ContainerTypeQuestion, schema.ContainerTypeProblem, schema.ContainerTypeKnowledgeNode,
		schema.ContainerTypeKnowledgeBit, schema.ContainerTypeDefinition, schema.ContainerTypeAction,
		schema.ContainerTypeRepetitiveTask, schema.ContainerTypeState,
		schema.ContainerTypeLongTask, schema.ContainerTypeDirection:
		return true
	default:
		return false
	}
}

// AddContainerVariable handles POST /container-variables
// @Summary      Add variable to container
// @Description  Creates a variable entry for the given container (creates variables stack if needed)
// @Tags         container variables
// @Accept       json
// @Produce      json
// @Param        request  body      AddContainerVariableRequest  true  "Variable creation request"
// @Success      200      {object}  ent.ContainerVariables
// @Failure      400      {object}  map[string]string
// @Failure      409      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /container-variables [post]
func (h *Handler) AddContainerVariable(c *gin.Context) {
	var req AddContainerVariableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if !validateContainerType(req.ContainerType) {
		c.JSON(400, gin.H{"error": "Invalid container type"})
		return
	}

	ctx := c.Request.Context()
	variable, err := h.App.ContainerVariablesRepository.AddVariableWithValue(
		ctx,
		req.ContainerType,
		req.ContainerID,
		req.VariableName,
		req.VariableValue,
	)
	if err != nil {
		if ent.IsConstraintError(err) {
			c.JSON(409, gin.H{"error": "Variable with this name already exists for container"})
			return
		}
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, variable)
}

// PatchContainerVariable handles PATCH /container-variables/:id
// @Summary      Update container variable
// @Description  Partially updates a container variable's name and/or value
// @Tags         container variables
// @Accept       json
// @Produce      json
// @Param        id       path      int                              true  "Container variable ID"
// @Param        request  body      PatchContainerVariableRequest  true  "Fields to update"
// @Success      200      {object}  ent.ContainerVariables
// @Failure      400      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Failure      409      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /container-variables/{id} [patch]
func (h *Handler) PatchContainerVariable(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid variable ID"})
		return
	}

	var req PatchContainerVariableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if req.VariableName == nil && req.VariableValue == nil {
		c.JSON(400, gin.H{"error": "At least one of variableName or variableValue must be provided"})
		return
	}

	ctx := c.Request.Context()
	variable, err := h.App.ContainerVariablesRepository.UpdateVariable(ctx, id, req.VariableName, req.VariableValue)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(404, gin.H{"error": "Variable not found"})
			return
		}
		if ent.IsConstraintError(err) {
			c.JSON(409, gin.H{"error": "Variable with this name already exists for container"})
			return
		}
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, variable)
}

// RemoveContainerVariable handles DELETE /container-variables/:id
// @Summary      Remove container variable
// @Description  Deletes a container variable by ID
// @Tags         container variables
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Container variable ID"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /container-variables/{id} [delete]
func (h *Handler) RemoveContainerVariable(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid variable ID"})
		return
	}

	ctx := c.Request.Context()
	err = h.App.ContainerVariablesRepository.RemoveVariable(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(404, gin.H{"error": "Variable not found"})
			return
		}
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"status": "ok"})
}
