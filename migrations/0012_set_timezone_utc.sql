-- +goose Up
-- Устанавливаем временную зону в UTC для текущей сессии
-- В продакшене рекомендуется настроить postgresql.conf

-- Проверяем текущую временную зону
SELECT current_setting('timezone') as current_timezone;

-- Устанавливаем UTC для текущей сессии
SET timezone = 'UTC';

-- Проверяем результат
SELECT current_setting('timezone') as new_timezone,
       NOW() as current_time,
       NOW() AT TIME ZONE 'UTC' as utc_time;

-- +goose Down
-- Возвращаем предыдущую временную зону (если нужно)
-- В реальности лучше оставить UTC