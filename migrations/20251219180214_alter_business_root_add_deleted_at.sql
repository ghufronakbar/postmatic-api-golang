-- +goose Up
-- +goose StatementBegin
ALTER TABLE business_roots
ADD COLUMN deleted_at TIMESTAMPTZ;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE business_roots
DROP COLUMN deleted_at;
-- +goose StatementEnd
