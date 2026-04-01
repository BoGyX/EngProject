-- Добавляем поля для хранения прогресса по режимам обучения
ALTER TABLE user_cards 
ADD COLUMN IF NOT EXISTS mode_view BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS mode_with_photo BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS mode_without_photo BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS mode_russian BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS mode_constructor BOOLEAN DEFAULT FALSE;

-- Комментарии для полей
COMMENT ON COLUMN user_cards.mode_view IS 'Пройден режим просмотра';
COMMENT ON COLUMN user_cards.mode_with_photo IS 'Пройден режим "Угадай по фото"';
COMMENT ON COLUMN user_cards.mode_without_photo IS 'Пройден режим "Угадай перевод"';
COMMENT ON COLUMN user_cards.mode_russian IS 'Пройден режим "Переведи на английский"';
COMMENT ON COLUMN user_cards.mode_constructor IS 'Пройден режим "Собери слово"';
