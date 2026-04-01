package services

import (
	"context"
	"english-learning/internal/models"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PersonalVocabularyService struct {
	db *pgxpool.Pool
}

func NewPersonalVocabularyService(db *pgxpool.Pool) *PersonalVocabularyService {
	return &PersonalVocabularyService{db: db}
}

// GetAllPersonalVocabulary возвращает список всех слов в личном словаре
func (s *PersonalVocabularyService) GetAllPersonalVocabulary() ([]models.PersonalVocabulary, error) {
	rows, err := s.db.Query(context.Background(),
		`SELECT id, user_id, word, translation, phonetic, audio_url, example, notes, tags, 
		 status, correct_count, wrong_count, last_seen, next_review, created_at, updated_at 
		 FROM personal_vocabulary ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vocabularies []models.PersonalVocabulary
	for rows.Next() {
		var vocab models.PersonalVocabulary
		err := rows.Scan(&vocab.ID, &vocab.UserID, &vocab.Word, &vocab.Translation, &vocab.Phonetic, 
			&vocab.AudioURL, &vocab.Example, &vocab.Notes, &vocab.Tags, &vocab.Status, 
			&vocab.CorrectCount, &vocab.WrongCount, &vocab.LastSeen, &vocab.NextReview, 
			&vocab.CreatedAt, &vocab.UpdatedAt)
		if err != nil {
			return nil, err
		}
		vocabularies = append(vocabularies, vocab)
	}

	return vocabularies, nil
}

// GetPersonalVocabularyByID возвращает слово по ID
func (s *PersonalVocabularyService) GetPersonalVocabularyByID(vocabID int64) (*models.PersonalVocabulary, error) {
	var vocab models.PersonalVocabulary
	err := s.db.QueryRow(context.Background(),
		`SELECT id, user_id, word, translation, phonetic, audio_url, example, notes, tags, 
		 status, correct_count, wrong_count, last_seen, next_review, created_at, updated_at 
		 FROM personal_vocabulary WHERE id = $1`,
		vocabID,
	).Scan(&vocab.ID, &vocab.UserID, &vocab.Word, &vocab.Translation, &vocab.Phonetic, 
		&vocab.AudioURL, &vocab.Example, &vocab.Notes, &vocab.Tags, &vocab.Status, 
		&vocab.CorrectCount, &vocab.WrongCount, &vocab.LastSeen, &vocab.NextReview, 
		&vocab.CreatedAt, &vocab.UpdatedAt)

	if err != nil {
		return nil, errors.New("vocabulary word not found")
	}

	return &vocab, nil
}

// GetPersonalVocabularyByUserID возвращает все слова для конкретного пользователя
func (s *PersonalVocabularyService) GetPersonalVocabularyByUserID(userID uuid.UUID) ([]models.PersonalVocabulary, error) {
	rows, err := s.db.Query(context.Background(),
		`SELECT id, user_id, word, translation, phonetic, audio_url, example, notes, tags, 
		 status, correct_count, wrong_count, last_seen, next_review, created_at, updated_at 
		 FROM personal_vocabulary WHERE user_id = $1 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vocabularies []models.PersonalVocabulary
	for rows.Next() {
		var vocab models.PersonalVocabulary
		err := rows.Scan(&vocab.ID, &vocab.UserID, &vocab.Word, &vocab.Translation, &vocab.Phonetic, 
			&vocab.AudioURL, &vocab.Example, &vocab.Notes, &vocab.Tags, &vocab.Status, 
			&vocab.CorrectCount, &vocab.WrongCount, &vocab.LastSeen, &vocab.NextReview, 
			&vocab.CreatedAt, &vocab.UpdatedAt)
		if err != nil {
			return nil, err
		}
		vocabularies = append(vocabularies, vocab)
	}

	return vocabularies, nil
}

// CreatePersonalVocabulary создает новое слово в личном словаре
func (s *PersonalVocabularyService) CreatePersonalVocabulary(userID uuid.UUID, word string, translation string, phonetic *string, audioURL *string, example *string, notes *string, tags []string, status string) (*models.PersonalVocabulary, error) {
	var vocab models.PersonalVocabulary
	err := s.db.QueryRow(context.Background(),
		`INSERT INTO personal_vocabulary (user_id, word, translation, phonetic, audio_url, example, notes, tags, status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 RETURNING id, user_id, word, translation, phonetic, audio_url, example, notes, tags, 
		 status, correct_count, wrong_count, last_seen, next_review, created_at, updated_at`,
		userID, word, translation, phonetic, audioURL, example, notes, tags, status,
	).Scan(&vocab.ID, &vocab.UserID, &vocab.Word, &vocab.Translation, &vocab.Phonetic, 
		&vocab.AudioURL, &vocab.Example, &vocab.Notes, &vocab.Tags, &vocab.Status, 
		&vocab.CorrectCount, &vocab.WrongCount, &vocab.LastSeen, &vocab.NextReview, 
		&vocab.CreatedAt, &vocab.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return &vocab, nil
}

// UpdatePersonalVocabulary обновляет слово в личном словаре
func (s *PersonalVocabularyService) UpdatePersonalVocabulary(vocabID int64, word string, translation string, phonetic *string, audioURL *string, example *string, notes *string, tags []string, status string) (*models.PersonalVocabulary, error) {
	var vocab models.PersonalVocabulary
	err := s.db.QueryRow(context.Background(),
		`UPDATE personal_vocabulary 
		 SET word = $1, translation = $2, phonetic = $3, audio_url = $4, example = $5, 
		 notes = $6, tags = $7, status = $8, updated_at = NOW()
		 WHERE id = $9
		 RETURNING id, user_id, word, translation, phonetic, audio_url, example, notes, tags, 
		 status, correct_count, wrong_count, last_seen, next_review, created_at, updated_at`,
		word, translation, phonetic, audioURL, example, notes, tags, status, vocabID,
	).Scan(&vocab.ID, &vocab.UserID, &vocab.Word, &vocab.Translation, &vocab.Phonetic, 
		&vocab.AudioURL, &vocab.Example, &vocab.Notes, &vocab.Tags, &vocab.Status, 
		&vocab.CorrectCount, &vocab.WrongCount, &vocab.LastSeen, &vocab.NextReview, 
		&vocab.CreatedAt, &vocab.UpdatedAt)

	if err != nil {
		return nil, errors.New("vocabulary word not found")
	}

	return &vocab, nil
}

// UpdatePersonalVocabularyStats обновляет статистику изучения слова
func (s *PersonalVocabularyService) UpdatePersonalVocabularyStats(vocabID int64, isCorrect bool) (*models.PersonalVocabulary, error) {
	var vocab models.PersonalVocabulary
	
	// Обновляем счетчики и last_seen
	updateQuery := `UPDATE personal_vocabulary 
		SET `
	if isCorrect {
		updateQuery += `correct_count = correct_count + 1, `
	} else {
		updateQuery += `wrong_count = wrong_count + 1, `
	}
	updateQuery += `last_seen = NOW(), updated_at = NOW()
		WHERE id = $1
		RETURNING id, user_id, word, translation, phonetic, audio_url, example, notes, tags, 
		status, correct_count, wrong_count, last_seen, next_review, created_at, updated_at`
	
	err := s.db.QueryRow(context.Background(), updateQuery, vocabID,
	).Scan(&vocab.ID, &vocab.UserID, &vocab.Word, &vocab.Translation, &vocab.Phonetic, 
		&vocab.AudioURL, &vocab.Example, &vocab.Notes, &vocab.Tags, &vocab.Status, 
		&vocab.CorrectCount, &vocab.WrongCount, &vocab.LastSeen, &vocab.NextReview, 
		&vocab.CreatedAt, &vocab.UpdatedAt)

	if err != nil {
		return nil, errors.New("vocabulary word not found")
	}

	return &vocab, nil
}

// DeletePersonalVocabulary удаляет слово из личного словаря
func (s *PersonalVocabularyService) DeletePersonalVocabulary(vocabID int64) error {
	result, err := s.db.Exec(context.Background(),
		"DELETE FROM personal_vocabulary WHERE id = $1",
		vocabID,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("vocabulary word not found")
	}

	return nil
}

// GetPersonalVocabularyForReview возвращает слова, которые нужно повторить
func (s *PersonalVocabularyService) GetPersonalVocabularyForReview(userID uuid.UUID) ([]models.PersonalVocabulary, error) {
	now := time.Now()
	rows, err := s.db.Query(context.Background(),
		`SELECT id, user_id, word, translation, phonetic, audio_url, example, notes, tags, 
		 status, correct_count, wrong_count, last_seen, next_review, created_at, updated_at 
		 FROM personal_vocabulary 
		 WHERE user_id = $1 AND (next_review IS NULL OR next_review <= $2)
		 ORDER BY next_review ASC NULLS FIRST, created_at ASC`,
		userID, now,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vocabularies []models.PersonalVocabulary
	for rows.Next() {
		var vocab models.PersonalVocabulary
		err := rows.Scan(&vocab.ID, &vocab.UserID, &vocab.Word, &vocab.Translation, &vocab.Phonetic, 
			&vocab.AudioURL, &vocab.Example, &vocab.Notes, &vocab.Tags, &vocab.Status, 
			&vocab.CorrectCount, &vocab.WrongCount, &vocab.LastSeen, &vocab.NextReview, 
			&vocab.CreatedAt, &vocab.UpdatedAt)
		if err != nil {
			return nil, err
		}
		vocabularies = append(vocabularies, vocab)
	}

	return vocabularies, nil
}
