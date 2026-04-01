package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// MoodleService сервис для работы с Moodle Web Services API
type MoodleService struct {
	baseURL string
	token   string
	service string
	client  *http.Client
	testMode bool // Режим разработки - работает без реального сервера
}

// MoodleUser представляет пользователя из Moodle
type MoodleUser struct {
	ID        int    `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
	FullName  string `json:"fullname"`
}

// MoodleAuthResponse ответ от Moodle API при авторизации
type MoodleAuthResponse struct {
	Token       string      `json:"token"`
	PrivateToken string     `json:"privatetoken,omitempty"`
	User        MoodleUser  `json:"user"`
}

// MoodleError ошибка от Moodle API
type MoodleError struct {
	Exception string `json:"exception"`
	ErrorCode string `json:"errorcode"`
	Message   string `json:"message"`
}

// NewMoodleService создает новый сервис для работы с Moodle
func NewMoodleService(baseURL, token, service string) *MoodleService {
	return &MoodleService{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		token:   token,
		service: service,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		testMode: false,
	}
}

// NewMoodleServiceTestMode создает сервис в тестовом режиме (без реального сервера)
func NewMoodleServiceTestMode() *MoodleService {
	return &MoodleService{
		baseURL:  "test-mode",
		token:     "test-token",
		service:   "moodle_mobile_app",
		testMode: true,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// MoodleAuthResult результат авторизации в Moodle
type MoodleAuthResult struct {
	User  *MoodleUser
	Token string
	Role  string
}

// Authenticate проверяет учетные данные пользователя через Moodle
func (s *MoodleService) Authenticate(username, password string) (*MoodleAuthResult, error) {
	// Тестовый режим - работает без реального сервера
	if s.testMode {
		// В тестовом режиме принимаем любые учетные данные
		// Создаем mock пользователя на основе username
		
		// Определяем роль на основе username
		role := "user"
		if strings.Contains(strings.ToLower(username), "admin") {
			role = "admin"
		}
		
		return &MoodleAuthResult{
			User: &MoodleUser{
				ID:        1,
				Username:  username,
				Email:     username + "@moodle.test",
				FirstName: "Test",
				LastName:  "User",
				FullName:  "Test User",
			},
			Token: "test-token",
			Role:  role,
		}, nil
	}

	if s.baseURL == "" || s.token == "" {
		return nil, errors.New("Moodle не настроен")
	}

	// Используем core_auth_user_login для проверки учетных данных
	endpoint := fmt.Sprintf("%s/login/token.php", s.baseURL)
	
	// Формируем запрос
	data := url.Values{}
	data.Set("username", username)
	data.Set("password", password)
	data.Set("service", s.service)

	req, err := http.NewRequest("POST", endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса к Moodle: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	// Парсим ответ - login/token.php возвращает только токен или ошибку
	var tokenResp struct {
		Token   string `json:"token"`
		Error   string `json:"error"`
		ErrorCode string `json:"errorcode"`
	}
	
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		// Пробуем парсить как MoodleError
		var moodleErr MoodleError
		if json.Unmarshal(body, &moodleErr) == nil {
			return nil, fmt.Errorf("Moodle ошибка: %s", moodleErr.Message)
		}
		return nil, fmt.Errorf("ошибка парсинга ответа Moodle: %w. Ответ: %s", err, string(body))
	}

	// Проверяем наличие ошибки
	if tokenResp.Error != "" {
		return nil, fmt.Errorf("Moodle ошибка: %s", tokenResp.Error)
	}

	// Проверяем наличие токена
	if tokenResp.Token == "" {
		return nil, errors.New("неверные учетные данные Moodle")
	}

	// Получаем информацию о пользователе через Web Services API используя полученный токен
	user, err := s.GetUserByToken(tokenResp.Token)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить информацию о пользователе: %w", err)
	}

	fmt.Printf("👤 Получена информация о пользователе: ID=%d, Username=%s, Email=%s\n", user.ID, user.Username, user.Email)

	// Получаем роль пользователя
	fmt.Println("🔍 Проверяем роль пользователя...")
	role, err := s.GetUserRole(tokenResp.Token)
	if err != nil {
		// Если не удалось получить роль, используем user по умолчанию
		fmt.Printf("⚠️  Не удалось получить роль пользователя: %v, используем 'user'\n", err)
		role = "user"
	}
	
	fmt.Printf("🔑 Moodle авторизация успешна: username=%s, role=%s\n", user.Username, role)

	return &MoodleAuthResult{
		User:  user,
		Token: tokenResp.Token,
		Role:  role,
	}, nil
}

// GetUserByToken получает информацию о пользователе по токену через Web Services API
func (s *MoodleService) GetUserByToken(token string) (*MoodleUser, error) {
	if s.baseURL == "" || s.token == "" {
		return nil, errors.New("Moodle не настроен")
	}

	// Используем core_webservice_get_site_info для получения информации о пользоватеle
	endpoint := fmt.Sprintf("%s/webservice/rest/server.php", s.baseURL)
	
	// Формируем запрос
	data := url.Values{}
	data.Set("wstoken", token)
	data.Set("wsfunction", "core_webservice_get_site_info")
	data.Set("moodlewsrestformat", "json")

	req, err := http.NewRequest("POST", endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса к Moodle: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	// Парсим ответ
	var siteInfo struct {
		UserID    int    `json:"userid"`
		Username  string `json:"username"`
		FirstName string `json:"firstname"`
		LastName  string `json:"lastname"`
		FullName  string `json:"fullname"`
		Email     string `json:"useremail"`
	}

	if err := json.Unmarshal(body, &siteInfo); err != nil {
		// Проверяем, может быть это ошибка
		var moodleErr MoodleError
		if json.Unmarshal(body, &moodleErr) == nil {
			return nil, fmt.Errorf("Moodle ошибка: %s", moodleErr.Message)
		}
		return nil, fmt.Errorf("ошибка парсинга ответа: %w", err)
	}

	// Получаем полную информацию о пользователе
	user, err := s.GetUserByID(siteInfo.UserID, token)
	if err != nil {
		// Если не удалось, используем данные из site_info
		return &MoodleUser{
			ID:        siteInfo.UserID,
			Username:  siteInfo.Username,
			Email:     siteInfo.Email,
			FirstName: siteInfo.FirstName,
			LastName:  siteInfo.LastName,
			FullName:  siteInfo.FullName,
		}, nil
	}

	return user, nil
}

// GetUserByID получает информацию о пользователе по ID через Web Services API
func (s *MoodleService) GetUserByID(userID int, token string) (*MoodleUser, error) {
	if s.baseURL == "" || s.token == "" {
		return nil, errors.New("Moodle не настроен")
	}

	endpoint := fmt.Sprintf("%s/webservice/rest/server.php", s.baseURL)
	
	// Формируем запрос
	data := url.Values{}
	data.Set("wstoken", token)
	data.Set("wsfunction", "core_user_get_users_by_field")
	data.Set("moodlewsrestformat", "json")
	data.Set("field", "id")
	data.Set(fmt.Sprintf("values[0]"), fmt.Sprintf("%d", userID))

	req, err := http.NewRequest("POST", endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса к Moodle: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	// Парсим ответ (возвращается массив пользователей)
	var users []MoodleUser
	if err := json.Unmarshal(body, &users); err != nil {
		// Проверяем, может быть это ошибка
		var moodleErr MoodleError
		if json.Unmarshal(body, &moodleErr) == nil {
			return nil, fmt.Errorf("Moodle ошибка: %s", moodleErr.Message)
		}
		return nil, fmt.Errorf("ошибка парсинга ответа: %w", err)
	}

	if len(users) == 0 {
		return nil, errors.New("пользователь не найден")
	}

	return &users[0], nil
}

// ValidateToken проверяет валидность токена Moodle
func (s *MoodleService) ValidateToken(token string) (*MoodleUser, error) {
	return s.GetUserByToken(token)
}

// IsUserAdmin проверяет, является ли пользователь администратором в Moodle
func (s *MoodleService) IsUserAdmin(userToken string) (bool, error) {
	if s.testMode {
		// В тестовом режиме проверяем по username
		return false, nil
	}

	if s.baseURL == "" || s.token == "" {
		return false, errors.New("Moodle не настроен")
	}

	// Получаем информацию о сайте, которая включает информацию о правах пользователя
	endpoint := fmt.Sprintf("%s/webservice/rest/server.php", s.baseURL)
	
	data := url.Values{}
	data.Set("wstoken", userToken)
	data.Set("wsfunction", "core_webservice_get_site_info")
	data.Set("moodlewsrestformat", "json")

	req, err := http.NewRequest("POST", endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return false, fmt.Errorf("ошибка создания запроса: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.client.Do(req)
	if err != nil {
		return false, fmt.Errorf("ошибка запроса к Moodle: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	fmt.Printf("📄 Полный ответ от Moodle site_info:\n%s\n", string(body))

	// Парсим ответ как map для просмотра всех полей
	var siteInfoMap map[string]interface{}
	if err := json.Unmarshal(body, &siteInfoMap); err == nil {
		fmt.Printf("📋 Все поля site_info:\n")
		for key, value := range siteInfoMap {
			fmt.Printf("  %s: %v\n", key, value)
		}
	}

	// Парсим ответ
	var siteInfo struct {
		UserID       int    `json:"userid"`
		SiteAdmin    bool   `json:"usercansiteadmin"`
		UserCanManageOwnFiles bool `json:"usercanmanageownfiles"`
	}

	if err := json.Unmarshal(body, &siteInfo); err != nil {
		fmt.Printf("❌ Ошибка парсинга site_info: %v\n", err)
		return false, fmt.Errorf("ошибка парсинга ответа: %w", err)
	}

	fmt.Printf("📊 Site info: UserID=%d, SiteAdmin=%v\n", siteInfo.UserID, siteInfo.SiteAdmin)

	// Проверяем, является ли пользователь администратором сайта
	return siteInfo.SiteAdmin, nil
}

// GetUserRole получает роль пользователя из Moodle (admin или user)
func (s *MoodleService) GetUserRole(userToken string) (string, error) {
	isAdmin, err := s.IsUserAdmin(userToken)
	if err != nil {
		fmt.Printf("⚠️  Ошибка проверки роли admin: %v\n", err)
		return "user", err
	}
	
	if isAdmin {
		fmt.Println("✅ Пользователь является администратором в Moodle")
		return "admin", nil
	}
	
	fmt.Println("ℹ️  Пользователь не является администратором в Moodle")
	return "user", nil
}
