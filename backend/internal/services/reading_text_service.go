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

func NewReadingTextService(db *pgxpool.Pool) *ReadingTextService {
	return &ReadingTextService{db: db}
}

func (s *ReadingTextService) GetAllByUserID(userID uuid.UUID) ([]models.ReadingText, error) {
	rows, err := s.db.Query(context.Background(),
		`SELECT id, user_id, title, content, created_at, updated_at 
		 FROM reading_texts WHERE user_id = $1 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var texts []models.ReadingText
	for rows.Next() {
		var text models.ReadingText
		err := rows.Scan(&text.ID, &text.UserID, &text.Title, &text.Content, &text.CreatedAt, &text.UpdatedAt)
		if err != nil {
			return nil, err
		}
		texts = append(texts, text)
	}

	return texts, nil
}

func (s *ReadingTextService) GetByID(textID int64, userID uuid.UUID) (*models.ReadingText, error) {
	var text models.ReadingText
	err := s.db.QueryRow(context.Background(),
		`SELECT id, user_id, title, content, created_at, updated_at 
		 FROM reading_texts WHERE id = $1 AND user_id = $2`,
		textID, userID,
	).Scan(&text.ID, &text.UserID, &text.Title, &text.Content, &text.CreatedAt, &text.UpdatedAt)

	if err != nil {
		return nil, errors.New("text not found")
	}

	return &text, nil
}

func (s *ReadingTextService) Create(userID uuid.UUID, title, content string) (*models.ReadingText, error) {
	var text models.ReadingText
	err := s.db.QueryRow(context.Background(),
		`INSERT INTO reading_texts (user_id, title, content)
		 VALUES ($1, $2, $3)
		 RETURNING id, user_id, title, content, created_at, updated_at`,
		userID, title, content,
	).Scan(&text.ID, &text.UserID, &text.Title, &text.Content, &text.CreatedAt, &text.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return &text, nil
}

func (s *ReadingTextService) Delete(textID int64, userID uuid.UUID) error {
	result, err := s.db.Exec(context.Background(),
		"DELETE FROM reading_texts WHERE id = $1 AND user_id = $2",
		textID, userID,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("text not found")
	}

	return nil
}
