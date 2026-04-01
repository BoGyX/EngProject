package services

import (
	"context"
	"english-learning/internal/models"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TrainingSessionService struct {
	db              *pgxpool.Pool
	cardService     *CardService
	userDeckService *UserDeckService
	userCardService *UserCardService
}

func NewTrainingSessionService(db *pgxpool.Pool, cardService *CardService, userDeckService *UserDeckService, userCardService *UserCardService) *TrainingSessionService {
	return &TrainingSessionService{
		db:              db,
		cardService:     cardService,
		userDeckService: userDeckService,
		userCardService: userCardService,
	}
}

// GetAllTrainingSessions возвращает все сессии тренировок
func (s *TrainingSessionService) GetAllTrainingSessions() ([]models.TrainingSession, error) {
	rows, err := s.db.Query(context.Background(),
		`SELECT id, user_id, course_id, deck_id, started_at, finished_at 
		 FROM training_sessions ORDER BY started_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []models.TrainingSession
	for rows.Next() {
		var ts models.TrainingSession
		err := rows.Scan(&ts.ID, &ts.UserID, &ts.CourseID, &ts.DeckID, &ts.StartedAt, &ts.FinishedAt)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, ts)
	}

	return sessions, rows.Err()
}

// GetTrainingSessionByID возвращает сессию тренировки по ID
func (s *TrainingSessionService) GetTrainingSessionByID(id int64) (*models.TrainingSession, error) {
	var ts models.TrainingSession
	err := s.db.QueryRow(context.Background(),
		`SELECT id, user_id, course_id, deck_id, started_at, finished_at 
		 FROM training_sessions WHERE id = $1`,
		id,
	).Scan(&ts.ID, &ts.UserID, &ts.CourseID, &ts.DeckID, &ts.StartedAt, &ts.FinishedAt)

	if err != nil {
		return nil, errors.New("training session not found")
	}

	return &ts, nil
}

// GetTrainingSessionsByUserID возвращает все сессии пользователя
func (s *TrainingSessionService) GetTrainingSessionsByUserID(userID uuid.UUID) ([]models.TrainingSession, error) {
	rows, err := s.db.Query(context.Background(),
		`SELECT id, user_id, course_id, deck_id, started_at, finished_at 
		 FROM training_sessions WHERE user_id = $1 ORDER BY started_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []models.TrainingSession
	for rows.Next() {
		var ts models.TrainingSession
		err := rows.Scan(&ts.ID, &ts.UserID, &ts.CourseID, &ts.DeckID, &ts.StartedAt, &ts.FinishedAt)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, ts)
	}

	return sessions, rows.Err()
}

// StartTrainingSession создает новую сессию тренировки
func (s *TrainingSessionService) StartTrainingSession(userID *uuid.UUID, courseID *int64, deckID *int64) (*models.TrainingSession, error) {
	var ts models.TrainingSession
	err := s.db.QueryRow(context.Background(),
		`INSERT INTO training_sessions (user_id, course_id, deck_id, started_at)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, user_id, course_id, deck_id, started_at, finished_at`,
		userID, courseID, deckID, time.Now(),
	).Scan(&ts.ID, &ts.UserID, &ts.CourseID, &ts.DeckID, &ts.StartedAt, &ts.FinishedAt)

	if err != nil {
		return nil, err
	}

	return &ts, nil
}

// FinishTrainingSession завершает сессию тренировки
func (s *TrainingSessionService) FinishTrainingSession(id int64) (*models.TrainingSession, error) {
	now := time.Now()
	var ts models.TrainingSession
	err := s.db.QueryRow(context.Background(),
		`UPDATE training_sessions 
		 SET finished_at = $1
		 WHERE id = $2
		 RETURNING id, user_id, course_id, deck_id, started_at, finished_at`,
		&now, id,
	).Scan(&ts.ID, &ts.UserID, &ts.CourseID, &ts.DeckID, &ts.StartedAt, &ts.FinishedAt)

	if err != nil {
		return nil, errors.New("training session not found")
	}

	return &ts, nil
}

// DeleteTrainingSession удаляет сессию тренировки
func (s *TrainingSessionService) DeleteTrainingSession(id int64) error {
	result, err := s.db.Exec(context.Background(),
		"DELETE FROM training_sessions WHERE id = $1",
		id,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("training session not found")
	}

	return nil
}
