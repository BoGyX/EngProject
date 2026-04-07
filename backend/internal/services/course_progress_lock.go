package services

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type CourseProgressLockError struct {
	BlockingUserCourseID int64
	BlockingCourseID     int64
	BlockingCourseTitle  string
	ProgressPercentage   float64
}

func (e *CourseProgressLockError) Error() string {
	progress := int(math.Round(e.ProgressPercentage))
	if progress < 0 {
		progress = 0
	}
	if progress > 100 {
		progress = 100
	}

	if e.BlockingCourseTitle != "" {
		return fmt.Sprintf("Сначала завершите курс \"%s\" на 100%%. Сейчас: %d%%.", e.BlockingCourseTitle, progress)
	}

	return fmt.Sprintf("Сначала завершите текущий курс на 100%%. Сейчас: %d%%.", progress)
}

func IsCourseProgressLockError(err error) bool {
	var lockErr *CourseProgressLockError
	return errors.As(err, &lockErr)
}

type courseProgressLockQuerier interface {
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
}

func findBlockingCourseProgress(ctx context.Context, q courseProgressLockQuerier, userID uuid.UUID, targetCourseID int64) (*CourseProgressLockError, error) {
	var lockErr CourseProgressLockError

	err := q.QueryRow(ctx, `
		SELECT
			uc.id,
			uc.course_id,
			c.title,
			COALESCE(uc.progress_percentage, 0)
		FROM user_courses uc
		JOIN courses c ON c.id = uc.course_id
		JOIN (
			SELECT course_id, MAX(attempt_number) AS max_attempt_number
			FROM user_courses
			WHERE user_id = $1
			GROUP BY course_id
		) latest
		  ON latest.course_id = uc.course_id
		 AND latest.max_attempt_number = uc.attempt_number
		WHERE uc.user_id = $1
		  AND uc.course_id <> $2
		  AND COALESCE(uc.progress_percentage, 0) < 100
		ORDER BY uc.is_active DESC, uc.last_opened_at DESC NULLS LAST, uc.started_at DESC, uc.id DESC
		LIMIT 1
	`, userID, targetCourseID).Scan(
		&lockErr.BlockingUserCourseID,
		&lockErr.BlockingCourseID,
		&lockErr.BlockingCourseTitle,
		&lockErr.ProgressPercentage,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &lockErr, nil
}
