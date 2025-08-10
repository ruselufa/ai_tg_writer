-- +goose Up
ALTER TABLE subscriptions
  ADD COLUMN yk_customer_id VARCHAR(64),
  ADD COLUMN yk_payment_method_id VARCHAR(64),
  ADD COLUMN yk_last_payment_id VARCHAR(64);

CREATE INDEX IF NOT EXISTS idx_subscriptions_yk_customer ON subscriptions(yk_customer_id);

-- +goose Down
ALTER TABLE subscriptions
  DROP COLUMN IF EXISTS yk_customer_id,
  DROP COLUMN IF EXISTS yk_payment_method_id,
  DROP COLUMN IF EXISTS yk_last_payment_id;
