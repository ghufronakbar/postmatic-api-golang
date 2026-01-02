-- +goose Up
-- +goose StatementBegin
CREATE TYPE app_role AS ENUM ('admin', 'user');

ALTER TABLE profiles
ADD COLUMN role app_role NOT NULL DEFAULT 'user';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE profiles
DROP COLUMN IF EXISTS role;

DROP TYPE app_role;
-- +goose StatementEnd
