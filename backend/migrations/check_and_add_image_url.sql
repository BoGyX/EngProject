-- Проверка и добавление колонки image_url в таблицу cards
-- Выполните этот SQL запрос в вашей базе данных PostgreSQL

-- Проверяем, существует ли колонка
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 
        FROM information_schema.columns 
        WHERE table_name = 'cards' 
        AND column_name = 'image_url'
    ) THEN
        -- Добавляем колонку, если её нет
        ALTER TABLE cards 
        ADD COLUMN image_url TEXT;
        
        RAISE NOTICE 'Колонка image_url успешно добавлена в таблицу cards';
    ELSE
        RAISE NOTICE 'Колонка image_url уже существует в таблице cards';
    END IF;
END $$;

-- Комментарий к колонке
COMMENT ON COLUMN cards.image_url IS 'URL изображения для карточки (опционально)';

-- Проверяем результат
SELECT column_name, data_type, is_nullable
FROM information_schema.columns
WHERE table_name = 'cards' 
AND column_name IN ('image_url', 'audio_url', 'phonetic', 'example')
ORDER BY column_name;
