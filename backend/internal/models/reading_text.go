package models

import (
	"time"

	"github.com/google/uuid"
)

type ReadingText struct {
	ID          int64     `json:"id" db:"id"`
	UserID      uuid.UUID `json:"user_id" db:"user_id"`
	CourseID    int64     `json:"course_id" db:"course_id"`
	CourseTitle string    `json:"course_title,omitempty" db:"course_title"`
	CourseSlug  string    `json:"course_slug,omitempty" db:"course_slug"`
	Title       string    `json:"title" db:"title"`
	Content     string    `json:"content" db:"content"`
	AudioURL    string    `json:"audio_url" db:"audio_url"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type ReadingPodcast struct {
	ID          int64     `json:"id" db:"id"`
	UserID      uuid.UUID `json:"user_id" db:"user_id"`
	UserName    string    `json:"user_name" db:"user_name"`
	UserEmail   string    `json:"user_email" db:"user_email"`
	CourseID    int64     `json:"course_id" db:"course_id"`
	CourseTitle string   `json:"course_title" db:"course_title"`
	CourseSlug  string   `json:"course_slug" db:"course_slug"`
	Title       string    `json:"title" db:"title"`
	AudioURL    string    `json:"audio_url" db:"audio_url"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}
