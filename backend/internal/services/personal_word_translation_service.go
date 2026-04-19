package services

import (
	"context"
	"english-learning/internal/models"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PersonalWordTranslationService struct {
	db *pgxpool.Pool
}

func NewPersonalWordTranslationService(db *pgxpool.Pool) *PersonalWordTranslationService {
	return &PersonalWordTranslationService{db: db}
}

func (s *PersonalWordTranslationService) GetByUserID(userID uuid.UUID) ([]models.PersonalWordTranslation, error) {
	rows, err := s.db.Query(context.Background(),
		`SELECT id, user_id, word, translation, created_at, updated_at
		 FROM personal_word_translations
		 WHERE user_id = $1
		 ORDER BY LOWER(word) ASC, created_at DESC, id DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var translations []models.PersonalWordTranslation
	for rows.Next() {
		var translation models.PersonalWordTranslation
		if err := rows.Scan(
			&translation.ID,
			&translation.UserID,
			&translation.Word,
			&translation.Translation,
			&translation.CreatedAt,
			&translation.UpdatedAt,
		); err != nil {
			return nil, err
		}
		translations = append(translations, translation)
	}

	return translations, rows.Err()
}

func (s *PersonalWordTranslationService) GetByUserIDAndWord(userID uuid.UUID, word string) ([]models.PersonalWordTranslation, error) {
	normalizedWord := normalizePersonalTranslationWord(word)
	if normalizedWord == "" {
		return []models.PersonalWordTranslation{}, nil
	}

	rows, err := s.db.Query(context.Background(),
		`SELECT id, user_id, word, translation, created_at, updated_at
		 FROM personal_word_translations
		 WHERE user_id = $1
		   AND LOWER(word) = LOWER($2)
		 ORDER BY created_at DESC, id DESC`,
		userID, normalizedWord,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var translations []models.PersonalWordTranslation
	for rows.Next() {
		var translation models.PersonalWordTranslation
		if err := rows.Scan(
			&translation.ID,
			&translation.UserID,
			&translation.Word,
			&translation.Translation,
			&translation.CreatedAt,
			&translation.UpdatedAt,
		); err != nil {
			return nil, err
		}
		translations = append(translations, translation)
	}

	return translations, rows.Err()
}

func (s *PersonalWordTranslationService) Create(userID uuid.UUID, word string, translation string) (*models.PersonalWordTranslation, error) {
	normalizedWord := normalizePersonalTranslationWord(word)
	normalizedTranslation := strings.TrimSpace(translation)
	if normalizedWord == "" || normalizedTranslation == "" {
		return nil, errors.New("word and translation are required")
	}

	var exists bool
	if err := s.db.QueryRow(context.Background(),
		`SELECT EXISTS(
			SELECT 1
			FROM personal_word_translations
			WHERE user_id = $1
			  AND LOWER(word) = LOWER($2)
			  AND LOWER(translation) = LOWER($3)
		)`,
		userID, normalizedWord, normalizedTranslation,
	).Scan(&exists); err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("personal translation already exists")
	}

	var created models.PersonalWordTranslation
	if err := s.db.QueryRow(context.Background(),
		`INSERT INTO personal_word_translations (user_id, word, translation)
		 VALUES ($1, $2, $3)
		 RETURNING id, user_id, word, translation, created_at, updated_at`,
		userID, normalizedWord, normalizedTranslation,
	).Scan(
		&created.ID,
		&created.UserID,
		&created.Word,
		&created.Translation,
		&created.CreatedAt,
		&created.UpdatedAt,
	); err != nil {
		return nil, err
	}

	return &created, nil
}

func (s *PersonalWordTranslationService) Delete(id int64, userID uuid.UUID) error {
	result, err := s.db.Exec(context.Background(),
		`DELETE FROM personal_word_translations
		 WHERE id = $1
		   AND user_id = $2`,
		id, userID,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return errors.New("personal translation not found")
	}

	return nil
}

func normalizePersonalTranslationWord(word string) string {
	return strings.ToLower(strings.TrimSpace(word))
}
