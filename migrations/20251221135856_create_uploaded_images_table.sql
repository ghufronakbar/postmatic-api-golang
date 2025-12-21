-- +goose Up
-- +goose StatementBegin
CREATE TYPE image_provider AS ENUM ('cloudinary');

CREATE TABLE uploaded_images (
    id BIGSERIAL PRIMARY KEY,
    hashkey VARCHAR(255) NOT NULL UNIQUE,
    public_id VARCHAR(255) NOT NULL,
    size BIGINT NOT NULL,
    image_url VARCHAR(255) NOT NULL,
    provider image_provider NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE uploaded_images;
DROP TYPE image_provider;
-- +goose StatementEnd
