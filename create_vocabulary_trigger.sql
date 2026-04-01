-- Триггер для автоматического добавления слов в личный словарь
-- после прохождения всех режимов обучения

CREATE OR REPLACE FUNCTION add_to_personal_vocabulary()
RETURNS TRIGGER AS $$
DECLARE
    card_data RECORD;
BEGIN
    -- Проверяем, пройдены ли все режимы
    IF NEW.mode_view = TRUE AND 
       NEW.mode_with_photo = TRUE AND 
       NEW.mode_without_photo = TRUE AND 
       NEW.mode_russian = TRUE AND 
       NEW.mode_constructor = TRUE THEN
        
        -- Получаем данные карточки
        SELECT word, translation, phonetic, audio_url, image_url, example
        INTO card_data
        FROM cards
        WHERE id = NEW.card_id;
        
        -- Добавляем в личный словарь, если еще нет
        INSERT INTO personal_vocabulary (
            user_id, 
            word, 
            translation, 
            phonetic, 
            audio_url, 
            example, 
            status,
            created_at,
            updated_at
        )
        VALUES (
            NEW.user_id,
            card_data.word,
            card_data.translation,
            card_data.phonetic,
            card_data.audio_url,
            card_data.example,
            'learned',
            NOW(),
            NOW()
        )
        ON CONFLICT (user_id, word) 
        DO UPDATE SET
            status = 'learned',
            updated_at = NOW();
            
        RAISE NOTICE 'Word "%" added to personal vocabulary for user %', card_data.word, NEW.user_id;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Удаляем старый триггер, если существует
DROP TRIGGER IF EXISTS trigger_add_to_vocabulary ON user_cards;

-- Создаем триггер
CREATE TRIGGER trigger_add_to_vocabulary
    AFTER INSERT OR UPDATE ON user_cards
    FOR EACH ROW
    EXECUTE FUNCTION add_to_personal_vocabulary();

SELECT 'Trigger created successfully!' AS status;
