package services

import (
	"context"
	"english-learning/internal/models"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CardService struct {
	db *pgxpool.Pool
}

func NewCardService(db *pgxpool.Pool) *CardService {
	return &CardService{db: db}
}

// GetAllCards возвращает список всех cards
func (s *CardService) GetAllCards() ([]models.Card, error) {
	rows, err := s.db.Query(context.Background(),
		"SELECT id, deck_id, word, translation, phonetic, audio_url, image_url, example, created_by, is_custom, created_at FROM cards ORDER BY created_at DESC",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []models.Card
	for rows.Next() {
		var card models.Card
		err := rows.Scan(&card.ID, &card.DeckID, &card.Word, &card.Translation, &card.Phonetic, &card.AudioURL, &card.ImageURL, &card.Example, &card.CreatedBy, &card.IsCustom, &card.CreatedAt)
		if err != nil {
			return nil, err
		}
		cards = append(cards, card)
	}

	return cards, nil
}

// GetCardByID возвращает card по ID
func (s *CardService) GetCardByID(cardID int64) (*models.Card, error) {
	var card models.Card
	err := s.db.QueryRow(context.Background(),
		"SELECT id, deck_id, word, translation, phonetic, audio_url, image_url, example, created_by, is_custom, created_at FROM cards WHERE id = $1",
		cardID,
	).Scan(&card.ID, &card.DeckID, &card.Word, &card.Translation, &card.Phonetic, &card.AudioURL, &card.ImageURL, &card.Example, &card.CreatedBy, &card.IsCustom, &card.CreatedAt)

	if err != nil {
		return nil, errors.New("card not found")
	}

	return &card, nil
}

// GetCardsByDeckID возвращает все cards для конкретного deck
func (s *CardService) GetCardsByDeckID(deckID int64) ([]models.Card, error) {
	rows, err := s.db.Query(context.Background(),
		"SELECT id, deck_id, word, translation, phonetic, audio_url, image_url, example, created_by, is_custom, created_at FROM cards WHERE deck_id = $1 ORDER BY created_at DESC",
		deckID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []models.Card
	for rows.Next() {
		var card models.Card
		err := rows.Scan(&card.ID, &card.DeckID, &card.Word, &card.Translation, &card.Phonetic, &card.AudioURL, &card.ImageURL, &card.Example, &card.CreatedBy, &card.IsCustom, &card.CreatedAt)
		if err != nil {
			return nil, err
		}
		cards = append(cards, card)
	}

	return cards, nil
}

// GetCardsByIDs возвращает карточки по набору id с сохранением входного порядка.
func (s *CardService) GetCardsByIDs(ids []int64) ([]models.Card, error) {
	if len(ids) == 0 {
		return []models.Card{}, nil
	}

	args := make([]interface{}, 0, len(ids))
	placeholders := make([]string, 0, len(ids))
	for i, id := range ids {
		args = append(args, id)
		placeholders = append(placeholders, fmt.Sprintf("$%d", i+1))
	}

	query := fmt.Sprintf(
		`SELECT id, deck_id, word, translation, phonetic, audio_url, image_url, example, created_by, is_custom, created_at
		 FROM cards
		 WHERE id IN (%s)`,
		strings.Join(placeholders, ", "),
	)

	rows, err := s.db.Query(context.Background(), query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cardByID := make(map[int64]models.Card, len(ids))
	for rows.Next() {
		var card models.Card
		if err := rows.Scan(&card.ID, &card.DeckID, &card.Word, &card.Translation, &card.Phonetic, &card.AudioURL, &card.ImageURL, &card.Example, &card.CreatedBy, &card.IsCustom, &card.CreatedAt); err != nil {
			return nil, err
		}
		cardByID[card.ID] = card
	}

	ordered := make([]models.Card, 0, len(ids))
	for _, id := range ids {
		if card, ok := cardByID[id]; ok {
			ordered = append(ordered, card)
		}
	}

	return ordered, nil
}

// CreateCard создает новый card
func (s *CardService) CreateCard(deckID int64, word string, translation string, phonetic *string, audioURL *string, imageURL *string, example *string, createdBy *uuid.UUID, isCustom bool) (*models.Card, error) {
	var card models.Card
	err := s.db.QueryRow(context.Background(),
		`INSERT INTO cards (deck_id, word, translation, phonetic, audio_url, image_url, example, created_by, is_custom)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 RETURNING id, deck_id, word, translation, phonetic, audio_url, image_url, example, created_by, is_custom, created_at`,
		deckID, word, translation, phonetic, audioURL, imageURL, example, createdBy, isCustom,
	).Scan(&card.ID, &card.DeckID, &card.Word, &card.Translation, &card.Phonetic, &card.AudioURL, &card.ImageURL, &card.Example, &card.CreatedBy, &card.IsCustom, &card.CreatedAt)

	if err != nil {
		return nil, err
	}

	return &card, nil
}

func (s *CardService) GetCustomCardByWord(deckID int64, userID uuid.UUID, word string) (*models.Card, error) {
	var card models.Card
	err := s.db.QueryRow(context.Background(),
		`SELECT id, deck_id, word, translation, phonetic, audio_url, image_url, example, created_by, is_custom, created_at
		 FROM cards
		 WHERE deck_id = $1
		   AND created_by = $2
		   AND is_custom = true
		   AND LOWER(word) = LOWER($3)
		 LIMIT 1`,
		deckID, userID, word,
	).Scan(&card.ID, &card.DeckID, &card.Word, &card.Translation, &card.Phonetic, &card.AudioURL, &card.ImageURL, &card.Example, &card.CreatedBy, &card.IsCustom, &card.CreatedAt)
	if err != nil {
		return nil, errors.New("card not found")
	}

	return &card, nil
}

func (s *CardService) CreateCustomCard(deckID int64, userID uuid.UUID, word string, translation string, phonetic *string, audioURL *string, imageURL *string, example *string) (*models.Card, error) {
	if existing, err := s.GetCustomCardByWord(deckID, userID, word); err == nil {
		return existing, nil
	}

	return s.CreateCard(deckID, word, translation, phonetic, audioURL, imageURL, example, &userID, true)
}

// UpdateCard обновляет card
func (s *CardService) UpdateCard(cardID int64, word string, translation string, phonetic *string, audioURL *string, imageURL *string, example *string) (*models.Card, error) {
	var card models.Card
	err := s.db.QueryRow(context.Background(),
		`UPDATE cards 
		 SET word = $1, translation = $2, phonetic = $3, audio_url = $4, image_url = $5, example = $6
		 WHERE id = $7
		 RETURNING id, deck_id, word, translation, phonetic, audio_url, image_url, example, created_by, is_custom, created_at`,
		word, translation, phonetic, audioURL, imageURL, example, cardID,
	).Scan(&card.ID, &card.DeckID, &card.Word, &card.Translation, &card.Phonetic, &card.AudioURL, &card.ImageURL, &card.Example, &card.CreatedBy, &card.IsCustom, &card.CreatedAt)

	if err != nil {
		return nil, errors.New("card not found")
	}

	return &card, nil
}

// DeleteCard удаляет card
func (s *CardService) DeleteCard(cardID int64) error {
	ctx := context.Background()
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err = tx.Exec(ctx, "DELETE FROM training_answers WHERE card_id = $1", cardID); err != nil {
		return err
	}

	result, err := tx.Exec(ctx,
		"DELETE FROM cards WHERE id = $1",
		cardID,
	)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("card not found")
	}

	return tx.Commit(ctx)
}
