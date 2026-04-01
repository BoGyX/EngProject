ALTER TABLE courses
ADD COLUMN IF NOT EXISTS slug TEXT;

UPDATE courses
SET slug = COALESCE(
    NULLIF(REGEXP_REPLACE(LOWER(title), '[^a-z0-9]+', '-', 'g'), ''),
    'course-' || id
)
WHERE slug IS NULL;

WITH ranked_courses AS (
    SELECT id, slug, ROW_NUMBER() OVER (PARTITION BY slug ORDER BY id) AS rn
    FROM courses
)
UPDATE courses c
SET slug = c.slug || '-' || c.id
FROM ranked_courses rc
WHERE c.id = rc.id
  AND rc.rn > 1;

ALTER TABLE courses
ALTER COLUMN slug SET NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_courses_slug_unique ON courses(slug);

ALTER TABLE decks
ADD COLUMN IF NOT EXISTS slug TEXT;

UPDATE decks
SET slug = COALESCE(
    NULLIF(REGEXP_REPLACE(LOWER(title), '[^a-z0-9]+', '-', 'g'), ''),
    'deck-' || id
)
WHERE slug IS NULL;

WITH ranked_decks AS (
    SELECT id, course_id, slug, ROW_NUMBER() OVER (PARTITION BY course_id, slug ORDER BY id) AS rn
    FROM decks
)
UPDATE decks d
SET slug = d.slug || '-' || d.id
FROM ranked_decks rd
WHERE d.id = rd.id
  AND rd.rn > 1;

ALTER TABLE decks
ALTER COLUMN slug SET NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_decks_course_slug_unique ON decks(course_id, slug);

ALTER TABLE user_courses
ADD COLUMN IF NOT EXISTS last_opened_at TIMESTAMP DEFAULT NOW(),
ADD COLUMN IF NOT EXISTS is_active BOOLEAN DEFAULT false;

UPDATE user_courses
SET last_opened_at = COALESCE(last_opened_at, started_at, NOW())
WHERE last_opened_at IS NULL;

ALTER TABLE user_decks
ADD COLUMN IF NOT EXISTS last_opened_at TIMESTAMP DEFAULT NOW(),
ADD COLUMN IF NOT EXISTS is_active BOOLEAN DEFAULT false;

UPDATE user_decks
SET last_opened_at = COALESCE(last_opened_at, started_at, created_at, NOW())
WHERE last_opened_at IS NULL;

ALTER TABLE user_cards
ADD COLUMN IF NOT EXISTS mode_choice BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS current_mode TEXT DEFAULT 'view',
ADD COLUMN IF NOT EXISTS progress_percentage INT DEFAULT 0;

UPDATE user_cards
SET current_mode = CASE
        WHEN COALESCE(mode_view, false) = false THEN 'view'
        WHEN COALESCE(mode_choice, false) = false THEN 'choice'
        WHEN COALESCE(mode_with_photo, false) = false THEN 'with_photo'
        WHEN COALESCE(mode_russian, false) = false THEN 'russian'
        WHEN COALESCE(mode_constructor, false) = false THEN 'constructor'
        ELSE 'completed'
    END,
    progress_percentage = (
        (CASE WHEN COALESCE(mode_view, false) THEN 1 ELSE 0 END) +
        (CASE WHEN COALESCE(mode_choice, false) THEN 1 ELSE 0 END) +
        (CASE WHEN COALESCE(mode_with_photo, false) THEN 1 ELSE 0 END) +
        (CASE WHEN COALESCE(mode_russian, false) THEN 1 ELSE 0 END) +
        (CASE WHEN COALESCE(mode_constructor, false) THEN 1 ELSE 0 END)
    ) * 20;

ALTER TABLE user_cards
ALTER COLUMN current_mode SET NOT NULL,
ALTER COLUMN progress_percentage SET NOT NULL;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'chk_user_cards_current_mode'
    ) THEN
        ALTER TABLE user_cards
        ADD CONSTRAINT chk_user_cards_current_mode
        CHECK (current_mode IN ('view', 'choice', 'with_photo', 'russian', 'constructor', 'completed'));
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS training_session_cards (
    id BIGSERIAL PRIMARY KEY,
    session_id BIGINT NOT NULL REFERENCES training_sessions(id) ON DELETE CASCADE,
    card_id BIGINT NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
    position INT NOT NULL,
    UNIQUE(session_id, card_id),
    UNIQUE(session_id, position)
);

ALTER TABLE training_sessions
ADD COLUMN IF NOT EXISTS user_deck_id BIGINT REFERENCES user_decks(id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_user_courses_active ON user_courses(user_id, is_active);
CREATE INDEX IF NOT EXISTS idx_user_decks_active ON user_decks(user_id, is_active);
CREATE INDEX IF NOT EXISTS idx_user_cards_progress ON user_cards(user_id, progress_percentage);
CREATE INDEX IF NOT EXISTS idx_training_session_cards_session ON training_session_cards(session_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_cards_custom_user_deck_word
ON cards (deck_id, created_by, LOWER(word))
WHERE is_custom = true AND created_by IS NOT NULL;
