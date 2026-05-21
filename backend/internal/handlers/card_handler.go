package handlers

import (
	"english-learning/internal/models"
	"english-learning/internal/services"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CardHandler struct {
	cardService                *services.CardService
	userDeckService            *services.UserDeckService
	deckService                *services.DeckService
	courseService              *services.CourseService
	courseAccessService        *services.CourseAccessService
	personalTranslationService *services.PersonalWordTranslationService
}

func NewCardHandler(
	cardService *services.CardService,
	userDeckService *services.UserDeckService,
	deckService *services.DeckService,
	courseService *services.CourseService,
	courseAccessService *services.CourseAccessService,
	personalTranslationService *services.PersonalWordTranslationService,
) *CardHandler {
	return &CardHandler{
		cardService:                cardService,
		userDeckService:            userDeckService,
		deckService:                deckService,
		courseService:              courseService,
		courseAccessService:        courseAccessService,
		personalTranslationService: personalTranslationService,
	}
}

type CreateCardRequest struct {
	DeckID      int64   `json:"deck_id" binding:"required"`
	Position    *int    `json:"position,omitempty"`
	Word        string  `json:"word" binding:"required"`
	Translation string  `json:"translation" binding:"required"`
	Phonetic    *string `json:"phonetic,omitempty"`
	AudioURL    *string `json:"audio_url,omitempty"`
	ImageURL    *string `json:"image_url,omitempty"`
	Example     *string `json:"example,omitempty"`
	CreatedBy   *string `json:"created_by,omitempty"`
	IsCustom    bool    `json:"is_custom"`
}

type UpdateCardRequest struct {
	Position    *int    `json:"position,omitempty"`
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

func (h *CardHandler) GetAllCards(c *gin.Context) {
	deckIDParam := c.Query("deck_id")

	if deckIDParam != "" {
		deckID, err := strconv.ParseInt(deckIDParam, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid deck_id format"})
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

		cards, err := h.cardService.GetCardsByDeckID(deckID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if err := h.applyPersonalTranslationsFromContext(c, cards); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, cards)
		return
	}

	if !isAdminRequest(c) {
		c.JSON(http.StatusOK, []any{})
		return
	}

	cards, err := h.cardService.GetAllCards()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := h.applyPersonalTranslationsFromContext(c, cards); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, cards)
}

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

	deck, err := h.deckService.GetDeckByID(card.DeckID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "deck not found"})
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

	if err := h.applyPersonalTranslationsFromContext(c, card); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, card)
}

func (h *CardHandler) CreateCard(c *gin.Context) {
	var req CreateCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.ImageURL != nil {
		c.Header("X-Debug-ImageURL", *req.ImageURL)
	}

	var createdBy *uuid.UUID
	if req.CreatedBy != nil && *req.CreatedBy != "" && *req.CreatedBy != "string" {
		uuidVal, err := uuid.Parse(*req.CreatedBy)
		if err == nil {
			createdBy = &uuidVal
		}
	}

	normalizeString := func(value *string) *string {
		if value == nil || *value == "" || *value == "null" || *value == "string" {
			return nil
		}
		return value
	}

	phonetic := normalizeString(req.Phonetic)
	audioURL := normalizeString(req.AudioURL)
	imageURL := normalizeString(req.ImageURL)
	example := normalizeString(req.Example)

	if imageURL != nil {
		c.Header("X-Debug-ImageURL-Normalized", *imageURL)
	}

	card, err := h.cardService.CreateCard(req.DeckID, req.Position, req.Word, req.Translation, phonetic, audioURL, imageURL, example, createdBy, req.IsCustom)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

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

	normalizeString := func(value *string) *string {
		if value == nil || *value == "" {
			return nil
		}
		return value
	}

	card, err := h.cardService.UpdateCard(
		cardID,
		req.Position,
		req.Word,
		req.Translation,
		normalizeString(req.Phonetic),
		normalizeString(req.AudioURL),
		normalizeString(req.ImageURL),
		normalizeString(req.Example),
	)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, card)
}

func (h *CardHandler) DeleteCard(c *gin.Context) {
	idParam := c.Param("id")
	cardID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid card ID format"})
		return
	}

	if err := h.cardService.DeleteCard(cardID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Card deleted successfully"})
}

func (h *CardHandler) applyPersonalTranslationsFromContext(c *gin.Context, target any) error {
	if h.personalTranslationService == nil {
		return nil
	}

	userID, err := getRequestUserID(c)
	if err != nil || userID == nil {
		return err
	}

	switch value := target.(type) {
	case []models.Card:
		return h.applyPersonalTranslationsToCards(*userID, value)
	case *models.Card:
		return h.applyPersonalTranslationToCard(*userID, value)
	default:
		return nil
	}
}

func (h *CardHandler) applyPersonalTranslationsToCards(userID uuid.UUID, cards []models.Card) error {
	if len(cards) == 0 {
		return nil
	}

	words := make([]string, 0, len(cards))
	for _, card := range cards {
		words = append(words, card.Word)
	}

	translationMap, err := h.personalTranslationService.GetLatestTranslationMap(userID, words)
	if err != nil {
		return err
	}

	for index := range cards {
		if translation, exists := translationMap[normalizeTranslationWord(cards[index].Word)]; exists {
			cards[index].Translation = translation
		}
	}

	return nil
}

func (h *CardHandler) applyPersonalTranslationToCard(userID uuid.UUID, card *models.Card) error {
	if card == nil {
		return nil
	}

	translationMap, err := h.personalTranslationService.GetLatestTranslationMap(userID, []string{card.Word})
	if err != nil {
		return err
	}

	if translation, exists := translationMap[normalizeTranslationWord(card.Word)]; exists {
		card.Translation = translation
	}

	return nil
}

func normalizeTranslationWord(word string) string {
	return strings.ToLower(strings.TrimSpace(word))
}
