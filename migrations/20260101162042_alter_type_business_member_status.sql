-- +goose Up
-- +goose StatementBegin
ALTER TYPE business_member_status ADD VALUE IF NOT EXISTS 'expired';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- +goose StatementEnd
