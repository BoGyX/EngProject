package handlers

import (
	"english-learning/internal/models"
	"english-learning/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CardHandler struct {
	cardService     *services.CardService
	userDeckService *services.UserDeckService
}

func NewCardHandler(cardService *services.CardService, userDeckService *services.UserDeckService) *CardHandler {
	return &CardHandler{cardService: cardService, userDeckService: userDeckService}
}

// CreateCardRequest запрос на создание card
type CreateCardRequest struct {
	DeckID      int64   `json:"deck_id" binding:"required"`
	Word        string  `json:"word" binding:"required"`
	Translation string  `json:"translation" binding:"required"`
	Phonetic    *string `json:"phonetic,omitempty"`
	AudioURL    *string `json:"audio_url,omitempty"`
	ImageURL    *string `json:"image_url,omitempty"`
	Example     *string `json:"example,omitempty"`
	CreatedBy   *string `json:"created_by,omitempty"` // UUID в виде строки
	IsCustom    bool    `json:"is_custom"`
}

// UpdateCardRequest запрос на обновление card
type UpdateCardRequest struct {
	Word        string  `json:"word" binding:"required"`
	Translation string  `json:"translation" binding:"required"`
	Phonetic    *string `json:"phonetic,omitempty"`
	AudioURL    *string `json:"audio_url,omitempty"`
	ImageURL    *string `json:"image_url,omitempty"`
	Example     *string `json:"example,omitempty"`
}

type CreateCustomCardRequest struct {
	DeckID      *int64  `json:"deck_id,omitempty"`
	Word        string  `json:"word" binding:"required"`
	Translation string  `json:"translation" binding:"required"`
	Phonetic    *string `json:"phonetic,omitempty"`
	AudioURL    *string `json:"audio_url,omitempty"`
	ImageURL    *string `json:"image_url,omitempty"`
	Example     *string `json:"example,omitempty"`
}

// GetAllCards возвращает список всех cards или cards для конкретного deck
// @Summary Получить все cards
// @Description Возвращает список всех cards в системе или cards для конкретного deck (если указан deck_id)
// @Tags cards
// @Produce json
// @Param deck_id query int false "Deck ID для фильтрации cards"
// @Success 200 {array} models.Card
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /cards [get]
func (h *CardHandler) GetAllCards(c *gin.Context) {
	deckIDParam := c.Query("deck_id")

	// Если указан deck_id, возвращаем cards для этого deck
	if deckIDParam != "" {
		deckID, err := strconv.ParseInt(deckIDParam, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid deck_id format"})
			return
		}

		cards, err := h.cardService.GetCardsByDeckID(deckID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, cards)
		return
	}

	// Иначе возвращаем все cards
	cards, err := h.cardService.GetAllCards()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, cards)
}

// GetCardByID возвращает card по ID
// @Summary Получить card по ID
// @Description Возвращает информацию о card по его ID
// @Tags cards
// @Produce json
// @Param id path int true "Card ID"
// @Success 200 {object} models.Card
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /cards/{id} [get]
func (h *CardHandler) GetCardByID(c *gin.Context) {
	idParam := c.Param("id")
	cardID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid card ID format"})
		return
	}

	card, err := h.cardService.GetCardByID(cardID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, card)
}

// CreateCard создает новый card
// @Summary Создать card
// @Description Создает новый card в системе
// @Tags cards
// @Accept json
// @Produce json
// @Param card body CreateCardRequest true "Данные для создания card"
// @Success 201 {object} models.Card
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /cards [post]
func (h *CardHandler) CreateCard(c *gin.Context) {
	var req CreateCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Логирование для отладки
	if req.ImageURL != nil {
		c.Header("X-Debug-ImageURL", *req.ImageURL)
	}

	var createdBy *uuid.UUID
	if req.CreatedBy != nil && *req.CreatedBy != "" && *req.CreatedBy != "string" {
		uuidVal, err := uuid.Parse(*req.CreatedBy)
		if err != nil {
			// Если не удалось распарсить UUID, просто игнорируем поле (не обязательное)
			createdBy = nil
		} else {
			createdBy = &uuidVal
		}
	}

	// Нормализуем пустые строки в nil для опциональных полей
	normalizeString := func(s *string) *string {
		if s == nil || *s == "" || *s == "null" || *s == "string" {
			return nil
		}
		return s
	}

	phonetic := normalizeString(req.Phonetic)
	audioURL := normalizeString(req.AudioURL)
	imageURL := normalizeString(req.ImageURL)
	example := normalizeString(req.Example)

	// Логирование для отладки
	if imageURL != nil {
		c.Header("X-Debug-ImageURL-Normalized", *imageURL)
	}

	card, err := h.cardService.CreateCard(req.DeckID, req.Word, req.Translation, phonetic, audioURL, imageURL, example, createdBy, req.IsCustom)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Логирование результата
	if card.ImageURL != nil {
		c.Header("X-Debug-ImageURL-Saved", *card.ImageURL)
	}

	c.JSON(http.StatusCreated, card)
}

func (h *CardHandler) CreateCustomCard(c *gin.Context) {
	var req CreateCustomCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User is not authenticated"})
		return
	}

	userID, err := uuid.Parse(userIDValue.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	normalizeString := func(value *string) *string {
		if value == nil || *value == "" || *value == "null" || *value == "string" {
			return nil
		}
		return value
	}

	var activeUserDeck *models.UserDeck
	if req.DeckID != nil {
		activeUserDeck, err = h.userDeckService.ActivateDeck(userID, *req.DeckID)
	} else {
		activeUserDeck, err = h.userDeckService.GetActiveUserDeck(userID)
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	card, err := h.cardService.CreateCustomCard(
		activeUserDeck.DeckID,
		userID,
		req.Word,
		req.Translation,
		normalizeString(req.Phonetic),
		normalizeString(req.AudioURL),
		normalizeString(req.ImageURL),
		normalizeString(req.Example),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"card":      card,
		"user_deck": activeUserDeck,
	})
}

// UpdateCard обновляет card
// @Summary Обновить card
// @Description Обновляет информацию о card
// @Tags cards
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Card ID"
// @Param card body UpdateCardRequest true "Данные для обновления card"
// @Success 200 {object} models.Card
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /admin/cards/{id} [put]
func (h *CardHandler) UpdateCard(c *gin.Context) {
	idParam := c.Param("id")
	cardID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid card ID format"})
		return
	}

	var req UpdateCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Нормализуем пустые строки в nil для опциональных полей
	normalizeString := func(s *string) *string {
		if s == nil || *s == "" {
			return nil
		}
		return s
	}

	phonetic := normalizeString(req.Phonetic)
	audioURL := normalizeString(req.AudioURL)
	imageURL := normalizeString(req.ImageURL)
	example := normalizeString(req.Example)

	card, err := h.cardService.UpdateCard(cardID, req.Word, req.Translation, phonetic, audioURL, imageURL, example)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, card)
}

// DeleteCard удаляет card
// @Summary Удалить card
// @Description Удаляет card из системы
// @Tags cards
// @Produce json
// @Security BearerAuth
// @Param id path int true "Card ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /admin/cards/{id} [delete]
func (h *CardHandler) DeleteCard(c *gin.Context) {
	idParam := c.Param("id")
	cardID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid card ID format"})
		return
	}

	err = h.cardService.DeleteCard(cardID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Card deleted successfully"})
}
