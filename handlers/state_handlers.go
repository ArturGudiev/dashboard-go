package handlers

import (
	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetStateByID handles GET /states/:id
// @Summary      Get state by ID
// @Description  Returns a state by its ID
// @Tags         states
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "State ID"
// @Success      200  {object}  models.StateFull
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /states/{id} [get]
func (h *Handler) GetStateByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid state ID"})
		return
	}

	stateFull, err := h.App.StatesService.GetStateFull(c.Request.Context(), id)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(404, gin.H{"error": "State not found"})
			return
		}
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, stateFull)
}

// GetAllStates handles GET /states
// @Summary      Get all states
// @Description  Returns all states
// @Tags         states
// @Accept       json
// @Produce      json
// @Success      200  {array}  models.StateFull
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /states [get]
func (h *Handler) GetAllStates(c *gin.Context) {
	statesFull, err := h.App.StatesService.GetAllStatesFull(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, statesFull)
}

// AddState handles POST /states
// @Summary      Add a state
// @Description  Adds a new state
// @Tags         states
// @Accept       json
// @Produce      json
// @Param        state  body      models.NewStateRequest  true  "State to add"
// @Success      200  {object}  models.StateFull
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /states [post]
func (h *Handler) AddState(c *gin.Context) {
	var req models.NewStateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	newState, err := h.App.StatesService.AddState(ctx, req.State, req.Parent)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, newState)	
}

// AddStateRequirement handles POST /states/:id/requirements
// @Summary      Add a state requirement
// @Description  Adds a new state requirement
// @Tags         states
// @Accept       json
// @Produce      json
// @Param        id       path      int                            true  "State ID"
// @Param        request  body      models.StateRequirementShort   true  "State requirement"
// @Success      200      {object}  models.StateRequirementFull
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /states/{id}/requirements [post]
func (h *Handler) AddStateRequirement(c *gin.Context) {
	stateID := c.Param("id")
	stateIDInt, err := strconv.Atoi(stateID)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid state ID"})
		return
	}

	var req models.StateRequirementShort
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	newStateRequirement, err := h.App.StatesService.AddStateRequirement(c.Request.Context(), stateIDInt, req.Description, req.OnceInDays)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, newStateRequirement)	
}

// GetStateRequirementsByStateID handles GET /states/:id/requirements
// @Summary      Get state requirements by state ID
// @Description  Returns state requirements by state ID
// @Tags         states
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "State ID"
// @Success      200  {array}  models.StateRequirementFull
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /states/{id}/requirements [get]
func (h *Handler) GetStateRequirementsByStateID(c *gin.Context) {	
	stateID := c.Param("id")
	stateIDInt, err := strconv.Atoi(stateID)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid state ID"})
		return
	}

	stateRequirements, err := h.App.StatesService.GetStateRequirementsFullByStateID(c.Request.Context(), stateIDInt)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, stateRequirements)
}