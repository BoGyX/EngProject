-- Применение миграции для добавления колонок режимов обучения в user_cards
-- Этот скрипт безопасно добавляет колонки, если их еще нет

DO $$ 
BEGIN
    -- Добавляем mode_view
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'user_cards' AND column_name = 'mode_view'
    ) THEN
        ALTER TABLE user_cards ADD COLUMN mode_view BOOLEAN DEFAULT FALSE;
        RAISE NOTICE 'Column mode_view added';
    ELSE
        RAISE NOTICE 'Column mode_view already exists';
    END IF;

    -- Добавляем mode_with_photo
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'user_cards' AND column_name = 'mode_with_photo'
    ) THEN
        ALTER TABLE user_cards ADD COLUMN mode_with_photo BOOLEAN DEFAULT FALSE;
        RAISE NOTICE 'Column mode_with_photo added';
    ELSE
        RAISE NOTICE 'Column mode_with_photo already exists';
    END IF;

    -- Добавляем mode_without_photo
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'user_cards' AND column_name = 'mode_without_photo'
    ) THEN
        ALTER TABLE user_cards ADD COLUMN mode_without_photo BOOLEAN DEFAULT FALSE;
        RAISE NOTICE 'Column mode_without_photo added';
    ELSE
        RAISE NOTICE 'Column mode_without_photo already exists';
    END IF;

    -- Добавляем mode_russian
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'user_cards' AND column_name = 'mode_russian'
    ) THEN
        ALTER TABLE user_cards ADD COLUMN mode_russian BOOLEAN DEFAULT FALSE;
        RAISE NOTICE 'Column mode_russian added';
    ELSE
        RAISE NOTICE 'Column mode_russian already exists';
    END IF;

    -- Добавляем mode_constructor
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'user_cards' AND column_name = 'mode_constructor'
    ) THEN
        ALTER TABLE user_cards ADD COLUMN mode_constructor BOOLEAN DEFAULT FALSE;
        RAISE NOTICE 'Column mode_constructor added';
    ELSE
        RAISE NOTICE 'Column mode_constructor already exists';
    END IF;
END $$;

-- Комментарии для полей
COMMENT ON COLUMN user_cards.mode_view IS 'Пройден режим просмотра';
COMMENT ON COLUMN user_cards.mode_with_photo IS 'Пройден режим "Угадай по фото"';
COMMENT ON COLUMN user_cards.mode_without_photo IS 'Пройден режим "Угадай перевод"';
COMMENT ON COLUMN user_cards.mode_russian IS 'Пройден режим "Переведи на английский"';
COMMENT ON COLUMN user_cards.mode_constructor IS 'Пройден режим "Собери слово"';

-- Обновляем существующие записи, устанавливая FALSE для новых колонок
UPDATE user_cards 
SET 
    mode_view = COALESCE(mode_view, FALSE),
    mode_with_photo = COALESCE(mode_with_photo, FALSE),
    mode_without_photo = COALESCE(mode_without_photo, FALSE),
    mode_russian = COALESCE(mode_russian, FALSE),
    mode_constructor = COALESCE(mode_constructor, FALSE)
WHERE 
    mode_view IS NULL OR 
    mode_with_photo IS NULL OR 
    mode_without_photo IS NULL OR 
    mode_russian IS NULL OR 
    mode_constructor IS NULL;

SELECT 'Migration completed successfully!' AS status;
