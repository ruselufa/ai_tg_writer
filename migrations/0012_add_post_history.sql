-- +goose Up
-- Таблица истории создания постов
CREATE TABLE IF NOT EXISTS post_history (
    id SERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    
    -- Голосовое сообщение
    voice_text TEXT NOT NULL,
    voice_file_id VARCHAR(255),
    voice_duration INTEGER,
    voice_file_size INTEGER,
    
    -- Тайминги
    voice_sent_at TIMESTAMP WITH TIME ZONE NOT NULL,           -- Время отправки в Whisper
    voice_received_at TIMESTAMP WITH TIME ZONE,                -- Время получения от Whisper
    ai_sent_at TIMESTAMP WITH TIME ZONE,                       -- Время отправки в AI LLM
    ai_received_at TIMESTAMP WITH TIME ZONE,                   -- Время получения от AI LLM
    
    -- AI ответ
    ai_response TEXT,                                           -- Полный ответ от AI
    ai_model VARCHAR(100) DEFAULT 'deepseek',                   -- Тип нейросети
    ai_tokens_used INTEGER,                                     -- Количество токенов
    ai_cost DECIMAL(10,6),                                     -- Стоимость запроса (если доступно)
    
    -- Статус
    is_saved BOOLEAN DEFAULT FALSE,                             -- Сохранил ли пользователь пост
    saved_at TIMESTAMP WITH TIME ZONE,                          -- Когда сохранил
    
    -- Метаданные
    processing_duration_ms INTEGER,                             -- Общее время обработки в мс
    whisper_duration_ms INTEGER,                                -- Время транскрипции в мс
    ai_generation_duration_ms INTEGER,                          -- Время генерации AI в мс
    
    -- Системные поля
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Индексы для оптимизации
CREATE INDEX IF NOT EXISTS idx_post_history_user_id ON post_history(user_id);
CREATE INDEX IF NOT EXISTS idx_post_history_created_at ON post_history(created_at);
CREATE INDEX IF NOT EXISTS idx_post_history_is_saved ON post_history(is_saved);
CREATE INDEX IF NOT EXISTS idx_post_history_ai_model ON post_history(ai_model);
CREATE INDEX IF NOT EXISTS idx_post_history_voice_sent_at ON post_history(voice_sent_at);

-- 

-- +goose Down
-- DROP TRIGGER IF EXISTS update_post_history_updated_at ON post_history;
-- DROP FUNCTION IF EXISTS update_updated_at_column();
-- DROP INDEX IF EXISTS idx_post_history_voice_sent_at;
-- DROP INDEX IF EXISTS idx_post_history_ai_model;
-- DROP INDEX IF EXISTS idx_post_history_is_saved;
-- DROP INDEX IF EXISTS idx_post_history_created_at;
-- DROP INDEX IF EXISTS idx_post_history_user_id;
DROP TABLE IF EXISTS post_history;
