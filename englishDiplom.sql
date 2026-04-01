-- ===================================
-- EXTENSIONS
-- ===================================
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ===================================
-- USERS (users + admins)
-- ===================================
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    name TEXT,
    role TEXT CHECK(role IN ('user','admin')) DEFAULT 'user',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- ===================================
-- COURSES (например: School / Life / IT)
-- ===================================
CREATE TABLE courses (
    id BIGSERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    slug TEXT UNIQUE NOT NULL,
    description TEXT,
    image_url TEXT,
    is_published BOOLEAN DEFAULT false,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP DEFAULT NOW()
);

-- ===================================
-- DECKS / LESSONS (подразделы курса)
-- Например: урок биологии, химии в курсе "Школа"
-- ===================================
CREATE TABLE decks (
    id BIGSERIAL PRIMARY KEY,
    course_id BIGINT REFERENCES courses(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    slug TEXT NOT NULL,
    description TEXT,
    position INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(course_id, slug)
);

-- ===================================
-- CARDS (flashcards)
-- ===================================
CREATE TABLE cards (
    id BIGSERIAL PRIMARY KEY,
    deck_id BIGINT REFERENCES decks(id) ON DELETE CASCADE,
    
    word TEXT NOT NULL,
    translation TEXT NOT NULL,
    phonetic TEXT,
    
    -- Аудио: либо URL от API, либо MP3 в БД
    audio_url TEXT,  -- если есть в API
    audio_mp3 BYTEA, -- если добавляется пользователем вручную
    
    -- Изображение
    image BYTEA, -- картинка хранится в БД
    image_url TEXT, -- URL изображения (например, из Unsplash или другого источника)
    
    example TEXT,
    
    -- Кто создал (NULL = системная, UUID = пользовательская)
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    is_custom BOOLEAN DEFAULT false,
    
    created_at TIMESTAMP DEFAULT NOW()
);

-- ===================================
-- PERSONAL_VOCABULARY (личный словарь пользователя)
-- Слова, которые пользователь добавляет вне курсов
-- ===================================
CREATE TABLE personal_vocabulary (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    word TEXT NOT NULL,
    translation TEXT NOT NULL,
    phonetic TEXT,
    
    -- Аудио: либо URL от API, либо MP3 в БД
    audio_url TEXT,
    audio_mp3 BYTEA,
    
    -- Изображение
    image BYTEA,
    
    example TEXT,
    
    -- Дополнительные поля для личного словаря
    notes TEXT, -- личные заметки пользователя
    tags TEXT[], -- теги для категоризации (массив строк)
    
    -- Статус изучения (можно синхронизировать с user_cards при необходимости)
    status TEXT CHECK(status IN ('new','learning','learned')) DEFAULT 'new',
    correct_count INT DEFAULT 0,
    wrong_count INT DEFAULT 0,
    last_seen TIMESTAMP,
    next_review TIMESTAMP,
    
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    
    UNIQUE(user_id, word) -- один пользователь не может добавить одно слово дважды
);

-- ===================================
-- USER_COURSES (подписка на курс / общий прогресс)
-- Содержит общий прогресс по курсу
-- ===================================
CREATE TABLE user_courses (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    course_id BIGINT REFERENCES courses(id) ON DELETE CASCADE,
    started_at TIMESTAMP DEFAULT NOW(),
    last_opened_at TIMESTAMP DEFAULT NOW(),
    completed_at TIMESTAMP,
    is_active BOOLEAN DEFAULT false,
    attempt_number INT DEFAULT 1,
    
    -- Прогресс курса (можно считать динамически, но храним для производительности)
    completed_decks_count INT DEFAULT 0, -- сколько decks пройдено
    total_decks_count INT DEFAULT 0, -- всего decks в курсе (на момент начала)
    progress_percentage DECIMAL(5,2) DEFAULT 0.00, -- процент выполнения (0-100)
    
    UNIQUE(user_id, course_id, attempt_number)
);

-- ===================================
-- USER_DECKS (прогресс по каждому подкурсу/deck)
-- Отслеживает прогресс пользователя по каждому уроку (биология, химия и т.д.)
-- ===================================
CREATE TABLE user_decks (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    deck_id BIGINT NOT NULL REFERENCES decks(id) ON DELETE CASCADE,
    user_course_id BIGINT REFERENCES user_courses(id) ON DELETE CASCADE, -- связь с попыткой прохождения курса
    
    -- Прогресс по deck
    status TEXT CHECK(status IN ('not_started','in_progress','completed')) DEFAULT 'not_started',
    learned_cards_count INT DEFAULT 0, -- сколько карточек выучено (status='learned')
    total_cards_count INT DEFAULT 0, -- всего карточек в deck
    progress_percentage DECIMAL(5,2) DEFAULT 0.00, -- процент выполнения deck
    
    started_at TIMESTAMP,
    last_opened_at TIMESTAMP DEFAULT NOW(),
    completed_at TIMESTAMP,
    is_active BOOLEAN DEFAULT false,
    
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    
    UNIQUE(user_id, deck_id, user_course_id)
);

-- ===================================
-- USER_CARDS (прогресс пользователя по карточкам)
-- ===================================
CREATE TABLE user_cards (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    card_id BIGINT NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
    user_deck_id BIGINT REFERENCES user_decks(id) ON DELETE CASCADE, -- связь с прогрессом deck
    
    status TEXT CHECK(status IN ('new','learning','learned')) DEFAULT 'new',
    correct_count INT DEFAULT 0,
    wrong_count INT DEFAULT 0,
    last_seen TIMESTAMP,
    next_review TIMESTAMP,
    
    -- Режимы обучения
    mode_view BOOLEAN DEFAULT FALSE,
    mode_choice BOOLEAN DEFAULT FALSE,
    mode_with_photo BOOLEAN DEFAULT FALSE,
    mode_without_photo BOOLEAN DEFAULT FALSE,
    mode_russian BOOLEAN DEFAULT FALSE,
    mode_constructor BOOLEAN DEFAULT FALSE,
    current_mode TEXT NOT NULL DEFAULT 'view',
    progress_percentage INT NOT NULL DEFAULT 0,
    
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    
    UNIQUE(user_id, card_id, user_deck_id)
);

-- ===================================
-- TRAINING_SESSIONS (тренировки)
-- ===================================
CREATE TABLE training_sessions (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    course_id BIGINT REFERENCES courses(id) ON DELETE CASCADE,
    deck_id BIGINT REFERENCES decks(id), -- добавил deck для статистики по урокам
    user_deck_id BIGINT REFERENCES user_decks(id) ON DELETE CASCADE,
    started_at TIMESTAMP DEFAULT NOW(),
    finished_at TIMESTAMP
);

CREATE TABLE training_session_cards (
    id BIGSERIAL PRIMARY KEY,
    session_id BIGINT NOT NULL REFERENCES training_sessions(id) ON DELETE CASCADE,
    card_id BIGINT NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
    position INT NOT NULL,
    sequence_number INT NOT NULL,
    current_mode TEXT NOT NULL DEFAULT 'view',
    progress_percentage INT NOT NULL DEFAULT 0,
    is_completed BOOLEAN NOT NULL DEFAULT FALSE,
    last_answer_correct BOOLEAN,
    correct_answers INT NOT NULL DEFAULT 0,
    wrong_answers INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(session_id, card_id),
    UNIQUE(session_id, position),
    UNIQUE(session_id, sequence_number)
);

-- ===================================
-- TRAINING_ANSWERS (ответы на карточки в сессии)
-- ===================================
CREATE TABLE training_answers (
    id BIGSERIAL PRIMARY KEY,
    session_id BIGINT REFERENCES training_sessions(id) ON DELETE CASCADE,
    card_id BIGINT REFERENCES cards(id) ON DELETE CASCADE,
    is_correct BOOLEAN,
    answered_at TIMESTAMP DEFAULT NOW()
);

-- ===================================
-- REFRESH_TOKENS (JWT для авторизации)
-- ===================================
CREATE TABLE refresh_tokens (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    token TEXT UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- ===================================
-- INDEXES для производительности
-- ===================================
CREATE INDEX idx_courses_published ON courses(is_published);
CREATE INDEX idx_courses_slug ON courses(slug);
CREATE INDEX idx_decks_course ON decks(course_id);
CREATE INDEX idx_decks_course_slug ON decks(course_id, slug);
CREATE INDEX idx_cards_deck ON cards(deck_id);
CREATE INDEX idx_cards_created_by ON cards(created_by);
CREATE INDEX idx_cards_is_custom ON cards(is_custom);
CREATE UNIQUE INDEX idx_cards_custom_user_deck_word ON cards (deck_id, created_by, LOWER(word))
WHERE is_custom = true AND created_by IS NOT NULL;

-- Индексы для личного словаря
CREATE INDEX idx_personal_vocabulary_user ON personal_vocabulary(user_id);
CREATE INDEX idx_personal_vocabulary_status ON personal_vocabulary(status);
CREATE INDEX idx_personal_vocabulary_next_review ON personal_vocabulary(next_review);
CREATE INDEX idx_personal_vocabulary_word ON personal_vocabulary(word);
-- Индекс для поиска по тегам (GIN индекс для массивов)
CREATE INDEX idx_personal_vocabulary_tags ON personal_vocabulary USING GIN(tags);

CREATE INDEX idx_user_courses_user ON user_courses(user_id);
CREATE INDEX idx_user_courses_user_course ON user_courses(user_id, course_id);
CREATE INDEX idx_user_courses_attempt ON user_courses(user_id, course_id, attempt_number);
CREATE INDEX idx_user_courses_active ON user_courses(user_id, is_active);

CREATE INDEX idx_user_decks_user ON user_decks(user_id);
CREATE INDEX idx_user_decks_deck ON user_decks(deck_id);
CREATE INDEX idx_user_decks_user_deck ON user_decks(user_id, deck_id);
CREATE INDEX idx_user_decks_user_course ON user_decks(user_course_id);
CREATE INDEX idx_user_decks_status ON user_decks(status);
CREATE INDEX idx_user_decks_active ON user_decks(user_id, is_active);

CREATE INDEX idx_user_cards_user ON user_cards(user_id);
CREATE INDEX idx_user_cards_card ON user_cards(card_id);
CREATE INDEX idx_user_cards_next_review ON user_cards(next_review);
CREATE INDEX idx_user_cards_user_deck ON user_cards(user_deck_id);
CREATE INDEX idx_user_cards_status ON user_cards(status);
CREATE INDEX idx_user_cards_progress ON user_cards(user_id, progress_percentage);

CREATE INDEX idx_training_sessions_user ON training_sessions(user_id);
CREATE INDEX idx_training_sessions_deck ON training_sessions(deck_id);
CREATE INDEX idx_training_session_cards_session ON training_session_cards(session_id);

-- ===================================
-- TRIGGERS для updated_at
-- ===================================
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_users_updated
BEFORE UPDATE ON users
FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_personal_vocabulary_updated
BEFORE UPDATE ON personal_vocabulary
FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_user_decks_updated
BEFORE UPDATE ON user_decks
FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_user_cards_updated
BEFORE UPDATE ON user_cards
FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ===================================
-- TRIGGER для автоматического обновления прогресса deck
-- Обновляет прогресс deck при изменении статуса карточки
-- ===================================
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
    
    -- Обновляем прогресс deck
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
      AND user_deck_id = NEW.user_deck_id;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ОТКЛЮЧЕН: Триггер вызывает ошибки с user_deck_id
-- CREATE TRIGGER trg_update_deck_progress
-- AFTER INSERT OR UPDATE ON user_cards
-- FOR EACH ROW EXECUTE FUNCTION update_deck_progress();

-- ===================================
-- TRIGGER для автоматического обновления прогресса курса
-- Обновляет общий прогресс курса при изменении прогресса deck
-- ===================================
CREATE OR REPLACE FUNCTION update_course_progress()
RETURNS TRIGGER AS $$
DECLARE
    course_id_var BIGINT;
    completed_decks INT;
    total_decks INT;
    progress_pct DECIMAL(5,2);
BEGIN
    -- Получаем course_id из deck_id
    SELECT course_id INTO course_id_var FROM decks WHERE id = NEW.deck_id;
    
    -- Подсчитываем завершенные decks
    SELECT COUNT(*) INTO completed_decks
    FROM user_decks
    WHERE user_id = NEW.user_id 
      AND user_course_id = NEW.user_course_id
      AND status = 'completed';
    
    -- Всего decks в курсе
    SELECT COUNT(*) INTO total_decks
    FROM decks
    WHERE course_id = course_id_var;
    
    -- Рассчитываем процент
    IF total_decks > 0 THEN
        progress_pct := (completed_decks::DECIMAL / total_decks::DECIMAL) * 100;
    ELSE
        progress_pct := 0;
    END IF;
    
    -- Обновляем прогресс курса
    UPDATE user_courses
    SET completed_decks_count = completed_decks,
        total_decks_count = total_decks,
        progress_percentage = progress_pct,
        completed_at = CASE 
            WHEN completed_decks = total_decks AND total_decks > 0 AND completed_at IS NULL THEN NOW()
            ELSE completed_at
        END
    WHERE user_id = NEW.user_id 
      AND course_id = course_id_var
      AND id = NEW.user_course_id;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_update_course_progress
AFTER INSERT OR UPDATE ON user_decks
FOR EACH ROW EXECUTE FUNCTION update_course_progress();

-- ===================================
-- Таблица для текстов для чтения (Ридер)
-- ===================================
CREATE TABLE IF NOT EXISTS reading_texts (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_reading_texts_user_id ON reading_texts(user_id);
