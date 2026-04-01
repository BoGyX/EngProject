package handlers

import (
	"english-learning/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PersonalVocabularyHandler struct {
	vocabService *services.PersonalVocabularyService
}

func NewPersonalVocabularyHandler(vocabService *services.PersonalVocabularyService) *PersonalVocabularyHandler {
	return &PersonalVocabularyHandler{vocabService: vocabService}
}

// CreatePersonalVocabularyRequest запрос на создание слова в личном словаре
type CreatePersonalVocabularyRequest struct {
	UserID      string   `json:"user_id" binding:"required"` // UUID в виде строки
	Word        string   `json:"word" binding:"required"`
	Translation string   `json:"translation" binding:"required"`
	Phonetic    *string  `json:"phonetic,omitempty"`
	AudioURL    *string  `json:"audio_url,omitempty"`
	Example     *string  `json:"example,omitempty"`
	Notes       *string  `json:"notes,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Status      string   `json:"status"` // 'new', 'learning', 'learned'
}

// UpdatePersonalVocabularyRequest запрос на обновление слова
type UpdatePersonalVocabularyRequest struct {
	Word        string   `json:"word" binding:"required"`
	Translation string   `json:"translation" binding:"required"`
	Phonetic    *string  `json:"phonetic,omitempty"`
	AudioURL    *string  `json:"audio_url,omitempty"`
	Example     *string  `json:"example,omitempty"`
	Notes       *string  `json:"notes,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Status      string   `json:"status"` // 'new', 'learning', 'learned'
}

// UpdateStatsRequest запрос на обновление статистики
type UpdateStatsRequest struct {
	IsCorrect bool `json:"is_correct" binding:"required"`
}

// GetAllPersonalVocabulary возвращает список всех слов в личном словаре или для конкретного пользователя
// @Summary Получить все слова личного словаря
// @Description Возвращает список всех слов в личном словаре или слова для конкретного пользователя (если указан user_id)
// @Tags personal-vocabulary
// @Produce json
// @Param user_id query string false "User ID (UUID) для фильтрации слов"
// @Success 200 {array} models.PersonalVocabulary
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /vocabulary [get]
func (h *PersonalVocabularyHandler) GetAllPersonalVocabulary(c *gin.Context) {
	userIDParam := c.Query("user_id")
	
	// Если указан user_id, возвращаем слова для этого пользователя
	if userIDParam != "" {
		userID, err := uuid.Parse(userIDParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_id format"})
			return
		}
		
		vocabularies, err := h.vocabService.GetPersonalVocabularyByUserID(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, vocabularies)
		return
	}
	
	// Иначе возвращаем все слова
	vocabularies, err := h.vocabService.GetAllPersonalVocabulary()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, vocabularies)
}

// GetPersonalVocabularyByID возвращает слово по ID
// @Summary Получить слово по ID
// @Description Возвращает информацию о слове в личном словаре по его ID
// @Tags personal-vocabulary
// @Produce json
// @Param id path int true "Vocabulary ID"
// @Success 200 {object} models.PersonalVocabulary
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /vocabulary/{id} [get]
func (h *PersonalVocabularyHandler) GetPersonalVocabularyByID(c *gin.Context) {
	idParam := c.Param("id")
	vocabID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid vocabulary ID format"})
		return
	}

	vocab, err := h.vocabService.GetPersonalVocabularyByID(vocabID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, vocab)
}

// GetPersonalVocabularyForReview возвращает слова для повторения
// @Summary Получить слова для повторения
// @Description Возвращает слова, которые нужно повторить (next_review <= now)
// @Tags personal-vocabulary
// @Produce json
// @Param user_id query string true "User ID (UUID)"
// @Success 200 {array} models.PersonalVocabulary
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /vocabulary/review [get]
func (h *PersonalVocabularyHandler) GetPersonalVocabularyForReview(c *gin.Context) {
	userIDParam := c.Query("user_id")
	if userIDParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	userID, err := uuid.Parse(userIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_id format"})
		return
	}

	vocabularies, err := h.vocabService.GetPersonalVocabularyForReview(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, vocabularies)
}

// CreatePersonalVocabulary создает новое слово в личном словаре
// @Summary Создать слово в личном словаре
// @Description Создает новое слово в личном словаре пользователя
// @Tags personal-vocabulary
// @Accept json
// @Produce json
// @Param vocabulary body CreatePersonalVocabularyRequest true "Данные для создания слова"
// @Success 201 {object} models.PersonalVocabulary
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /vocabulary [post]
func (h *PersonalVocabularyHandler) CreatePersonalVocabulary(c *gin.Context) {
	var req CreatePersonalVocabularyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_id UUID format"})
		return
	}

	// Устанавливаем статус по умолчанию, если не указан
	status := req.Status
	if status == "" {
		status = "new"
	}

	vocab, err := h.vocabService.CreatePersonalVocabulary(userID, req.Word, req.Translation, req.Phonetic, req.AudioURL, req.Example, req.Notes, req.Tags, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, vocab)
}

// UpdatePersonalVocabulary обновляет слово в личном словаре
// @Summary Обновить слово в личном словаре
// @Description Обновляет информацию о слове в личном словаре
// @Tags personal-vocabulary
// @Accept json
// @Produce json
// @Param id path int true "Vocabulary ID"
// @Param vocabulary body UpdatePersonalVocabularyRequest true "Данные для обновления слова"
// @Success 200 {object} models.PersonalVocabulary
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /vocabulary/{id} [put]
func (h *PersonalVocabularyHandler) UpdatePersonalVocabulary(c *gin.Context) {
	idParam := c.Param("id")
	vocabID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid vocabulary ID format"})
		return
	}

	var req UpdatePersonalVocabularyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Устанавливаем статус по умолчанию, если не указан
	status := req.Status
	if status == "" {
		status = "new"
	}

	vocab, err := h.vocabService.UpdatePersonalVocabulary(vocabID, req.Word, req.Translation, req.Phonetic, req.AudioURL, req.Example, req.Notes, req.Tags, status)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, vocab)
}

// UpdatePersonalVocabularyStats обновляет статистику изучения слова
// @Summary Обновить статистику слова
// @Description Обновляет счетчики правильных/неправильных ответов для слова
// @Tags personal-vocabulary
// @Accept json
// @Produce json
// @Param id path int true "Vocabulary ID"
// @Param stats body UpdateStatsRequest true "Данные для обновления статистики"
// @Success 200 {object} models.PersonalVocabulary
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /vocabulary/{id}/stats [put]
func (h *PersonalVocabularyHandler) UpdatePersonalVocabularyStats(c *gin.Context) {
	idParam := c.Param("id")
	vocabID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid vocabulary ID format"})
		return
	}

	var req UpdateStatsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	vocab, err := h.vocabService.UpdatePersonalVocabularyStats(vocabID, req.IsCorrect)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, vocab)
}

// DeletePersonalVocabulary удаляет слово из личного словаря
// @Summary Удалить слово из личного словаря
// @Description Удаляет слово из личного словаря пользователя
// @Tags personal-vocabulary
// @Produce json
// @Param id path int true "Vocabulary ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /vocabulary/{id} [delete]
func (h *PersonalVocabularyHandler) DeletePersonalVocabulary(c *gin.Context) {
	idParam := c.Param("id")
	vocabID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid vocabulary ID format"})
		return
	}

	err = h.vocabService.DeletePersonalVocabulary(vocabID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Vocabulary word deleted successfully"})
}
