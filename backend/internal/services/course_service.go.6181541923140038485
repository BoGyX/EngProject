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

// GetAllCourses возвращает список всех курсов
func (s *CourseService) GetAllCourses() ([]models.Course, error) {
	rows, err := s.db.Query(context.Background(),
		"SELECT id, title, slug, description, image_url, is_published, created_by, created_at FROM courses ORDER BY created_at DESC",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var courses []models.Course
	for rows.Next() {
		var course models.Course
		err := rows.Scan(&course.ID, &course.Title, &course.Slug, &course.Description, &course.ImageURL, &course.IsPublished, &course.CreatedBy, &course.CreatedAt)
		if err != nil {
			return nil, err
		}
		courses = append(courses, course)
	}

	return courses, nil
}

// GetPublishedCourses возвращает только опубликованные курсы
func (s *CourseService) GetPublishedCourses() ([]models.Course, error) {
	rows, err := s.db.Query(context.Background(),
		"SELECT id, title, slug, description, image_url, is_published, created_by, created_at FROM courses WHERE is_published = true ORDER BY created_at DESC",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var courses []models.Course
	for rows.Next() {
		var course models.Course
		err := rows.Scan(&course.ID, &course.Title, &course.Slug, &course.Description, &course.ImageURL, &course.IsPublished, &course.CreatedBy, &course.CreatedAt)
		if err != nil {
			return nil, err
		}
		courses = append(courses, course)
	}

	return courses, nil
}

// GetCourseByID возвращает курс по ID
func (s *CourseService) GetCourseByID(courseID int64) (*models.Course, error) {
	var course models.Course
	err := s.db.QueryRow(context.Background(),
		"SELECT id, title, slug, description, image_url, is_published, created_by, created_at FROM courses WHERE id = $1",
		courseID,
	).Scan(&course.ID, &course.Title, &course.Slug, &course.Description, &course.ImageURL, &course.IsPublished, &course.CreatedBy, &course.CreatedAt)

	if err != nil {
		return nil, errors.New("course not found")
	}

	return &course, nil
}

// CreateCourse создает новый курс
func (s *CourseService) CreateCourse(title string, slug string, description *string, imageURL *string, isPublished bool, createdBy *uuid.UUID) (*models.Course, error) {
	uniqueSlug, err := s.ensureUniqueSlug(context.Background(), slug, title, nil)
	if err != nil {
		return nil, err
	}

	var course models.Course
	err = s.db.QueryRow(context.Background(),
		`INSERT INTO courses (title, slug, description, image_url, is_published, created_by)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, title, slug, description, image_url, is_published, created_by, created_at`,
		title, uniqueSlug, description, imageURL, isPublished, createdBy,
	).Scan(&course.ID, &course.Title, &course.Slug, &course.Description, &course.ImageURL, &course.IsPublished, &course.CreatedBy, &course.CreatedAt)

	if err != nil {
		return nil, err
	}

	return &course, nil
}

// UpdateCourse обновляет курс
func (s *CourseService) UpdateCourse(courseID int64, title string, slug string, description *string, imageURL *string, isPublished bool) (*models.Course, error) {
	uniqueSlug, err := s.ensureUniqueSlug(context.Background(), slug, title, &courseID)
	if err != nil {
		return nil, err
	}

	var course models.Course
	err = s.db.QueryRow(context.Background(),
		`UPDATE courses 
		 SET title = $1, slug = $2, description = $3, image_url = $4, is_published = $5
		 WHERE id = $6
		 RETURNING id, title, slug, description, image_url, is_published, created_by, created_at`,
		title, uniqueSlug, description, imageURL, isPublished, courseID,
	).Scan(&course.ID, &course.Title, &course.Slug, &course.Description, &course.ImageURL, &course.IsPublished, &course.CreatedBy, &course.CreatedAt)

	if err != nil {
		return nil, errors.New("course not found")
	}

	return &course, nil
}

// GetCourseBySlug возвращает курс по slug
func (s *CourseService) GetCourseBySlug(slug string) (*models.Course, error) {
	var course models.Course
	err := s.db.QueryRow(context.Background(),
		"SELECT id, title, slug, description, image_url, is_published, created_by, created_at FROM courses WHERE slug = $1",
		slug,
	).Scan(&course.ID, &course.Title, &course.Slug, &course.Description, &course.ImageURL, &course.IsPublished, &course.CreatedBy, &course.CreatedAt)

	if err != nil {
		return nil, errors.New("course not found")
	}

	return &course, nil
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

// DeleteCourse удаляет курс
func (s *CourseService) DeleteCourse(courseID int64) error {
	result, err := s.db.Exec(context.Background(),
		"DELETE FROM courses WHERE id = $1",
		courseID,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("course not found")
	}

	return nil
}
