package services

import (
	"context"
	"english-learning/internal/models"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ReadingTextService struct {
	db *pgxpool.Pool
}

type readingTextScanner interface {
	Scan(dest ...any) error
}

func NewReadingTextService(db *pgxpool.Pool) *ReadingTextService {
	return &ReadingTextService{db: db}
}

func scanReadingText(scanner readingTextScanner, text *models.ReadingText) error {
	return scanner.Scan(
		&text.ID,
		&text.UserID,
		&text.CourseID,
		&text.CourseTitle,
		&text.CourseSlug,
		&text.Title,
		&text.Content,
		&text.AudioURL,
		&text.CreatedAt,
		&text.UpdatedAt,
	)
}

func (s *ReadingTextService) GetAllAccessible(userID uuid.UUID, isAdmin bool) ([]models.ReadingText, error) {
	var (
		query string
		args  []any
	)

	if isAdmin {
		query = `SELECT
			rt.id,
			rt.user_id,
			COALESCE(rt.course_id, 0)::BIGINT AS course_id,
			COALESCE(c.title, '') AS course_title,
			COALESCE(c.slug, '') AS course_slug,
			rt.title,
			rt.content,
			COALESCE(rt.audio_url, '') AS audio_url,
			rt.created_at,
			rt.updated_at
		 FROM reading_texts rt
		 LEFT JOIN courses c ON c.id = rt.course_id
		 ORDER BY rt.created_at DESC`
	} else {
		query = `SELECT
			rt.id,
			rt.user_id,
			COALESCE(rt.course_id, 0)::BIGINT AS course_id,
			COALESCE(c.title, '') AS course_title,
			COALESCE(c.slug, '') AS course_slug,
			rt.title,
			rt.content,
			COALESCE(rt.audio_url, '') AS audio_url,
			rt.created_at,
			rt.updated_at
		 FROM reading_texts rt
		 JOIN courses c ON c.id = rt.course_id
		 JOIN course_accesses ca
		   ON ca.course_id = rt.course_id
		  AND ca.user_id = $1
		 WHERE c.is_published = true
		 ORDER BY rt.created_at DESC`
		args = append(args, userID)
	}

	rows, err := s.db.Query(context.Background(), query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var texts []models.ReadingText
	for rows.Next() {
		var text models.ReadingText
		if err := scanReadingText(rows, &text); err != nil {
			return nil, err
		}
		texts = append(texts, text)
	}

	return texts, rows.Err()
}

func (s *ReadingTextService) GetAccessibleByID(textID int64, userID uuid.UUID, isAdmin bool) (*models.ReadingText, error) {
	var (
		query string
		args  []any
	)

	if isAdmin {
		query = `SELECT
			rt.id,
			rt.user_id,
			COALESCE(rt.course_id, 0)::BIGINT AS course_id,
			COALESCE(c.title, '') AS course_title,
			COALESCE(c.slug, '') AS course_slug,
			rt.title,
			rt.content,
			COALESCE(rt.audio_url, '') AS audio_url,
			rt.created_at,
			rt.updated_at
		 FROM reading_texts rt
		 LEFT JOIN courses c ON c.id = rt.course_id
		 WHERE rt.id = $1`
		args = []any{textID}
	} else {
		query = `SELECT
			rt.id,
			rt.user_id,
			COALESCE(rt.course_id, 0)::BIGINT AS course_id,
			COALESCE(c.title, '') AS course_title,
			COALESCE(c.slug, '') AS course_slug,
			rt.title,
			rt.content,
			COALESCE(rt.audio_url, '') AS audio_url,
			rt.created_at,
			rt.updated_at
		 FROM reading_texts rt
		 JOIN courses c ON c.id = rt.course_id
		 JOIN course_accesses ca
		   ON ca.course_id = rt.course_id
		  AND ca.user_id = $2
		 WHERE rt.id = $1
		   AND c.is_published = true`
		args = []any{textID, userID}
	}

	var text models.ReadingText
	if err := scanReadingText(s.db.QueryRow(context.Background(), query, args...), &text); err != nil {
		return nil, errors.New("text not found")
	}

	return &text, nil
}

func (s *ReadingTextService) Create(userID uuid.UUID, courseID int64, title, content, audioURL string) (*models.ReadingText, error) {
	var textID int64
	if err := s.db.QueryRow(context.Background(),
		`INSERT INTO reading_texts (user_id, course_id, title, content, audio_url)
		 VALUES ($1, NULLIF($2, 0), $3, $4, $5)
		 RETURNING id`,
		userID, courseID, title, content, audioURL,
	).Scan(&textID); err != nil {
		return nil, err
	}

	return s.GetAccessibleByID(textID, userID, true)
}

func (s *ReadingTextService) Delete(textID int64) error {
	result, err := s.db.Exec(context.Background(),
		`DELETE FROM reading_texts
		 WHERE id = $1`,
		textID,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return errors.New("text not found")
	}

	return nil
}

func (s *ReadingTextService) UpdateAudio(textID int64, audioURL string) (*models.ReadingText, error) {
	var updatedID int64
	if err := s.db.QueryRow(context.Background(),
		`UPDATE reading_texts
		 SET audio_url = $2, updated_at = NOW()
		 WHERE id = $1
		 RETURNING id`,
		textID, audioURL,
	).Scan(&updatedID); err != nil {
		return nil, errors.New("text not found")
	}

	var userID uuid.UUID
	if err := s.db.QueryRow(context.Background(),
		`SELECT user_id
		 FROM reading_texts
		 WHERE id = $1`,
		updatedID,
	).Scan(&userID); err != nil {
		return nil, errors.New("text not found")
	}

	return s.GetAccessibleByID(updatedID, userID, true)
}

func (s *ReadingTextService) GetPodcasts(courseID int64) ([]models.ReadingPodcast, error) {
	query := `
SELECT
	rt.id,
	rt.user_id,
	COALESCE(u.name, '') AS user_name,
	COALESCE(u.email, '') AS user_email,
	COALESCE(rt.course_id, 0)::BIGINT AS course_id,
	COALESCE(c.title, '') AS course_title,
	COALESCE(c.slug, '') AS course_slug,
	rt.title,
	COALESCE(rt.audio_url, '') AS audio_url,
	rt.created_at,
	rt.updated_at
FROM reading_texts rt
JOIN users u ON u.id = rt.user_id
LEFT JOIN courses c ON c.id = rt.course_id
WHERE COALESCE(rt.audio_url, '') <> ''`

	args := []any{}
	if courseID > 0 {
		query += " AND rt.course_id = $1"
		args = append(args, courseID)
	}

	query += `
ORDER BY
	CASE WHEN COALESCE(c.title, '') = '' THEN 1 ELSE 0 END,
	COALESCE(c.title, ''),
	rt.created_at DESC`

	rows, err := s.db.Query(context.Background(), query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var podcasts []models.ReadingPodcast
	for rows.Next() {
		var podcast models.ReadingPodcast
		if err := rows.Scan(
			&podcast.ID,
			&podcast.UserID,
			&podcast.UserName,
			&podcast.UserEmail,
			&podcast.CourseID,
			&podcast.CourseTitle,
			&podcast.CourseSlug,
			&podcast.Title,
			&podcast.AudioURL,
			&podcast.CreatedAt,
			&podcast.UpdatedAt,
		); err != nil {
			return nil, err
		}
		podcasts = append(podcasts, podcast)
	}

	return podcasts, rows.Err()
}
