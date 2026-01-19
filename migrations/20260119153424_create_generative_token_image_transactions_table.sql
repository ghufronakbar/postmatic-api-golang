-- +goose Up
-- +goose StatementBegin
CREATE TYPE token_transaction_type AS ENUM ('in', 'out');
CREATE TABLE IF NOT EXISTS generative_token_image_transactions (
    id BIGSERIAL PRIMARY KEY,
    -- type of transaction
    type token_transaction_type NOT NULL,
    amount BIGINT NOT NULL,

    -- profile that take action
    profile_id UUID NOT NULL,
    FOREIGN KEY (profile_id) REFERENCES profiles (id),
    
    -- business that take action
    business_root_id BIGINT NOT NULL,
    FOREIGN KEY (business_root_id) REFERENCES business_roots (id),
    
    -- track from where token type 'in' or 'out'
    -- in
    payment_history_id UUID,
    FOREIGN KEY (payment_history_id) REFERENCES payment_histories (id),
    -- out (TODO later)
    

    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);
CREATE TRIGGER trigger_generative_token_image_transactions_updated_at
BEFORE UPDATE ON generative_token_image_transactions
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trigger_generative_token_image_transactions_updated_at ON generative_token_image_transactions;
DROP TABLE IF EXISTS generative_token_image_transactions;
DROP TYPE IF EXISTS token_transaction_type;
-- +goose StatementEnd
