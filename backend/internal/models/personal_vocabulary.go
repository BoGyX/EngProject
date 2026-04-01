package models

import (
	"time"

	"github.com/google/uuid"
)

// PersonalVocabulary представляет слово в личном словаре пользователя
type PersonalVocabulary struct {
	ID           int64      `json:"id" db:"id"`
	UserID       uuid.UUID  `json:"user_id" db:"user_id"`
	Word         string     `json:"word" db:"word"`
	Translation  string     `json:"translation" db:"translation"`
	Phonetic     *string    `json:"phonetic,omitempty" db:"phonetic"`
	AudioURL     *string    `json:"audio_url,omitempty" db:"audio_url"`
	Example      *string    `json:"example,omitempty" db:"example"`
	Notes        *string    `json:"notes,omitempty" db:"notes"`
	Tags         []string   `json:"tags,omitempty" db:"tags"`
	Status       string     `json:"status" db:"status"` // 'new', 'learning', 'learned'
	CorrectCount int        `json:"correct_count" db:"correct_count"`
	WrongCount   int        `json:"wrong_count" db:"wrong_count"`
	LastSeen     *time.Time `json:"last_seen,omitempty" db:"last_seen"`
	NextReview   *time.Time `json:"next_review,omitempty" db:"next_review"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}
