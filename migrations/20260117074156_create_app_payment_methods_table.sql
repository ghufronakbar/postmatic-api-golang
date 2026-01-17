-- +goose Up
-- +goose StatementBegin

CREATE TYPE action_change_type AS ENUM ('create', 'update', 'delete');

CREATE TYPE app_payment_admin_type AS ENUM ('fixed', 'percentage');
CREATE TYPE app_payment_method_type AS ENUM ('bank', 'ewallet');
CREATE TABLE IF NOT EXISTS app_payment_methods (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(20) NOT NULL,

    -- information
    name VARCHAR(255) NOT NULL,
    type app_payment_method_type NOT NULL,
    image VARCHAR(255),

    -- functional
    -- tax must be percentage (0-100)
    tax_fee BIGINT NOT NULL,
    -- admin fee can be fixed or percentage
    admin_type app_payment_admin_type NOT NULL,
    admin_fee BIGINT NOT NULL,
    -- switch active state
    is_active BOOLEAN NOT NULL,

    -- timestamp
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX app_payment_methods_code_key 
ON app_payment_methods (code) 
WHERE deleted_at IS NULL;

CREATE TRIGGER trigger_app_payment_methods_updated_at
BEFORE UPDATE ON app_payment_methods
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TABLE IF NOT EXISTS app_payment_method_changes (
    id BIGSERIAL PRIMARY KEY,
    action action_change_type NOT NULL,

    -- actioner (for seeder no need to input this)
    profile_id UUID NOT NULL,
    FOREIGN KEY (profile_id) REFERENCES profiles (id),

    -- payment method that has been changed
    payment_method_id BIGINT NOT NULL,
    FOREIGN KEY (payment_method_id) REFERENCES app_payment_methods (id),

    -- before (jika create, before = after)
    before_code VARCHAR(20) NOT NULL,
    before_name VARCHAR(255) NOT NULL,
    before_type app_payment_method_type NOT NULL,
    before_image VARCHAR(255),
    before_admin_type app_payment_admin_type NOT NULL,
    before_admin_fee BIGINT NOT NULL,
    before_tax_fee BIGINT NOT NULL,
    before_is_active BOOLEAN NOT NULL,

    -- after
    after_code VARCHAR(20) NOT NULL,
    after_name VARCHAR(255) NOT NULL,
    after_type app_payment_method_type NOT NULL,
    after_image VARCHAR(255),
    after_admin_type app_payment_admin_type NOT NULL,
    after_admin_fee BIGINT NOT NULL,
    after_tax_fee BIGINT NOT NULL,
    after_is_active BOOLEAN NOT NULL,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);
CREATE TRIGGER trigger_app_payment_method_changes_updated_at
BEFORE UPDATE ON app_payment_method_changes
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

-- trigger: when changes inserted/updated, touch rules.updated_at
CREATE OR REPLACE FUNCTION touch_app_payment_method_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  UPDATE app_payment_methods
  SET updated_at = now()
  WHERE id = NEW.payment_method_id;

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_touch_app_payment_method_updated_at
AFTER INSERT OR UPDATE ON app_payment_method_changes
FOR EACH ROW
EXECUTE FUNCTION touch_app_payment_method_updated_at();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- 1. Trigger & Function Log
DROP TRIGGER IF EXISTS trigger_touch_app_payment_method_updated_at ON app_payment_method_changes;
DROP FUNCTION IF EXISTS touch_app_payment_method_updated_at;

DROP TRIGGER IF EXISTS trigger_app_payment_method_changes_updated_at ON app_payment_method_changes;
DROP TABLE IF EXISTS app_payment_method_changes;

-- 2. Trigger & Table Utama
DROP TRIGGER IF EXISTS trigger_app_payment_methods_updated_at ON app_payment_methods;
DROP INDEX IF EXISTS app_payment_methods_code_key; 
DROP TABLE IF EXISTS app_payment_methods;

-- 3. Enum
DROP TYPE IF EXISTS action_change_type;
DROP TYPE IF EXISTS app_payment_admin_type;
DROP TYPE IF EXISTS app_payment_method_type;
-- +goose StatementEnd