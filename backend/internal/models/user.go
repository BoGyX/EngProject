package models

import (
	"time"

	"github.com/google/uuid"
)

// User представляет пользователя системы
type User struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Email     string    `json:"email" db:"email"`
	Name      *string   `json:"name,omitempty" db:"name"`
	Role      string    `json:"role" db:"role"`        // 'user' или 'admin' (в БД)
	RoleID    int       `json:"role_id"`               // 1 для user, 2 для admin (вычисляемое)
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// GetRoleID возвращает числовой ID роли (1 = user, 2 = admin)
func (u *User) GetRoleID() int {
	if u.Role == "admin" {
		return 2
	}
	return 1 // user
}

// SetRoleID устанавливает роль по числовому ID
func (u *User) SetRoleID(roleID int) {
	if roleID == 2 {
		u.Role = "admin"
	} else {
		u.Role = "user"
	}
	u.RoleID = roleID
}

// UserWithPassword для внутреннего использования (с паролем)
type UserWithPassword struct {
	User
	PasswordHash string `db:"password_hash"`
}
