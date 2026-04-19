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
	courseService       *services.CourseService
	deckService         *services.DeckService
	cardService         *services.CardService
	dictionaryService   *services.DictionaryService
	courseAccessService *services.CourseAccessService
}

func NewCourseHandler(
	courseService *services.CourseService,
	deckService *services.DeckService,
	cardService *services.CardService,
	dictionaryService *services.DictionaryService,
	courseAccessService *services.CourseAccessService,
) *CourseHandler {
	return &CourseHandler{
		courseService:       courseService,
		deckService:         deckService,
		cardService:         cardService,
		dictionaryService:   dictionaryService,
		courseAccessService: courseAccessService,
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

type CreateCourseRequest struct {
	Title       string  `json:"title" binding:"required"`
	Slug        string  `json:"slug"`
	Description *string `json:"description,omitempty"`
	ImageURL    *string `json:"image_url,omitempty"`
	IsPublished bool    `json:"is_published"`
	CreatedBy   *string `json:"created_by,omitempty"`
}

type UpdateCourseRequest struct {
	Title       string  `json:"title" binding:"required"`
	Slug        string  `json:"slug"`
	Description *string `json:"description,omitempty"`
	ImageURL    *string `json:"image_url,omitempty"`
	IsPublished bool    `json:"is_published"`
}

func (h *CourseHandler) GetAllCourses(c *gin.Context) {
	var courses []models.Course
	var err error

	if isAdminRequest(c) {
		courses, err = h.courseService.GetAllCourses()
	} else {
		userID, userIDErr := getRequestUserID(c)
		if userIDErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
			return
		}

		if userID == nil {
			courses = []models.Course{}
		} else {
			courses, err = h.courseAccessService.GetAccessibleCoursesByUserID(*userID, false)
		}
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, courses)
}

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

func (h *CourseHandler) DeleteCourse(c *gin.Context) {
	idParam := c.Param("id")
	courseID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID format"})
		return
	}

	if err := h.courseService.DeleteCourse(courseID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Course deleted successfully"})
}

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

	newStatus := !course.IsPublished
	course, err = h.courseService.UpdateCourse(courseID, course.Title, course.Slug, course.Description, course.ImageURL, newStatus)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, course)
}
