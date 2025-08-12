-- +goose Up
ALTER TABLE subscriptions 
ADD COLUMN failed_attempts INTEGER DEFAULT 0,
ADD COLUMN next_retry TIMESTAMP,
ADD COLUMN suspended_at TIMESTAMP;

-- +goose Down
ALTER TABLE subscriptions 
DROP COLUMN failed_attempts,
DROP COLUMN next_retry,
DROP COLUMN suspended_at;
