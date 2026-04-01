package services

import (
	"context"
	"english-learning/internal/models"
	"english-learning/internal/utils"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserService struct {
	db *pgxpool.Pool
}

func NewUserService(db *pgxpool.Pool) *UserService {
	return &UserService{db: db}
}

// GetAllUsers возвращает список всех пользователей
func (s *UserService) GetAllUsers() ([]models.User, error) {
	rows, err := s.db.Query(context.Background(),
		"SELECT id, email, name, role, created_at, updated_at FROM users ORDER BY created_at DESC",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.ID, &user.Email, &user.Name, &user.Role, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, err
		}
		user.RoleID = user.GetRoleID() // Вычисляем RoleID
		users = append(users, user)
	}

	return users, nil
}

// GetUserByID возвращает пользователя по ID
func (s *UserService) GetUserByID(userID uuid.UUID) (*models.User, error) {
	var user models.User
	err := s.db.QueryRow(context.Background(),
		"SELECT id, email, name, role, created_at, updated_at FROM users WHERE id = $1",
		userID,
	).Scan(&user.ID, &user.Email, &user.Name, &user.Role, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, errors.New("user not found")
	}

	user.RoleID = user.GetRoleID() // Вычисляем RoleID
	return &user, nil
}

// GetUserByEmail возвращает пользователя по email
func (s *UserService) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	err := s.db.QueryRow(context.Background(),
		"SELECT id, email, name, role, created_at, updated_at FROM users WHERE email = $1",
		email,
	).Scan(&user.ID, &user.Email, &user.Name, &user.Role, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, errors.New("user not found")
	}

	user.RoleID = user.GetRoleID() // Вычисляем RoleID
	return &user, nil
}

// CreateUser создает нового пользователя с ролью 'user'
func (s *UserService) CreateUser(email, password, name string) (*models.User, error) {
	return s.createUserWithRole(email, password, name, "user")
}

// CreateAdmin создает нового пользователя с ролью 'admin'
func (s *UserService) CreateAdmin(email, password, name string) (*models.User, error) {
	return s.createUserWithRole(email, password, name, "admin")
}

// createUserWithRole создает пользователя с указанной ролью
func (s *UserService) createUserWithRole(email, password, name, role string) (*models.User, error) {
	// Проверяем, существует ли пользователь
	var exists bool
	err := s.db.QueryRow(context.Background(),
		"SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", email).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("user with this email already exists")
	}

	// Хешируем пароль
	passwordHash, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}

	// Создаем пользователя
	var user models.User
	var namePtr *string
	if name != "" {
		namePtr = &name
	}

	err = s.db.QueryRow(context.Background(),
		`INSERT INTO users (email, password_hash, name, role)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, email, name, role, created_at, updated_at`,
		email, passwordHash, namePtr, role,
	).Scan(&user.ID, &user.Email, &user.Name, &user.Role, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, err
	}

	user.RoleID = user.GetRoleID() // Вычисляем RoleID
	return &user, nil
}

// Authenticate проверяет email и пароль, возвращает пользователя
func (s *UserService) Authenticate(email, password string) (*models.User, error) {
	var user models.UserWithPassword

	err := s.db.QueryRow(context.Background(),
		"SELECT id, email, password_hash, name, role, created_at, updated_at FROM users WHERE email = $1",
		email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.Role, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if !utils.CheckPasswordHash(password, user.PasswordHash) {
		return nil, errors.New("invalid credentials")
	}

	user.User.RoleID = user.User.GetRoleID() // Вычисляем RoleID
	return &user.User, nil
}

// SaveRefreshToken сохраняет refresh token в БД
func (s *UserService) SaveRefreshToken(userID uuid.UUID, token string, expiry time.Duration) error {
	_, err := s.db.Exec(context.Background(),
		"INSERT INTO refresh_tokens (user_id, token, expires_at) VALUES ($1, $2, $3)",
		userID, token, time.Now().Add(expiry),
	)
	return err
}

// ValidateRefreshToken проверяет refresh token
func (s *UserService) ValidateRefreshToken(token string) (uuid.UUID, error) {
	var userID uuid.UUID
	var expiresAt time.Time

	err := s.db.QueryRow(context.Background(),
		"SELECT user_id, expires_at FROM refresh_tokens WHERE token = $1",
		token,
	).Scan(&userID, &expiresAt)

	if err != nil {
		return uuid.Nil, errors.New("invalid refresh token")
	}

	if time.Now().After(expiresAt) {
		return uuid.Nil, errors.New("refresh token expired")
	}

	return userID, nil
}

// DeleteRefreshToken удаляет refresh token
func (s *UserService) DeleteRefreshToken(token string) error {
	_, err := s.db.Exec(context.Background(),
		"DELETE FROM refresh_tokens WHERE token = $1",
		token,
	)
	return err
}

// UpdateUserRole обновляет роль пользователя
func (s *UserService) UpdateUserRole(userID uuid.UUID, role string) error {
	_, err := s.db.Exec(context.Background(),
		"UPDATE users SET role = $1, updated_at = NOW() WHERE id = $2",
		role, userID,
	)
	return err
}
