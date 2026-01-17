-- +goose Up
-- +goose StatementBegin

-- Provider khusus text (hanya openai dan google)
CREATE TYPE app_generative_text_model_provider_type AS ENUM ('openai', 'google');

CREATE TABLE IF NOT EXISTS app_generative_text_models (
    id BIGSERIAL PRIMARY KEY,
    -- ex: "gpt-4-turbo"
    model VARCHAR(255) NOT NULL,

    -- information
    -- ex: "GPT-4 Turbo"
    label VARCHAR(255) NOT NULL,
    -- path to icon/image (optional, consistent with image models)
    image VARCHAR(255),

    -- functional
    provider app_generative_text_model_provider_type NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,

    -- timestamp
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

-- Partial Index untuk Unique Model (Soft Delete friendly)
CREATE UNIQUE INDEX app_generative_text_models_model_key 
ON app_generative_text_models (model) 
WHERE deleted_at IS NULL;

-- Trigger updated_at main table
CREATE TRIGGER trigger_app_generative_text_models_updated_at
BEFORE UPDATE ON app_generative_text_models
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

-- -- -- LOGGING / AUDIT TRAIL -- -- --

CREATE TABLE IF NOT EXISTS app_generative_text_model_changes (
    id BIGSERIAL PRIMARY KEY,
    action action_change_type NOT NULL, -- Enum shared

    -- actioner
    profile_id UUID NOT NULL,
    FOREIGN KEY (profile_id) REFERENCES profiles (id),

    -- relation
    generative_text_model_id BIGINT NOT NULL,
    FOREIGN KEY (generative_text_model_id) REFERENCES app_generative_text_models (id),

    -- SNAPSHOT BEFORE (Create: Before = After)
    before_model VARCHAR(255) NOT NULL,
    before_label VARCHAR(255) NOT NULL,
    before_image VARCHAR(255),
    before_provider app_generative_text_model_provider_type NOT NULL,
    before_is_active BOOLEAN NOT NULL,

    -- SNAPSHOT AFTER
    after_model VARCHAR(255) NOT NULL,
    after_label VARCHAR(255) NOT NULL,
    after_image VARCHAR(255),
    after_provider app_generative_text_model_provider_type NOT NULL,
    after_is_active BOOLEAN NOT NULL,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

-- Trigger updated_at log table
CREATE TRIGGER trigger_app_generative_text_model_changes_updated_at
BEFORE UPDATE ON app_generative_text_model_changes
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

-- Trigger Update Parent Timestamp
CREATE OR REPLACE FUNCTION touch_app_generative_text_model_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  UPDATE app_generative_text_models
  SET updated_at = now()
  WHERE id = NEW.generative_text_model_id;

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_touch_app_generative_text_model_updated_at
AFTER INSERT OR UPDATE ON app_generative_text_model_changes
FOR EACH ROW
EXECUTE FUNCTION touch_app_generative_text_model_updated_at();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- 1. Clean Log Related
DROP TRIGGER IF EXISTS trigger_touch_app_generative_text_model_updated_at ON app_generative_text_model_changes;
DROP FUNCTION IF EXISTS touch_app_generative_text_model_updated_at;

DROP TRIGGER IF EXISTS trigger_app_generative_text_model_changes_updated_at ON app_generative_text_model_changes;
DROP TABLE IF EXISTS app_generative_text_model_changes;

-- 2. Clean Main Table
DROP TRIGGER IF EXISTS trigger_app_generative_text_models_updated_at ON app_generative_text_models;
-- Drop index manual optional karena table di-drop, tapi good practice
DROP INDEX IF EXISTS app_generative_text_models_model_key;
DROP TABLE IF EXISTS app_generative_text_models;

-- 3. Clean Enum (Only specific for text)
DROP TYPE IF EXISTS app_generative_text_model_provider_type;
-- +goose StatementEnd