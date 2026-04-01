package handlers

import (
	"english-learning/internal/models"
	"english-learning/internal/services"
	"english-learning/internal/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthHandler struct {
	userService   *services.UserService
	moodleService *services.MoodleService
	jwtSecret     string
	jwtExpiry     int
	moodleEnabled bool
	moodleOnlyAuth bool
	moodleAutoCreate bool
}

func NewAuthHandler(userService *services.UserService, moodleService *services.MoodleService, jwtSecret string, jwtExpiry int, moodleEnabled, moodleOnlyAuth, moodleAutoCreate bool) *AuthHandler {
	return &AuthHandler{
		userService:      userService,
		moodleService:    moodleService,
		jwtSecret:        jwtSecret,
		jwtExpiry:        jwtExpiry,
		moodleEnabled:    moodleEnabled,
		moodleOnlyAuth:   moodleOnlyAuth,
		moodleAutoCreate: moodleAutoCreate,
	}
}

func (h *AuthHandler) rejectIfMoodleOnly(c *gin.Context, message string) bool {
	if !h.moodleOnlyAuth {
		return false
	}

	c.JSON(http.StatusForbidden, gin.H{
		"error":   "Only Moodle authentication is enabled",
		"message": message,
	})
	return true
}

// RegisterRequest запрос на регистрацию
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name"`
}

// LoginRequest запрос на вход
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse ответ с токенами
type AuthResponse struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	User         interface{} `json:"user"`
}

// Register регистрация нового пользователя
// @Summary Регистрация
// @Description Создание нового аккаунта пользователя
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Данные для регистрации"
// @Success 201 {object} AuthResponse
// @Failure 400 {object} map[string]string
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	if h.rejectIfMoodleOnly(c, "Use Moodle login instead of local registration.") {
		return
	}

	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userService.CreateUser(req.Email, req.Password, req.Name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Генерируем токены
	accessToken, err := utils.GenerateToken(user.ID.String(), user.Role, h.jwtSecret, h.jwtExpiry)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	refreshToken := uuid.New().String()
	err = h.userService.SaveRefreshToken(user.ID, refreshToken, time.Hour*24*7) // 7 дней
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save refresh token"})
		return
	}

	c.JSON(http.StatusCreated, AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user,
	})
}

// RegisterAdmin регистрация нового администратора
// @Summary Регистрация администратора
// @Description Создание нового аккаунта администратора
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Данные для регистрации"
// @Success 201 {object} AuthResponse
// @Failure 400 {object} map[string]string
// @Router /auth/register/admin [post]
func (h *AuthHandler) RegisterAdmin(c *gin.Context) {
	if h.rejectIfMoodleOnly(c, "Admin accounts are created from Moodle roles. Use Moodle login.") {
		return
	}

	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userService.CreateAdmin(req.Email, req.Password, req.Name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Генерируем токены
	accessToken, err := utils.GenerateToken(user.ID.String(), user.Role, h.jwtSecret, h.jwtExpiry)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	refreshToken := uuid.New().String()
	err = h.userService.SaveRefreshToken(user.ID, refreshToken, time.Hour*24*7) // 7 дней
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save refresh token"})
		return
	}

	c.JSON(http.StatusCreated, AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user,
	})
}

// Login вход в систему
// @Summary Вход
// @Description Авторизация пользователя
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Данные для входа"
// @Success 200 {object} AuthResponse
// @Failure 401 {object} map[string]string
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	if h.rejectIfMoodleOnly(c, "Use Moodle login on /api/auth/login/moodle.") {
		return
	}

	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userService.Authenticate(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Генерируем токены
	accessToken, err := utils.GenerateToken(user.ID.String(), user.Role, h.jwtSecret, h.jwtExpiry)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	refreshToken := uuid.New().String()
	err = h.userService.SaveRefreshToken(user.ID, refreshToken, time.Hour*24*7) // 7 дней
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save refresh token"})
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user,
	})
}

