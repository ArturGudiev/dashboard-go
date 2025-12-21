package handlers

import "github.com/gin-gonic/gin"

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
