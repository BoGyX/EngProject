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

func (s *CardService) resolveCardPosition(ctx context.Context, deckID int64, requestedPosition *int) (int, error) {
	if requestedPosition != nil && *requestedPosition > 0 {
		return *requestedPosition, nil
	}

	var nextPosition int
	err := s.db.QueryRow(ctx, "SELECT COALESCE(MAX(position), 0) + 1 FROM cards WHERE deck_id = $1", deckID).Scan(&nextPosition)
	if err != nil {
		return 0, err
	}

	return nextPosition, nil
}

func (s *CardService) scanCard(scanner interface{ Scan(dest ...any) error }, card *models.Card) error {
	return scanner.Scan(
		&card.ID,
		&card.DeckID,
		&card.Position,
		&card.Word,
		&card.Translation,
		&card.Phonetic,
		&card.AudioURL,
		&card.ImageURL,
		&card.Example,
		&card.CreatedBy,
		&card.IsCustom,
		&card.CreatedAt,
	)
}

// GetAllCards возвращает список всех cards
func (s *CardService) GetAllCards() ([]models.Card, error) {
	rows, err := s.db.Query(context.Background(),
		`SELECT id, deck_id, position, word, translation, phonetic, audio_url, image_url, example, created_by, is_custom, created_at
		 FROM cards
		 ORDER BY deck_id ASC, position ASC, created_at ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []models.Card
	for rows.Next() {
		var card models.Card
		if err := s.scanCard(rows, &card); err != nil {
			return nil, err
		}
		cards = append(cards, card)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return cards, nil
}

// GetCardByID возвращает card по ID
func (s *CardService) GetCardByID(cardID int64) (*models.Card, error) {
	var card models.Card
	err := s.scanCard(
		s.db.QueryRow(context.Background(),
			`SELECT id, deck_id, position, word, translation, phonetic, audio_url, image_url, example, created_by, is_custom, created_at
			 FROM cards
			 WHERE id = $1`,
			cardID,
		),
		&card,
	)
	if err != nil {
		return nil, errors.New("card not found")
	}

	return &card, nil
}

// GetCardsByDeckID возвращает все cards для конкретного deck
func (s *CardService) GetCardsByDeckID(deckID int64) ([]models.Card, error) {
	rows, err := s.db.Query(context.Background(),
		`SELECT id, deck_id, position, word, translation, phonetic, audio_url, image_url, example, created_by, is_custom, created_at
		 FROM cards
		 WHERE deck_id = $1
		 ORDER BY position ASC, created_at ASC`,
		deckID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []models.Card
	for rows.Next() {
		var card models.Card
		if err := s.scanCard(rows, &card); err != nil {
			return nil, err
		}
		cards = append(cards, card)
	}

	if err := rows.Err(); err != nil {
		return nil, err
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
		`SELECT id, deck_id, position, word, translation, phonetic, audio_url, image_url, example, created_by, is_custom, created_at
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
		if err := s.scanCard(rows, &card); err != nil {
			return nil, err
		}
		cardByID[card.ID] = card
	}

	if err := rows.Err(); err != nil {
		return nil, err
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
func (s *CardService) CreateCard(deckID int64, position *int, word string, translation string, phonetic *string, audioURL *string, imageURL *string, example *string, createdBy *uuid.UUID, isCustom bool) (*models.Card, error) {
	ctx := context.Background()
	resolvedPosition, err := s.resolveCardPosition(ctx, deckID, position)
	if err != nil {
		return nil, err
	}

	var card models.Card
	err = s.scanCard(
		s.db.QueryRow(ctx,
			`INSERT INTO cards (deck_id, position, word, translation, phonetic, audio_url, image_url, example, created_by, is_custom)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			 RETURNING id, deck_id, position, word, translation, phonetic, audio_url, image_url, example, created_by, is_custom, created_at`,
			deckID, resolvedPosition, word, translation, phonetic, audioURL, imageURL, example, createdBy, isCustom,
		),
		&card,
	)
	if err != nil {
		return nil, err
	}

	return &card, nil
}

func (s *CardService) GetCustomCardByWord(deckID int64, userID uuid.UUID, word string) (*models.Card, error) {
	var card models.Card
	err := s.scanCard(
		s.db.QueryRow(context.Background(),
			`SELECT id, deck_id, position, word, translation, phonetic, audio_url, image_url, example, created_by, is_custom, created_at
			 FROM cards
			 WHERE deck_id = $1
			   AND created_by = $2
			   AND is_custom = true
			   AND LOWER(word) = LOWER($3)
			 LIMIT 1`,
			deckID, userID, word,
		),
		&card,
	)
	if err != nil {
		return nil, errors.New("card not found")
	}

	return &card, nil
}

func (s *CardService) CreateCustomCard(deckID int64, userID uuid.UUID, word string, translation string, phonetic *string, audioURL *string, imageURL *string, example *string) (*models.Card, error) {
	if existing, err := s.GetCustomCardByWord(deckID, userID, word); err == nil {
		return existing, nil
	}

	return s.CreateCard(deckID, nil, word, translation, phonetic, audioURL, imageURL, example, &userID, true)
}

// UpdateCard обновляет card
func (s *CardService) UpdateCard(cardID int64, position *int, word string, translation string, phonetic *string, audioURL *string, imageURL *string, example *string) (*models.Card, error) {
	ctx := context.Background()

	var currentPosition int
	if err := s.db.QueryRow(ctx, "SELECT position FROM cards WHERE id = $1", cardID).Scan(&currentPosition); err != nil {
		return nil, errors.New("card not found")
	}

	resolvedPosition := currentPosition
	if position != nil && *position > 0 {
		resolvedPosition = *position
	}

	var card models.Card
	err := s.scanCard(
		s.db.QueryRow(ctx,
			`UPDATE cards
			 SET position = $1, word = $2, translation = $3, phonetic = $4, audio_url = $5, image_url = $6, example = $7
			 WHERE id = $8
			 RETURNING id, deck_id, position, word, translation, phonetic, audio_url, image_url, example, created_by, is_custom, created_at`,
			resolvedPosition, word, translation, phonetic, audioURL, imageURL, example, cardID,
		),
		&card,
	)
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

	result, err := tx.Exec(ctx, "DELETE FROM cards WHERE id = $1", cardID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("card not found")
	}

	return tx.Commit(ctx)
}
