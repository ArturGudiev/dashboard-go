package handlers

import (
	"log"
	"strconv"

	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/models"

	"github.com/gin-gonic/gin"
)

// GetQuestionByID handles GET /question/:id
// @Summary      Get question by ID
// @Description  Returns a question by its ID
// @Tags         questions
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Question ID"
// @Success      200  {object}  models.QuestionFull
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /question/{id} [get]
func (h *Handler) GetQuestionByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid question ID"})
		return
	}

	questionFull, err := h.App.QuestionsService.GetQuestionFull(c.Request.Context(), id)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(404, gin.H{"error": "Question not found"})
			return
		}
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, questionFull)
}

// GetQuestionsByIDs handles POST /get-questions
// @Summary      Get questions by IDs
// @Description  Returns multiple questions by their IDs
// @Tags         questions
// @Accept       json
// @Produce      json
// @Param        request  body      IDsRequest  true  "List of question IDs"
// @Success      200      {array}   models.QuestionFull
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /get-questions [post]
func (h *Handler) GetQuestionsByIDs(c *gin.Context) {
	var req IDsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	questionsFull, err := h.App.QuestionsService.GetQuestionsFull(c.Request.Context(), req.IDs)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(404, gin.H{"error": "Questions not found"})
			return
		}
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, questionsFull)
}

// NewQuestion handles POST /new-question
// @Summary      Create new question
// @Description  Creates a new question with optional parent relationship
// @Tags         questions
// @Accept       json
// @Produce      json
// @Param        request  body      NewQuestionRequest  true  "Question creation request"
// @Success      200      {object}  models.QuestionFull
// @Failure      400      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /new-question [post]
func (h *Handler) NewQuestion(c *gin.Context) {
	var req NewQuestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	newQuestion, err := h.App.QuestionsService.AddQuestion(ctx, req.Question, req.Parent)

	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(404, gin.H{"error": "Parent container not found"})
			return
		}
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, newQuestion)
}

// UpdateQuestion handles PUT /update-question
// @Summary      Update question
// @Description  Updates an existing question by ID
// @Tags         questions
// @Accept       json
// @Produce      json
// @Param        request  body      models.QuestionPartial  true  "Question update request"
// @Success      200      {object}  models.QuestionFull
// @Failure      400      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /update-question [put]
func (h *Handler) UpdateQuestion(c *gin.Context) {
	var req models.QuestionPartial
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()

	questionFull, err := h.App.QuestionsService.UpdateQuestion(ctx, req)

	if err != nil {
		log.Printf("Error updating question: %v", err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, questionFull)
}

// AnswerQuestion handles POST /answer-question/:id
// @Summary      Answer question
// @Description  Sets the answer for a question
// @Tags         questions
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Question ID"
// @Param        request  body      AnswerQuestionRequest  true  "Answer request"
// @Success      200  {object}  models.QuestionFull
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /answer-question/{id} [post]
func (h *Handler) AnswerQuestion(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid question ID"})
		return
	}

	var req AnswerQuestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	questionFull, err := h.App.QuestionsService.AnswerQuestion(ctx, id, req.Answer)

	if err != nil {
		log.Printf("Error answering question %d: %v", id, err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, questionFull)
}
