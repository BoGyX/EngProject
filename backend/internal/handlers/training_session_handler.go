package handlers

import (
	"english-learning/internal/models"
	"english-learning/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TrainingSessionHandler struct {
	trainingSessionService *services.TrainingSessionService
}

func NewTrainingSessionHandler(trainingSessionService *services.TrainingSessionService) *TrainingSessionHandler {
	return &TrainingSessionHandler{trainingSessionService: trainingSessionService}
}

// StartTrainingSessionRequest запрос на начало сессии
type StartTrainingSessionRequest struct {
	UserID   *string `json:"user_id,omitempty"`   // UUID, опционально
	CourseID *int64  `json:"course_id,omitempty"` // int64, опционально
	DeckID   *int64  `json:"deck_id,omitempty"`   // int64, опционально
}

// GetAllTrainingSessions возвращает все сессии тренировок
// @Summary Получить все сессии тренировок
// @Description Возвращает список всех сессий тренировок
// @Tags training-sessions
// @Produce json
// @Success 200 {array} models.TrainingSession
// @Failure 500 {object} map[string]string
// @Router /training-sessions [get]
func (h *TrainingSessionHandler) GetAllTrainingSessions(c *gin.Context) {
	sessions, err := h.trainingSessionService.GetAllTrainingSessions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, sessions)
}

// GetTrainingSessionByID возвращает сессию тренировки по ID
// @Summary Получить сессию тренировки по ID
// @Description Возвращает информацию о сессии тренировки
// @Tags training-sessions
// @Produce json
// @Param id path int true "Training Session ID"
// @Success 200 {object} models.TrainingSession
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /training-sessions/{id} [get]
func (h *TrainingSessionHandler) GetTrainingSessionByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid training session ID format"})
		return
	}

	var session *models.TrainingSessionState
	userIDValue, hasUserID := c.Get("user_id")
	if hasUserID {
		userID, err := uuid.Parse(userIDValue.(string))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
			return
		}
		session, err = h.trainingSessionService.GetTrainingSessionByIDForUser(id, userID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
	} else {
		session, err = h.trainingSessionService.GetTrainingSessionStateByID(id, nil)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, session)
}

// GetTrainingSessionsByUserID возвращает все сессии пользователя
// @Summary Получить сессии пользователя
// @Description Возвращает список всех сессий тренировок конкретного пользователя
// @Tags training-sessions
// @Produce json
// @Param user_id path string true "User ID (UUID)"
// @Success 200 {array} models.TrainingSession
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /training-sessions/user/{user_id} [get]
func (h *TrainingSessionHandler) GetTrainingSessionsByUserID(c *gin.Context) {
	userIDParam := c.Param("user_id")
	userID, err := uuid.Parse(userIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	sessions, err := h.trainingSessionService.GetTrainingSessionsByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, sessions)
}

// StartTrainingSession создает новую сессию тренировки
// @Summary Начать сессию тренировки
// @Description Создает новую сессию тренировки
// @Tags training-sessions
// @Accept json
// @Produce json
// @Param body body StartTrainingSessionRequest true "Данные для начала сессии"
// @Success 201 {object} models.TrainingSession
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /training-sessions [post]
func (h *TrainingSessionHandler) StartTrainingSession(c *gin.Context) {
	var req StartTrainingSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var userID uuid.UUID
	userIDValue, hasUserID := c.Get("user_id")
	if hasUserID {
		parsed, err := uuid.Parse(userIDValue.(string))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
			return
		}
		userID = parsed
	} else {
		if req.UserID == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
			return
		}
		parsed, err := uuid.Parse(*req.UserID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
			return
		}
		userID = parsed
	}

	session, err := h.trainingSessionService.StartScopedTrainingSession(userID, req.CourseID, req.DeckID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, session)
}

// FinishTrainingSession завершает сессию тренировки
// @Summary Завершить сессию тренировки
// @Description Завершает сессию тренировки, устанавливая finished_at
// @Tags training-sessions
// @Param id path int true "Training Session ID"
// @Success 200 {object} models.TrainingSession
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /training-sessions/{id}/finish [put]
func (h *TrainingSessionHandler) FinishTrainingSession(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid training session ID format"})
		return
	}

	session, err := h.trainingSessionService.FinishTrainingSession(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, session)
}

// DeleteTrainingSession удаляет сессию тренировки
// @Summary Удалить сессию тренировки
// @Description Удаляет сессию тренировки
// @Tags training-sessions
// @Param id path int true "Training Session ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /training-sessions/{id} [delete]
func (h *TrainingSessionHandler) DeleteTrainingSession(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid training session ID format"})
		return
	}

	if err := h.trainingSessionService.DeleteTrainingSession(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *TrainingSessionHandler) SubmitAnswer(c *gin.Context) {
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

	sessionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid training session ID format"})
		return
	}

	var req struct {
		CardID int64  `json:"card_id" binding:"required"`
		Answer string `json:"answer"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	state, isCorrect, err := h.trainingSessionService.SubmitScopedAnswer(userID, sessionID, req.CardID, req.Answer)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"is_correct": isCorrect,
		"session":    state,
	})
}
