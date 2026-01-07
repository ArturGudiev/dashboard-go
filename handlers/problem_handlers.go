package handlers

import (
	"arturgudiev/dashboard/models"
	"log"
	"strconv"

	"arturgudiev/dashboard/ent"

	"github.com/gin-gonic/gin"
)

// GetProblemByID handles GET /problem/:id
// @Summary      Get problem by ID
// @Description  Returns a problem by its ID
// @Tags         problems
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Problem ID"
// @Success      200  {object}  models.ProblemFull
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /problem/{id} [get]
func (h *Handler) GetProblemByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid problem ID"})
		return
	}

	problemFull, err := h.App.ProblemService.GetProblemFull(c.Request.Context(), id)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(404, gin.H{"error": "Problem not found"})
			return
		}
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, problemFull)
}

// GetProblemsByIDs handles POST /get-problems
// @Summary      Get problems by IDs
// @Description  Returns multiple problems by their IDs
// @Tags         problems
// @Accept       json
// @Produce      json
// @Param        request  body      IDsRequest  true  "List of problem IDs"
// @Success      200      {array}   models.ProblemFull
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /get-problems [post]
func (h *Handler) GetProblemsByIDs(c *gin.Context) {
	var req IDsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	problemsFull, err := h.App.ProblemService.GetProblemsFull(c.Request.Context(), req.IDs)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(404, gin.H{"error": "Problems not found"})
			return
		}
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, problemsFull)
}

// SolveProblem handles PUT /solve-problem/:id
// @Summary      Solve problem
// @Description  Sets the solution for a problem
// @Tags         problems
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Problem ID"
// @Param        request  body      SolveProblemRequest  true  "Solution request"
// @Success      200  {object}  models.ProblemFull
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /solve-problem/{id} [put]
func (h *Handler) SolveProblem(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid problem ID"})
		return
	}

	var req SolveProblemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	problemFull, err := h.App.ProblemService.SolveProblem(ctx, id, req.Solution)

	if err != nil {
		log.Printf("Error solving problem %d: %v", id, err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, problemFull)
}

// NewProblem handles POST /new-problem
// @Summary      Create new problem
// @Description  Creates a new problem with optional parent relationship
// @Tags         problems
// @Accept       json
// @Produce      json
// @Param        request  body      NewProblemRequest  true  "Problem creation request"
// @Success      200      {object}  models.ProblemFull
// @Failure      400      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /new-problem [post]
func (h *Handler) NewProblem(c *gin.Context) {
	var req NewProblemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	newProblem, err := h.App.ProblemService.AddProblem(ctx, req.Problem, req.Parent)

	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(404, gin.H{"error": "Parent container not found"})
			return
		}
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, newProblem)
}

// UpdateProblem handles PUT /update-problem
// @Summary      Update problem
// @Description  Updates an existing problem by ID
// @Tags         problems
// @Accept       json
// @Produce      json
// @Param        request  body      models.ProblemPartial  true  "Problem update request"
// @Success      200      {object}  ProblemResponse
// @Failure      400      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /update-problem [put]
func (h *Handler) UpdateProblem(c *gin.Context) {
	var req models.ProblemPartial
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()

	problemFull, err := h.App.ProblemService.UpdateProblem(ctx, req)

	if err != nil {
		log.Printf("Error updating problem: %v", err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, problemFull)
}
