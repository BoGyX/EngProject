package services

import (
	"context"
	"english-learning/internal/models"
	"errors"
	"math/rand"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *TrainingSessionService) GetTrainingSessionStateByID(id int64, userID *uuid.UUID) (*models.TrainingSessionState, error) {
	session, err := s.getSessionModelByID(id, userID)
	if err != nil {
		return nil, err
	}

	rows, err := s.db.Query(context.Background(),
		`SELECT
			tsc.id,
			tsc.card_id,
			tsc.sequence_number,
			tsc.current_mode,
			tsc.progress_percentage,
			tsc.is_completed,
			c.deck_id,
			c.word,
			c.translation,
			c.phonetic,
			c.audio_url,
			c.image_url,
			c.example,
			c.is_custom
		 FROM training_session_cards tsc
		 JOIN cards c ON c.id = tsc.card_id
		 WHERE tsc.session_id = $1
		 ORDER BY tsc.sequence_number ASC`,
		id,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []models.TrainingSessionCardState
	for rows.Next() {
		var card models.TrainingSessionCardState
		if err := rows.Scan(
			&card.SessionCardID,
			&card.CardID,
			&card.SequenceNumber,
			&card.CurrentMode,
			&card.ProgressPercentage,
			&card.IsCompleted,
			&card.DeckID,
			&card.Word,
			&card.Translation,
			&card.Phonetic,
			&card.AudioURL,
			&card.ImageURL,
			&card.Example,
			&card.IsCustom,
		); err != nil {
			return nil, err
		}
		cards = append(cards, card)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	deckTranslations, err := s.getDeckTranslations(session.DeckID)
	if err != nil {
		return nil, err
	}
	fallbackTranslations, err := s.getFallbackTranslations(session.DeckID, 12)
	if err != nil {
		return nil, err
	}

	var currentCard *models.TrainingSessionCardState
	remaining := 0
	for index := range cards {
		if cards[index].CurrentMode == "choice" && !cards[index].IsCompleted {
			cards[index].Options = buildChoiceOptions(cards[index].Translation, cards, deckTranslations, fallbackTranslations)
		}
		if !cards[index].IsCompleted {
			remaining++
			if currentCard == nil {
				currentCard = &cards[index]
			}
		}
	}

	return &models.TrainingSessionState{
		Session:        *session,
		Cards:          cards,
		CurrentCard:    currentCard,
		RemainingCards: remaining,
	}, nil
}

func (s *TrainingSessionService) GetTrainingSessionByIDForUser(id int64, userID uuid.UUID) (*models.TrainingSessionState, error) {
	return s.GetTrainingSessionStateByID(id, &userID)
}

func (s *TrainingSessionService) StartScopedTrainingSession(userID uuid.UUID, courseID *int64, deckID *int64) (*models.TrainingSessionState, error) {
	userDeck, resolvedCourseID, err := s.resolveScopedUserDeck(userID, courseID, deckID)
	if err != nil {
		return nil, err
	}
	if openSessionID, err := s.getLatestOpenScopedSessionID(userID, userDeck.ID); err != nil {
		return nil, err
	} else if openSessionID != 0 {
		existingState, err := s.GetTrainingSessionStateByID(openSessionID, &userID)
		if err == nil && existingState != nil {
			if existingState.CurrentCard != nil {
				return existingState, nil
			}
			if _, finishErr := s.FinishTrainingSession(openSessionID); finishErr != nil {
				return nil, finishErr
			}
		}
	}

	cardIDs, err := s.getAvailableScopedCardIDs(userID, userDeck)
	if err != nil {
		return nil, err
	}
	if len(cardIDs) == 0 {
		return nil, errors.New("no unfinished cards available for training")
	}

	ctx := context.Background()
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var session models.TrainingSession
	err = tx.QueryRow(ctx,
		`INSERT INTO training_sessions (user_id, course_id, deck_id, user_deck_id, started_at)
		 VALUES ($1, $2, $3, $4, NOW())
		 RETURNING id, user_id, course_id, deck_id, user_deck_id, started_at, finished_at`,
		userID, resolvedCourseID, userDeck.DeckID, userDeck.ID,
	).Scan(&session.ID, &session.UserID, &session.CourseID, &session.DeckID, &session.UserDeckID, &session.StartedAt, &session.FinishedAt)
	if err != nil {
		return nil, err
	}

	for index, cardID := range cardIDs {
		card, err := s.cardService.GetCardByID(cardID)
		if err != nil {
			return nil, err
		}

		currentMode := deriveCurrentTrainingModeForCard(nil, card)
		progress := deriveTrainingProgressForCard(nil, card)
		if userCard, err := s.userCardService.GetUserCardByCardAndDeck(userID, cardID, &userDeck.ID); err == nil {
			currentMode = deriveCurrentTrainingModeForCard(userCard, card)
			progress = deriveTrainingProgressForCard(userCard, card)
		}

		if _, err := tx.Exec(ctx,
			`INSERT INTO training_session_cards (
				session_id, card_id, position, sequence_number, current_mode, progress_percentage, is_completed,
				correct_answers, wrong_answers, created_at, updated_at
			) VALUES ($1, $2, $3, $3, $4, $5, $6, 0, 0, NOW(), NOW())`,
			session.ID, cardID, index+1, currentMode, progress, progress >= 100,
		); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return s.GetTrainingSessionStateByID(session.ID, &userID)
}

func (s *TrainingSessionService) SubmitScopedAnswer(userID uuid.UUID, sessionID int64, cardID int64, answer string) (*models.TrainingSessionState, bool, error) {
	ctx := context.Background()
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, false, err
	}
	defer tx.Rollback(ctx)

	session, err := s.getSessionForUpdate(ctx, tx, sessionID)
	if err != nil {
		return nil, false, err
	}
	if session.UserID == nil || *session.UserID != userID {
		return nil, false, errors.New("training session not found")
	}
	if session.UserDeckID == nil {
		return nil, false, errors.New("training session is missing active deck")
	}
	if session.FinishedAt != nil {
		return nil, false, errors.New("training session is already finished")
	}

	sessionCard, err := s.getSessionCardForUpdate(ctx, tx, sessionID, cardID)
	if err != nil {
		return nil, false, err
	}

	card, err := s.cardService.GetCardByID(cardID)
	if err != nil {
		return nil, false, err
	}

	userCard, err := s.userCardService.GetUserCardByCardAndDeck(userID, cardID, session.UserDeckID)
	if err != nil {
		userCard, err = s.userCardService.CreateUserCardWithModes(
			userID,
			cardID,
			session.UserDeckID,
			"new",
			0,
			0,
			false,
			false,
			false,
			false,
			false,
			false,
			deriveCurrentTrainingModeForCard(nil, card),
			0,
		)
		if err != nil {
			return nil, false, err
		}
	}

	currentMode := deriveCurrentTrainingModeForCard(userCard, card)
	progress := deriveTrainingProgressForCard(userCard, card)
	isCorrect := evaluateTrainingAnswer(currentMode, answer, card)
	lastSeen := time.Now()
	nextReview := lastSeen.Add(24 * time.Hour)

	if isCorrect {
		setTrainingModeCompleted(userCard, currentMode)
		userCard.CorrectCount++
	} else {
		userCard.WrongCount++
	}

	progress = deriveTrainingProgressForCard(userCard, card)
	nextMode := deriveCurrentTrainingModeForCard(userCard, card)
	status := "learning"
	if progress >= 100 || nextMode == "completed" {
		progress = 100
		nextMode = "completed"
		status = "learned"
	} else if userCard.CorrectCount == 0 {
		status = "new"
	}

	userCard, err = s.userCardService.UpdateUserCardWithModes(
		userCard.ID,
		status,
		userCard.CorrectCount,
		userCard.WrongCount,
		&lastSeen,
		&nextReview,
		&userCard.ModeView,
		&userCard.ModeChoice,
		&userCard.ModeWithPhoto,
		nil,
		&userCard.ModeRussian,
		&userCard.ModeConstructor,
		&nextMode,
		&progress,
	)
	if err != nil {
		return nil, false, err
	}

	if err := s.updateTrainingSessionCard(ctx, tx, sessionCard.ID, nextMode, progress, progress >= 100, &isCorrect, boolToInt(isCorrect), boolToInt(!isCorrect)); err != nil {
		return nil, false, err
	}
	if err := s.refreshScopedDeckAndCourseProgress(ctx, tx, userID, *session.UserDeckID); err != nil {
		return nil, false, err
	}
	if err := s.finishScopedSessionIfCompleted(ctx, tx, sessionID); err != nil {
		return nil, false, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, false, err
	}

	state, err := s.GetTrainingSessionStateByID(sessionID, &userID)
	if err != nil {
		return nil, false, err
	}

	return state, isCorrect, nil
}

func (s *TrainingSessionService) getSessionModelByID(id int64, userID *uuid.UUID) (*models.TrainingSession, error) {
	query := `SELECT id, user_id, course_id, deck_id, user_deck_id, started_at, finished_at
	          FROM training_sessions
	          WHERE id = $1`
	args := []interface{}{id}
	if userID != nil {
		query += ` AND user_id = $2`
		args = append(args, *userID)
	}

	var session models.TrainingSession
	err := s.db.QueryRow(context.Background(), query, args...).Scan(
		&session.ID,
		&session.UserID,
		&session.CourseID,
		&session.DeckID,
		&session.UserDeckID,
		&session.StartedAt,
		&session.FinishedAt,
	)
	if err != nil {
		return nil, errors.New("training session not found")
	}

	return &session, nil
}

func (s *TrainingSessionService) getSessionForUpdate(ctx context.Context, tx pgx.Tx, sessionID int64) (*models.TrainingSession, error) {
	var session models.TrainingSession
	err := tx.QueryRow(ctx,
		`SELECT id, user_id, course_id, deck_id, user_deck_id, started_at, finished_at
		 FROM training_sessions
		 WHERE id = $1
		 FOR UPDATE`,
		sessionID,
	).Scan(&session.ID, &session.UserID, &session.CourseID, &session.DeckID, &session.UserDeckID, &session.StartedAt, &session.FinishedAt)
	if err != nil {
		return nil, errors.New("training session not found")
	}

	return &session, nil
}

func (s *TrainingSessionService) getSessionCardForUpdate(ctx context.Context, tx pgx.Tx, sessionID int64, cardID int64) (*models.TrainingSessionCard, error) {
	var sessionCard models.TrainingSessionCard
	err := tx.QueryRow(ctx,
		`SELECT id, session_id, card_id, sequence_number, current_mode, progress_percentage,
		        is_completed, last_answer_correct, correct_answers, wrong_answers, created_at, updated_at
		 FROM training_session_cards
		 WHERE session_id = $1 AND card_id = $2
		 FOR UPDATE`,
		sessionID, cardID,
	).Scan(
		&sessionCard.ID,
		&sessionCard.SessionID,
		&sessionCard.CardID,
		&sessionCard.SequenceNumber,
		&sessionCard.CurrentMode,
		&sessionCard.ProgressPercentage,
		&sessionCard.IsCompleted,
		&sessionCard.LastAnswerCorrect,
		&sessionCard.CorrectAnswers,
		&sessionCard.WrongAnswers,
		&sessionCard.CreatedAt,
		&sessionCard.UpdatedAt,
	)
	if err != nil {
		return nil, errors.New("training session card not found")
	}

	return &sessionCard, nil
}

func (s *TrainingSessionService) getLatestOpenScopedSessionID(userID uuid.UUID, userDeckID int64) (int64, error) {
	var sessionID int64
	err := s.db.QueryRow(context.Background(),
		`SELECT id
		 FROM training_sessions
		 WHERE user_id = $1
		   AND user_deck_id = $2
		   AND finished_at IS NULL
		 ORDER BY started_at DESC, id DESC
		 LIMIT 1`,
		userID, userDeckID,
	).Scan(&sessionID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}

	return sessionID, nil
}

func (s *TrainingSessionService) resolveScopedUserDeck(userID uuid.UUID, courseID *int64, deckID *int64) (*models.UserDeck, *int64, error) {
	if deckID != nil {
		userDeck, err := s.userDeckService.ActivateDeck(userID, *deckID)
		if err != nil {
			return nil, nil, err
		}
		resolvedCourseID, err := s.getCourseIDByDeck(userDeck.DeckID)
		if err != nil {
			return nil, nil, err
		}
		return userDeck, &resolvedCourseID, nil
	}

	if courseID != nil {
		firstDeckID, err := s.getFirstDeckIDByCourse(*courseID)
		if err != nil {
			return nil, nil, err
		}
		userDeck, err := s.userDeckService.ActivateDeck(userID, firstDeckID)
		if err != nil {
			return nil, nil, err
		}
		return userDeck, courseID, nil
	}

	userDeck, err := s.userDeckService.GetActiveUserDeck(userID)
	if err != nil {
		return nil, nil, err
	}
	resolvedCourseID, err := s.getCourseIDByDeck(userDeck.DeckID)
	if err != nil {
		return nil, nil, err
	}
	return userDeck, &resolvedCourseID, nil
}

func (s *TrainingSessionService) getAvailableScopedCardIDs(userID uuid.UUID, userDeck *models.UserDeck) ([]int64, error) {
	rows, err := s.db.Query(context.Background(),
		`SELECT
			c.id,
			c.is_custom,
			c.image_url,
			uc.id,
			COALESCE(uc.status, 'new'),
			COALESCE(uc.correct_count, 0),
			COALESCE(uc.wrong_count, 0),
			COALESCE(uc.mode_view, false),
			COALESCE(uc.mode_choice, false),
			COALESCE(uc.mode_with_photo, false),
			COALESCE(uc.mode_without_photo, false),
			COALESCE(uc.mode_russian, false),
			COALESCE(uc.mode_constructor, false),
			COALESCE(uc.current_mode, ''),
			COALESCE(uc.progress_percentage, 0)
		 FROM cards c
		 LEFT JOIN user_cards uc
		   ON uc.card_id = c.id
		  AND uc.user_id = $1
		  AND uc.user_deck_id = $2
		 WHERE c.deck_id = $3
		 ORDER BY c.id ASC`,
		userID, userDeck.ID, userDeck.DeckID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var (
			cardID          int64
			isCustom        bool
			imageURL        *string
			userCardID      *int64
			status          string
			correctCount    int
			wrongCount      int
			modeView        bool
			modeChoice      bool
			modeWithPhoto   bool
			modeWithoutPhoto bool
			modeRussian     bool
			modeConstructor bool
			currentMode     string
			progress        int
		)
		if err := rows.Scan(
			&cardID,
			&isCustom,
			&imageURL,
			&userCardID,
			&status,
			&correctCount,
			&wrongCount,
			&modeView,
			&modeChoice,
			&modeWithPhoto,
			&modeWithoutPhoto,
			&modeRussian,
			&modeConstructor,
			&currentMode,
			&progress,
		); err != nil {
			return nil, err
		}

		card := &models.Card{
			ID:       cardID,
			DeckID:   userDeck.DeckID,
			IsCustom: isCustom,
			ImageURL: imageURL,
		}

		var userCard *models.UserCard
		if userCardID != nil {
			userCard = &models.UserCard{
				ID:               *userCardID,
				UserID:           userID,
				CardID:           cardID,
				UserDeckID:       &userDeck.ID,
				Status:           status,
				CorrectCount:     correctCount,
				WrongCount:       wrongCount,
				ModeView:         modeView,
				ModeChoice:       modeChoice,
				ModeWithPhoto:    modeWithPhoto,
				ModeWithoutPhoto: modeWithoutPhoto,
				ModeRussian:      modeRussian,
				ModeConstructor:  modeConstructor,
				CurrentMode:      currentMode,
				ProgressPercentage: progress,
			}
		}

		if deriveTrainingProgressForCard(userCard, card) < 100 {
			ids = append(ids, cardID)
		}
	}

	return ids, rows.Err()
}

func (s *TrainingSessionService) getFirstDeckIDByCourse(courseID int64) (int64, error) {
	var deckID int64
	err := s.db.QueryRow(context.Background(),
		`SELECT id
		 FROM decks
		 WHERE course_id = $1
		 ORDER BY position ASC, id ASC
		 LIMIT 1`,
		courseID,
	).Scan(&deckID)
	if err != nil {
		return 0, errors.New("deck not found for course")
	}
	return deckID, nil
}

func (s *TrainingSessionService) getCourseIDByDeck(deckID int64) (int64, error) {
	var courseID int64
	err := s.db.QueryRow(context.Background(), "SELECT course_id FROM decks WHERE id = $1", deckID).Scan(&courseID)
	if err != nil {
		return 0, errors.New("deck not found")
	}
	return courseID, nil
}

func (s *TrainingSessionService) updateTrainingSessionCard(ctx context.Context, tx pgx.Tx, sessionCardID int64, currentMode string, progress int, isCompleted bool, lastAnswerCorrect *bool, correctDelta int, wrongDelta int) error {
	_, err := tx.Exec(ctx,
		`UPDATE training_session_cards
		 SET current_mode = $1,
		     progress_percentage = $2,
		     is_completed = $3,
		     last_answer_correct = $4,
		     correct_answers = correct_answers + $5,
		     wrong_answers = wrong_answers + $6,
		     updated_at = NOW()
		 WHERE id = $7`,
		currentMode, progress, isCompleted, lastAnswerCorrect, correctDelta, wrongDelta, sessionCardID,
	)
	return err
}

func (s *TrainingSessionService) refreshScopedDeckAndCourseProgress(ctx context.Context, tx pgx.Tx, userID uuid.UUID, userDeckID int64) error {
	var deckID int64
	var userCourseID *int64
	if err := tx.QueryRow(ctx,
		`SELECT deck_id, user_course_id
		 FROM user_decks
		 WHERE id = $1`,
		userDeckID,
	).Scan(&deckID, &userCourseID); err != nil {
		return err
	}

	var totalCards int
	if err := tx.QueryRow(ctx, "SELECT COUNT(*) FROM cards WHERE deck_id = $1", deckID).Scan(&totalCards); err != nil {
		return err
	}

	var learnedCards int
	if err := tx.QueryRow(ctx,
		`SELECT COUNT(*)
		 FROM cards c
		 JOIN user_cards uc ON uc.card_id = c.id
		 WHERE c.deck_id = $1
		   AND uc.user_id = $2
		   AND uc.user_deck_id = $3
		   AND COALESCE(uc.progress_percentage, 0) >= 100`,
		deckID, userID, userDeckID,
	).Scan(&learnedCards); err != nil {
		return err
	}

	progress := 0.0
	status := "in_progress"
	if totalCards > 0 {
		progress = float64(learnedCards) * 100 / float64(totalCards)
	}
	if totalCards == 0 {
		status = "not_started"
	}
	if totalCards > 0 && learnedCards >= totalCards {
		status = "completed"
	}

	if _, err := tx.Exec(ctx,
		`UPDATE user_decks
		 SET learned_cards_count = $1,
		     total_cards_count = $2,
		     progress_percentage = $3,
		     status = $4,
		     completed_at = CASE WHEN $4 = 'completed' THEN COALESCE(completed_at, NOW()) ELSE completed_at END,
		     last_opened_at = NOW(),
		     updated_at = NOW()
		 WHERE id = $5`,
		learnedCards, totalCards, progress, status, userDeckID,
	); err != nil {
		return err
	}

	if userCourseID == nil {
		return nil
	}

	var totalDecks int
	if err := tx.QueryRow(ctx, "SELECT COUNT(*) FROM user_decks WHERE user_course_id = $1", *userCourseID).Scan(&totalDecks); err != nil {
		return err
	}
	var completedDecks int
	if err := tx.QueryRow(ctx, "SELECT COUNT(*) FROM user_decks WHERE user_course_id = $1 AND status = 'completed'", *userCourseID).Scan(&completedDecks); err != nil {
		return err
	}

	courseProgress := 0.0
	if totalDecks > 0 {
		courseProgress = float64(completedDecks) * 100 / float64(totalDecks)
	}

	_, err := tx.Exec(ctx,
		`UPDATE user_courses
		 SET completed_decks_count = $1,
		     total_decks_count = $2,
		     progress_percentage = $3,
		     last_opened_at = NOW()
		 WHERE id = $4`,
		completedDecks, totalDecks, courseProgress, *userCourseID,
	)
	return err
}

func (s *TrainingSessionService) finishScopedSessionIfCompleted(ctx context.Context, tx pgx.Tx, sessionID int64) error {
	var remaining int
	if err := tx.QueryRow(ctx,
		`SELECT COUNT(*)
		 FROM training_session_cards
		 WHERE session_id = $1 AND is_completed = false`,
		sessionID,
	).Scan(&remaining); err != nil {
		return err
	}
	if remaining > 0 {
		return nil
	}

	_, err := tx.Exec(ctx,
		`UPDATE training_sessions
		 SET finished_at = COALESCE(finished_at, NOW())
		 WHERE id = $1`,
		sessionID,
	)
	return err
}

func (s *TrainingSessionService) getDeckTranslations(deckID *int64) ([]string, error) {
	if deckID == nil {
		return []string{}, nil
	}

	rows, err := s.db.Query(context.Background(),
		`SELECT translation
		 FROM cards
		 WHERE deck_id = $1`,
		*deckID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var translations []string
	for rows.Next() {
		var translation string
		if err := rows.Scan(&translation); err != nil {
			return nil, err
		}
		translations = append(translations, translation)
	}

	return translations, rows.Err()
}

func (s *TrainingSessionService) getFallbackTranslations(deckID *int64, limit int) ([]string, error) {
	if limit <= 0 {
		return []string{}, nil
	}

	query := `
		SELECT DISTINCT translation
		FROM cards
		WHERE ($1::bigint IS NULL OR deck_id <> $1)
		ORDER BY translation ASC
		LIMIT $2`

	rows, err := s.db.Query(context.Background(), query, deckID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var translations []string
	for rows.Next() {
		var translation string
		if err := rows.Scan(&translation); err != nil {
			return nil, err
		}
		translations = append(translations, translation)
	}

	return translations, rows.Err()
}

func deriveCurrentTrainingModeForCard(userCard *models.UserCard, card *models.Card) string {
	for _, mode := range getTrainingModeSequence(card) {
		if !isTrainingModeCompleted(userCard, mode) {
			return mode
		}
	}
	return "completed"
}

func deriveTrainingProgressForCard(userCard *models.UserCard, card *models.Card) int {
	modes := getTrainingModeSequence(card)
	if len(modes) == 0 {
		return 100
	}

	completedModes := 0
	for _, mode := range modes {
		if isTrainingModeCompleted(userCard, mode) {
			completedModes++
		}
	}

	return completedModes * 100 / len(modes)
}

func getTrainingModeSequence(card *models.Card) []string {
	if card != nil && card.IsCustom {
		return []string{"choice", "constructor"}
	}

	modes := []string{"view", "choice"}
	if hasTrainingImage(card) {
		modes = append(modes, "with_photo")
	}
	modes = append(modes, "russian", "constructor")

	return modes
}

func hasTrainingImage(card *models.Card) bool {
	return card != nil && card.ImageURL != nil && strings.TrimSpace(*card.ImageURL) != ""
}

func isTrainingModeCompleted(userCard *models.UserCard, mode string) bool {
	if userCard == nil {
		return false
	}

	switch mode {
	case "view":
		return userCard.ModeView
	case "choice":
		return userCard.ModeChoice
	case "with_photo":
		return userCard.ModeWithPhoto
	case "russian":
		return userCard.ModeRussian
	case "constructor":
		return userCard.ModeConstructor
	default:
		return false
	}
}

func setTrainingModeCompleted(userCard *models.UserCard, mode string) {
	if userCard == nil {
		return
	}

	switch mode {
	case "view":
		userCard.ModeView = true
	case "choice":
		userCard.ModeChoice = true
	case "with_photo":
		userCard.ModeWithPhoto = true
	case "russian":
		userCard.ModeRussian = true
	case "constructor":
		userCard.ModeConstructor = true
	}
}

func evaluateTrainingAnswer(mode string, answer string, card *models.Card) bool {
	switch mode {
	case "view":
		return true
	case "choice":
		return matchesTrainingTranslation(answer, card.Translation)
	case "with_photo", "russian", "constructor":
		return normalizeTrainingText(answer) == normalizeTrainingText(card.Word)
	default:
		return false
	}
}

func normalizeTrainingText(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return ""
	}

	parts := strings.FieldsFunc(value, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r) && !unicode.IsSpace(r)
	})

	return strings.Join(strings.Fields(strings.Join(parts, " ")), " ")
}

func matchesTrainingTranslation(answer string, translation string) bool {
	normalizedAnswer := normalizeTrainingText(answer)
	if normalizedAnswer == "" {
		return false
	}
	if normalizedAnswer == normalizeTrainingText(translation) {
		return true
	}

	parts := strings.FieldsFunc(translation, func(r rune) bool {
		return r == ',' || r == ';' || r == '/' || r == '|'
	})
	for _, part := range parts {
		if normalizedAnswer == normalizeTrainingText(part) {
			return true
		}
	}

	return false
}

func buildChoiceOptions(correctTranslation string, sessionCards []models.TrainingSessionCardState, deckTranslations []string, fallbackTranslations []string) []string {
	seen := map[string]struct{}{normalizeTrainingText(correctTranslation): {}}
	distractors := make([]string, 0, 4)

	appendDistinct := func(candidate string) {
		normalized := normalizeTrainingText(candidate)
		if normalized == "" {
			return
		}
		if _, exists := seen[normalized]; exists {
			return
		}
		seen[normalized] = struct{}{}
		distractors = append(distractors, candidate)
	}

	for _, card := range sessionCards {
		appendDistinct(card.Translation)
	}
	for _, translation := range deckTranslations {
		if len(distractors) >= 2 {
			break
		}
		appendDistinct(translation)
	}
	for _, translation := range fallbackTranslations {
		if len(distractors) >= 2 {
			break
		}
		appendDistinct(translation)
	}

	shuffleStringValues(distractors)
	if len(distractors) > 2 {
		distractors = distractors[:2]
	}

	options := append(distractors, correctTranslation)
	shuffleStringValues(options)
	return options
}

func shuffleInt64Values(values []int64) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(values), func(i, j int) {
		values[i], values[j] = values[j], values[i]
	})
}

func shuffleStringValues(values []string) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(values), func(i, j int) {
		values[i], values[j] = values[j], values[i]
	})
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
