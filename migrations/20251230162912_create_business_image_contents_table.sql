-- +goose Up
-- +goose StatementBegin
CREATE TYPE business_image_content_type   AS ENUM ('personal', 'generated');

-- 1 to many with business_root_id
-- many to one with business_product_id (nullable)
CREATE TABLE IF NOT EXISTS business_image_contents (
	id BIGSERIAL PRIMARY KEY,
    image_urls VARCHAR(255)[] NOT NULL,
    caption TEXT,
    type business_image_content_type NOT NULL,
    ready_to_post BOOLEAN NOT NULL,
    category VARCHAR(255) NOT NULL,

    business_product_id BIGINT NULL,
    FOREIGN KEY (business_product_id) REFERENCES business_products (id) ON DELETE CASCADE,

	business_root_id BIGSERIAL NOT NULL,
    FOREIGN KEY (business_root_id) REFERENCES business_roots (id) ON DELETE CASCADE,

	created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);
CREATE TRIGGER trg_business_image_contents_set_updated_at
BEFORE UPDATE ON business_image_contents
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE business_image_contents;
DROP TYPE business_image_content_type;
-- +goose StatementEnd
