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

	log.Printf("Подключение к БД: %s@%s:%s/%s", cfg.User, cfg.Host, cfg.Port, cfg.Name)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать пул подключений: %w", err)
	}

	ctxPing, cancelPing := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelPing()

	if err := pool.Ping(ctxPing); err != nil {
		pool.Close()
		return nil, fmt.Errorf("не удалось подключиться к БД: %w. Проверьте параметры подключения в .env", err)
	}

	log.Printf("Успешное подключение к БД")

	if err := ensureTrainingSessionCompatibility(pool); err != nil {
		pool.Close()
		return nil, fmt.Errorf("не удалось привести схему обучения к актуальному виду: %w", err)
	}

	return pool, nil
}

func ensureTrainingSessionCompatibility(pool *pgxpool.Pool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	log.Println("Проверка схемы training_session_cards...")

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
`

	if _, err := pool.Exec(ctx, compatSQL); err != nil {
		return err
	}

	log.Println("Схема training_session_cards актуальна")
	return nil
}
