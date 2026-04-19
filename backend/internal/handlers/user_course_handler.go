package handlers

import (
	"english-learning/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserCourseHandler struct {
	userCourseService *services.UserCourseService
}

func NewUserCourseHandler(userCourseService *services.UserCourseService) *UserCourseHandler {
	return &UserCourseHandler{userCourseService: userCourseService}
}

// StartCourseRequest запрос на начало курса
type StartCourseRequest struct {
	UserID   string `json:"user_id" binding:"required"`   // UUID
	CourseID int64  `json:"course_id" binding:"required"` // int64
}

// UpdateUserCourseRequest запрос на обновление прогресса
type UpdateUserCourseRequest struct {
	CompletedDecksCount int     `json:"completed_decks_count"`
	TotalDecksCount     int     `json:"total_decks_count"`
	ProgressPercentage  float64 `json:"progress_percentage"`
}

// GetAllUserCourses возвращает все записи прогресса по курсам
// @Summary Получить все записи прогресса по курсам
// @Description Возвращает список всех записей прогресса пользователей по курсам
// @Tags user-courses
// @Produce json
// @Success 200 {array} models.UserCourse
// @Failure 500 {object} map[string]string
// @Router /user-courses [get]
func (h *UserCourseHandler) GetAllUserCourses(c *gin.Context) {
	userCourses, err := h.userCourseService.GetAllUserCourses()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, userCourses)
}

// GetUserCourseByID возвращает прогресс по курсу по ID
// @Summary Получить прогресс по курсу по ID
// @Description Возвращает информацию о прогрессе пользователя по курсу
// @Tags user-courses
// @Produce json
// @Param id path int true "User Course ID"
// @Success 200 {object} models.UserCourse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /user-courses/{id} [get]
func (h *UserCourseHandler) GetUserCourseByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user course ID format"})
		return
	}

	userCourse, err := h.userCourseService.GetUserCourseByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, userCourse)
}

// GetUserCoursesByUserID возвращает все курсы пользователя
// @Summary Получить курсы пользователя
// @Description Возвращает список всех курсов конкретного пользователя
// @Tags user-courses
// @Produce json
// @Param user_id path string true "User ID (UUID)"
// @Success 200 {array} models.UserCourse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /user-courses/user/{user_id} [get]
func (h *UserCourseHandler) GetUserCoursesByUserID(c *gin.Context) {
	userIDParam := c.Param("user_id")
	userID, err := uuid.Parse(userIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	userCourses, err := h.userCourseService.GetUserCoursesByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, userCourses)
}

func (h *UserCourseHandler) GetActiveUserCourse(c *gin.Context) {
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

	userCourse, err := h.userCourseService.GetActiveUserCourse(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, userCourse)
}

func (h *UserCourseHandler) ActivateCourse(c *gin.Context) {
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

	courseID, err := strconv.ParseInt(c.Param("course_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID format"})
		return
	}

	userCourse, err := h.userCourseService.ActivateCourse(userID, courseID)
	if err != nil {
		if err.Error() == "course access denied" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if services.IsCourseProgressLockError(err) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, userCourse)
}

// StartCourse создает новую запись о начале курса
// @Summary Начать курс
// @Description Создает новую запись о начале прохождения курса пользователем
// @Tags user-courses
// @Accept json
// @Produce json
// @Param body body StartCourseRequest true "Данные для начала курса"
// @Success 201 {object} models.UserCourse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /user-courses [post]
func (h *UserCourseHandler) StartCourse(c *gin.Context) {
	var req StartCourseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	userCourse, err := h.userCourseService.StartCourse(userID, req.CourseID)
	if err != nil {
		if err.Error() == "course access denied" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if services.IsCourseProgressLockError(err) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, userCourse)
}

// UpdateUserCourse обновляет прогресс пользователя по курсу
// @Summary Обновить прогресс по курсу
// @Description Обновляет информацию о прогрессе пользователя по курсу
// @Tags user-courses
// @Accept json
// @Produce json
// @Param id path int true "User Course ID"
// @Param body body UpdateUserCourseRequest true "Данные для обновления"
// @Success 200 {object} models.UserCourse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /user-courses/{id} [put]
func (h *UserCourseHandler) UpdateUserCourse(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user course ID format"})
		return
	}

	var req UpdateUserCourseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userCourse, err := h.userCourseService.UpdateUserCourse(id, req.CompletedDecksCount, req.TotalDecksCount, req.ProgressPercentage)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, userCourse)
}

// DeleteUserCourse удаляет запись о прогрессе по курсу
// @Summary Удалить прогресс по курсу
// @Description Удаляет запись о прогрессе пользователя по курсу
// @Tags user-courses
// @Param id path int true "User Course ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /user-courses/{id} [delete]
func (h *UserCourseHandler) DeleteUserCourse(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user course ID format"})
		return
	}

	if err := h.userCourseService.DeleteUserCourse(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
