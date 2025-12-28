package handlers

import (
	"log"
	"strconv"

	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/ent/containerchild"
	"arturgudiev/dashboard/ent/problem"
	"arturgudiev/dashboard/ent/schema"

	"github.com/gin-gonic/gin"
	"github.com/niemeyer/pretty"
)

// GetProblemByID handles GET /problem/:id
// @Summary      Get problem by ID
// @Description  Returns a problem by its ID
// @Tags         problems
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Problem ID"
// @Success      200  {object}  ProblemResponse
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

	problemEntity, err := h.App.Client.Problem.Get(c.Request.Context(), id)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(404, gin.H{"error": "Problem not found"})
			return
		}
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	pretty.Print(problemEntity)

	// Convert to custom response type to ensure all fields are included
	response := ProblemResponse{
		ID:               problemEntity.ID,
		Description:      problemEntity.Description,
		Tags:             problemEntity.Tags,
		Notes:            problemEntity.Notes,
		Problems:         problemEntity.Problems,
		Questions:        problemEntity.Questions,
		Actions:          problemEntity.Actions,
		Definitions:      problemEntity.Definitions,
		KnowledgeBits:    problemEntity.KnowledgeBits,
		ParentContainers: problemEntity.ParentContainers,
		KnowledgeNodes:   problemEntity.KnowledgeNodes,
		DoneDateTime:     problemEntity.DoneDateTime,
		Solution:         problemEntity.Solution,
	}

	c.JSON(200, response)
}

