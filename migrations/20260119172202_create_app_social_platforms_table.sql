-- +goose Up
-- +goose StatementBegin

-- 1. Create Enum for Platform Type
CREATE TYPE social_platform_type AS ENUM (
    'linked_in', 
    'facebook_page', 
    'instagram_business', 
    'whatsapp_business', 
    'tiktok', 
    'youtube', 
    'twitter', 
    'pinterest'
);

-- 2. Create Main Table
CREATE TABLE IF NOT EXISTS app_social_platforms (
    id BIGSERIAL PRIMARY KEY,
    
    -- Enum Column
    platform_code social_platform_type NOT NULL,

    -- Basic Info
    logo VARCHAR(255), -- Nullable
    name VARCHAR(255) NOT NULL,
    hint TEXT NOT NULL, -- Text & Not Null

    -- State
    is_active BOOLEAN NOT NULL DEFAULT true,

    -- Timestamp
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

-- Unique Index (Soft Delete Friendly)
CREATE UNIQUE INDEX app_social_platforms_platform_code_key 
ON app_social_platforms (platform_code) 
WHERE deleted_at IS NULL;

-- Trigger Auto Updated At (Main Table)
CREATE TRIGGER trigger_app_social_platforms_updated_at
BEFORE UPDATE ON app_social_platforms
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

-- 3. Create Changes / Log Table
CREATE TABLE IF NOT EXISTS app_social_platform_changes (
    id BIGSERIAL PRIMARY KEY,
    action action_change_type NOT NULL, -- Menggunakan Shared Enum

    -- Actioner
    profile_id UUID NOT NULL,
    FOREIGN KEY (profile_id) REFERENCES profiles (id),

    -- Reference to Main Table
    social_platform_id BIGINT NOT NULL,
    FOREIGN KEY (social_platform_id) REFERENCES app_social_platforms (id),

    -- SNAPSHOT BEFORE
    before_platform_code social_platform_type NOT NULL,
    before_logo VARCHAR(255), -- Nullable mengikuti table asli
    before_name VARCHAR(255) NOT NULL,
    before_hint TEXT NOT NULL,
    before_is_active BOOLEAN NOT NULL,

    -- SNAPSHOT AFTER
    after_platform_code social_platform_type NOT NULL,
    after_logo VARCHAR(255), -- Nullable mengikuti table asli
    after_name VARCHAR(255) NOT NULL,
    after_hint TEXT NOT NULL,
    after_is_active BOOLEAN NOT NULL,

    -- Timestamp Log
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

-- Trigger Auto Updated At (Log Table)
CREATE TRIGGER trigger_app_social_platform_changes_updated_at
BEFORE UPDATE ON app_social_platform_changes
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

-- 4. Trigger Touch Parent Updated At
-- Update 'updated_at' di tabel utama saat ada log baru
CREATE OR REPLACE FUNCTION touch_app_social_platform_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  UPDATE app_social_platforms
  SET updated_at = now()
  WHERE id = NEW.social_platform_id;

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_touch_app_social_platform_updated_at
AFTER INSERT OR UPDATE ON app_social_platform_changes
FOR EACH ROW
EXECUTE FUNCTION touch_app_social_platform_updated_at();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- 1. Drop Touch Triggers & Functions
DROP TRIGGER IF EXISTS trigger_touch_app_social_platform_updated_at ON app_social_platform_changes;
DROP FUNCTION IF EXISTS touch_app_social_platform_updated_at;

-- 2. Drop Changes Table & Triggers
DROP TRIGGER IF EXISTS trigger_app_social_platform_changes_updated_at ON app_social_platform_changes;
DROP TABLE IF EXISTS app_social_platform_changes;

-- 3. Drop Main Table & Triggers
DROP TRIGGER IF EXISTS trigger_app_social_platforms_updated_at ON app_social_platforms;
DROP INDEX IF EXISTS app_social_platforms_platform_code_key; -- Optional explicit drop
DROP TABLE IF EXISTS app_social_platforms;

-- 4. Drop Enum (Only specific to this feature)
DROP TYPE IF EXISTS social_platform_type;

-- NOTE: action_change_type TIDAK didrop karena kemungkinan dipakai tabel lain
-- +goose StatementEnd