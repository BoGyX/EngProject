package services

import (
	"context"
	"english-learning/internal/models"
	"english-learning/internal/utils"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserService struct {
	db *pgxpool.Pool
}

type CreateUserParams struct {
	Email       string
	Password    string
	Name        string
	Role        string
	MoodleLogin string
}

var moodleLoginEmailSanitizer = regexp.MustCompile(`[^a-z0-9._-]+`)

func NewUserService(db *pgxpool.Pool) *UserService {
	return &UserService{db: db}
}

// GetAllUsers returns all users ordered by creation date.
func (s *UserService) GetAllUsers() ([]models.User, error) {
	rows, err := s.db.Query(context.Background(),
		"SELECT id, email, moodle_login, name, role, created_at, updated_at FROM users ORDER BY created_at DESC",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.ID, &user.Email, &user.MoodleLogin, &user.Name, &user.Role, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, err
		}
		user.RoleID = user.GetRoleID()
		users = append(users, user)
	}

	return users, rows.Err()
}

// GetUserByID returns a user by ID.
func (s *UserService) GetUserByID(userID uuid.UUID) (*models.User, error) {
	var user models.User
	err := s.db.QueryRow(context.Background(),
		"SELECT id, email, moodle_login, name, role, created_at, updated_at FROM users WHERE id = $1",
		userID,
	).Scan(&user.ID, &user.Email, &user.MoodleLogin, &user.Name, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, errors.New("user not found")
	}

	user.RoleID = user.GetRoleID()
	return &user, nil
}

// GetUserByEmail returns a user by email.
func (s *UserService) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	err := s.db.QueryRow(context.Background(),
		"SELECT id, email, moodle_login, name, role, created_at, updated_at FROM users WHERE email = $1",
		strings.TrimSpace(email),
	).Scan(&user.ID, &user.Email, &user.MoodleLogin, &user.Name, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, errors.New("user not found")
	}

	user.RoleID = user.GetRoleID()
	return &user, nil
}

// GetUserByMoodleLogin returns a user by Moodle login.
func (s *UserService) GetUserByMoodleLogin(moodleLogin string) (*models.User, error) {
	var user models.User
	err := s.db.QueryRow(context.Background(),
		"SELECT id, email, moodle_login, name, role, created_at, updated_at FROM users WHERE moodle_login = $1",
		strings.TrimSpace(moodleLogin),
	).Scan(&user.ID, &user.Email, &user.MoodleLogin, &user.Name, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, errors.New("user not found")
	}

	user.RoleID = user.GetRoleID()
	return &user, nil
}

// CreateUser creates a local user account with the "user" role.
func (s *UserService) CreateUser(email, password, name string) (*models.User, error) {
	return s.CreateUserWithProfile(CreateUserParams{
		Email:    email,
		Password: password,
		Name:     name,
		Role:     "user",
	})
}

// CreateAdmin creates a local admin account.
func (s *UserService) CreateAdmin(email, password, name string) (*models.User, error) {
	return s.CreateUserWithProfile(CreateUserParams{
		Email:    email,
		Password: password,
		Name:     name,
		Role:     "admin",
	})
}

// CreateManagedUser creates a platform user that is linked by Moodle login.
func (s *UserService) CreateManagedUser(name, moodleLogin, password string) (*models.User, error) {
	return s.CreateUserWithProfile(CreateUserParams{
		Name:        name,
		Password:    password,
		Role:        "user",
		MoodleLogin: moodleLogin,
	})
}

