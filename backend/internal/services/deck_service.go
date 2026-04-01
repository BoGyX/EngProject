package services

import (
	"context"
	"english-learning/internal/models"
	"errors"
	"fmt"
	"strings"

	"english-learning/internal/utils"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DeckService struct {
	db *pgxpool.Pool
}

func NewDeckService(db *pgxpool.Pool) *DeckService {
	return &DeckService{db: db}
}

// GetAllDecks возвращает список всех decks
func (s *DeckService) GetAllDecks() ([]models.Deck, error) {
	rows, err := s.db.Query(context.Background(),
		"SELECT id, course_id, title, slug, description, position, created_at FROM decks ORDER BY position ASC, created_at DESC",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var decks []models.Deck
	for rows.Next() {
		var deck models.Deck
		err := rows.Scan(&deck.ID, &deck.CourseID, &deck.Title, &deck.Slug, &deck.Description, &deck.Position, &deck.CreatedAt)
		if err != nil {
			return nil, err
		}
		decks = append(decks, deck)
	}

	return decks, nil
}

// GetDeckByID возвращает deck по ID
func (s *DeckService) GetDeckByID(deckID int64) (*models.Deck, error) {
	var deck models.Deck
	err := s.db.QueryRow(context.Background(),
		"SELECT id, course_id, title, slug, description, position, created_at FROM decks WHERE id = $1",
		deckID,
	).Scan(&deck.ID, &deck.CourseID, &deck.Title, &deck.Slug, &deck.Description, &deck.Position, &deck.CreatedAt)

	if err != nil {
		return nil, errors.New("deck not found")
	}

	return &deck, nil
}

// GetDecksByCourseID возвращает все decks для конкретного курса
func (s *DeckService) GetDecksByCourseID(courseID int64) ([]models.Deck, error) {
	rows, err := s.db.Query(context.Background(),
		"SELECT id, course_id, title, slug, description, position, created_at FROM decks WHERE course_id = $1 ORDER BY position ASC, created_at DESC",
		courseID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var decks []models.Deck
	for rows.Next() {
		var deck models.Deck
		err := rows.Scan(&deck.ID, &deck.CourseID, &deck.Title, &deck.Slug, &deck.Description, &deck.Position, &deck.CreatedAt)
		if err != nil {
			return nil, err
		}
		decks = append(decks, deck)
	}

	return decks, nil
}

// CreateDeck создает новый deck
func (s *DeckService) CreateDeck(courseID int64, title string, slug string, description *string, position int) (*models.Deck, error) {
	uniqueSlug, err := s.ensureUniqueSlug(context.Background(), courseID, slug, title, nil)
	if err != nil {
		return nil, err
	}

	var deck models.Deck
	err = s.db.QueryRow(context.Background(),
		`INSERT INTO decks (course_id, title, slug, description, position)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, course_id, title, slug, description, position, created_at`,
		courseID, title, uniqueSlug, description, position,
	).Scan(&deck.ID, &deck.CourseID, &deck.Title, &deck.Slug, &deck.Description, &deck.Position, &deck.CreatedAt)

	if err != nil {
		return nil, err
	}

	return &deck, nil
}

// UpdateDeck обновляет deck
func (s *DeckService) UpdateDeck(deckID int64, title string, slug string, description *string, position int) (*models.Deck, error) {
	var courseID int64
	if err := s.db.QueryRow(context.Background(), "SELECT course_id FROM decks WHERE id = $1", deckID).Scan(&courseID); err != nil {
		return nil, errors.New("deck not found")
	}

	uniqueSlug, err := s.ensureUniqueSlug(context.Background(), courseID, slug, title, &deckID)
	if err != nil {
		return nil, err
	}

	var deck models.Deck
	err = s.db.QueryRow(context.Background(),
		`UPDATE decks 
		 SET title = $1, slug = $2, description = $3, position = $4
		 WHERE id = $5
		 RETURNING id, course_id, title, slug, description, position, created_at`,
		title, uniqueSlug, description, position, deckID,
	).Scan(&deck.ID, &deck.CourseID, &deck.Title, &deck.Slug, &deck.Description, &deck.Position, &deck.CreatedAt)

	if err != nil {
		return nil, errors.New("deck not found")
	}

	return &deck, nil
}

// GetDeckBySlug возвращает deck по slug курса и slug дека
func (s *DeckService) GetDeckBySlug(courseSlug string, deckSlug string) (*models.Deck, error) {
	var deck models.Deck
	err := s.db.QueryRow(context.Background(),
		`SELECT d.id, d.course_id, d.title, d.slug, d.description, d.position, d.created_at
		 FROM decks d
		 JOIN courses c ON c.id = d.course_id
		 WHERE c.slug = $1 AND d.slug = $2`,
		courseSlug, deckSlug,
	).Scan(&deck.ID, &deck.CourseID, &deck.Title, &deck.Slug, &deck.Description, &deck.Position, &deck.CreatedAt)

	if err != nil {
		return nil, errors.New("deck not found")
	}

	return &deck, nil
}

func (s *DeckService) ensureUniqueSlug(ctx context.Context, courseID int64, requestedSlug string, title string, excludeID *int64) (string, error) {
	baseSlug := strings.TrimSpace(requestedSlug)
	if baseSlug == "" {
		baseSlug = utils.Slugify(title, "deck")
	} else {
		baseSlug = utils.Slugify(baseSlug, "deck")
	}

	if baseSlug == "" {
		baseSlug = "deck"
	}

	candidate := baseSlug
	suffix := 2

	for {
		var exists bool
		err := s.db.QueryRow(ctx,
			`SELECT EXISTS(
				SELECT 1
				FROM decks
				WHERE course_id = $1
				  AND slug = $2
				  AND ($3::bigint IS NULL OR id <> $3)
			)`,
			courseID, candidate, excludeID,
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

// DeleteDeck удаляет deck
func (s *DeckService) DeleteDeck(deckID int64) error {
	result, err := s.db.Exec(context.Background(),
		"DELETE FROM decks WHERE id = $1",
		deckID,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("deck not found")
	}

	return nil
}
