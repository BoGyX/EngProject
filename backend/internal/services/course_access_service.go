package services

import (
	"context"
	"english-learning/internal/models"
	"errors"
	"sort"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CourseAccessService struct {
	db *pgxpool.Pool
}

func NewCourseAccessService(db *pgxpool.Pool) *CourseAccessService {
	return &CourseAccessService{db: db}
}

func (s *CourseAccessService) GetAllCourseAccesses() ([]models.CourseAccess, error) {
	rows, err := s.db.Query(context.Background(),
		`SELECT user_id, course_id, granted_by, created_at
		 FROM course_accesses
		 ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accesses []models.CourseAccess
	for rows.Next() {
		var access models.CourseAccess
		if err := rows.Scan(&access.UserID, &access.CourseID, &access.GrantedBy, &access.CreatedAt); err != nil {
			return nil, err
		}
		accesses = append(accesses, access)
	}

	return accesses, rows.Err()
}

func (s *CourseAccessService) GetAssignedCourseIDsByUserID(userID uuid.UUID) ([]int64, error) {
	rows, err := s.db.Query(context.Background(),
		`SELECT course_id
		 FROM course_accesses
		 WHERE user_id = $1
		 ORDER BY created_at DESC, course_id ASC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var courseIDs []int64
	for rows.Next() {
		var courseID int64
		if err := rows.Scan(&courseID); err != nil {
			return nil, err
		}
		courseIDs = append(courseIDs, courseID)
	}

	return courseIDs, rows.Err()
}

func (s *CourseAccessService) GetAccessibleCoursesByUserID(userID uuid.UUID, includeUnpublished bool) ([]models.Course, error) {
	rows, err := s.db.Query(context.Background(),
		`SELECT c.id, c.title, c.slug, c.description, c.image_url, c.is_published,
		        NULLIF(BTRIM(c.created_by::text), '') AS created_by,
		        c.created_at
		 FROM courses c
		 JOIN course_accesses ca
		   ON ca.course_id = c.id
		 WHERE ca.user_id = $1
		   AND ($2::boolean = true OR c.is_published = true)
		 ORDER BY c.created_at DESC`,
		userID, includeUnpublished,
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

func (s *CourseAccessService) UserHasAccess(userID uuid.UUID, courseID int64) (bool, error) {
	var hasAccess bool
	err := s.db.QueryRow(context.Background(),
		`SELECT EXISTS(
			SELECT 1
			FROM users u
			WHERE u.id = $1
			  AND (
				u.role = 'admin'
				OR EXISTS(
					SELECT 1
					FROM course_accesses
					WHERE user_id = $1 AND course_id = $2
				)
			  )
		)`,
		userID, courseID,
	).Scan(&hasAccess)
	return hasAccess, err
}

func (s *CourseAccessService) EnsureUserHasAccess(userID uuid.UUID, courseID int64) error {
	hasAccess, err := s.UserHasAccess(userID, courseID)
	if err != nil {
		return err
	}
	if !hasAccess {
		return errors.New("course access denied")
	}
	return nil
}

func (s *CourseAccessService) ReplaceUserCourseAccess(userID uuid.UUID, courseIDs []int64, grantedBy *uuid.UUID) ([]int64, error) {
	normalized := normalizeCourseIDs(courseIDs)
	ctx := context.Background()

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if len(normalized) == 0 {
		if _, err := tx.Exec(ctx,
			`DELETE FROM course_accesses
			 WHERE user_id = $1`,
			userID,
		); err != nil {
			return nil, err
		}
	} else {
		if _, err := tx.Exec(ctx,
			`DELETE FROM course_accesses
			 WHERE user_id = $1
			   AND NOT (course_id = ANY($2::bigint[]))`,
			userID, normalized,
		); err != nil {
			return nil, err
		}

		for _, courseID := range normalized {
			if _, err := tx.Exec(ctx,
				`INSERT INTO course_accesses (user_id, course_id, granted_by)
				 VALUES ($1, $2, $3)
				 ON CONFLICT (user_id, course_id)
				 DO UPDATE SET granted_by = EXCLUDED.granted_by`,
				userID, courseID, grantedBy,
			); err != nil {
				return nil, err
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return normalized, nil
}

func normalizeCourseIDs(courseIDs []int64) []int64 {
	unique := make(map[int64]struct{}, len(courseIDs))
	for _, courseID := range courseIDs {
		if courseID <= 0 {
			continue
		}
		unique[courseID] = struct{}{}
	}

	normalized := make([]int64, 0, len(unique))
	for courseID := range unique {
		normalized = append(normalized, courseID)
	}

	sort.Slice(normalized, func(i, j int) bool {
		return normalized[i] < normalized[j]
	})

	return normalized
}

func (s *CourseAccessService) GrantCourseAccess(userID uuid.UUID, courseID int64, grantedBy *uuid.UUID) (*models.CourseAccess, error) {
	var access models.CourseAccess
	err := s.db.QueryRow(context.Background(),
		`INSERT INTO course_accesses (user_id, course_id, granted_by)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (user_id, course_id)
		 DO UPDATE SET granted_by = EXCLUDED.granted_by
		 RETURNING user_id, course_id, granted_by, created_at`,
		userID, courseID, grantedBy,
	).Scan(&access.UserID, &access.CourseID, &access.GrantedBy, &access.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &access, nil
}

func (s *CourseAccessService) RevokeCourseAccess(userID uuid.UUID, courseID int64) error {
	result, err := s.db.Exec(context.Background(),
		`DELETE FROM course_accesses
		 WHERE user_id = $1 AND course_id = $2`,
		userID, courseID,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return errors.New("course access not found")
	}

	return nil
}
