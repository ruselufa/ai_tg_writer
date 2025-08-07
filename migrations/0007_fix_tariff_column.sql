-- +goose Up
-- Добавляем поле tariff в таблицу users
ALTER TABLE users ADD COLUMN IF NOT EXISTS tariff VARCHAR(50) DEFAULT 'free';

-- Обновляем существующих пользователей на тариф 'free'
UPDATE users SET tariff = 'free' WHERE tariff IS NULL;

-- +goose Down
-- Удаляем поле tariff из таблицы users
ALTER TABLE users DROP COLUMN IF EXISTS tariff; 