// GetProblemsByIDs handles POST /get-problems
// @Summary      Get problems by IDs
// @Description  Returns multiple problems by their IDs
// @Tags         problems
// @Accept       json
// @Produce      json
// @Param        request  body      IDsRequest  true  "List of problem IDs"
// @Success      200      {array}   ProblemResponse
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /get-problems [post]
func (h *Handler) GetProblemsByIDs(c *gin.Context) {
	var req IDsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	problems, err := h.App.Client.Problem.Query().Where(problem.IDIn(req.IDs...)).All(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// Convert to slice of ProblemResponse to ensure all fields are included
	responses := make([]ProblemResponse, len(problems))
	for i, p := range problems {
		responses[i] = ProblemResponse{
			ID:               p.ID,
			Description:      p.Description,
			Tags:             p.Tags,
			Notes:            p.Notes,
			Problems:         p.Problems,
			Questions:        p.Questions,
			Actions:          p.Actions,
			Definitions:      p.Definitions,
			KnowledgeBits:    p.KnowledgeBits,
			ParentContainers: p.ParentContainers,
			KnowledgeNodes:   p.KnowledgeNodes,
			DoneDateTime:     p.DoneDateTime,
			Solution:         p.Solution,
		}
	}

	c.JSON(200, responses)
}

// SolveProblem handles PUT /solve-problem/:id
// @Summary      Solve problem
// @Description  Sets the solution for a problem
// @Tags         problems
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Problem ID"
// @Param        request  body      SolveProblemRequest  true  "Solution request"
// @Success      200  {object}  ProblemResponse
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

	// Verify problem exists
	_, err = h.App.Client.Problem.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(404, gin.H{"error": "Problem not found"})
			return
		}
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// Update problem with solution
	updatedProblem, err := h.App.Client.Problem.UpdateOneID(id).
		SetSolution(req.Solution).
		Save(ctx)
	if err != nil {
		log.Printf("Error solving problem %d: %v", id, err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// Convert to custom response type to ensure all fields are included
	response := ProblemResponse{
		ID:               updatedProblem.ID,
		Description:      updatedProblem.Description,
		Tags:             updatedProblem.Tags,
		Notes:            updatedProblem.Notes,
		Problems:         updatedProblem.Problems,
		Questions:        updatedProblem.Questions,
		Actions:          updatedProblem.Actions,
		Definitions:      updatedProblem.Definitions,
		KnowledgeBits:    updatedProblem.KnowledgeBits,
		ParentContainers: updatedProblem.ParentContainers,
		KnowledgeNodes:   updatedProblem.KnowledgeNodes,
		DoneDateTime:     updatedProblem.DoneDateTime,
		Solution:         updatedProblem.Solution,
	}

	c.JSON(200, response)
}

// NewProblem handles POST /new-problem
// @Summary      Create new problem
// @Description  Creates a new problem with optional parent relationship
// @Tags         problems
// @Accept       json
// @Produce      json
// @Param        request  body      NewProblemRequest  true  "Problem creation request"
// @Success      200      {object}  ProblemResponse
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

	problemBuilder := h.App.Client.Problem.Create().
		SetDescription(req.Problem.Description)

	if req.Problem.Tags != nil {
		problemBuilder = problemBuilder.SetTags(req.Problem.Tags)
	}
	if req.Problem.Notes != "" {
		problemBuilder = problemBuilder.SetNotes(req.Problem.Notes)
	}
	if req.Problem.Problems != nil {
		problemBuilder = problemBuilder.SetProblems(req.Problem.Problems)
	}
	if req.Problem.Questions != nil {
		problemBuilder = problemBuilder.SetQuestions(req.Problem.Questions)
	}
	if req.Problem.Actions != nil {
		problemBuilder = problemBuilder.SetActions(req.Problem.Actions)
	}
	if req.Problem.Definitions != nil {
		problemBuilder = problemBuilder.SetDefinitions(req.Problem.Definitions)
	}
	if req.Problem.KnowledgeBits != nil {
		problemBuilder = problemBuilder.SetKnowledgeBits(req.Problem.KnowledgeBits)
	}
	if req.Problem.KnowledgeNodes != nil {
		problemBuilder = problemBuilder.SetKnowledgeNodes(req.Problem.KnowledgeNodes)
	}
	if req.Problem.ParentContainers != nil {
		problemBuilder = problemBuilder.SetParentContainers(req.Problem.ParentContainers)
	}
	if req.Problem.DoneDateTime != nil {
		problemBuilder = problemBuilder.SetDoneDateTime(*req.Problem.DoneDateTime)
	}
	problemBuilder = problemBuilder.SetNillableSolution(req.Problem.Solution)

	newProblem, err := problemBuilder.Save(ctx)
	if err != nil {
		log.Printf("Error creating problem: %v", err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// Handle parent-child relationship if parent is provided
	if req.Parent != nil {
		var parentID int
		var parentType schema.ContainerType

		// Determine parent type and get parent entity
		switch req.Parent.Type {
		case "task":
			parentTask, err := h.App.Client.Task.Get(ctx, req.Parent.Obj.ID)
			if err != nil {
				if ent.IsNotFound(err) {
					c.JSON(404, gin.H{"error": "Parent task not found"})
					return
				}
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
			parentID = parentTask.ID
			parentType = schema.ContainerTypeTask
		case "problem":
			parentProblem, err := h.App.Client.Problem.Get(ctx, req.Parent.Obj.ID)
			if err != nil {
				if ent.IsNotFound(err) {
					c.JSON(404, gin.H{"error": "Parent problem not found"})
					return
				}
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
			parentID = parentProblem.ID
			parentType = schema.ContainerTypeProblem
		default:
			c.JSON(400, gin.H{"error": "Unsupported parent type"})
			return
		}

		// Check if relationship already exists
		exists, err := h.App.Client.ContainerChild.Query().
			Where(
				containerchild.ParentTypeEQ(parentType),
				containerchild.ParentID(parentID),
				containerchild.ChildTypeEQ(schema.ContainerTypeProblem),
				containerchild.ChildID(newProblem.ID),
			).
			Exist(ctx)
		if err != nil {
			log.Printf("Error checking relationship: %v", err)
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		if !exists {
			childCount, err := h.App.Client.ContainerChild.Query().
				Where(
					containerchild.ParentTypeEQ(parentType),
					containerchild.ParentID(parentID),
					containerchild.ChildTypeEQ(schema.ContainerTypeProblem),
				).
				Count(ctx)
			if err != nil {
				log.Printf("Error counting children: %v", err)
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}

			parentCount, err := h.App.Client.ContainerChild.Query().
				Where(
					containerchild.ChildTypeEQ(schema.ContainerTypeProblem),
					containerchild.ChildID(newProblem.ID),
					containerchild.ParentTypeEQ(parentType),
				).
				Count(ctx)
			if err != nil {
				log.Printf("Error counting parents: %v", err)
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}

			_, err = h.App.Client.ContainerChild.Create().
				SetParentType(parentType).
				SetParentID(parentID).
				SetChildType(schema.ContainerTypeProblem).
				SetChildID(newProblem.ID).
				SetChildOrder(childCount).
				SetParentOrder(parentCount).
				Save(ctx)
			if err != nil {
				log.Printf("Error creating relationship: %v", err)
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
		}
	}

	// Convert to custom response type to ensure all fields are included
	response := ProblemResponse{
		ID:               newProblem.ID,
		Description:      newProblem.Description,
		Tags:             newProblem.Tags,
		Notes:            newProblem.Notes,
		Problems:         newProblem.Problems,
		Questions:        newProblem.Questions,
		Actions:          newProblem.Actions,
		Definitions:      newProblem.Definitions,
		KnowledgeBits:    newProblem.KnowledgeBits,
		ParentContainers: newProblem.ParentContainers,
		KnowledgeNodes:   newProblem.KnowledgeNodes,
		DoneDateTime:     newProblem.DoneDateTime,
		Solution:         newProblem.Solution,
	}

	c.JSON(200, response)
}

// UpdateProblem handles PUT /update-problem
// @Summary      Update problem
// @Description  Updates an existing problem by ID
// @Tags         problems
// @Accept       json
// @Produce      json
// @Param        request  body      UpdateProblemRequest  true  "Problem update request"
// @Success      200      {object}  ProblemResponse
// @Failure      400      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /update-problem [put]
func (h *Handler) UpdateProblem(c *gin.Context) {
	var req UpdateProblemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()

	_, err := h.App.Client.Problem.Get(ctx, req.ID)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(404, gin.H{"error": "Problem not found"})
			return
		}
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	problemBuilder := h.App.Client.Problem.UpdateOneID(req.ID).
		SetDescription(req.Description)

	if req.Tags != nil {
		problemBuilder = problemBuilder.SetTags(req.Tags)
	}
	if req.Notes != "" {
		problemBuilder = problemBuilder.SetNotes(req.Notes)
	}
	if req.Problems != nil {
		problemBuilder = problemBuilder.SetProblems(req.Problems)
	}
	if req.Questions != nil {
		problemBuilder = problemBuilder.SetQuestions(req.Questions)
	}
	if req.Actions != nil {
		problemBuilder = problemBuilder.SetActions(req.Actions)
	}
	if req.Definitions != nil {
		problemBuilder = problemBuilder.SetDefinitions(req.Definitions)
	}
	if req.KnowledgeBits != nil {
		problemBuilder = problemBuilder.SetKnowledgeBits(req.KnowledgeBits)
	}
	if req.KnowledgeNodes != nil {
		problemBuilder = problemBuilder.SetKnowledgeNodes(req.KnowledgeNodes)
	}
	if req.ParentContainers != nil {
		problemBuilder = problemBuilder.SetParentContainers(req.ParentContainers)
	}
	if req.DoneDateTime != nil {
		problemBuilder = problemBuilder.SetDoneDateTime(*req.DoneDateTime)
	}
	problemBuilder = problemBuilder.SetNillableSolution(req.Solution)

	updatedProblem, err := problemBuilder.Save(ctx)
	if err != nil {
		log.Printf("Error updating problem: %v", err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// Convert to custom response type to ensure all fields are included
	response := ProblemResponse{
		ID:               updatedProblem.ID,
		Description:      updatedProblem.Description,
		Tags:             updatedProblem.Tags,
		Notes:            updatedProblem.Notes,
		Problems:         updatedProblem.Problems,
		Questions:        updatedProblem.Questions,
		Actions:          updatedProblem.Actions,
		Definitions:      updatedProblem.Definitions,
		KnowledgeBits:    updatedProblem.KnowledgeBits,
		ParentContainers: updatedProblem.ParentContainers,
		KnowledgeNodes:   updatedProblem.KnowledgeNodes,
		DoneDateTime:     updatedProblem.DoneDateTime,
		Solution:         updatedProblem.Solution,
	}

	c.JSON(200, response)
}
