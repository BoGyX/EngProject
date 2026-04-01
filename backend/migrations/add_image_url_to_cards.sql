-- Миграция: добавление поля image_url в таблицу cards
-- Выполните этот SQL запрос в вашей базе данных PostgreSQL

ALTER TABLE cards 
ADD COLUMN IF NOT EXISTS image_url TEXT;

-- Комментарий к колонке
COMMENT ON COLUMN cards.image_url IS 'URL изображения для карточки (опционально)';
