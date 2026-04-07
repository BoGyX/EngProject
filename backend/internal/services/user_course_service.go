package services

import (
	"context"
	"english-learning/internal/models"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserCourseService struct {
	db *pgxpool.Pool
}

func NewUserCourseService(db *pgxpool.Pool) *UserCourseService {
	return &UserCourseService{db: db}
}

// GetAllUserCourses возвращает все записи прогресса пользователей по курсам
func (s *UserCourseService) GetAllUserCourses() ([]models.UserCourse, error) {
	rows, err := s.db.Query(context.Background(),
		`SELECT id, user_id, course_id, started_at, last_opened_at, completed_at, is_active, attempt_number, 
		 completed_decks_count, total_decks_count, progress_percentage 
		 FROM user_courses ORDER BY started_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userCourses []models.UserCourse
	for rows.Next() {
		var uc models.UserCourse
		err := rows.Scan(&uc.ID, &uc.UserID, &uc.CourseID, &uc.StartedAt, &uc.LastOpenedAt, &uc.CompletedAt,
			&uc.IsActive, &uc.AttemptNumber, &uc.CompletedDecksCount, &uc.TotalDecksCount, &uc.ProgressPercentage)
		if err != nil {
			return nil, err
		}
		userCourses = append(userCourses, uc)
	}

	return userCourses, rows.Err()
}

// GetUserCourseByID возвращает прогресс по курсу по ID
func (s *UserCourseService) GetUserCourseByID(id int64) (*models.UserCourse, error) {
	var uc models.UserCourse
	err := s.db.QueryRow(context.Background(),
		`SELECT id, user_id, course_id, started_at, last_opened_at, completed_at, is_active, attempt_number, 
		 completed_decks_count, total_decks_count, progress_percentage 
		 FROM user_courses WHERE id = $1`,
		id,
	).Scan(&uc.ID, &uc.UserID, &uc.CourseID, &uc.StartedAt, &uc.LastOpenedAt, &uc.CompletedAt,
		&uc.IsActive, &uc.AttemptNumber, &uc.CompletedDecksCount, &uc.TotalDecksCount, &uc.ProgressPercentage)

	if err != nil {
		return nil, errors.New("user course not found")
	}

	return &uc, nil
}

// GetUserCoursesByUserID возвращает все курсы пользователя
func (s *UserCourseService) GetUserCoursesByUserID(userID uuid.UUID) ([]models.UserCourse, error) {
	rows, err := s.db.Query(context.Background(),
		`SELECT id, user_id, course_id, started_at, last_opened_at, completed_at, is_active, attempt_number, 
		 completed_decks_count, total_decks_count, progress_percentage 
		 FROM user_courses WHERE user_id = $1 ORDER BY started_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userCourses []models.UserCourse
	for rows.Next() {
		var uc models.UserCourse
		err := rows.Scan(&uc.ID, &uc.UserID, &uc.CourseID, &uc.StartedAt, &uc.LastOpenedAt, &uc.CompletedAt,
			&uc.IsActive, &uc.AttemptNumber, &uc.CompletedDecksCount, &uc.TotalDecksCount, &uc.ProgressPercentage)
		if err != nil {
			return nil, err
		}
		userCourses = append(userCourses, uc)
	}

	return userCourses, rows.Err()
}

// StartCourse создает новую запись о начале курса пользователем
func (s *UserCourseService) StartCourse(userID uuid.UUID, courseID int64) (*models.UserCourse, error) {
	if lockErr, err := findBlockingCourseProgress(context.Background(), s.db, userID, courseID); err != nil {
		return nil, err
	} else if lockErr != nil {
		return nil, lockErr
	}
	// Получаем количество дек в курсе
	var totalDecks int
	err := s.db.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM decks WHERE course_id = $1",
		courseID,
	).Scan(&totalDecks)
	if err != nil {
		return nil, err
	}

	// Получаем номер попытки
	var attemptNumber int
	err = s.db.QueryRow(context.Background(),
		"SELECT COALESCE(MAX(attempt_number), 0) + 1 FROM user_courses WHERE user_id = $1 AND course_id = $2",
		userID, courseID,
	).Scan(&attemptNumber)
	if err != nil {
		attemptNumber = 1
	}

	var uc models.UserCourse
	err = s.db.QueryRow(context.Background(),
		`INSERT INTO user_courses (user_id, course_id, started_at, attempt_number, 
		 last_opened_at, is_active, completed_decks_count, total_decks_count, progress_percentage)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 RETURNING id, user_id, course_id, started_at, last_opened_at, completed_at, is_active, attempt_number, 
		 completed_decks_count, total_decks_count, progress_percentage`,
		userID, courseID, time.Now(), attemptNumber, time.Now(), false, 0, totalDecks, 0.0,
	).Scan(&uc.ID, &uc.UserID, &uc.CourseID, &uc.StartedAt, &uc.LastOpenedAt, &uc.CompletedAt,
		&uc.IsActive, &uc.AttemptNumber, &uc.CompletedDecksCount, &uc.TotalDecksCount, &uc.ProgressPercentage)

	if err != nil {
		return nil, err
	}

	return &uc, nil
}

// UpdateUserCourse обновляет прогресс пользователя по курсу
func (s *UserCourseService) UpdateUserCourse(id int64, completedDecksCount int, totalDecksCount int, progressPercentage float64) (*models.UserCourse, error) {
	var uc models.UserCourse
	var completedAt *time.Time

	// Если все деки завершены, устанавливаем completed_at
	if completedDecksCount >= totalDecksCount && totalDecksCount > 0 {
		now := time.Now()
		completedAt = &now
	}

	err := s.db.QueryRow(context.Background(),
		`UPDATE user_courses 
		 SET completed_decks_count = $1, total_decks_count = $2, progress_percentage = $3, completed_at = $4, last_opened_at = NOW()
		 WHERE id = $5
		 RETURNING id, user_id, course_id, started_at, last_opened_at, completed_at, is_active, attempt_number, 
		 completed_decks_count, total_decks_count, progress_percentage`,
		completedDecksCount, totalDecksCount, progressPercentage, completedAt, id,
	).Scan(&uc.ID, &uc.UserID, &uc.CourseID, &uc.StartedAt, &uc.LastOpenedAt, &uc.CompletedAt,
		&uc.IsActive, &uc.AttemptNumber, &uc.CompletedDecksCount, &uc.TotalDecksCount, &uc.ProgressPercentage)

	if err != nil {
		return nil, errors.New("user course not found")
	}

	return &uc, nil
}

// DeleteUserCourse удаляет запись о прогрессе по курсу
func (s *UserCourseService) DeleteUserCourse(id int64) error {
	result, err := s.db.Exec(context.Background(),
		"DELETE FROM user_courses WHERE id = $1",
		id,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("user course not found")
	}

	return nil
}

func (s *UserCourseService) GetLatestUserCourseByCourseID(userID uuid.UUID, courseID int64) (*models.UserCourse, error) {
	var uc models.UserCourse
	err := s.db.QueryRow(context.Background(),
		`SELECT id, user_id, course_id, started_at, last_opened_at, completed_at, is_active, attempt_number,
		        completed_decks_count, total_decks_count, progress_percentage
		 FROM user_courses
		 WHERE user_id = $1 AND course_id = $2
		 ORDER BY attempt_number DESC
		 LIMIT 1`,
		userID, courseID,
	).Scan(&uc.ID, &uc.UserID, &uc.CourseID, &uc.StartedAt, &uc.LastOpenedAt, &uc.CompletedAt, &uc.IsActive,
		&uc.AttemptNumber, &uc.CompletedDecksCount, &uc.TotalDecksCount, &uc.ProgressPercentage)
	if err != nil {
		return nil, errors.New("user course not found")
	}

	return &uc, nil
}

func (s *UserCourseService) GetActiveUserCourse(userID uuid.UUID) (*models.UserCourse, error) {
	var uc models.UserCourse
	err := s.db.QueryRow(context.Background(),
		`SELECT id, user_id, course_id, started_at, last_opened_at, completed_at, is_active, attempt_number,
		        completed_decks_count, total_decks_count, progress_percentage
		 FROM user_courses
		 WHERE user_id = $1 AND is_active = true
		 ORDER BY last_opened_at DESC NULLS LAST, started_at DESC
		 LIMIT 1`,
		userID,
	).Scan(&uc.ID, &uc.UserID, &uc.CourseID, &uc.StartedAt, &uc.LastOpenedAt, &uc.CompletedAt, &uc.IsActive,
		&uc.AttemptNumber, &uc.CompletedDecksCount, &uc.TotalDecksCount, &uc.ProgressPercentage)
	if err != nil {
		return nil, errors.New("active user course not found")
	}

	return &uc, nil
}

func (s *UserCourseService) ActivateCourse(userID uuid.UUID, courseID int64) (*models.UserCourse, error) {
	ctx := context.Background()
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	if lockErr, err := findBlockingCourseProgress(ctx, tx, userID, courseID); err != nil {
		return nil, err
	} else if lockErr != nil {
		return nil, lockErr
	}

	var course models.UserCourse
	err = tx.QueryRow(ctx,
		`SELECT id, user_id, course_id, started_at, last_opened_at, completed_at, is_active, attempt_number,
		        completed_decks_count, total_decks_count, progress_percentage
		 FROM user_courses
		 WHERE user_id = $1 AND course_id = $2
		 ORDER BY attempt_number DESC
		 LIMIT 1`,
		userID, courseID,
	).Scan(&course.ID, &course.UserID, &course.CourseID, &course.StartedAt, &course.LastOpenedAt, &course.CompletedAt, &course.IsActive,
		&course.AttemptNumber, &course.CompletedDecksCount, &course.TotalDecksCount, &course.ProgressPercentage)
	if err != nil {
		var totalDecks int
		if err := tx.QueryRow(ctx, "SELECT COUNT(*) FROM decks WHERE course_id = $1", courseID).Scan(&totalDecks); err != nil {
			return nil, err
		}
		err = tx.QueryRow(ctx,
			`INSERT INTO user_courses (
				user_id, course_id, started_at, last_opened_at, is_active, attempt_number,
				completed_decks_count, total_decks_count, progress_percentage
			) VALUES ($1, $2, NOW(), NOW(), true, 1, 0, $3, 0)
			RETURNING id, user_id, course_id, started_at, last_opened_at, completed_at, is_active, attempt_number,
			          completed_decks_count, total_decks_count, progress_percentage`,
			userID, courseID, totalDecks,
		).Scan(&course.ID, &course.UserID, &course.CourseID, &course.StartedAt, &course.LastOpenedAt, &course.CompletedAt, &course.IsActive,
			&course.AttemptNumber, &course.CompletedDecksCount, &course.TotalDecksCount, &course.ProgressPercentage)
		if err != nil {
			return nil, err
		}
	} else {
		if _, err := tx.Exec(ctx, "UPDATE user_courses SET is_active = false WHERE user_id = $1", userID); err != nil {
			return nil, err
		}
		err = tx.QueryRow(ctx,
			`UPDATE user_courses
			 SET is_active = true, last_opened_at = NOW()
			 WHERE id = $1
			 RETURNING id, user_id, course_id, started_at, last_opened_at, completed_at, is_active, attempt_number,
			           completed_decks_count, total_decks_count, progress_percentage`,
			course.ID,
		).Scan(&course.ID, &course.UserID, &course.CourseID, &course.StartedAt, &course.LastOpenedAt, &course.CompletedAt, &course.IsActive,
			&course.AttemptNumber, &course.CompletedDecksCount, &course.TotalDecksCount, &course.ProgressPercentage)
		if err != nil {
			return nil, err
		}
	}

	if _, err := tx.Exec(ctx, "UPDATE user_courses SET is_active = false WHERE user_id = $1 AND id <> $2", userID, course.ID); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &course, nil
}
