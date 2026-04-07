package handlers

import (
	"english-learning/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserDeckHandler struct {
	userDeckService *services.UserDeckService
}

func NewUserDeckHandler(userDeckService *services.UserDeckService) *UserDeckHandler {
	return &UserDeckHandler{userDeckService: userDeckService}
}

// StartDeckRequest запрос на начало деки
type StartDeckRequest struct {
	UserID      string  `json:"user_id" binding:"required"`      // UUID
	DeckID      int64   `json:"deck_id" binding:"required"`      // int64
	UserCourseID *int64 `json:"user_course_id,omitempty"`        // int64, опционально
}

// UpdateUserDeckRequest запрос на обновление прогресса
type UpdateUserDeckRequest struct {
	Status            string  `json:"status" binding:"required"` // 'not_started', 'in_progress', 'completed'
	LearnedCardsCount int     `json:"learned_cards_count"`
	TotalCardsCount   int     `json:"total_cards_count"`
	ProgressPercentage float64 `json:"progress_percentage"`
}

// GetAllUserDecks возвращает все записи прогресса по декам
// @Summary Получить все записи прогресса по декам
// @Description Возвращает список всех записей прогресса пользователей по декам
// @Tags user-decks
// @Produce json
// @Success 200 {array} models.UserDeck
// @Failure 500 {object} map[string]string
// @Router /user-decks [get]
func (h *UserDeckHandler) GetAllUserDecks(c *gin.Context) {
	userDecks, err := h.userDeckService.GetAllUserDecks()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, userDecks)
}

// GetUserDeckByID возвращает прогресс по деку по ID
// @Summary Получить прогресс по деку по ID
// @Description Возвращает информацию о прогрессе пользователя по деку
// @Tags user-decks
// @Produce json
// @Param id path int true "User Deck ID"
// @Success 200 {object} models.UserDeck
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /user-decks/{id} [get]
func (h *UserDeckHandler) GetUserDeckByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user deck ID format"})
		return
	}

	userDeck, err := h.userDeckService.GetUserDeckByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, userDeck)
}

// GetUserDecksByUserID возвращает все деки пользователя
// @Summary Получить деки пользователя
// @Description Возвращает список всех дек конкретного пользователя
// @Tags user-decks
// @Produce json
// @Param user_id path string true "User ID (UUID)"
// @Success 200 {array} models.UserDeck
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /user-decks/user/{user_id} [get]
func (h *UserDeckHandler) GetUserDecksByUserID(c *gin.Context) {
	userIDParam := c.Param("user_id")
	userID, err := uuid.Parse(userIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	userDecks, err := h.userDeckService.GetUserDecksByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, userDecks)
}

func (h *UserDeckHandler) GetActiveUserDeck(c *gin.Context) {
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

	userDeck, err := h.userDeckService.GetActiveUserDeck(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, userDeck)
}

func (h *UserDeckHandler) ActivateDeck(c *gin.Context) {
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

	deckID, err := strconv.ParseInt(c.Param("deck_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid deck ID format"})
		return
	}

	userDeck, err := h.userDeckService.ActivateDeck(userID, deckID)
	if err != nil {
		if services.IsCourseProgressLockError(err) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, userDeck)
}

// StartDeck создает новую запись о начале деки
// @Summary Начать деку
// @Description Создает новую запись о начале прохождения деки пользователем
// @Tags user-decks
// @Accept json
// @Produce json
// @Param body body StartDeckRequest true "Данные для начала деки"
// @Success 201 {object} models.UserDeck
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /user-decks [post]
func (h *UserDeckHandler) StartDeck(c *gin.Context) {
	var req StartDeckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	userDeck, err := h.userDeckService.StartDeck(userID, req.DeckID, req.UserCourseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, userDeck)
}

// UpdateUserDeck обновляет прогресс пользователя по деку
// @Summary Обновить прогресс по деку
// @Description Обновляет информацию о прогрессе пользователя по деку
// @Tags user-decks
// @Accept json
// @Produce json
// @Param id path int true "User Deck ID"
// @Param body body UpdateUserDeckRequest true "Данные для обновления"
// @Success 200 {object} models.UserDeck
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /user-decks/{id} [put]
func (h *UserDeckHandler) UpdateUserDeck(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user deck ID format"})
		return
	}

	var req UpdateUserDeckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userDeck, err := h.userDeckService.UpdateUserDeck(id, req.Status, req.LearnedCardsCount, req.TotalCardsCount, req.ProgressPercentage)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, userDeck)
}

// DeleteUserDeck удаляет запись о прогрессе по деку
// @Summary Удалить прогресс по деку
// @Description Удаляет запись о прогрессе пользователя по деку
// @Tags user-decks
// @Param id path int true "User Deck ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /user-decks/{id} [delete]
func (h *UserDeckHandler) DeleteUserDeck(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user deck ID format"})
		return
	}

	if err := h.userDeckService.DeleteUserDeck(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
