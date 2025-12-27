-- +goose Up
ALTER TYPE image_provider ADD VALUE IF NOT EXISTS 's3';

-- +goose Down
-- no-op (Postgres tidak mendukung drop enum value secara langsung)