// CreateUserWithProfile creates a user with the provided profile details.
func (s *UserService) CreateUserWithProfile(params CreateUserParams) (*models.User, error) {
	email := strings.TrimSpace(params.Email)
	password := strings.TrimSpace(params.Password)
	name := strings.TrimSpace(params.Name)
	role := strings.TrimSpace(params.Role)
	moodleLogin := strings.TrimSpace(params.MoodleLogin)

	if role == "" {
		role = "user"
	}
	if role != "user" && role != "admin" {
		return nil, errors.New("invalid role")
	}
	if password == "" {
		return nil, errors.New("password is required")
	}
	if email == "" {
		if moodleLogin == "" {
			return nil, errors.New("email is required")
		}
		email = buildSyntheticEmailFromMoodleLogin(moodleLogin)
	}

	var exists bool
	err := s.db.QueryRow(context.Background(),
		"SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)",
		email,
	).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("user with this email already exists")
	}

	if moodleLogin != "" {
		err = s.db.QueryRow(context.Background(),
			"SELECT EXISTS(SELECT 1 FROM users WHERE moodle_login = $1)",
			moodleLogin,
		).Scan(&exists)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.New("user with this Moodle login already exists")
		}
	}

	passwordHash, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}

	var user models.User
	var namePtr *string
	if name != "" {
		namePtr = &name
	}
	var moodleLoginPtr *string
	if moodleLogin != "" {
		moodleLoginPtr = &moodleLogin
	}

	err = s.db.QueryRow(context.Background(),
		`INSERT INTO users (email, password_hash, name, role, moodle_login)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, email, moodle_login, name, role, created_at, updated_at`,
		email, passwordHash, namePtr, role, moodleLoginPtr,
	).Scan(&user.ID, &user.Email, &user.MoodleLogin, &user.Name, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}

	user.RoleID = user.GetRoleID()
	return &user, nil
}

// Authenticate verifies email and password.
func (s *UserService) Authenticate(email, password string) (*models.User, error) {
	var user models.UserWithPassword

	err := s.db.QueryRow(context.Background(),
		"SELECT id, email, moodle_login, password_hash, name, role, created_at, updated_at FROM users WHERE email = $1",
		strings.TrimSpace(email),
	).Scan(&user.ID, &user.Email, &user.MoodleLogin, &user.PasswordHash, &user.Name, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if !utils.CheckPasswordHash(password, user.PasswordHash) {
		return nil, errors.New("invalid credentials")
	}

	user.User.RoleID = user.User.GetRoleID()
	return &user.User, nil
}

func (s *UserService) UpdateMoodleLogin(userID uuid.UUID, moodleLogin string) error {
	normalized := strings.TrimSpace(moodleLogin)
	if normalized == "" {
		return errors.New("moodle login is required")
	}

	result, err := s.db.Exec(context.Background(),
		"UPDATE users SET moodle_login = $1, updated_at = NOW() WHERE id = $2",
		normalized, userID,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return errors.New("user not found")
	}

	return nil
}

// SaveRefreshToken stores the refresh token in the database.
func (s *UserService) SaveRefreshToken(userID uuid.UUID, token string, expiry time.Duration) error {
	_, err := s.db.Exec(context.Background(),
		"INSERT INTO refresh_tokens (user_id, token, expires_at) VALUES ($1, $2, $3)",
		userID, token, time.Now().Add(expiry),
	)
	return err
}

// ValidateRefreshToken validates the refresh token.
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

// DeleteRefreshToken removes the refresh token.
func (s *UserService) DeleteRefreshToken(token string) error {
	_, err := s.db.Exec(context.Background(),
		"DELETE FROM refresh_tokens WHERE token = $1",
		token,
	)
	return err
}

// UpdateUserRole updates the role of the user.
func (s *UserService) UpdateUserRole(userID uuid.UUID, role string) error {
	if role != "user" && role != "admin" {
		return errors.New("invalid role")
	}

	result, err := s.db.Exec(context.Background(),
		"UPDATE users SET role = $1, updated_at = NOW() WHERE id = $2",
		role, userID,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return errors.New("user not found")
	}

	return nil
}

func buildSyntheticEmailFromMoodleLogin(moodleLogin string) string {
	normalized := strings.ToLower(strings.TrimSpace(moodleLogin))
	normalized = moodleLoginEmailSanitizer.ReplaceAllString(normalized, "-")
	normalized = strings.Trim(normalized, "-.")
	if normalized == "" {
		normalized = uuid.NewString()
	}
	return fmt.Sprintf("%s@moodle.local", normalized)
}
