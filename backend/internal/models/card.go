package models

import (
	"time"

	"github.com/google/uuid"
)

// Card представляет карточку для изучения слова
type Card struct {
	ID          int64      `json:"id" db:"id"`
	DeckID      int64      `json:"deck_id" db:"deck_id"`
	Position    int        `json:"position" db:"position"`
	Word        string     `json:"word" db:"word"`
	Translation string    `json:"translation" db:"translation"`
	Phonetic    *string    `json:"phonetic,omitempty" db:"phonetic"`
	AudioURL    *string    `json:"audio_url,omitempty" db:"audio_url"`
	ImageURL    *string    `json:"image_url,omitempty" db:"image_url"`
	Example     *string    `json:"example,omitempty" db:"example"`
	CreatedBy   *uuid.UUID `json:"created_by,omitempty" db:"created_by"`
	IsCustom    bool       `json:"is_custom" db:"is_custom"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
}
