package handlers

import (
	"strconv"

	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/models"

	"github.com/gin-gonic/gin"
)

// GetStoryByID handles GET /story/:id
// @Summary      Get story by ID
// @Description  Returns a story by its ID
// @Tags         stories
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Story ID"
// @Success      200  {object}  models.StoryFull
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /story/{id} [get]
func (h *Handler) GetStoryByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid story ID"})
		return
	}

	storyFull, err := h.App.StoriesService.GetStoryFull(c.Request.Context(), id)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(404, gin.H{"error": "Story not found"})
			return
		}
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, storyFull)
}

// GetStoriesByIDs handles POST /get-stories
// @Summary      Get stories by IDs
// @Description  Returns multiple stories by their IDs
// @Tags         stories
// @Accept       json
// @Produce      json
// @Param        request  body      IDsRequest  true  "List of story IDs"
// @Success      200      {array}   models.StoryFull
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /get-stories [post]
func (h *Handler) GetStoriesByIDs(c *gin.Context) {
	var req IDsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	storiesFull, err := h.App.StoriesService.GetStoriesFull(c.Request.Context(), req.IDs)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(404, gin.H{"error": "Stories not found"})
			return
		}
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, storiesFull)
}

// NewStory handles POST /new-story
// @Summary      Create new story
// @Description  Creates a new story with optional parent relationship
// @Tags         stories
// @Accept       json
// @Produce      json
// @Param        request  body      models.NewStoryRequest  true  "Story creation request"
// @Success      200      {object}  models.StoryFull
// @Failure      400      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /new-story [post]
func (h *Handler) NewStory(c *gin.Context) {
	var req models.NewStoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	newStory, err := h.App.StoriesService.AddStory(ctx, req.Story, req.Parent)

	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(404, gin.H{"error": "Parent container not found"})
			return
		}
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, newStory)
}

// UpdateStory handles PUT /update-story
// @Summary      Update story
// @Description  Updates an existing story by ID
// @Tags         stories
// @Accept       json
// @Produce      json
// @Param        request  body      models.StoryPartial  true  "Story update request"
// @Success      200      {object}  models.StoryFull
// @Failure      400      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /update-story [put]
func (h *Handler) UpdateStory(c *gin.Context) {
	var req models.StoryPartial
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()

	storyFull, err := h.App.StoriesService.UpdateStory(ctx, req)

	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(404, gin.H{"error": "Story not found"})
			return
		}
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, storyFull)
}
