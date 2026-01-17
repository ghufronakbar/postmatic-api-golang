-- +goose Up
-- +goose StatementBegin
CREATE TYPE app_generative_image_model_provider_type AS ENUM ('openai', 'google');
CREATE TABLE IF NOT EXISTS app_generative_image_models (
    id BIGSERIAL PRIMARY KEY,
    -- ex: "gpt-image-1" UNIQUE (partial with deleted)
    model VARCHAR(255) NOT NULL,

    -- information
    -- ex:"GPT Image 1"
    label VARCHAR(255) NOT NULL,
    image VARCHAR(255),

    -- functional
    -- provider (for switch case generate image)
    provider app_generative_image_model_provider_type NOT NULL,
    -- switch active state
    is_active BOOLEAN NOT NULL,
    -- valid rations (string[])
    valid_ratios TEXT[] NOT NULL,
    -- image_sizes (string[]), if null: models are not supported image sizes
    image_sizes TEXT[],
    

    -- timestamp
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX app_generative_image_models_model_key 
ON app_generative_image_models (model) 
WHERE deleted_at IS NULL;

CREATE TRIGGER trigger_app_generative_image_models_updated_at
BEFORE UPDATE ON app_generative_image_models
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TABLE IF NOT EXISTS app_generative_image_model_changes (
    id BIGSERIAL PRIMARY KEY,
    action action_change_type NOT NULL,

    -- actioner (for seeder no need to input this)
    profile_id UUID NOT NULL,
    FOREIGN KEY (profile_id) REFERENCES profiles (id),

    -- generative image model that has been changed
    generative_image_model_id BIGINT NOT NULL,
    FOREIGN KEY (generative_image_model_id) REFERENCES app_generative_image_models (id),

    -- before (jika create, before = after)
    before_model VARCHAR(255) NOT NULL,
    before_label VARCHAR(255) NOT NULL,
    before_image VARCHAR(255),
    before_provider app_generative_image_model_provider_type NOT NULL,
    before_is_active BOOLEAN NOT NULL,
    before_valid_ratios TEXT[] NOT NULL,
    before_image_sizes TEXT[],

    -- after
    after_model VARCHAR(255) NOT NULL,
    after_label VARCHAR(255) NOT NULL,
    after_image VARCHAR(255),
    after_provider app_generative_image_model_provider_type NOT NULL,
    after_is_active BOOLEAN NOT NULL,
    after_valid_ratios TEXT[] NOT NULL,
    after_image_sizes TEXT[],
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);
CREATE TRIGGER trigger_app_generative_image_model_changes_updated_at
BEFORE UPDATE ON app_generative_image_model_changes
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

-- trigger: when changes inserted/updated, touch rules.updated_at
CREATE OR REPLACE FUNCTION touch_app_generative_image_model_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  UPDATE app_generative_image_models
  SET updated_at = now()
  WHERE id = NEW.generative_image_model_id;

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_touch_app_generative_image_model_updated_at
AFTER INSERT OR UPDATE ON app_generative_image_model_changes
FOR EACH ROW
EXECUTE FUNCTION touch_app_generative_image_model_updated_at();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- 1. Trigger & Function Log
DROP TRIGGER IF EXISTS trigger_touch_app_generative_image_model_updated_at ON app_generative_image_model_changes;
DROP FUNCTION IF EXISTS touch_app_generative_image_model_updated_at;

DROP TRIGGER IF EXISTS trigger_app_generative_image_model_changes_updated_at ON app_generative_image_model_changes;
DROP TABLE IF EXISTS app_generative_image_model_changes;

-- 2. Trigger & Table Utama
DROP TRIGGER IF EXISTS trigger_app_generative_image_models_updated_at ON app_generative_image_models;
DROP INDEX IF EXISTS app_generative_image_models_model_key; 
DROP TABLE IF EXISTS app_generative_image_models;

-- 3. Enum
DROP TYPE IF EXISTS app_generative_image_model_provider_type;
-- +goose StatementEnd