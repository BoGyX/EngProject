package services

import (
	"context"
	"english-learning/internal/models"
	"errors"
	"sort"
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

	ctx := context.Background()
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var existingID int64
	err = tx.QueryRow(ctx,
		`SELECT id
		 FROM personal_word_translations
		 WHERE user_id = $1
		   AND LOWER(word) = LOWER($2)
		 ORDER BY updated_at DESC, created_at DESC, id DESC
		 LIMIT 1`,
		userID, normalizedWord,
	).Scan(&existingID)

	var result models.PersonalWordTranslation
	if err == nil {
		if err := tx.QueryRow(ctx,
			`UPDATE personal_word_translations
			 SET translation = $1,
			     updated_at = NOW()
			 WHERE id = $2
			 RETURNING id, user_id, word, translation, created_at, updated_at`,
			normalizedTranslation, existingID,
		).Scan(
			&result.ID,
			&result.UserID,
			&result.Word,
			&result.Translation,
			&result.CreatedAt,
			&result.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if _, err := tx.Exec(ctx,
			`DELETE FROM personal_word_translations
			 WHERE user_id = $1
			   AND LOWER(word) = LOWER($2)
			   AND id <> $3`,
			userID, normalizedWord, existingID,
		); err != nil {
			return nil, err
		}
	} else {
		if err := tx.QueryRow(ctx,
			`INSERT INTO personal_word_translations (user_id, word, translation)
			 VALUES ($1, $2, $3)
			 RETURNING id, user_id, word, translation, created_at, updated_at`,
			userID, normalizedWord, normalizedTranslation,
		).Scan(
			&result.ID,
			&result.UserID,
			&result.Word,
			&result.Translation,
			&result.CreatedAt,
			&result.UpdatedAt,
		); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &result, nil
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

func (s *PersonalWordTranslationService) GetLatestTranslationMap(userID uuid.UUID, words []string) (map[string]string, error) {
	normalizedWords := normalizePersonalTranslationWords(words)
	if len(normalizedWords) == 0 {
		return map[string]string{}, nil
	}

	rows, err := s.db.Query(context.Background(),
		`SELECT DISTINCT ON (LOWER(word))
			word, translation
		 FROM personal_word_translations
		 WHERE user_id = $1
		   AND LOWER(word) = ANY($2::text[])
		 ORDER BY LOWER(word), updated_at DESC, created_at DESC, id DESC`,
		userID, normalizedWords,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]string, len(normalizedWords))
	for rows.Next() {
		var word string
		var translation string
		if err := rows.Scan(&word, &translation); err != nil {
			return nil, err
		}
		result[normalizePersonalTranslationWord(word)] = translation
	}

	return result, rows.Err()
}

func normalizePersonalTranslationWords(words []string) []string {
	seen := make(map[string]struct{}, len(words))
	result := make([]string, 0, len(words))
	for _, word := range words {
		normalized := normalizePersonalTranslationWord(word)
		if normalized == "" {
			continue
		}
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}

	sort.Strings(result)
	return result
}
