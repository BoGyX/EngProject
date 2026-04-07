-- Таблица для хранения текстов для чтения
CREATE TABLE IF NOT EXISTS reading_texts (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    course_id BIGINT REFERENCES courses(id) ON DELETE SET NULL,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    audio_url TEXT DEFAULT '',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_reading_texts_user_id ON reading_texts(user_id);
CREATE INDEX IF NOT EXISTS idx_reading_texts_course_id ON reading_texts(course_id);
