package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"english-learning/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(cfg config.DatabaseConfig) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode,
	)

	log.Printf("Connecting to DB: %s@%s:%s/%s", cfg.User, cfg.Host, cfg.Port, cfg.Name)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	ctxPing, cancelPing := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelPing()

	if err := pool.Ping(ctxPing); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to connect to DB: %w. Check your .env connection settings", err)
	}

	log.Printf("DB connection established")

	if err := ensureTrainingSessionCompatibility(pool); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to update training schema: %w", err)
	}

	if err := ensureReadingTextCompatibility(pool); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to update reading_texts schema: %w", err)
	}

	return pool, nil
}

func ensureTrainingSessionCompatibility(pool *pgxpool.Pool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	log.Println("Checking training_session_cards schema...")

	compatSQL := `
ALTER TABLE training_session_cards
    ADD COLUMN IF NOT EXISTS sequence_number INT,
    ADD COLUMN IF NOT EXISTS current_mode TEXT,
    ADD COLUMN IF NOT EXISTS progress_percentage INT,
    ADD COLUMN IF NOT EXISTS is_completed BOOLEAN,
    ADD COLUMN IF NOT EXISTS last_answer_correct BOOLEAN,
    ADD COLUMN IF NOT EXISTS correct_answers INT,
    ADD COLUMN IF NOT EXISTS wrong_answers INT,
    ADD COLUMN IF NOT EXISTS created_at TIMESTAMP DEFAULT NOW(),
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT NOW();

UPDATE training_session_cards
SET sequence_number = position
WHERE sequence_number IS NULL;

UPDATE training_session_cards
SET current_mode = 'view'
WHERE current_mode IS NULL;

UPDATE training_session_cards
SET progress_percentage = 0
WHERE progress_percentage IS NULL;

UPDATE training_session_cards
SET is_completed = false
WHERE is_completed IS NULL;

UPDATE training_session_cards
SET correct_answers = 0
WHERE correct_answers IS NULL;

UPDATE training_session_cards
SET wrong_answers = 0
WHERE wrong_answers IS NULL;

UPDATE training_session_cards
SET created_at = NOW()
WHERE created_at IS NULL;

UPDATE training_session_cards
SET updated_at = NOW()
WHERE updated_at IS NULL;

ALTER TABLE training_session_cards
    ALTER COLUMN sequence_number SET NOT NULL,
    ALTER COLUMN current_mode SET NOT NULL,
    ALTER COLUMN progress_percentage SET NOT NULL,
    ALTER COLUMN is_completed SET NOT NULL,
    ALTER COLUMN correct_answers SET NOT NULL,
    ALTER COLUMN wrong_answers SET NOT NULL,
    ALTER COLUMN created_at SET NOT NULL,
    ALTER COLUMN updated_at SET NOT NULL,
    ALTER COLUMN current_mode SET DEFAULT 'view',
    ALTER COLUMN progress_percentage SET DEFAULT 0,
    ALTER COLUMN is_completed SET DEFAULT false,
    ALTER COLUMN correct_answers SET DEFAULT 0,
    ALTER COLUMN wrong_answers SET DEFAULT 0,
    ALTER COLUMN created_at SET DEFAULT NOW(),
    ALTER COLUMN updated_at SET DEFAULT NOW();

ALTER TABLE training_session_cards
    DROP CONSTRAINT IF EXISTS training_session_cards_session_id_card_id_key;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'training_session_cards_session_id_sequence_number_key'
    ) THEN
        ALTER TABLE training_session_cards
            ADD CONSTRAINT training_session_cards_session_id_sequence_number_key
            UNIQUE (session_id, sequence_number);
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'chk_training_session_cards_current_mode'
    ) THEN
        ALTER TABLE training_session_cards
            ADD CONSTRAINT chk_training_session_cards_current_mode
            CHECK (current_mode IN ('view', 'choice', 'with_photo', 'russian', 'constructor', 'completed'));
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_training_session_cards_session_card
    ON training_session_cards(session_id, card_id);
`

	if _, err := pool.Exec(ctx, compatSQL); err != nil {
		return err
	}

	log.Println("training_session_cards schema is up to date")
	return nil
}

func ensureReadingTextCompatibility(pool *pgxpool.Pool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	log.Println("Checking reading_texts schema...")

	compatSQL := `
CREATE TABLE IF NOT EXISTS reading_texts (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    audio_url TEXT DEFAULT '',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE reading_texts
    ADD COLUMN IF NOT EXISTS audio_url TEXT;

UPDATE reading_texts
SET audio_url = ''
WHERE audio_url IS NULL;

ALTER TABLE reading_texts
    ALTER COLUMN audio_url SET DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_reading_texts_user_id ON reading_texts(user_id);
`

	if _, err := pool.Exec(ctx, compatSQL); err != nil {
		return err
	}

	log.Println("reading_texts schema is up to date")
	return nil
}
