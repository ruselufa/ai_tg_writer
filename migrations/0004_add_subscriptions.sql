-- +goose Up
CREATE TABLE subscriptions (
    id SERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id),
    subscription_id INTEGER UNIQUE,
    tariff VARCHAR(50),
    status VARCHAR(20) DEFAULT 'pending',
    amount DECIMAL(10,2),
    next_payment TIMESTAMP,
    last_payment TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    cancelled_at TIMESTAMP,
    active BOOLEAN DEFAULT TRUE
);

-- Индексы
CREATE INDEX idx_subscriptions_user_id ON subscriptions(user_id);
CREATE INDEX idx_subscriptions_status ON subscriptions(status);
CREATE INDEX idx_subscriptions_next_payment ON subscriptions(next_payment);
CREATE INDEX idx_subscriptions_active ON subscriptions(active);

-- +goose Down
DROP INDEX IF EXISTS idx_subscriptions_active;
DROP INDEX IF EXISTS idx_subscriptions_next_payment;
DROP INDEX IF EXISTS idx_subscriptions_status;
DROP INDEX IF EXISTS idx_subscriptions_user_id;
DROP TABLE IF EXISTS subscriptions; 