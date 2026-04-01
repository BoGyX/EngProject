package handlers

import (
	"english-learning/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type DeckHandler struct {
	deckService *services.DeckService
}

func NewDeckHandler(deckService *services.DeckService) *DeckHandler {
	return &DeckHandler{deckService: deckService}
}

// CreateDeckRequest запрос на создание deck
type CreateDeckRequest struct {
	CourseID    int64   `json:"course_id" binding:"required"`
	Title       string  `json:"title" binding:"required"`
	Slug        string  `json:"slug"`
	Description *string `json:"description,omitempty"`
	Position    int     `json:"position"`
}

// UpdateDeckRequest запрос на обновление deck
type UpdateDeckRequest struct {
	Title       string  `json:"title" binding:"required"`
	Slug        string  `json:"slug"`
	Description *string `json:"description,omitempty"`
	Position    int     `json:"position"`
}

// GetAllDecks возвращает список всех decks или decks для конкретного курса
// @Summary Получить все decks
// @Description Возвращает список всех decks в системе или decks для конкретного курса (если указан course_id)
// @Tags decks
// @Produce json
// @Param course_id query int false "Course ID для фильтрации decks"
// @Success 200 {array} models.Deck
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /decks [get]
func (h *DeckHandler) GetAllDecks(c *gin.Context) {
	courseIDParam := c.Query("course_id")
	
	// Если указан course_id, возвращаем decks для этого курса
	if courseIDParam != "" {
		courseID, err := strconv.ParseInt(courseIDParam, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course_id format"})
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
	
	// Иначе возвращаем все decks
	decks, err := h.deckService.GetAllDecks()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, decks)
}

// GetDeckByID возвращает deck по ID
// @Summary Получить deck по ID
// @Description Возвращает информацию о deck по его ID
// @Tags decks
// @Produce json
// @Param id path int true "Deck ID"
// @Success 200 {object} models.Deck
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /decks/{id} [get]
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

	c.JSON(http.StatusOK, deck)
}


// CreateDeck создает новый deck
// @Summary Создать deck
// @Description Создает новый deck в системе
// @Tags decks
// @Accept json
// @Produce json
// @Param deck body CreateDeckRequest true "Данные для создания deck"
// @Success 201 {object} models.Deck
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /decks [post]
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

// UpdateDeck обновляет deck
// @Summary Обновить deck
// @Description Обновляет информацию о deck
// @Tags decks
// @Accept json
// @Produce json
// @Param id path int true "Deck ID"
// @Param deck body UpdateDeckRequest true "Данные для обновления deck"
// @Success 200 {object} models.Deck
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /decks/{id} [put]
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

// DeleteDeck удаляет deck
// @Summary Удалить deck
// @Description Удаляет deck из системы
// @Tags decks
// @Produce json
// @Param id path int true "Deck ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /decks/{id} [delete]
func (h *DeckHandler) DeleteDeck(c *gin.Context) {
	idParam := c.Param("id")
	deckID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid deck ID format"})
		return
	}

	err = h.deckService.DeleteDeck(deckID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Deck deleted successfully"})
}
