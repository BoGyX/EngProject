package handlers

import (
	"english-learning/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ReadingTextHandler struct {
	textService *services.ReadingTextService
}

func NewReadingTextHandler(textService *services.ReadingTextService) *ReadingTextHandler {
	return &ReadingTextHandler{textService: textService}
}

type CreateReadingTextRequest struct {
	UserID   string `json:"user_id" binding:"required"`
	CourseID int64  `json:"course_id" binding:"required"`
	Title    string `json:"title" binding:"required"`
	Content  string `json:"content" binding:"required"`
	AudioURL string `json:"audio_url"`
}

type UpdateReadingTextAudioRequest struct {
	UserID   string `json:"user_id" binding:"required"`
	AudioURL string `json:"audio_url" binding:"required"`
}

// GetAllReadingTexts возвращает все тексты пользователя
func (h *ReadingTextHandler) GetAllReadingTexts(c *gin.Context) {
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

	texts, err := h.textService.GetAllByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, texts)
}

// GetReadingTextByID возвращает текст по ID
func (h *ReadingTextHandler) GetReadingTextByID(c *gin.Context) {
	idParam := c.Param("id")
	textID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid text ID format"})
		return
	}

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

	text, err := h.textService.GetByID(textID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, text)
}

// CreateReadingText создает новый текст
func (h *ReadingTextHandler) CreateReadingText(c *gin.Context) {
	var req CreateReadingTextRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_id format"})
		return
	}

	if req.CourseID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "course_id is required"})
		return
	}

	text, err := h.textService.Create(userID, req.CourseID, req.Title, req.Content, req.AudioURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, text)
}

// DeleteReadingText удаляет текст
func (h *ReadingTextHandler) DeleteReadingText(c *gin.Context) {
	idParam := c.Param("id")
	textID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid text ID format"})
		return
	}

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

	err = h.textService.Delete(textID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Text deleted successfully"})
}

func (h *ReadingTextHandler) UpdateReadingTextAudio(c *gin.Context) {
	idParam := c.Param("id")
	textID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid text ID format"})
		return
	}

	var req UpdateReadingTextAudioRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_id format"})
		return
	}

	text, err := h.textService.UpdateAudio(textID, userID, req.AudioURL)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, text)
}

func (h *ReadingTextHandler) GetAdminPodcasts(c *gin.Context) {
	courseIDParam := c.Query("course_id")
	var courseID int64

	if courseIDParam != "" {
		parsedCourseID, err := strconv.ParseInt(courseIDParam, 10, 64)
		if err != nil || parsedCourseID <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course_id format"})
			return
		}
		courseID = parsedCourseID
	}

	podcasts, err := h.textService.GetPodcasts(courseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, podcasts)
}
