package main

import (
	"english-learning/internal/config"
	"english-learning/internal/database"
	"english-learning/internal/handlers"
	"english-learning/internal/middleware"
	"english-learning/internal/services"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	// Импорт для инициализации Swagger документации
	_ "english-learning/docs"
)

// @title English Learning Platform API
// @version 1.0
// @description API для платформы изучения английского языка. Используется Dictionary API для получения данных о словах.
//
// ## Таблица всех эндпоинтов (API Routes)
//
// ### Общие эндпоинты
// | Метод | Путь | Описание |
// |-------|------|----------|
// | GET | `/health` | Проверка работоспособности сервера и БД |
// | GET | `/swagger/*any` | Swagger документация |
// | GET | `/api` | Информация об API |
// | GET | `/api/greeting` | Приветствие |
//
// ### Авторизация (`/api/auth`)
// | Метод | Путь | Описание |
// |-------|------|----------|
// | POST | `/api/auth/register` | Регистрация пользователя |
// | POST | `/api/auth/register/admin` | Регистрация администратора |
// | POST | `/api/auth/register/moodle` | Регистрация через Moodle LMS |
// | POST | `/api/auth/login` | Вход в систему |
// | POST | `/api/auth/login/moodle` | Вход через Moodle LMS |
// | POST | `/api/auth/refresh` | Обновление токена |
// | POST | `/api/auth/logout` | Выход из системы |
//
// ### Пользователи (`/api/users`)
// | Метод | Путь | Описание |
// |-------|------|----------|
// | GET | `/api/users` | Получить список всех пользователей |
// | GET | `/api/users/:id` | Получить пользователя по ID |
//
// ### Курсы (`/api/courses`)
// | Метод | Путь | Описание |
// |-------|------|----------|
// | GET | `/api/courses` | Получить список всех курсов |
// | GET | `/api/courses/:id` | Получить курс по ID |
// | POST | `/api/courses` | Создать новый курс |
// | PUT | `/api/courses/:id` | Обновить курс |
// | DELETE | `/api/courses/:id` | Удалить курс |
//
// ### Деки (`/api/decks`)
// | Метод | Путь | Описание |
// |-------|------|----------|
// | GET | `/api/decks` | Получить список всех дек |
// | GET | `/api/decks/:id` | Получить деку по ID |
// | POST | `/api/decks` | Создать новую деку |
// | PUT | `/api/decks/:id` | Обновить деку |
// | DELETE | `/api/decks/:id` | Удалить деку |
//
// ### Карточки (`/api/cards`)
// | Метод | Путь | Описание |
// |-------|------|----------|
// | GET | `/api/cards` | Получить список всех карточек |
// | GET | `/api/cards/:id` | Получить карточку по ID |
// | POST | `/api/cards` | Создать новую карточку |
// | PUT | `/api/cards/:id` | Обновить карточку |
// | DELETE | `/api/cards/:id` | Удалить карточку |
//
// ### Личный словарь (`/api/vocabulary`)
// | Метод | Путь | Описание |
// |-------|------|----------|
// | GET | `/api/vocabulary` | Получить весь личный словарь |
// | GET | `/api/vocabulary/review` | Получить слова для повторения |
// | GET | `/api/vocabulary/:id` | Получить слово по ID |
// | POST | `/api/vocabulary` | Добавить слово в словарь |
// | PUT | `/api/vocabulary/:id` | Обновить слово в словаре |
// | PUT | `/api/vocabulary/:id/stats` | Обновить статистику слова |
// | DELETE | `/api/vocabulary/:id` | Удалить слово из словаря |
//
// ### Прогресс по курсам (`/api/user-courses`)
// | Метод | Путь | Описание |
// |-------|------|----------|
// | GET | `/api/user-courses` | Получить весь прогресс по курсам |
// | GET | `/api/user-courses/user/:user_id` | Получить курсы пользователя |
// | GET | `/api/user-courses/:id` | Получить прогресс по курсу по ID |
// | POST | `/api/user-courses` | Начать курс |
// | PUT | `/api/user-courses/:id` | Обновить прогресс по курсу |
// | DELETE | `/api/user-courses/:id` | Удалить прогресс по курсу |
//
// ### Прогресс по декам (`/api/user-decks`)
// | Метод | Путь | Описание |
// |-------|------|----------|
// | GET | `/api/user-decks` | Получить весь прогресс по декам |
// | GET | `/api/user-decks/user/:user_id` | Получить деки пользователя |
// | GET | `/api/user-decks/:id` | Получить прогресс по деку по ID |
// | POST | `/api/user-decks` | Начать деку |
// | PUT | `/api/user-decks/:id` | Обновить прогресс по деку |
// | DELETE | `/api/user-decks/:id` | Удалить прогресс по деку |
//
// ### Прогресс по карточкам (`/api/user-cards`)
// | Метод | Путь | Описание |
// |-------|------|----------|
// | GET | `/api/user-cards` | Получить весь прогресс по карточкам |
// | GET | `/api/user-cards/user/:user_id` | Получить карточки пользователя |
// | GET | `/api/user-cards/:id` | Получить прогресс по карточке по ID |
// | POST | `/api/user-cards` | Создать запись о прогрессе по карточке |
// | PUT | `/api/user-cards/:id` | Обновить прогресс по карточке |
// | DELETE | `/api/user-cards/:id` | Удалить прогресс по карточке |
//
// ### Сессии тренировок (`/api/training-sessions`)
// | Метод | Путь | Описание |
// |-------|------|----------|
// | GET | `/api/training-sessions` | Получить все сессии тренировок |
// | GET | `/api/training-sessions/user/:user_id` | Получить сессии пользователя |
// | GET | `/api/training-sessions/:id` | Получить сессию по ID |
// | POST | `/api/training-sessions` | Начать сессию тренировки |
// | PUT | `/api/training-sessions/:id/finish` | Завершить сессию тренировки |
// | DELETE | `/api/training-sessions/:id` | Удалить сессию тренировки |
//
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@example.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8000
// @BasePath /api
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	log.Println("🚀 Запуск сервера...")

	// Загружаем конфигурацию
	cfg := config.Load()
	log.Println("✅ Конфигурация загружена")

	// Подключаемся к БД
	log.Println("📦 Подключение к базе данных...")
	db, err := database.Connect(cfg.Database)
	if err != nil {
		log.Fatalf("❌ Ошибка подключения к БД: %v", err)
	}
	defer db.Close()
	log.Println("✅ Подключение к БД установлено")

	// Инициализируем сервисы
	userService := services.NewUserService(db)
	courseService := services.NewCourseService(db)
	courseAccessService := services.NewCourseAccessService(db)
	deckService := services.NewDeckService(db)
	cardService := services.NewCardService(db)
	vocabService := services.NewPersonalVocabularyService(db)
	userCourseService := services.NewUserCourseService(db)
	userDeckService := services.NewUserDeckService(db)
	userCardService := services.NewUserCardService(db)
	trainingSessionService := services.NewTrainingSessionService(db, cardService, userDeckService, userCardService)
	dictionaryService := services.NewDictionaryService()
	readingTextService := services.NewReadingTextService(db)

	// Инициализируем Moodle сервис (если включен)
	var moodleService *services.MoodleService
	moodleEnabled := cfg.Moodle.Enabled

	// Проверяем настройки Moodle
	if cfg.Moodle.Enabled {
		// Если включен тестовый режим, используем mock сервис
		if cfg.Moodle.TestMode {
			moodleService = services.NewMoodleServiceTestMode()
			log.Println("🧪 Moodle интеграция включена в ТЕСТОВОМ РЕЖИМЕ (без реального сервера)")
			log.Println("   Для использования реального Moodle установите MOODLE_TEST_MODE=false в .env")
		} else {
			baseURL := strings.TrimSpace(cfg.Moodle.BaseURL)
			token := strings.TrimSpace(cfg.Moodle.Token)

			// Проверяем, что URL и токен не пустые и не являются placeholder значениями
			if baseURL == "" || token == "" ||
				baseURL == "https://your-moodle-site.com" ||
				token == "your_moodle_web_service_token_here" {
				log.Println("⚠️  ВНИМАНИЕ: MOODLE_ENABLED=true, но MOODLE_BASE_URL или MOODLE_TOKEN не настроены!")
				log.Println("   Переключаемся в тестовый режим...")
				moodleService = services.NewMoodleServiceTestMode()
				log.Println("🧪 Moodle работает в ТЕСТОВОМ РЕЖИМЕ (без реального сервера)")
				log.Println("   Для использования реального Moodle:")
				log.Println("   - Установите MOODLE_BASE_URL=https://ваш-moodle-сервер.com")
				log.Println("   - Установите MOODLE_TOKEN=ваш_токен_из_админки_moodle")
				log.Println("   - Установите MOODLE_TEST_MODE=false")
			} else {
				moodleService = services.NewMoodleService(baseURL, token, cfg.Moodle.Service)
				log.Printf("✅ Moodle интеграция включена: %s", baseURL)
			}
		}
	} else {
		log.Println("ℹ️  Moodle интеграция отключена (MOODLE_ENABLED=false)")
	}

	// Инициализируем handlers
	userHandler := handlers.NewUserHandler(userService, courseAccessService)
	authHandler := handlers.NewAuthHandler(
		userService,
		moodleService,
		cfg.JWT.Secret,
		cfg.JWT.ExpiryHours,
		moodleEnabled, // Используем проверенное значение
		cfg.Moodle.OnlyAuth,
		cfg.Moodle.AutoCreate,
	)
	courseHandler := handlers.NewCourseHandler(courseService, deckService, cardService, dictionaryService, courseAccessService)
	deckHandler := handlers.NewDeckHandler(deckService, courseService, courseAccessService)
	cardHandler := handlers.NewCardHandler(cardService, userDeckService, deckService, courseService, courseAccessService)
	vocabHandler := handlers.NewPersonalVocabularyHandler(vocabService)
	userCourseHandler := handlers.NewUserCourseHandler(userCourseService)
	userDeckHandler := handlers.NewUserDeckHandler(userDeckService)
	userCardHandler := handlers.NewUserCardHandler(userCardService)
	trainingSessionHandler := handlers.NewTrainingSessionHandler(trainingSessionService)
	dictionaryHandler := handlers.NewDictionaryHandler(dictionaryService)
	uploadHandler := handlers.NewUploadHandler()
	readingTextHandler := handlers.NewReadingTextHandler(readingTextService)

	// Настраиваем роутер
	router := gin.Default()

	// Добавляем CORS middleware
	router.Use(middleware.CORS())

	// Health check endpoint с проверкой БД
	// @Summary Health check
	// @Description Проверка работоспособности сервера и подключения к БД
	// @Tags health
	// @Produce json
	// @Success 200 {object} map[string]interface{}
	// @Router /health [get]
	router.GET("/health", func(c *gin.Context) {
		// Проверяем подключение к БД
		ctx := c.Request.Context()
		if err := db.Ping(ctx); err != nil {
			c.JSON(500, gin.H{
				"status":  "error",
				"message": "Database connection failed",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "Server and database are running",
			"database": gin.H{
				"connected": true,
				"host":      cfg.Database.Host,
				"port":      cfg.Database.Port,
				"name":      cfg.Database.Name,
			},
		})
	})

	// Swagger документация (только для администраторов)
	// @Summary Swagger UI
	// @Description Интерактивная документация API (доступна только администраторам)
	router.GET("/swagger/*any", middleware.Auth(cfg.JWT.Secret), middleware.Admin(), ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API routes
	api := router.Group("/api")
	{
		// @Summary API info
		// @Description Информация об API
		// @Tags info
		// @Produce json
		// @Success 200 {object} map[string]string
		// @Router / [get]
		api.GET("", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "English Learning Platform API",
				"version": "1.0.0",
				"docs":    "/swagger/index.html",
			})
		})

		// Greeting endpoint
		api.GET("/greeting", handlers.GetGreeting)

		// Dictionary API endpoint (публичный)
		api.GET("/dictionary/:word", dictionaryHandler.GetWordInfo)

		// Auth endpoints (публичные)
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/register/admin", authHandler.RegisterAdmin)
			auth.POST("/register/moodle", authHandler.RegisterMoodle)
			auth.POST("/login", authHandler.Login)
			auth.POST("/login/moodle", authHandler.LoginMoodle)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.POST("/logout", authHandler.Logout)
		}

		// Users endpoints (пока без ограничений)
		api.GET("/users", userHandler.GetAllUsers)
		api.GET("/users/:id", userHandler.GetUserByID)

		// Courses endpoints (публичные - только чтение)
		// Используем OptionalAuth чтобы админы видели все курсы, а пользователи - только опубликованные
        api.GET("/courses", middleware.OptionalAuth(cfg.JWT.Secret), courseHandler.GetAllCourses)
		api.GET("/courses/by-slug/:slug", middleware.OptionalAuth(cfg.JWT.Secret), courseHandler.GetAccessibleCourseBySlug)
		api.GET("/courses/:id", middleware.OptionalAuth(cfg.JWT.Secret), courseHandler.GetAccessibleCourseByID)

		// Decks endpoints (публичные - только чтение)
		api.GET("/decks", middleware.OptionalAuth(cfg.JWT.Secret), deckHandler.GetAllDecks)
		api.GET("/decks/by-slug/:course_slug/:deck_slug", middleware.OptionalAuth(cfg.JWT.Secret), deckHandler.GetDeckBySlug)
		api.GET("/decks/:id", middleware.OptionalAuth(cfg.JWT.Secret), deckHandler.GetDeckByID)

		// Cards endpoints (публичные - чтение и создание)
		api.GET("/cards", middleware.OptionalAuth(cfg.JWT.Secret), cardHandler.GetAllCards)
		api.GET("/cards/:id", middleware.OptionalAuth(cfg.JWT.Secret), cardHandler.GetCardByID)
		api.POST("/cards", cardHandler.CreateCard)

		// Personal Vocabulary endpoints
		api.GET("/vocabulary", vocabHandler.GetAllPersonalVocabulary)
		api.GET("/vocabulary/review", vocabHandler.GetPersonalVocabularyForReview)
		api.GET("/vocabulary/:id", vocabHandler.GetPersonalVocabularyByID)
		api.POST("/vocabulary", vocabHandler.CreatePersonalVocabulary)
		api.PUT("/vocabulary/:id", vocabHandler.UpdatePersonalVocabulary)
		api.PUT("/vocabulary/:id/stats", vocabHandler.UpdatePersonalVocabularyStats)
		api.DELETE("/vocabulary/:id", vocabHandler.DeletePersonalVocabulary)

		// Reading Texts endpoints
		api.GET("/reading-texts", readingTextHandler.GetAllReadingTexts)
		api.GET("/reading-texts/:id", readingTextHandler.GetReadingTextByID)
		api.POST("/reading-texts", readingTextHandler.CreateReadingText)
		api.PUT("/reading-texts/:id/audio", readingTextHandler.UpdateReadingTextAudio)
		api.DELETE("/reading-texts/:id", readingTextHandler.DeleteReadingText)

		// User Courses endpoints (прогресс пользователя по курсам)
		authenticated := api.Group("")
		authenticated.Use(middleware.Auth(cfg.JWT.Secret))
		{
			authenticated.GET("/user-courses/active", userCourseHandler.GetActiveUserCourse)
			authenticated.POST("/user-courses/:course_id/activate", userCourseHandler.ActivateCourse)

			authenticated.GET("/user-decks/active", userDeckHandler.GetActiveUserDeck)
			authenticated.POST("/user-decks/:deck_id/activate", userDeckHandler.ActivateDeck)

			authenticated.POST("/cards/custom", cardHandler.CreateCustomCard)

			authenticated.POST("/training-sessions/start", trainingSessionHandler.StartTrainingSession)
			authenticated.POST("/training-sessions/:id/answer", trainingSessionHandler.SubmitAnswer)
		}

		api.GET("/user-courses", userCourseHandler.GetAllUserCourses)
		api.GET("/user-courses/user/:user_id", userCourseHandler.GetUserCoursesByUserID)
		api.GET("/user-courses/:id", userCourseHandler.GetUserCourseByID)
		api.POST("/user-courses", userCourseHandler.StartCourse)
		api.PUT("/user-courses/:id", userCourseHandler.UpdateUserCourse)
		api.DELETE("/user-courses/:id", userCourseHandler.DeleteUserCourse)

		// User Decks endpoints (прогресс пользователя по декам)
		api.GET("/user-decks", userDeckHandler.GetAllUserDecks)
		api.GET("/user-decks/user/:user_id", userDeckHandler.GetUserDecksByUserID)
		api.GET("/user-decks/:id", userDeckHandler.GetUserDeckByID)
		api.POST("/user-decks", userDeckHandler.StartDeck)
		api.PUT("/user-decks/:id", userDeckHandler.UpdateUserDeck)
		api.DELETE("/user-decks/:id", userDeckHandler.DeleteUserDeck)

		// User Cards endpoints (прогресс пользователя по карточкам)
		api.GET("/user-cards", userCardHandler.GetAllUserCards)
		api.GET("/user-cards/user/:user_id", userCardHandler.GetUserCardsByUserID)
		api.GET("/user-cards/:id", userCardHandler.GetUserCardByID)
		api.POST("/user-cards", userCardHandler.CreateUserCard)
		api.PUT("/user-cards/:id", userCardHandler.UpdateUserCard)
		api.DELETE("/user-cards/:id", userCardHandler.DeleteUserCard)

		// Training Sessions endpoints (сессии тренировок)
		api.GET("/training-sessions", trainingSessionHandler.GetAllTrainingSessions)
		api.GET("/training-sessions/user/:user_id", trainingSessionHandler.GetTrainingSessionsByUserID)
		api.GET("/training-sessions/:id", middleware.OptionalAuth(cfg.JWT.Secret), trainingSessionHandler.GetTrainingSessionByID)
		api.POST("/training-sessions", trainingSessionHandler.StartTrainingSession)
		api.PUT("/training-sessions/:id/finish", trainingSessionHandler.FinishTrainingSession)
		api.DELETE("/training-sessions/:id", trainingSessionHandler.DeleteTrainingSession)

		// Upload endpoints (требуют авторизации)
		upload := api.Group("/upload")
		upload.Use(middleware.Auth(cfg.JWT.Secret))
		{
			upload.POST("/image", uploadHandler.UploadImage)
			upload.POST("/audio", uploadHandler.UploadAudio)
			upload.DELETE("/delete", uploadHandler.DeleteFile)
		}

		// Статические файлы для загруженных файлов
		router.Static("/uploads", "./uploads")

		// Admin endpoints (требуют авторизации и роль admin)
		admin := api.Group("/admin")
		admin.Use(middleware.Auth(cfg.JWT.Secret))
		admin.Use(middleware.Admin())
		{
			// Admin Users
			admin.GET("/users", userHandler.GetAllUsers)
			admin.PUT("/users/:id/role", userHandler.UpdateUserRole)
			admin.GET("/users/course-accesses", userHandler.GetAllCourseAccesses)
			admin.POST("/users/:id/course-accesses/:course_id", userHandler.GrantCourseAccess)
			admin.DELETE("/users/:id/course-accesses/:course_id", userHandler.RevokeCourseAccess)

			// Admin Courses
			admin.GET("/courses", courseHandler.GetAdminCourses)
			admin.POST("/courses", courseHandler.CreateCourse)
			admin.POST("/courses/import", courseHandler.CreateCourseFromImport)
			admin.PUT("/courses/:id", courseHandler.UpdateCourse)
			admin.DELETE("/courses/:id", courseHandler.DeleteCourse)
			admin.POST("/courses/:id/publish", courseHandler.PublishCourse)

			// Admin Decks
			admin.POST("/decks", deckHandler.CreateDeck)
			admin.PUT("/decks/:id", deckHandler.UpdateDeck)
			admin.DELETE("/decks/:id", deckHandler.DeleteDeck)

			// Admin Cards
			admin.POST("/cards", cardHandler.CreateCard)
			admin.PUT("/cards/:id", cardHandler.UpdateCard)
			admin.DELETE("/cards/:id", cardHandler.DeleteCard)

			// Admin Podcasts
			admin.GET("/podcasts", readingTextHandler.GetAdminPodcasts)
		}
	}

	// Запускаем сервер
	addr := ":" + cfg.Server.Port
	log.Printf("🌐 Сервер запущен на %s", addr)
	log.Printf("📚 Swagger документация: http://localhost%s/swagger/index.html", addr)
	log.Printf("❤️  Health check: http://localhost%s/health", addr)

	if err := router.Run(addr); err != nil {
		log.Fatalf("❌ Ошибка запуска сервера: %v", err)
	}
}
