-- +goose Up
CREATE TABLE IF NOT EXISTS users (
    id BIGINT PRIMARY KEY,
    username VARCHAR(255),
    is_admin BOOLEAN DEFAULT FALSE,
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    tariff VARCHAR(50) DEFAULT 'free',
    usage_count INTEGER DEFAULT 0,
    last_usage TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    referral_code VARCHAR(20) UNIQUE,
    referred_by BIGINT REFERENCES users(id)
);

-- Индексы
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_tariff ON users(tariff);

-- +goose Down
DROP INDEX IF EXISTS idx_users_tariff;
DROP INDEX IF EXISTS idx_users_username;
DROP TABLE IF EXISTS users;