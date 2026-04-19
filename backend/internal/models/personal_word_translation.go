package models

import (
	"time"

	"github.com/google/uuid"
)

type PersonalWordTranslation struct {
	ID          int64     `json:"id" db:"id"`
	UserID      uuid.UUID `json:"user_id" db:"user_id"`
	Word        string    `json:"word" db:"word"`
	Translation string    `json:"translation" db:"translation"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}
