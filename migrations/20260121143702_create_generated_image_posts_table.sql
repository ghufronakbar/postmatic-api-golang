-- +goose Up
-- +goose StatementBegin
CREATE TYPE generated_image_post_status AS ENUM ('pending','processing','retry','success', 'failed');
CREATE TYPE generated_image_post_mode AS ENUM ('generate', 'regenerate', 'rss', 'mask');
-- root table / table utama dari setiap request untuk generate image
CREATE TABLE IF NOT EXISTS generated_image_posts (
    id BIGSERIAL PRIMARY KEY,
    status generated_image_post_status NOT NULL,
    error_log TEXT,
    generated_image_prompt TEXT,
    generated_caption_prompt TEXT,

    -- business
    business_root_id BIGINT NOT NULL,
    FOREIGN KEY (business_root_id) REFERENCES business_roots(id) ON DELETE CASCADE,

    -- product knowledge
    business_product_id BIGINT NOT NULL,
    FOREIGN KEY (business_product_id) REFERENCES business_products(id) ON DELETE CASCADE,
    
    -- model
    -- image model (wajib)
    app_generative_image_model_id BIGINT NOT NULL,
    FOREIGN KEY (app_generative_image_model_id) REFERENCES app_generative_image_models(id) ON DELETE CASCADE,
    -- text model (tidak wajib)
    app_generative_text_model_id BIGINT,
    FOREIGN KEY (app_generative_text_model_id) REFERENCES app_generative_text_models(id) ON DELETE CASCADE,

    -- model name and provider (record current value)
    recorded_model_name VARCHAR(255) NOT NULL,
    recorded_model_provider VARCHAR(255) NOT NULL,

    -- from request
    mode generated_image_post_mode NOT NULL,
    num_of_images INT NOT NULL,
    ratio VARCHAR(255) NOT NULL,
    additional_prompt TEXT,
    design_style VARCHAR(40),
    category VARCHAR(40),
    reference_image_url TEXT,
    mask_image_url TEXT,
    image_size VARCHAR(40),
    current_caption TEXT,

    -- advanced generate
    -- checklist business knowledge
    adv_bk_name BOOLEAN NOT NULL,
    adv_bk_category BOOLEAN NOT NULL,
    adv_bk_description BOOLEAN NOT NULL,
    adv_bk_location BOOLEAN NOT NULL,
    adv_bk_logo BOOLEAN NOT NULL,
    adv_bk_unique_selling_point BOOLEAN NOT NULL,
    adv_bk_website BOOLEAN NOT NULL,
    adv_bk_vision_mission BOOLEAN NOT NULL,
    adv_bk_color_tone BOOLEAN NOT NULL,
    -- checklist product knowledge
    adv_pd_name BOOLEAN NOT NULL,
    adv_pd_category BOOLEAN NOT NULL,
    adv_pd_description BOOLEAN NOT NULL,
    adv_pd_price BOOLEAN NOT NULL,
    -- checklist role knowledge
    adv_rl_hashtags BOOLEAN NOT NULL,

    -- rss
    rss_title TEXT,
    rss_url TEXT,
    rss_published_at TIMESTAMPTZ,
    rss_image_url TEXT,
    rss_summary TEXT,
    rss_publisher TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

-- table untuk menyimpan setiap gambar yang dihasilkan (1 to many mandatory)
CREATE TABLE IF NOT EXISTS generated_image_post_items (
    id BIGSERIAL PRIMARY KEY,
    generated_image_post_id BIGINT NOT NULL,
    FOREIGN KEY (generated_image_post_id) REFERENCES generated_image_posts(id) ON DELETE CASCADE,

    image_url TEXT NOT NULL,
    token_used INT NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

-- table untuk menyimpan setiap caption yang dihasilkan (1 to 1 not mandatory)
CREATE TABLE IF NOT EXISTS generated_image_post_captions (
    id BIGSERIAL PRIMARY KEY,
    generated_image_post_id BIGINT NOT NULL UNIQUE,
    FOREIGN KEY (generated_image_post_id) REFERENCES generated_image_posts(id) ON DELETE CASCADE,

    caption_text TEXT NOT NULL,
    token_used INT NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

CREATE TRIGGER trigger_generated_image_post_updated_at
BEFORE UPDATE ON generated_image_posts
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trigger_generated_image_post_items_updated_at
BEFORE UPDATE ON generated_image_post_items
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trigger_generated_image_post_captions_updated_at
BEFORE UPDATE ON generated_image_post_captions
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

-- Trigger logic: Saat child (items/captions) di-insert/update, update parent timestamp
CREATE OR REPLACE FUNCTION touch_generated_image_post_parent_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  UPDATE generated_image_posts
  SET updated_at = NOW()
  WHERE id = NEW.generated_image_post_id;
  
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Pasang Trigger di Table Items
CREATE TRIGGER trigger_touch_generated_image_post_items
AFTER INSERT OR UPDATE ON generated_image_post_items
FOR EACH ROW
EXECUTE FUNCTION touch_generated_image_post_parent_updated_at();

-- Pasang Trigger di Table Captions
CREATE TRIGGER trigger_touch_generated_image_post_captions
AFTER INSERT OR UPDATE ON generated_image_post_captions
FOR EACH ROW
EXECUTE FUNCTION touch_generated_image_post_parent_updated_at();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- 1. Drop Touch Triggers & Function
DROP TRIGGER IF EXISTS trigger_touch_generated_image_post_captions ON generated_image_post_captions;
DROP TRIGGER IF EXISTS trigger_touch_generated_image_post_items ON generated_image_post_items;
DROP FUNCTION IF EXISTS touch_generated_image_post_parent_updated_at;

-- 2. Drop Standard Triggers
DROP TRIGGER IF EXISTS trigger_generated_image_post_updated_at ON generated_image_posts;
DROP TRIGGER IF EXISTS trigger_generated_image_post_items_updated_at ON generated_image_post_items;
DROP TRIGGER IF EXISTS trigger_generated_image_post_captions_updated_at ON generated_image_post_captions;

-- 3. Drop Tables
DROP TABLE IF EXISTS generated_image_post_captions;
DROP TABLE IF EXISTS generated_image_post_items;
DROP TABLE IF EXISTS generated_image_posts;

-- 4. Drop Types
DROP TYPE IF EXISTS generated_image_post_mode;
DROP TYPE IF EXISTS generated_image_post_status;
-- +goose StatementEnd
