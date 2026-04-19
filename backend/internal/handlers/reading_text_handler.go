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
	CourseID int64  `json:"course_id" binding:"required"`
	Title    string `json:"title" binding:"required"`
	Content  string `json:"content" binding:"required"`
	AudioURL string `json:"audio_url"`
}

type UpdateReadingTextAudioRequest struct {
	AudioURL string `json:"audio_url" binding:"required"`
}

func (h *ReadingTextHandler) GetAllReadingTexts(c *gin.Context) {
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

	texts, err := h.textService.GetAllAccessible(userID, isAdminRequest(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, texts)
}

func (h *ReadingTextHandler) GetReadingTextByID(c *gin.Context) {
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

	textID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid text ID format"})
		return
	}

	text, err := h.textService.GetAccessibleByID(textID, userID, isAdminRequest(c))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, text)
}

func (h *ReadingTextHandler) CreateReadingText(c *gin.Context) {
	if !isAdminRequest(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can create reader texts"})
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

	var req CreateReadingTextRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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

func (h *ReadingTextHandler) DeleteReadingText(c *gin.Context) {
	if !isAdminRequest(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can delete reader texts"})
		return
	}

	textID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid text ID format"})
		return
	}

	if err := h.textService.Delete(textID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Text deleted successfully"})
}

func (h *ReadingTextHandler) UpdateReadingTextAudio(c *gin.Context) {
	if !isAdminRequest(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can update reader audio"})
		return
	}

	textID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid text ID format"})
		return
	}

	var req UpdateReadingTextAudioRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	text, err := h.textService.UpdateAudio(textID, req.AudioURL)
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
