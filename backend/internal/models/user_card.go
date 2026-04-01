package models

import (
	"time"

	"github.com/google/uuid"
)

// UserCard представляет прогресс пользователя по карточке
type UserCard struct {
	ID           int64      `json:"id" db:"id"`
	UserID       uuid.UUID  `json:"user_id" db:"user_id"`
	CardID       int64      `json:"card_id" db:"card_id"`
	UserDeckID   *int64     `json:"user_deck_id,omitempty" db:"user_deck_id"`
	Status       string     `json:"status" db:"status"` // 'new', 'learning', 'learned'
	CorrectCount int        `json:"correct_count" db:"correct_count"`
	WrongCount   int        `json:"wrong_count" db:"wrong_count"`
	LastSeen     *time.Time `json:"last_seen,omitempty" db:"last_seen"`
	NextReview   *time.Time `json:"next_review,omitempty" db:"next_review"`

	// Прогресс по режимам обучения
	ModeView           bool   `json:"mode_view" db:"mode_view"`
	ModeChoice         bool   `json:"mode_choice" db:"mode_choice"`
	ModeWithPhoto      bool   `json:"mode_with_photo" db:"mode_with_photo"`
	ModeWithoutPhoto   bool   `json:"mode_without_photo" db:"mode_without_photo"`
	ModeRussian        bool   `json:"mode_russian" db:"mode_russian"`
	ModeConstructor    bool   `json:"mode_constructor" db:"mode_constructor"`
	CurrentMode        string `json:"current_mode" db:"current_mode"`
	ProgressPercentage int    `json:"progress_percentage" db:"progress_percentage"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
