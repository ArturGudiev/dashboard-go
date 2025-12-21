package handlers

import "github.com/gin-gonic/gin"

// TestResponse represents a test response
type TestResponse struct {
	ID   int      `json:"id"`
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

// GetTests handles GET /tests
// @Summary      Get all tests
// @Description  Returns a list of all tests
// @Tags         tests
// @Accept       json
// @Produce      json
// @Success      200  {array}   TestResponse
// @Failure      500  {object}  map[string]string
// @Router       /tests [get]
func (h *Handler) GetTests(c *gin.Context) {
	tests, err := h.App.Client.Test.Query().All(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	response := make([]TestResponse, len(tests))
	for i, t := range tests {
		response[i] = TestResponse{
			ID:   t.ID,
			Name: t.Name,
			Tags: t.Tags,
		}
	}

	c.JSON(200, response)
}
