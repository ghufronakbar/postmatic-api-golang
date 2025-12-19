-- +goose Up
-- +goose StatementBegin
ALTER TABLE business_members ADD COLUMN deleted_at TIMESTAMPTZ;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE business_members DROP COLUMN deleted_at;
-- +goose StatementEnd
