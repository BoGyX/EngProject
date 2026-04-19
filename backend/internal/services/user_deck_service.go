package services

import (
	"context"
	"english-learning/internal/models"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserDeckService struct {
	db *pgxpool.Pool
}

func NewUserDeckService(db *pgxpool.Pool) *UserDeckService {
	return &UserDeckService{db: db}
}

func (s *UserDeckService) GetAllUserDecks() ([]models.UserDeck, error) {
	rows, err := s.db.Query(context.Background(),
		`SELECT id, user_id, deck_id, user_course_id, status, learned_cards_count,
		        total_cards_count, progress_percentage, started_at, last_opened_at, completed_at, is_active, created_at, updated_at
		 FROM user_decks
		 ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userDecks []models.UserDeck
	for rows.Next() {
		var ud models.UserDeck
		if err := rows.Scan(&ud.ID, &ud.UserID, &ud.DeckID, &ud.UserCourseID, &ud.Status,
			&ud.LearnedCardsCount, &ud.TotalCardsCount, &ud.ProgressPercentage,
			&ud.StartedAt, &ud.LastOpenedAt, &ud.CompletedAt, &ud.IsActive, &ud.CreatedAt, &ud.UpdatedAt); err != nil {
			return nil, err
		}
		userDecks = append(userDecks, ud)
	}

	return userDecks, rows.Err()
}

func (s *UserDeckService) GetUserDeckByID(id int64) (*models.UserDeck, error) {
	var ud models.UserDeck
	err := s.db.QueryRow(context.Background(),
		`SELECT id, user_id, deck_id, user_course_id, status, learned_cards_count,
		        total_cards_count, progress_percentage, started_at, last_opened_at, completed_at, is_active, created_at, updated_at
		 FROM user_decks
		 WHERE id = $1`,
		id,
	).Scan(&ud.ID, &ud.UserID, &ud.DeckID, &ud.UserCourseID, &ud.Status,
		&ud.LearnedCardsCount, &ud.TotalCardsCount, &ud.ProgressPercentage,
		&ud.StartedAt, &ud.LastOpenedAt, &ud.CompletedAt, &ud.IsActive, &ud.CreatedAt, &ud.UpdatedAt)
	if err != nil {
		return nil, errors.New("user deck not found")
	}

	return &ud, nil
}

func (s *UserDeckService) GetUserDecksByUserID(userID uuid.UUID) ([]models.UserDeck, error) {
	rows, err := s.db.Query(context.Background(),
		`SELECT ud.id, ud.user_id, ud.deck_id, ud.user_course_id, ud.status, ud.learned_cards_count,
		        ud.total_cards_count, ud.progress_percentage, ud.started_at, ud.last_opened_at, ud.completed_at, ud.is_active, ud.created_at, ud.updated_at
		 FROM user_decks ud
		 JOIN decks d ON d.id = ud.deck_id
		 JOIN course_accesses ca
		   ON ca.user_id = ud.user_id
		  AND ca.course_id = d.course_id
		 WHERE ud.user_id = $1
		 ORDER BY ud.created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userDecks []models.UserDeck
	for rows.Next() {
		var ud models.UserDeck
		if err := rows.Scan(&ud.ID, &ud.UserID, &ud.DeckID, &ud.UserCourseID, &ud.Status,
			&ud.LearnedCardsCount, &ud.TotalCardsCount, &ud.ProgressPercentage,
			&ud.StartedAt, &ud.LastOpenedAt, &ud.CompletedAt, &ud.IsActive, &ud.CreatedAt, &ud.UpdatedAt); err != nil {
			return nil, err
		}
		userDecks = append(userDecks, ud)
	}

	return userDecks, rows.Err()
}

func (s *UserDeckService) StartDeck(userID uuid.UUID, deckID int64, userCourseID *int64) (*models.UserDeck, error) {
	ctx := context.Background()

	var courseID int64
	if err := s.db.QueryRow(ctx, "SELECT course_id FROM decks WHERE id = $1", deckID).Scan(&courseID); err != nil {
		return nil, errors.New("deck not found")
	}
	if err := s.ensureCourseAccess(ctx, userID, courseID); err != nil {
		return nil, err
	}

	var totalCards int
	if err := s.db.QueryRow(ctx, "SELECT COUNT(*) FROM cards WHERE deck_id = $1", deckID).Scan(&totalCards); err != nil {
		return nil, err
	}

	now := time.Now()
	var ud models.UserDeck
	err := s.db.QueryRow(ctx,
		`INSERT INTO user_decks (
			user_id, deck_id, user_course_id, status, learned_cards_count, total_cards_count,
			progress_percentage, started_at, last_opened_at, is_active, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, user_id, deck_id, user_course_id, status, learned_cards_count,
		          total_cards_count, progress_percentage, started_at, last_opened_at, completed_at, is_active, created_at, updated_at`,
		userID, deckID, userCourseID, "in_progress", 0, totalCards, 0.0, &now, &now, false, now, now,
	).Scan(&ud.ID, &ud.UserID, &ud.DeckID, &ud.UserCourseID, &ud.Status,
		&ud.LearnedCardsCount, &ud.TotalCardsCount, &ud.ProgressPercentage,
		&ud.StartedAt, &ud.LastOpenedAt, &ud.CompletedAt, &ud.IsActive, &ud.CreatedAt, &ud.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &ud, nil
}

func (s *UserDeckService) UpdateUserDeck(id int64, status string, learnedCardsCount int, totalCardsCount int, progressPercentage float64) (*models.UserDeck, error) {
	var completedAt *time.Time
	if status == "completed" {
		now := time.Now()
		completedAt = &now
	}

	var ud models.UserDeck
	err := s.db.QueryRow(context.Background(),
		`UPDATE user_decks
		 SET status = $1,
		     learned_cards_count = $2,
		     total_cards_count = $3,
		     progress_percentage = $4,
		     completed_at = $5,
		     last_opened_at = NOW(),
		     updated_at = $6
		 WHERE id = $7
		 RETURNING id, user_id, deck_id, user_course_id, status, learned_cards_count,
		           total_cards_count, progress_percentage, started_at, last_opened_at, completed_at, is_active, created_at, updated_at`,
		status, learnedCardsCount, totalCardsCount, progressPercentage, completedAt, time.Now(), id,
	).Scan(&ud.ID, &ud.UserID, &ud.DeckID, &ud.UserCourseID, &ud.Status,
		&ud.LearnedCardsCount, &ud.TotalCardsCount, &ud.ProgressPercentage,
		&ud.StartedAt, &ud.LastOpenedAt, &ud.CompletedAt, &ud.IsActive, &ud.CreatedAt, &ud.UpdatedAt)
	if err != nil {
		return nil, errors.New("user deck not found")
	}

	return &ud, nil
}

func (s *UserDeckService) DeleteUserDeck(id int64) error {
	result, err := s.db.Exec(context.Background(),
		"DELETE FROM user_decks WHERE id = $1",
		id,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return errors.New("user deck not found")
	}

	return nil
}

func (s *UserDeckService) GetActiveUserDeck(userID uuid.UUID) (*models.UserDeck, error) {
	var ud models.UserDeck
	err := s.db.QueryRow(context.Background(),
		`SELECT ud.id, ud.user_id, ud.deck_id, ud.user_course_id, ud.status, ud.learned_cards_count,
		        ud.total_cards_count, ud.progress_percentage, ud.started_at, ud.last_opened_at, ud.completed_at, ud.is_active, ud.created_at, ud.updated_at
		 FROM user_decks ud
		 JOIN decks d ON d.id = ud.deck_id
		 JOIN course_accesses ca
		   ON ca.user_id = ud.user_id
		  AND ca.course_id = d.course_id
		 WHERE ud.user_id = $1 AND ud.is_active = true
		 ORDER BY ud.last_opened_at DESC NULLS LAST, ud.created_at DESC
		 LIMIT 1`,
		userID,
	).Scan(&ud.ID, &ud.UserID, &ud.DeckID, &ud.UserCourseID, &ud.Status,
		&ud.LearnedCardsCount, &ud.TotalCardsCount, &ud.ProgressPercentage,
		&ud.StartedAt, &ud.LastOpenedAt, &ud.CompletedAt, &ud.IsActive, &ud.CreatedAt, &ud.UpdatedAt)
	if err != nil {
		return nil, errors.New("active user deck not found")
	}

	return &ud, nil
}

func (s *UserDeckService) ActivateDeck(userID uuid.UUID, deckID int64) (*models.UserDeck, error) {
	ctx := context.Background()
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var courseID int64
	if err := tx.QueryRow(ctx, "SELECT course_id FROM decks WHERE id = $1", deckID).Scan(&courseID); err != nil {
		return nil, errors.New("deck not found")
	}
	if err := s.ensureCourseAccessTx(ctx, tx, userID, courseID); err != nil {
		return nil, err
	}

	var userCourseID int64
	err = tx.QueryRow(ctx,
		`SELECT id
		 FROM user_courses
		 WHERE user_id = $1 AND course_id = $2
		 ORDER BY attempt_number DESC
		 LIMIT 1`,
		userID, courseID,
	).Scan(&userCourseID)
	if err != nil {
		var totalDecks int
		if err := tx.QueryRow(ctx, "SELECT COUNT(*) FROM decks WHERE course_id = $1", courseID).Scan(&totalDecks); err != nil {
			return nil, err
		}
		if err := tx.QueryRow(ctx,
			`INSERT INTO user_courses (
				user_id, course_id, started_at, last_opened_at, is_active, attempt_number,
				completed_decks_count, total_decks_count, progress_percentage
			) VALUES ($1, $2, NOW(), NOW(), true, 1, 0, $3, 0)
			RETURNING id`,
			userID, courseID, totalDecks,
		).Scan(&userCourseID); err != nil {
			return nil, err
		}
	}

	if err := s.ensureDeckUnlocked(ctx, tx, userID, userCourseID, courseID, deckID); err != nil {
		return nil, err
	}

	if _, err := tx.Exec(ctx, "UPDATE user_courses SET is_active = false WHERE user_id = $1", userID); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx,
		"UPDATE user_courses SET is_active = true, last_opened_at = NOW() WHERE id = $1",
		userCourseID,
	); err != nil {
		return nil, err
	}

	var ud models.UserDeck
	err = tx.QueryRow(ctx,
		`SELECT id, user_id, deck_id, user_course_id, status, learned_cards_count,
		        total_cards_count, progress_percentage, started_at, last_opened_at, completed_at, is_active, created_at, updated_at
		 FROM user_decks
		 WHERE user_id = $1 AND deck_id = $2 AND user_course_id = $3
		 LIMIT 1`,
		userID, deckID, userCourseID,
	).Scan(&ud.ID, &ud.UserID, &ud.DeckID, &ud.UserCourseID, &ud.Status,
		&ud.LearnedCardsCount, &ud.TotalCardsCount, &ud.ProgressPercentage,
		&ud.StartedAt, &ud.LastOpenedAt, &ud.CompletedAt, &ud.IsActive, &ud.CreatedAt, &ud.UpdatedAt)
	if err != nil {
		var totalCards int
		if err := tx.QueryRow(ctx, "SELECT COUNT(*) FROM cards WHERE deck_id = $1", deckID).Scan(&totalCards); err != nil {
			return nil, err
		}
		err = tx.QueryRow(ctx,
			`INSERT INTO user_decks (
				user_id, deck_id, user_course_id, status, learned_cards_count, total_cards_count,
				progress_percentage, started_at, last_opened_at, is_active, created_at, updated_at
			) VALUES ($1, $2, $3, 'in_progress', 0, $4, 0, NOW(), NOW(), true, NOW(), NOW())
			RETURNING id, user_id, deck_id, user_course_id, status, learned_cards_count,
			          total_cards_count, progress_percentage, started_at, last_opened_at, completed_at, is_active, created_at, updated_at`,
			userID, deckID, userCourseID, totalCards,
		).Scan(&ud.ID, &ud.UserID, &ud.DeckID, &ud.UserCourseID, &ud.Status,
			&ud.LearnedCardsCount, &ud.TotalCardsCount, &ud.ProgressPercentage,
			&ud.StartedAt, &ud.LastOpenedAt, &ud.CompletedAt, &ud.IsActive, &ud.CreatedAt, &ud.UpdatedAt)
		if err != nil {
			return nil, err
		}
	} else {
		if _, err := tx.Exec(ctx, "UPDATE user_decks SET is_active = false WHERE user_id = $1", userID); err != nil {
			return nil, err
		}
		err = tx.QueryRow(ctx,
			`UPDATE user_decks
			 SET is_active = true, last_opened_at = NOW()
			 WHERE id = $1
			 RETURNING id, user_id, deck_id, user_course_id, status, learned_cards_count,
			           total_cards_count, progress_percentage, started_at, last_opened_at, completed_at, is_active, created_at, updated_at`,
			ud.ID,
		).Scan(&ud.ID, &ud.UserID, &ud.DeckID, &ud.UserCourseID, &ud.Status,
			&ud.LearnedCardsCount, &ud.TotalCardsCount, &ud.ProgressPercentage,
			&ud.StartedAt, &ud.LastOpenedAt, &ud.CompletedAt, &ud.IsActive, &ud.CreatedAt, &ud.UpdatedAt)
		if err != nil {
			return nil, err
		}
	}

	if _, err := tx.Exec(ctx, "UPDATE user_decks SET is_active = false WHERE user_id = $1 AND id <> $2", userID, ud.ID); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &ud, nil
}

func (s *UserDeckService) ensureDeckUnlocked(ctx context.Context, tx pgx.Tx, userID uuid.UUID, userCourseID int64, courseID int64, deckID int64) error {
	var targetPosition int
	if err := tx.QueryRow(ctx,
		`SELECT position
		 FROM decks
		 WHERE id = $1 AND course_id = $2`,
		deckID, courseID,
	).Scan(&targetPosition); err != nil {
		return errors.New("deck not found")
	}

	var incompletePreviousDecks int
	if err := tx.QueryRow(ctx,
		`SELECT COUNT(*)
		 FROM decks d
		 LEFT JOIN user_decks ud
		   ON ud.deck_id = d.id
		  AND ud.user_id = $1
		  AND ud.user_course_id = $2
		 WHERE d.course_id = $3
		   AND (d.position < $4 OR (d.position = $4 AND d.id < $5))
		   AND COALESCE(ud.progress_percentage, 0) < 100`,
		userID, userCourseID, courseID, targetPosition, deckID,
	).Scan(&incompletePreviousDecks); err != nil {
		return err
	}

	if incompletePreviousDecks > 0 {
		return errors.New("deck is locked until previous decks are completed")
	}

	return nil
}

func (s *UserDeckService) ensureCourseAccess(ctx context.Context, userID uuid.UUID, courseID int64) error {
	var hasAccess bool
	if err := s.db.QueryRow(ctx,
		`SELECT EXISTS(
			SELECT 1
			FROM course_accesses
			WHERE user_id = $1 AND course_id = $2
		)`,
		userID, courseID,
	).Scan(&hasAccess); err != nil {
		return err
	}
	if !hasAccess {
		return errors.New("course access denied")
	}

	return nil
}

func (s *UserDeckService) ensureCourseAccessTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, courseID int64) error {
	var hasAccess bool
	if err := tx.QueryRow(ctx,
		`SELECT EXISTS(
			SELECT 1
			FROM course_accesses
			WHERE user_id = $1 AND course_id = $2
		)`,
		userID, courseID,
	).Scan(&hasAccess); err != nil {
		return err
	}
	if !hasAccess {
		return errors.New("course access denied")
	}

	return nil
}
