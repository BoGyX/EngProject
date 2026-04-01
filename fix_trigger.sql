CREATE OR REPLACE FUNCTION update_deck_progress()
RETURNS TRIGGER AS $$
DECLARE
    deck_id_var BIGINT;
    learned_count INT;
    total_count INT;
    progress_pct DECIMAL(5,2);
BEGIN
    -- Получаем deck_id из card_id
    SELECT deck_id INTO deck_id_var FROM cards WHERE id = NEW.card_id;
    
    -- Если user_deck_id NULL, пропускаем обновление
    IF NEW.user_deck_id IS NULL THEN
        RETURN NEW;
    END IF;
    
    -- Подсчитываем выученные карточки (status='learned')
    SELECT COUNT(*) INTO learned_count
    FROM user_cards uc
    JOIN cards c ON uc.card_id = c.id
    WHERE uc.user_id = NEW.user_id 
      AND uc.user_deck_id = NEW.user_deck_id
      AND uc.status = 'learned'
      AND c.deck_id = deck_id_var;
    
    -- Всего карточек в deck
    SELECT COUNT(*) INTO total_count
    FROM cards
    WHERE deck_id = deck_id_var;
    
    -- Рассчитываем процент
    IF total_count > 0 THEN
        progress_pct := (learned_count::DECIMAL / total_count::DECIMAL) * 100;
    ELSE
        progress_pct := 0;
    END IF;
    
    -- Обновляем прогресс deck (используем id вместо user_deck_id)
    UPDATE user_decks
    SET learned_cards_count = learned_count,
        total_cards_count = total_count,
        progress_percentage = progress_pct,
        status = CASE 
            WHEN learned_count = total_count AND total_count > 0 THEN 'completed'
            WHEN learned_count > 0 THEN 'in_progress'
            ELSE 'not_started'
        END,
        completed_at = CASE 
            WHEN learned_count = total_count AND total_count > 0 AND completed_at IS NULL THEN NOW()
            ELSE completed_at
        END
    WHERE user_id = NEW.user_id 
      AND deck_id = deck_id_var
      AND id = NEW.user_deck_id;  -- ИСПРАВЛЕНО: было user_deck_id = NEW.user_deck_id
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Пересоздаем триггер
CREATE TRIGGER trg_update_deck_progress
AFTER INSERT OR UPDATE ON user_cards
FOR EACH ROW
EXECUTE FUNCTION update_deck_progress();
