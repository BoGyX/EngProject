package handlers

import (
	"english-learning/internal/models"
	"english-learning/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CourseHandler struct {
	courseService     *services.CourseService
	deckService       *services.DeckService
	cardService       *services.CardService
	dictionaryService *services.DictionaryService
}

func NewCourseHandler(courseService *services.CourseService, deckService *services.DeckService, cardService *services.CardService, dictionaryService *services.DictionaryService) *CourseHandler {
	return &CourseHandler{
		courseService:     courseService,
		deckService:       deckService,
		cardService:       cardService,
		dictionaryService: dictionaryService,
	}
}

func isAdminRequest(c *gin.Context) bool {
	userRole, exists := c.Get("user_role")
	if !exists {
		return false
	}

	role, ok := userRole.(string)
	return ok && role == "admin"
}

// CreateCourseRequest запрос на создание курса
type CreateCourseRequest struct {
	Title       string  `json:"title" binding:"required"`
	Slug        string  `json:"slug"`
	Description *string `json:"description,omitempty"`
	ImageURL    *string `json:"image_url,omitempty"`
	IsPublished bool    `json:"is_published"`
	CreatedBy   *string `json:"created_by,omitempty"` // UUID в виде строки
}

// UpdateCourseRequest запрос на обновление курса
type UpdateCourseRequest struct {
	Title       string  `json:"title" binding:"required"`
	Slug        string  `json:"slug"`
	Description *string `json:"description,omitempty"`
	ImageURL    *string `json:"image_url,omitempty"`
	IsPublished bool    `json:"is_published"`
}

// GetAllCourses возвращает список всех курсов
// Для админов - все курсы, для обычных пользователей - только опубликованные
// @Summary Получить все курсы
// @Description Возвращает список всех курсов в системе. Админы видят все курсы, пользователи - только опубликованные
// @Tags courses
// @Produce json
// @Success 200 {array} models.Course
// @Failure 500 {object} map[string]string
// @Router /courses [get]
func (h *CourseHandler) GetAllCourses(c *gin.Context) {
	// Проверяем роль пользователя из контекста (если авторизован)
	userRole, exists := c.Get("user_role")

	var courses []models.Course
	var err error

	// Если пользователь авторизован и является админом, показываем все курсы
	// Иначе - только опубликованные (для неавторизованных и обычных пользователей)
	if exists && userRole == "admin" {
		courses, err = h.courseService.GetAllCourses()
	} else {
		courses, err = h.courseService.GetPublishedCourses()
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, courses)
}

// GetCourseByID возвращает курс по ID
// @Summary Получить курс по ID
// @Description Возвращает информацию о курсе по его ID
// @Tags courses
// @Produce json
// @Param id path int true "Course ID"
// @Success 200 {object} models.Course
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /courses/{id} [get]
func (h *CourseHandler) GetCourseByID(c *gin.Context) {
	idParam := c.Param("id")
	courseID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID format"})
		return
	}

	course, err := h.courseService.GetCourseByID(courseID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, course)
}

func (h *CourseHandler) GetCourseBySlug(c *gin.Context) {
	slug := c.Param("slug")

	course, err := h.courseService.GetCourseBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, course)
}

// CreateCourse создает новый курс
// @Summary Создать курс
// @Description Создает новый курс в системе
// @Tags courses
// @Accept json
// @Produce json
// @Param course body CreateCourseRequest true "Данные для создания курса"
// @Success 201 {object} models.Course
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /courses [post]
func (h *CourseHandler) CreateCourse(c *gin.Context) {
	var req CreateCourseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var createdBy *uuid.UUID
	if req.CreatedBy != nil && *req.CreatedBy != "" {
		uuidVal, err := uuid.Parse(*req.CreatedBy)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid created_by UUID format"})
			return
		}
		createdBy = &uuidVal
	}

	course, err := h.courseService.CreateCourse(req.Title, req.Slug, req.Description, req.ImageURL, req.IsPublished, createdBy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, course)
}

// UpdateCourse обновляет курс
// @Summary Обновить курс
// @Description Обновляет информацию о курсе
// @Tags courses
// @Accept json
// @Produce json
// @Param id path int true "Course ID"
// @Param course body UpdateCourseRequest true "Данные для обновления курса"
// @Success 200 {object} models.Course
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /courses/{id} [put]
func (h *CourseHandler) UpdateCourse(c *gin.Context) {
	idParam := c.Param("id")
	courseID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID format"})
		return
	}

	var req UpdateCourseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	course, err := h.courseService.UpdateCourse(courseID, req.Title, req.Slug, req.Description, req.ImageURL, req.IsPublished)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, course)
}

// DeleteCourse удаляет курс
// @Summary Удалить курс
// @Description Удаляет курс из системы
// @Tags courses
// @Produce json
// @Param id path int true "Course ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /courses/{id} [delete]
func (h *CourseHandler) DeleteCourse(c *gin.Context) {
	idParam := c.Param("id")
	courseID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID format"})
		return
	}

	err = h.courseService.DeleteCourse(courseID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Course deleted successfully"})
}

// PublishCourse переключает статус публикации курса
// @Summary Опубликовать/снять с публикации курс
// @Description Переключает статус публикации курса (опубликован/черновик)
// @Tags courses
// @Produce json
// @Param id path int true "Course ID"
// @Success 200 {object} models.Course
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /courses/{id}/publish [post]
func (h *CourseHandler) PublishCourse(c *gin.Context) {
	idParam := c.Param("id")
	courseID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID format"})
		return
	}

	course, err := h.courseService.GetCourseByID(courseID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Переключаем статус публикации
	newStatus := !course.IsPublished
	course, err = h.courseService.UpdateCourse(courseID, course.Title, course.Slug, course.Description, course.ImageURL, newStatus)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, course)
}
