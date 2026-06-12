package handlers

import (
	"log"
	"strconv"

	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/models"

	"github.com/gin-gonic/gin"
)

func bindPatchContainerRequest(c *gin.Context) (PatchContainerByIDRequest, bool) {
	var req PatchContainerByIDRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return req, false
	}
	if req.Description == nil && req.Notes == nil {
		c.JSON(400, gin.H{"error": "At least one of description or notes must be provided"})
		return req, false
	}
	return req, true
}

// PatchQuestionByID handles PATCH /question/:id
// @Summary      Patch question by ID
// @Description  Partially updates a question's description and/or notes
// @Tags         questions
// @Accept       json
// @Produce      json
// @Param        id       path      int                      true  "Question ID"
// @Param        request  body      PatchContainerByIDRequest  true  "Fields to update"
// @Success      200      {object}  models.QuestionFull
// @Failure      400      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /question/{id} [patch]
func (h *Handler) PatchQuestionByID(c *gin.Context) {
	id, ok := parsePatchID(c, "question")
	if !ok {
		return
	}
	req, ok := bindPatchContainerRequest(c)
	if !ok {
		return
	}

	ctx := c.Request.Context()
	questionFull, err := h.App.QuestionsService.UpdateQuestion(ctx, models.QuestionPartial{
		ID:          id,
		Description: req.Description,
		Notes:       req.Notes,
	})
	if err != nil {
		writePatchError(c, id, "question", err)
		return
	}
	c.JSON(200, questionFull)
}

// PatchProblemByID handles PATCH /problem/:id
// @Summary      Patch problem by ID
// @Description  Partially updates a problem's description and/or notes
// @Tags         problems
// @Accept       json
// @Produce      json
// @Param        id       path      int                      true  "Problem ID"
// @Param        request  body      PatchContainerByIDRequest  true  "Fields to update"
// @Success      200      {object}  models.ProblemFull
// @Failure      400      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /problem/{id} [patch]
func (h *Handler) PatchProblemByID(c *gin.Context) {
	id, ok := parsePatchID(c, "problem")
	if !ok {
		return
	}
	req, ok := bindPatchContainerRequest(c)
	if !ok {
		return
	}

	ctx := c.Request.Context()
	problemFull, err := h.App.ProblemService.UpdateProblem(ctx, models.ProblemPartial{
		ID:          id,
		Description: req.Description,
		Notes:       req.Notes,
	})
	if err != nil {
		writePatchError(c, id, "problem", err)
		return
	}
	c.JSON(200, problemFull)
}

// PatchStoryByID handles PATCH /story/:id
// @Summary      Patch story by ID
// @Description  Partially updates a story's description and/or notes
// @Tags         stories
// @Accept       json
// @Produce      json
// @Param        id       path      int                      true  "Story ID"
// @Param        request  body      PatchContainerByIDRequest  true  "Fields to update"
// @Success      200      {object}  models.StoryFull
// @Failure      400      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /story/{id} [patch]
func (h *Handler) PatchStoryByID(c *gin.Context) {
	id, ok := parsePatchID(c, "story")
	if !ok {
		return
	}
	req, ok := bindPatchContainerRequest(c)
	if !ok {
		return
	}

	ctx := c.Request.Context()
	storyFull, err := h.App.StoriesService.UpdateStory(ctx, models.StoryPartial{
		ID:          id,
		Description: req.Description,
		Notes:       req.Notes,
	})
	if err != nil {
		writePatchError(c, id, "story", err)
		return
	}
	c.JSON(200, storyFull)
}

// PatchEpicByID handles PATCH /epic/:id
// @Summary      Patch epic by ID
// @Description  Partially updates an epic's description and/or notes
// @Tags         epics
// @Accept       json
// @Produce      json
// @Param        id       path      int                      true  "Epic ID"
// @Param        request  body      PatchContainerByIDRequest  true  "Fields to update"
// @Success      200      {object}  models.EpicFull
// @Failure      400      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /epic/{id} [patch]
func (h *Handler) PatchEpicByID(c *gin.Context) {
	id, ok := parsePatchID(c, "epic")
	if !ok {
		return
	}
	req, ok := bindPatchContainerRequest(c)
	if !ok {
		return
	}

	ctx := c.Request.Context()
	epicFull, err := h.App.EpicsService.UpdateEpic(ctx, models.EpicPartial{
		ID:          id,
		Description: req.Description,
		Notes:       req.Notes,
	})
	if err != nil {
		writePatchError(c, id, "epic", err)
		return
	}
	c.JSON(200, epicFull)
}

