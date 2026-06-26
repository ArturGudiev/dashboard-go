package handlers

import (
	"strconv"
	"strings"

	"arturgudiev/dashboard/models"
	"github.com/gin-gonic/gin"
)

// GetStateRequirements handles GET /state-requirements
// @Summary      Get state requirements by their IDs
// @Description  Returns state requirements by their IDs
// @Tags         state-requirements
// @Accept       json
// @Produce      json
// @Param        ids   query      []int  true  "State requirement IDs (comma separated)"
// @Success      200  {array}  models.StateRequirementFull
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /state-requirements [get]
func (h *Handler) GetStateRequirementsByIDs(c *gin.Context) {
	idsParam := c.Query("ids")
	ids := strings.Split(idsParam, ",")
	idsInt := make([]int, len(ids))
	for i, id := range ids {
		idInt, err := strconv.Atoi(id)
		if err != nil {
			c.JSON(400, gin.H{"error": "Invalid state requirement ID"})
			return
		}
		idsInt[i] = idInt
	}

	stateRequirements, err := h.App.StatesService.GetStateRequirementsFullByIDs(c.Request.Context(), idsInt)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, stateRequirements)
}

// GetStateRequirementsChecksByStateRequirementID handles GET /state-requirements/:id/checks
// @Summary      Get state requirement checks by state requirement ID
// @Description  Returns state requirement checks by state requirement ID
// @Tags         state-requirements
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "State requirement ID"
// @Success      200  {array}  ent.StateRequirementCheck
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /state-requirements/{id}/checks [get]
func (h *Handler) GetStateRequirementsChecksByStateRequirementID(c *gin.Context) {
	stateRequirementID := c.Param("id")
	stateRequirementIDInt, err := strconv.Atoi(stateRequirementID)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid state requirement ID"})
		return
	}

	stateRequirementChecks, err := h.App.StateRequirementChecksRepository.GetStateRequirementChecksByStateRequirementID(c.Request.Context(), stateRequirementIDInt)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(200, stateRequirementChecks)
}

// AddStateRequirementCheck handles POST /state-requirements/:id/checks
// @Summary      Add a state requirement check
// @Description  Adds a new state requirement check
// @Tags         state-requirements
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "State requirement ID"
// @Param        request  body      models.StateRequirementCheckShort   true  "State requirement check"
// @Success      200  {object}  ent.StateRequirementCheck
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /state-requirements/{id}/checks [post]
func (h *Handler) AddStateRequirementCheck(c *gin.Context) {
	stateRequirementID := c.Param("id")
	stateRequirementIDInt, err := strconv.Atoi(stateRequirementID)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid state requirement ID"})
		return
	}

	var req models.StateRequirementCheckShort
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	newStateRequirementCheck, err := h.App.StateRequirementChecksRepository.AddStateRequirementCheck(c.Request.Context(), stateRequirementIDInt, req.IsFulfilled)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, newStateRequirementCheck)
}