DROP INDEX IF EXISTS idx_cards_custom_user_deck_word;
DROP INDEX IF EXISTS idx_training_session_cards_session;
DROP INDEX IF EXISTS idx_user_cards_progress;
DROP INDEX IF EXISTS idx_user_decks_active;
DROP INDEX IF EXISTS idx_user_courses_active;
DROP INDEX IF EXISTS idx_decks_course_slug_unique;
DROP INDEX IF EXISTS idx_courses_slug_unique;

DROP TABLE IF EXISTS training_session_cards;

ALTER TABLE training_sessions
DROP COLUMN IF EXISTS user_deck_id;

ALTER TABLE user_cards
DROP CONSTRAINT IF EXISTS chk_user_cards_current_mode;

ALTER TABLE user_cards
DROP COLUMN IF EXISTS progress_percentage,
DROP COLUMN IF EXISTS current_mode,
DROP COLUMN IF EXISTS mode_choice;

ALTER TABLE user_decks
DROP COLUMN IF EXISTS is_active,
DROP COLUMN IF EXISTS last_opened_at;

ALTER TABLE user_courses
DROP COLUMN IF EXISTS is_active,
DROP COLUMN IF EXISTS last_opened_at;

ALTER TABLE decks
DROP COLUMN IF EXISTS slug;

ALTER TABLE courses
DROP COLUMN IF EXISTS slug;
