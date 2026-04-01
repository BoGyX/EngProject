package models

import (
	"time"

	"github.com/google/uuid"
)

// TrainingSession представляет сессию тренировки
type TrainingSession struct {
	ID         int64      `json:"id" db:"id"`
	UserID     *uuid.UUID `json:"user_id,omitempty" db:"user_id"`
	CourseID   *int64     `json:"course_id,omitempty" db:"course_id"`
	DeckID     *int64     `json:"deck_id,omitempty" db:"deck_id"`
	UserDeckID *int64     `json:"user_deck_id,omitempty" db:"user_deck_id"`
	StartedAt  time.Time  `json:"started_at" db:"started_at"`
	FinishedAt *time.Time `json:"finished_at,omitempty" db:"finished_at"`
}

// TrainingAnswer представляет ответ на карточку в сессии
type TrainingAnswer struct {
	ID         int64     `json:"id" db:"id"`
	SessionID  int64     `json:"session_id" db:"session_id"`
	CardID     *int64    `json:"card_id,omitempty" db:"card_id"`
	IsCorrect  *bool     `json:"is_correct,omitempty" db:"is_correct"`
	AnsweredAt time.Time `json:"answered_at" db:"answered_at"`
}

type TrainingSessionCard struct {
	ID                 int64     `json:"id" db:"id"`
	SessionID          int64     `json:"session_id" db:"session_id"`
	CardID             int64     `json:"card_id" db:"card_id"`
	SequenceNumber     int       `json:"sequence_number" db:"sequence_number"`
	CurrentMode        string    `json:"current_mode" db:"current_mode"`
	ProgressPercentage int       `json:"progress_percentage" db:"progress_percentage"`
	IsCompleted        bool      `json:"is_completed" db:"is_completed"`
	LastAnswerCorrect  *bool     `json:"last_answer_correct,omitempty" db:"last_answer_correct"`
	CorrectAnswers     int       `json:"correct_answers" db:"correct_answers"`
	WrongAnswers       int       `json:"wrong_answers" db:"wrong_answers"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time `json:"updated_at" db:"updated_at"`
}

type TrainingSessionCardState struct {
	SessionCardID      int64    `json:"session_card_id"`
	CardID             int64    `json:"card_id"`
	DeckID             int64    `json:"deck_id"`
	Word               string   `json:"word"`
	Translation        string   `json:"translation"`
	Phonetic           *string  `json:"phonetic,omitempty"`
	AudioURL           *string  `json:"audio_url,omitempty"`
	ImageURL           *string  `json:"image_url,omitempty"`
	Example            *string  `json:"example,omitempty"`
	IsCustom           bool     `json:"is_custom"`
	SequenceNumber     int      `json:"sequence_number"`
	CurrentMode        string   `json:"current_mode"`
	ProgressPercentage int      `json:"progress_percentage"`
	IsCompleted        bool     `json:"is_completed"`
	Options            []string `json:"options,omitempty"`
}

type TrainingSessionState struct {
	Session        TrainingSession            `json:"session"`
	Cards          []TrainingSessionCardState `json:"cards"`
	CurrentCard    *TrainingSessionCardState  `json:"current_card,omitempty"`
	RemainingCards int                        `json:"remaining_cards"`
}
