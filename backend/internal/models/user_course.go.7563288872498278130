package models

import (
	"time"

	"github.com/google/uuid"
)

// UserCourse представляет прогресс пользователя по курсу
type UserCourse struct {
	ID                  int64      `json:"id" db:"id"`
	UserID              uuid.UUID  `json:"user_id" db:"user_id"`
	CourseID            int64      `json:"course_id" db:"course_id"`
	StartedAt           time.Time  `json:"started_at" db:"started_at"`
	LastOpenedAt        *time.Time `json:"last_opened_at,omitempty" db:"last_opened_at"`
	CompletedAt         *time.Time `json:"completed_at,omitempty" db:"completed_at"`
	IsActive            bool       `json:"is_active" db:"is_active"`
	AttemptNumber       int        `json:"attempt_number" db:"attempt_number"`
	CompletedDecksCount int        `json:"completed_decks_count" db:"completed_decks_count"`
	TotalDecksCount     int        `json:"total_decks_count" db:"total_decks_count"`
	ProgressPercentage  float64    `json:"progress_percentage" db:"progress_percentage"`
}
