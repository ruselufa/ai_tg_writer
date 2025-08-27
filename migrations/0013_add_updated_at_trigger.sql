-- +goose Up
-- Add trigger for automatic updated_at update

-- Function for automatic updated_at update (single line)
CREATE OR REPLACE FUNCTION update_updated_at_column() RETURNS TRIGGER AS 'BEGIN NEW.updated_at = CURRENT_TIMESTAMP; RETURN NEW; END;' language 'plpgsql';

-- Trigger for automatic updated_at update
DROP TRIGGER IF EXISTS update_post_history_updated_at ON post_history;
CREATE TRIGGER update_post_history_updated_at BEFORE UPDATE ON post_history FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- +goose Down
DROP TRIGGER IF EXISTS update_post_history_updated_at ON post_history;
DROP FUNCTION IF EXISTS update_updated_at_column();