// RefreshTokenRequest запрос на обновление токена
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RefreshToken обновление access token
// @Summary Обновить токен
// @Description Обновление access token используя refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "Refresh token"
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := h.userService.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	// Генерируем новый access token
	accessToken, err := utils.GenerateToken(user.ID.String(), user.Role, h.jwtSecret, h.jwtExpiry)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": accessToken,
	})
}

// LogoutRequest запрос на выход
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// Logout выход из системы
// @Summary Выход
// @Description Выход пользователя из системы (удаление refresh token)
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body LogoutRequest false "Refresh token (опционально)"
// @Success 200 {object} map[string]string
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	var req LogoutRequest
	if err := c.ShouldBindJSON(&req); err == nil && req.RefreshToken != "" {
		h.userService.DeleteRefreshToken(req.RefreshToken)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// MoodleLoginRequest запрос на вход через Moodle
type MoodleLoginRequest struct {
	Username string `json:"username" binding:"required"` // Может быть email или username
	Password string `json:"password" binding:"required"`
}

// LoginMoodle вход через Moodle LMS
// @Summary Вход через Moodle
// @Description Авторизация пользователя через Moodle LMS
// @Tags auth
// @Accept json
// @Produce json
// @Param request body MoodleLoginRequest true "Данные для входа через Moodle"
// @Success 200 {object} AuthResponse
// @Failure 401 {object} map[string]string
// @Failure 503 {object} map[string]string
// @Router /auth/login/moodle [post]
func (h *AuthHandler) LoginMoodle(c *gin.Context) {
	if !h.moodleEnabled || h.moodleService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Moodle авторизация отключена",
			"message": "Для включения Moodle авторизации:",
			"instructions": []string{
				"1. Создайте файл backend/.env (скопируйте из .env.example если есть)",
				"2. Установите MOODLE_ENABLED=true",
				"3. Укажите MOODLE_BASE_URL=https://ваш-moodle-сервер.com",
				"4. Получите токен в Moodle (Site administration → Web services → Manage tokens)",
				"5. Укажите MOODLE_TOKEN=ваш_токен",
				"6. Перезапустите сервер",
				"Подробная инструкция: backend/MOODLE_SETUP_GUIDE.md",
			},
		})
		return
	}

	var req MoodleLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Проверяем учетные данные через Moodle
	moodleResult, err := h.moodleService.Authenticate(req.Username, req.Password)
	if err != nil {
		// Логируем детальную ошибку для отладки
		println("❌ Moodle login error:", err.Error())
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Неверные учетные данные Moodle",
			"details": err.Error(),
		})
		return
	}

	moodleUser := moodleResult.User
	moodleRole := moodleResult.Role
	
	println("🔐 Moodle login успешен:", moodleUser.Username, "| Email:", moodleUser.Email, "| Role:", moodleRole)

	// Ищем пользователя в локальной БД по email
	var user *models.User
	user, err = h.userService.GetUserByEmail(moodleUser.Email)
	
	// Если пользователь не найден и включено автосоздание, создаем его
	if err != nil && h.moodleAutoCreate {
		// Создаем пользователя без пароля (так как авторизация через Moodle)
		// Используем случайный пароль, который никогда не будет использован
		randomPassword := uuid.New().String()
		name := moodleUser.FullName
		if name == "" {
			name = moodleUser.FirstName + " " + moodleUser.LastName
		}
		user, err = h.userService.CreateUser(moodleUser.Email, randomPassword, name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось создать пользователя"})
			return
		}
		
		// Назначаем роль из Moodle
		if moodleRole == "admin" {
			user.Role = "admin"
			// Обновляем роль в БД
			if err := h.userService.UpdateUserRole(user.ID, "admin"); err != nil {
				println("⚠️  Не удалось обновить роль пользователя:", err.Error())
			} else {
				println("✅ Роль admin назначена новому пользователю")
			}
		}
	} else if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не найден. Обратитесь к администратору"})
		return
	} else {
		// Пользователь существует - обновляем роль из Moodle если она изменилась
		println("👤 Существующий пользователь:", user.Email, "| Текущая роль:", user.Role, "| Роль из Moodle:", moodleRole)
		if moodleRole == "admin" && user.Role != "admin" {
			user.Role = "admin"
			if err := h.userService.UpdateUserRole(user.ID, "admin"); err != nil {
				println("⚠️  Не удалось обновить роль пользователя:", err.Error())
			} else {
				println("✅ Роль обновлена на admin")
			}
		} else if moodleRole != "admin" && user.Role == "admin" {
			// Если в Moodle больше не admin, но в локальной БД admin - оставляем admin
			// (не понижаем права автоматически для безопасности)
			println("ℹ️  Пользователь остается admin (не понижаем права автоматически)")
		}
	}

	// Генерируем токены
	accessToken, err := utils.GenerateToken(user.ID.String(), user.Role, h.jwtSecret, h.jwtExpiry)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	refreshToken := uuid.New().String()
	err = h.userService.SaveRefreshToken(user.ID, refreshToken, time.Hour*24*7) // 7 дней
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save refresh token"})
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user,
	})
}

