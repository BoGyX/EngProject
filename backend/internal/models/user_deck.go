package models

import (
	"time"

	"github.com/google/uuid"
)

// UserDeck представляет прогресс пользователя по deck (уроку)
type UserDeck struct {
	ID                 int64      `json:"id" db:"id"`
	UserID             uuid.UUID  `json:"user_id" db:"user_id"`
	DeckID             int64      `json:"deck_id" db:"deck_id"`
	UserCourseID       *int64     `json:"user_course_id,omitempty" db:"user_course_id"`
	Status             string     `json:"status" db:"status"` // 'not_started', 'in_progress', 'completed'
	LearnedCardsCount  int        `json:"learned_cards_count" db:"learned_cards_count"`
	TotalCardsCount    int        `json:"total_cards_count" db:"total_cards_count"`
	ProgressPercentage float64    `json:"progress_percentage" db:"progress_percentage"`
	StartedAt          *time.Time `json:"started_at,omitempty" db:"started_at"`
	LastOpenedAt       *time.Time `json:"last_opened_at,omitempty" db:"last_opened_at"`
	CompletedAt        *time.Time `json:"completed_at,omitempty" db:"completed_at"`
	IsActive           bool       `json:"is_active" db:"is_active"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at" db:"updated_at"`
}
