package handlers

import (
	"arturgudiev/dashboard/ent"

	"github.com/gin-gonic/gin"
)

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
