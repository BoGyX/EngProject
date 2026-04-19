package services

import (
	"context"
	"english-learning/internal/models"
	"english-learning/internal/utils"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CourseService struct {
	db *pgxpool.Pool
}

func NewCourseService(db *pgxpool.Pool) *CourseService {
	return &CourseService{db: db}
}

type courseScanner interface {
	Scan(dest ...any) error
}

func scanCourse(scanner courseScanner) (*models.Course, error) {
	var course models.Course
	var createdByRaw *string

	if err := scanner.Scan(
		&course.ID,
		&course.Title,
		&course.Slug,
		&course.Description,
		&course.ImageURL,
		&course.IsPublished,
		&createdByRaw,
		&course.CreatedAt,
	); err != nil {
		return nil, err
	}

	if createdByRaw != nil {
		if trimmed := strings.TrimSpace(*createdByRaw); trimmed != "" {
			if parsed, err := uuid.Parse(trimmed); err == nil {
				course.CreatedBy = &parsed
			}
		}
	}

	return &course, nil
}

func courseSelectQuery(base string) string {
	return base + `
SELECT id, title, slug, description, image_url, is_published,
       NULLIF(BTRIM(created_by::text), '') AS created_by,
       created_at
  FROM courses`
}

// GetAllCourses returns all courses.
func (s *CourseService) GetAllCourses() ([]models.Course, error) {
	rows, err := s.db.Query(context.Background(),
		courseSelectQuery("")+`
 ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var courses []models.Course
	for rows.Next() {
		course, err := scanCourse(rows)
		if err != nil {
			return nil, err
		}
		courses = append(courses, *course)
	}

	return courses, rows.Err()
}

// GetPublishedCourses returns only published courses.
func (s *CourseService) GetPublishedCourses() ([]models.Course, error) {
	rows, err := s.db.Query(context.Background(),
		courseSelectQuery("")+`
 WHERE is_published = true
 ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var courses []models.Course
	for rows.Next() {
		course, err := scanCourse(rows)
		if err != nil {
			return nil, err
		}
		courses = append(courses, *course)
	}

	return courses, rows.Err()
}

// GetCourseByID returns a course by ID.
func (s *CourseService) GetCourseByID(courseID int64) (*models.Course, error) {
	course, err := scanCourse(s.db.QueryRow(context.Background(),
		courseSelectQuery("")+`
 WHERE id = $1`,
		courseID,
	))
	if err != nil {
		return nil, errors.New("course not found")
	}

	return course, nil
}

// CreateCourse creates a new course.
func (s *CourseService) CreateCourse(title string, slug string, description *string, imageURL *string, isPublished bool, createdBy *uuid.UUID) (*models.Course, error) {
	uniqueSlug, err := s.ensureUniqueSlug(context.Background(), slug, title, nil)
	if err != nil {
		return nil, err
	}

	course, err := scanCourse(s.db.QueryRow(context.Background(),
		`INSERT INTO courses (title, slug, description, image_url, is_published, created_by)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, title, slug, description, image_url, is_published,
		           NULLIF(BTRIM(created_by::text), '') AS created_by,
		           created_at`,
		title, uniqueSlug, description, imageURL, isPublished, createdBy,
	))
	if err != nil {
		return nil, err
	}

	return course, nil
}

// UpdateCourse updates a course.
func (s *CourseService) UpdateCourse(courseID int64, title string, slug string, description *string, imageURL *string, isPublished bool) (*models.Course, error) {
	uniqueSlug, err := s.ensureUniqueSlug(context.Background(), slug, title, &courseID)
	if err != nil {
		return nil, err
	}

	course, err := scanCourse(s.db.QueryRow(context.Background(),
		`UPDATE courses
		    SET title = $1, slug = $2, description = $3, image_url = $4, is_published = $5
		  WHERE id = $6
		 RETURNING id, title, slug, description, image_url, is_published,
		           NULLIF(BTRIM(created_by::text), '') AS created_by,
		           created_at`,
		title, uniqueSlug, description, imageURL, isPublished, courseID,
	))
	if err != nil {
		return nil, errors.New("course not found")
	}

	return course, nil
}

// GetCourseBySlug returns a course by slug.
func (s *CourseService) GetCourseBySlug(slug string) (*models.Course, error) {
	course, err := scanCourse(s.db.QueryRow(context.Background(),
		courseSelectQuery("")+`
 WHERE slug = $1`,
		slug,
	))
	if err != nil {
		return nil, errors.New("course not found")
	}

	return course, nil
}

func (s *CourseService) ensureUniqueSlug(ctx context.Context, requestedSlug string, title string, excludeID *int64) (string, error) {
	baseSlug := strings.TrimSpace(requestedSlug)
	if baseSlug == "" {
		baseSlug = utils.Slugify(title, "course")
	} else {
		baseSlug = utils.Slugify(baseSlug, "course")
	}

	if baseSlug == "" {
		baseSlug = "course"
	}

	candidate := baseSlug
	suffix := 2

	for {
		var exists bool
		err := s.db.QueryRow(ctx,
			`SELECT EXISTS(
				SELECT 1
				FROM courses
				WHERE slug = $1
				  AND ($2::bigint IS NULL OR id <> $2)
			)`,
			candidate, excludeID,
		).Scan(&exists)
		if err != nil {
			return "", err
		}
		if !exists {
			return candidate, nil
		}
		candidate = fmt.Sprintf("%s-%d", baseSlug, suffix)
		suffix++
	}
}

// DeleteCourse deletes a course.
func (s *CourseService) DeleteCourse(courseID int64) error {
	ctx := context.Background()
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err = tx.Exec(ctx,
		`DELETE FROM training_answers ta
		 USING cards c
		 JOIN decks d ON d.id = c.deck_id
		 WHERE ta.card_id = c.id
		   AND d.course_id = $1`,
		courseID,
	); err != nil {
		return err
	}

	if _, err = tx.Exec(ctx,
		`DELETE FROM training_sessions
		 WHERE course_id = $1
		    OR deck_id IN (SELECT id FROM decks WHERE course_id = $1)`,
		courseID,
	); err != nil {
		return err
	}

	result, err := tx.Exec(ctx,
		"DELETE FROM courses WHERE id = $1",
		courseID,
	)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("course not found")
	}

	return tx.Commit(ctx)
}