// RegisterMoodle регистрация через Moodle LMS
// @Summary Регистрация через Moodle
// @Description Регистрация пользователя через Moodle LMS (проверяет учетные данные и создает пользователя)
// @Tags auth
// @Accept json
// @Produce json
// @Param request body MoodleLoginRequest true "Данные для регистрации через Moodle"
// @Success 201 {object} AuthResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 503 {object} map[string]string
// @Router /auth/register/moodle [post]
func (h *AuthHandler) RegisterMoodle(c *gin.Context) {
	if h.rejectIfMoodleOnly(c, "Use Moodle login. The first successful Moodle login will create the local profile automatically.") {
		return
	}

	if !h.moodleEnabled || h.moodleService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Moodle авторизация отключена",
			"message": "Для включения Moodle регистрации:",
			"instructions": []string{
				"1. Создайте файл backend/.env (скопируйте из .env.example если есть)",
				"2. Установите MOODLE_ENABLED=true",
				"3. Укажите MOODLE_BASE_URL=https://ваш-moodle-сервер.com",
				"4. Получите токен в Moodle (Site administration → Web services → Manage tokens)",
				"5. Укажите MOODLE_TOKEN=ваш_токен",
				"6. Перезапустите сервер",
				"Подробная инструкция: backend/MOODLE_SETUP_GUIDE.md",
			},
		})
		return
	}

	var req MoodleLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Проверяем учетные данные через Moodle
	moodleResult, err := h.moodleService.Authenticate(req.Username, req.Password)
	if err != nil {
		// Логируем детальную ошибку для отладки
		println("❌ Moodle register error:", err.Error())
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Неверные учетные данные Moodle",
			"details": err.Error(),
		})
		return
	}
	
	moodleUser := moodleResult.User
	moodleRole := moodleResult.Role
	
	existingUser, err := h.userService.GetUserByEmail(moodleUser.Email)
	if err == nil && existingUser != nil {
		// Пользователь уже существует - возвращаем ошибку
		c.JSON(http.StatusBadRequest, gin.H{"error": "Пользователь с таким email уже зарегистрирован. Используйте вход через Moodle."})
		return
	}

	// Создаем пользователя в локальной БД
	// Используем случайный пароль, который никогда не будет использован
	randomPassword := uuid.New().String()
	name := moodleUser.FullName
	if name == "" {
		name = moodleUser.FirstName + " " + moodleUser.LastName
	}
	user, err := h.userService.CreateUser(moodleUser.Email, randomPassword, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось создать пользователя: " + err.Error()})
		return
	}

	// Назначаем роль из Moodle
	if moodleRole == "admin" {
		user.Role = "admin"
		if err := h.userService.UpdateUserRole(user.ID, "admin"); err != nil {
			println("⚠️  Не удалось обновить роль пользователя:", err.Error())
		}
	}

	// Генерируем токены
	accessToken, err := utils.GenerateToken(user.ID.String(), user.Role, h.jwtSecret, h.jwtExpiry)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	refreshToken := uuid.New().String()
	err = h.userService.SaveRefreshToken(user.ID, refreshToken, time.Hour*24*7) // 7 дней
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save refresh token"})
		return
	}

	c.JSON(http.StatusCreated, AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user,
	})
}
