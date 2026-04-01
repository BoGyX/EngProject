package handlers

import (
	"english-learning/internal/services"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserCardHandler struct {
	userCardService *services.UserCardService
}

func NewUserCardHandler(userCardService *services.UserCardService) *UserCardHandler {
	return &UserCardHandler{userCardService: userCardService}
}

// CreateUserCardRequest запрос на создание записи о прогрессе по карточке
type CreateUserCardRequest struct {
	UserID             string `json:"user_id" binding:"required"` // UUID
	CardID             int64  `json:"card_id" binding:"required"` // int64
	UserDeckID         *int64 `json:"user_deck_id,omitempty"`     // int64, опционально
	Status             string `json:"status"`                     // 'new', 'learning', 'learned'
	CorrectCount       int    `json:"correct_count"`
	WrongCount         int    `json:"wrong_count"`
	ModeView           bool   `json:"mode_view"`
	ModeChoice         bool   `json:"mode_choice"`
	ModeWithPhoto      bool   `json:"mode_with_photo"`
	ModeWithoutPhoto   bool   `json:"mode_without_photo"`
	ModeRussian        bool   `json:"mode_russian"`
	ModeConstructor    bool   `json:"mode_constructor"`
	CurrentMode        string `json:"current_mode"`
	ProgressPercentage int    `json:"progress_percentage"`
}

// UpdateUserCardRequest запрос на обновление прогресса
type UpdateUserCardRequest struct {
	Status             string  `json:"status" binding:"required"` // 'new', 'learning', 'learned'
	CorrectCount       int     `json:"correct_count"`
	WrongCount         int     `json:"wrong_count"`
	LastSeen           *string `json:"last_seen,omitempty"`   // ISO 8601 timestamp
	NextReview         *string `json:"next_review,omitempty"` // ISO 8601 timestamp
	ModeView           *bool   `json:"mode_view,omitempty"`
	ModeChoice         *bool   `json:"mode_choice,omitempty"`
	ModeWithPhoto      *bool   `json:"mode_with_photo,omitempty"`
	ModeWithoutPhoto   *bool   `json:"mode_without_photo,omitempty"`
	ModeRussian        *bool   `json:"mode_russian,omitempty"`
	ModeConstructor    *bool   `json:"mode_constructor,omitempty"`
	CurrentMode        *string `json:"current_mode,omitempty"`
	ProgressPercentage *int    `json:"progress_percentage,omitempty"`
}

// GetAllUserCards возвращает все записи прогресса по карточкам
// @Summary Получить все записи прогресса по карточкам
// @Description Возвращает список всех записей прогресса пользователей по карточкам
// @Tags user-cards
// @Produce json
// @Success 200 {array} models.UserCard
// @Failure 500 {object} map[string]string
// @Router /user-cards [get]
func (h *UserCardHandler) GetAllUserCards(c *gin.Context) {
	userCards, err := h.userCardService.GetAllUserCards()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, userCards)
}

// GetUserCardByID возвращает прогресс по карточке по ID
// @Summary Получить прогресс по карточке по ID
// @Description Возвращает информацию о прогрессе пользователя по карточке
// @Tags user-cards
// @Produce json
// @Param id path int true "User Card ID"
// @Success 200 {object} models.UserCard
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /user-cards/{id} [get]
func (h *UserCardHandler) GetUserCardByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user card ID format"})
		return
	}

	userCard, err := h.userCardService.GetUserCardByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, userCard)
}

