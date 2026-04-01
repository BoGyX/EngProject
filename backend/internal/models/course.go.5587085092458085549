package models

import (
	"time"

	"github.com/google/uuid"
)

// Course представляет курс обучения
type Course struct {
	ID          int64      `json:"id" db:"id"`
	Title       string     `json:"title" db:"title"`
	Slug        string     `json:"slug" db:"slug"`
	Description *string    `json:"description,omitempty" db:"description"`
	ImageURL    *string    `json:"image_url,omitempty" db:"image_url"`
	IsPublished bool       `json:"is_published" db:"is_published"`
	CreatedBy   *uuid.UUID `json:"created_by,omitempty" db:"created_by"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
}
