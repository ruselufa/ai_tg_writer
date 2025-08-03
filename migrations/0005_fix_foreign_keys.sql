-- +goose Up
-- Удаляем старые таблицы если они существуют
DROP TABLE IF EXISTS usage_stats;
DROP TABLE IF EXISTS voice_messages;

-- Создаем таблицы с правильными внешними ключами
CREATE TABLE IF NOT EXISTS voice_messages (
    id SERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id),
    file_id VARCHAR(255),
    duration INTEGER,
    file_size INTEGER,
    text TEXT,
    rewritten TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS usage_stats (
    id SERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id),
    date DATE DEFAULT CURRENT_DATE,
    usage_count INTEGER DEFAULT 0,
    UNIQUE(user_id, date)
);

-- Индексы для оптимизации
CREATE INDEX IF NOT EXISTS idx_voice_messages_user_id ON voice_messages(user_id);
CREATE INDEX IF NOT EXISTS idx_voice_messages_created_at ON voice_messages(created_at);
CREATE INDEX IF NOT EXISTS idx_usage_stats_user_date ON usage_stats(user_id, date);

-- +goose Down
DROP INDEX IF EXISTS idx_usage_stats_user_date;
DROP INDEX IF EXISTS idx_voice_messages_created_at;
DROP INDEX IF EXISTS idx_voice_messages_user_id;
DROP TABLE IF EXISTS usage_stats;
DROP TABLE IF EXISTS voice_messages;
