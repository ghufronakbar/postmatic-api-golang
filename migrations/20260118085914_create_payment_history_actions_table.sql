-- +goose Up
-- +goose StatementBegin

-- Enum for action value type
CREATE TYPE payment_action_value_type AS ENUM ('image', 'link', 'text', 'claim');

-- Table: payment_history_actions
-- Stores actions (Gopay) or VA numbers (Bank Transfer) from Midtrans
CREATE TABLE IF NOT EXISTS payment_history_actions (
    id BIGSERIAL PRIMARY KEY,
    
    -- relation to payment_histories
    payment_history_id UUID NOT NULL,
    FOREIGN KEY (payment_history_id) REFERENCES payment_histories (id) ON DELETE CASCADE,
    
    -- action details
    name VARCHAR(100) NOT NULL,                              -- dari midtrans (generate-qr-code, deeplink-redirect, dll)
    label VARCHAR(255) NOT NULL,                             -- label readable (QR Code, Deeplink Redirect, Virtual Account)
    value TEXT NOT NULL,                                     -- url atau va number
    value_type payment_action_value_type NOT NULL,           -- image, link, text, claim
    payment_type app_payment_method_type NOT NULL,           -- bank atau ewallet (reuse existing enum)
    action_method VARCHAR(10) NOT NULL DEFAULT 'GET',        -- GET, POST (VA default GET)
    is_public BOOLEAN NOT NULL DEFAULT TRUE,                 -- filter mana yang ditampilkan ke user
    
    -- timestamp
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Index untuk query by payment_history_id
CREATE INDEX idx_payment_history_actions_payment_id 
ON payment_history_actions (payment_history_id);

-- Trigger for updated_at
CREATE TRIGGER trigger_payment_history_actions_updated_at
BEFORE UPDATE ON payment_history_actions
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TRIGGER IF EXISTS trigger_payment_history_actions_updated_at ON payment_history_actions;
DROP INDEX IF EXISTS idx_payment_history_actions_payment_id;
DROP TABLE IF EXISTS payment_history_actions;
DROP TYPE IF EXISTS payment_action_value_type;

-- +goose StatementEnd
