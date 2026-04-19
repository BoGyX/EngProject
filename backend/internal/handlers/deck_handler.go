package handlers

import (
	"english-learning/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type DeckHandler struct {
	deckService         *services.DeckService
	courseService       *services.CourseService
	courseAccessService *services.CourseAccessService
}

func NewDeckHandler(deckService *services.DeckService, courseService *services.CourseService, courseAccessService *services.CourseAccessService) *DeckHandler {
	return &DeckHandler{
		deckService:         deckService,
		courseService:       courseService,
		courseAccessService: courseAccessService,
	}
}

type CreateDeckRequest struct {
	CourseID    int64   `json:"course_id" binding:"required"`
	Title       string  `json:"title" binding:"required"`
	Slug        string  `json:"slug"`
	Description *string `json:"description,omitempty"`
	Position    int     `json:"position"`
}

type UpdateDeckRequest struct {
	Title       string  `json:"title" binding:"required"`
	Slug        string  `json:"slug"`
	Description *string `json:"description,omitempty"`
	Position    int     `json:"position"`
}

func (h *DeckHandler) GetAllDecks(c *gin.Context) {
	courseIDParam := c.Query("course_id")

	if courseIDParam != "" {
		courseID, err := strconv.ParseInt(courseIDParam, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course_id format"})
			return
		}

		course, err := h.courseService.GetCourseByID(courseID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if !ensureCourseAccess(c, h.courseAccessService, course) {
			return
		}

		decks, err := h.deckService.GetDecksByCourseID(courseID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, decks)
		return
	}

	if !isAdminRequest(c) {
		c.JSON(http.StatusOK, []any{})
		return
	}

	decks, err := h.deckService.GetAllDecks()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, decks)
}

func (h *DeckHandler) GetDeckByID(c *gin.Context) {
	idParam := c.Param("id")
	deckID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid deck ID format"})
		return
	}

	deck, err := h.deckService.GetDeckByID(deckID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	course, err := h.courseService.GetCourseByID(deck.CourseID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "course not found"})
		return
	}
	if !ensureCourseAccess(c, h.courseAccessService, course) {
		return
	}

	c.JSON(http.StatusOK, deck)
}

func (h *DeckHandler) GetDeckBySlug(c *gin.Context) {
	courseSlug := c.Param("course_slug")
	deckSlug := c.Param("deck_slug")

	deck, err := h.deckService.GetDeckBySlug(courseSlug, deckSlug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	course, err := h.courseService.GetCourseByID(deck.CourseID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "course not found"})
		return
	}
	if !ensureCourseAccess(c, h.courseAccessService, course) {
		return
	}

	c.JSON(http.StatusOK, deck)
}

func (h *DeckHandler) CreateDeck(c *gin.Context) {
	var req CreateDeckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	deck, err := h.deckService.CreateDeck(req.CourseID, req.Title, req.Slug, req.Description, req.Position)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, deck)
}

func (h *DeckHandler) UpdateDeck(c *gin.Context) {
	idParam := c.Param("id")
	deckID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid deck ID format"})
		return
	}

	var req UpdateDeckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	deck, err := h.deckService.UpdateDeck(deckID, req.Title, req.Slug, req.Description, req.Position)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, deck)
}

func (h *DeckHandler) DeleteDeck(c *gin.Context) {
	idParam := c.Param("id")
	deckID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid deck ID format"})
		return
	}

	if err := h.deckService.DeleteDeck(deckID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Deck deleted successfully"})
}
