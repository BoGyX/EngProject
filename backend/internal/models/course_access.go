package models

import (
	"time"

	"github.com/google/uuid"
)

// CourseAccess хранит выданный пользователю доступ к курсу.
type CourseAccess struct {
	UserID    uuid.UUID  `json:"user_id" db:"user_id"`
	CourseID  int64      `json:"course_id" db:"course_id"`
	GrantedBy *uuid.UUID `json:"granted_by,omitempty" db:"granted_by"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
}
