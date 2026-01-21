-- +goose Up
-- +goose StatementBegin
CREATE TABLE business_saved_template_creator_images (
    id BIGSERIAL PRIMARY KEY,
    
    -- foreign to business
    business_root_id BIGINT NOT NULL,
    FOREIGN KEY (business_root_id) REFERENCES business_roots(id) ON DELETE CASCADE, -- Tambahkan ON DELETE CASCADE (Optional tapi recommended)

    -- foreign to creator image
    creator_image_id BIGINT NOT NULL,
    FOREIGN KEY (creator_image_id) REFERENCES creator_images(id) ON DELETE CASCADE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

-- 1 business hanya boleh menyimpan 1 creator image yang sama
CREATE UNIQUE INDEX idx_unique_business_saved_creator_img 
ON business_saved_template_creator_images (business_root_id, creator_image_id)
WHERE deleted_at IS NULL;

CREATE TRIGGER trigger_business_saved_template_creator_images_updated_at
BEFORE UPDATE ON business_saved_template_creator_images
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trigger_business_saved_template_creator_images_updated_at ON business_saved_template_creator_images;

DROP INDEX IF EXISTS idx_unique_business_saved_creator_img; 

DROP TABLE IF EXISTS business_saved_template_creator_images;
-- +goose StatementEnd