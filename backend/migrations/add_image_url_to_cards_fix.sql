-- Добавляем колонку image_url в таблицу cards если её нет
DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 
        FROM information_schema.columns 
        WHERE table_name = 'cards' 
        AND column_name = 'image_url'
    ) THEN
        ALTER TABLE cards ADD COLUMN image_url TEXT;
        COMMENT ON COLUMN cards.image_url IS 'URL изображения (например, из Unsplash или другого источника)';
    END IF;
END $$;