// PatchRepetitiveTaskByID handles PATCH /repetitive-tasks/:id
// @Summary      Patch repetitive task by ID
// @Description  Partially updates a repetitive task's description and/or notes
// @Tags         repetitive-tasks
// @Accept       json
// @Produce      json
// @Param        id       path      int                      true  "Repetitive task ID"
// @Param        request  body      PatchContainerByIDRequest  true  "Fields to update"
// @Success      200      {object}  models.RepetitiveTaskResponse
// @Failure      400      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /repetitive-tasks/{id} [patch]
// PatchLongTaskByID handles PATCH /long-tasks/:id
// @Summary      Patch long task by ID
// @Description  Partially updates a long task's description and/or notes
// @Tags         long-tasks
// @Accept       json
// @Produce      json
// @Param        id       path      int                      true  "Long task ID"
// @Param        request  body      PatchContainerByIDRequest  true  "Fields to update"
// @Success      200      {object}  ent.LongTask
// @Failure      400      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /long-tasks/{id} [patch]
// PatchDirectionByID handles PATCH /directions/:id
// @Summary      Patch direction by ID
// @Description  Partially updates a direction's description, notes, and/or closed state
// @Tags         directions
// @Accept       json
// @Produce      json
// @Param        id       path      int                         true  "Direction ID"
// @Param        request  body      PatchDirectionByIDRequest  true  "Fields to update"
// @Success      200      {object}  models.DirectionFull
// @Failure      400      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /directions/{id} [patch]
func (h *Handler) PatchDirectionByID(c *gin.Context) {
	id, ok := parsePatchID(c, "direction")
	if !ok {
		return
	}

	var req PatchDirectionByIDRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if req.Description == nil && req.Notes == nil && req.Closed == nil {
		c.JSON(400, gin.H{"error": "At least one of description, notes, or closed must be provided"})
		return
	}

	ctx := c.Request.Context()
	directionFull, err := h.App.DirectionsService.UpdateDirection(ctx, models.DirectionPartial{
		ID:          id,
		Description: req.Description,
		Notes:       req.Notes,
		Closed:      req.Closed,
	})
	if err != nil {
		writePatchError(c, id, "direction", err)
		return
	}
	c.JSON(200, directionFull)
}

func (h *Handler) PatchLongTaskByID(c *gin.Context) {
	id, ok := parsePatchID(c, "long task")
	if !ok {
		return
	}
	req, ok := bindPatchContainerRequest(c)
	if !ok {
		return
	}

	ctx := c.Request.Context()
	longTask, err := h.App.LongTasksService.UpdateLongTask(ctx, models.LongTaskPartial{
		ID:          id,
		Description: req.Description,
		Notes:       req.Notes,
	})
	if err != nil {
		writePatchError(c, id, "long task", err)
		return
	}
	c.JSON(200, longTask)
}

func (h *Handler) PatchRepetitiveTaskByID(c *gin.Context) {
	id, ok := parsePatchID(c, "repetitive task")
	if !ok {
		return
	}
	req, ok := bindPatchContainerRequest(c)
	if !ok {
		return
	}

	ctx := c.Request.Context()
	repetitiveTask, err := h.App.RepetitiveTaskService.UpdateRepetitiveTask(ctx, models.RepetitiveTaskPartial{
		ID:          id,
		Description: req.Description,
		Notes:       req.Notes,
	})
	if err != nil {
		writePatchError(c, id, "repetitive task", err)
		return
	}
	c.JSON(200, repetitiveTask)
}

func parsePatchID(c *gin.Context, entityName string) (int, bool) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid " + entityName + " ID"})
		return 0, false
	}
	return id, true
}

func writePatchError(c *gin.Context, id int, entityName string, err error) {
	if ent.IsNotFound(err) {
		c.JSON(404, gin.H{"error": entityName + " not found"})
		return
	}
	log.Printf("Error patching %s %d: %v", entityName, id, err)
	c.JSON(500, gin.H{"error": err.Error()})
}