// GetUserCardsByUserID возвращает все карточки пользователя
// @Summary Получить карточки пользователя
// @Description Возвращает список всех карточек конкретного пользователя
// @Tags user-cards
// @Produce json
// @Param user_id path string true "User ID (UUID)"
// @Success 200 {array} models.UserCard
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /user-cards/user/{user_id} [get]
func (h *UserCardHandler) GetUserCardsByUserID(c *gin.Context) {
	userIDParam := c.Param("user_id")
	userID, err := uuid.Parse(userIDParam)
	if err != nil {
		log.Printf("❌ Invalid user ID format: %s, error: %v", userIDParam, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	log.Printf("📋 Getting user cards for user: %s", userID.String())
	userCards, err := h.userCardService.GetUserCardsByUserID(userID)
	if err != nil {
		log.Printf("❌ Error getting user cards: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("✅ Found %d user cards", len(userCards))
	c.JSON(http.StatusOK, userCards)
}

// CreateUserCard создает новую запись о прогрессе по карточке
// @Summary Создать запись о прогрессе по карточке
// @Description Создает новую запись о прогрессе пользователя по карточке
// @Tags user-cards
// @Accept json
// @Produce json
// @Param body body CreateUserCardRequest true "Данные для создания записи"
// @Success 201 {object} models.UserCard
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /user-cards [post]
func (h *UserCardHandler) CreateUserCard(c *gin.Context) {
	var req CreateUserCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Логируем полученные данные
	log.Printf("Creating user card: user_id=%s, card_id=%d, status=%s, modes: view=%v, choice=%v, photo=%v, no_photo=%v, russian=%v, constructor=%v",
		req.UserID, req.CardID, req.Status, req.ModeView, req.ModeChoice, req.ModeWithPhoto, req.ModeWithoutPhoto, req.ModeRussian, req.ModeConstructor)

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	// Устанавливаем статус по умолчанию, если не указан
	status := req.Status
	if status == "" {
		status = "new"
	}

	currentMode := req.CurrentMode
	if currentMode == "" {
		currentMode = "view"
	}

	userCard, err := h.userCardService.CreateUserCardWithModes(
		userID, req.CardID, req.UserDeckID,
		status, req.CorrectCount, req.WrongCount,
		req.ModeView, req.ModeChoice, req.ModeWithPhoto, req.ModeWithoutPhoto, req.ModeRussian, req.ModeConstructor,
		currentMode, req.ProgressPercentage,
	)
	if err != nil {
		log.Printf("Error creating user card: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("User card created successfully: id=%d", userCard.ID)
	c.JSON(http.StatusCreated, userCard)
}

// UpdateUserCard обновляет прогресс пользователя по карточке
// @Summary Обновить прогресс по карточке
// @Description Обновляет информацию о прогрессе пользователя по карточке
// @Tags user-cards
// @Accept json
// @Produce json
// @Param id path int true "User Card ID"
// @Param body body UpdateUserCardRequest true "Данные для обновления"
// @Success 200 {object} models.UserCard
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /user-cards/{id} [put]
func (h *UserCardHandler) UpdateUserCard(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user card ID format"})
		return
	}

	var req UpdateUserCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var lastSeen, nextReview *time.Time
	if req.LastSeen != nil {
		t, err := time.Parse(time.RFC3339, *req.LastSeen)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid last_seen format, use RFC3339"})
			return
		}
		lastSeen = &t
	}
	if req.NextReview != nil {
		t, err := time.Parse(time.RFC3339, *req.NextReview)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid next_review format, use RFC3339"})
			return
		}
		nextReview = &t
	}

	userCard, err := h.userCardService.UpdateUserCardWithModes(
		id,
		req.Status,
		req.CorrectCount,
		req.WrongCount,
		lastSeen,
		nextReview,
		req.ModeView,
		req.ModeChoice,
		req.ModeWithPhoto,
		req.ModeWithoutPhoto,
		req.ModeRussian,
		req.ModeConstructor,
		req.CurrentMode,
		req.ProgressPercentage,
	)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, userCard)
}

// DeleteUserCard удаляет запись о прогрессе по карточке
// @Summary Удалить прогресс по карточке
// @Description Удаляет запись о прогрессе пользователя по карточке
// @Tags user-cards
// @Param id path int true "User Card ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /user-cards/{id} [delete]
func (h *UserCardHandler) DeleteUserCard(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user card ID format"})
		return
	}

	if err := h.userCardService.DeleteUserCard(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
