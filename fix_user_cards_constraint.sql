-- Исправление constraint для user_cards
-- Проблема: текущий UNIQUE constraint требует user_deck_id, но он может быть NULL
-- Решение: создаем частичный индекс для случаев с NULL и без NULL

-- Удаляем старый constraint
ALTER TABLE user_cards DROP CONSTRAINT IF EXISTS user_cards_user_id_card_id_user_deck_id_key;

-- Создаем уникальный индекс для случаев, когда user_deck_id IS NULL
CREATE UNIQUE INDEX IF NOT EXISTS user_cards_user_card_null_deck_idx 
ON user_cards (user_id, card_id) 
WHERE user_deck_id IS NULL;

-- Создаем уникальный индекс для случаев, когда user_deck_id IS NOT NULL
CREATE UNIQUE INDEX IF NOT EXISTS user_cards_user_card_deck_idx 
ON user_cards (user_id, card_id, user_deck_id) 
WHERE user_deck_id IS NOT NULL;

-- Комментарий
COMMENT ON INDEX user_cards_user_card_null_deck_idx IS 'Уникальность для карточек без привязки к deck';
COMMENT ON INDEX user_cards_user_card_deck_idx IS 'Уникальность для карточек с привязкой к deck';

SELECT 'Constraint fixed successfully!' AS status;
