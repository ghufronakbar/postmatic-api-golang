-- +goose Up
-- +goose StatementBegin

-- ENUM (idempotent)
CREATE TYPE token_type AS ENUM ('image_token', 'video_token', 'livestream_token');


CREATE TABLE IF NOT EXISTS app_token_products (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

  token_type token_type NOT NULL,
--   ex: IDR, USD, etc
  currency_code VARCHAR(3) NOT NULL,
--   ex: Rp 1000
  price_amount BIGINT NOT NULL,
--   ex: 833 tokens
  token_amount BIGINT NOT NULL,
  is_active BOOLEAN NOT NULL,
  sort_order INT NOT NULL,

  created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMPTZ

  -- sanity checks
  CONSTRAINT app_token_products_price_positive CHECK (price_amount > 0),
  CONSTRAINT app_token_products_token_positive CHECK (token_amount > 0),
  CONSTRAINT app_token_products_currency_len CHECK (char_length(currency_code) = 3),
  CONSTRAINT app_token_products_currency_upper CHECK (currency_code = UPPER(currency_code))
);

CREATE TRIGGER trigger_app_token_products_updated_at
BEFORE UPDATE ON app_token_products
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

-- seed first row
INSERT INTO app_token_products (id, token_type, currency_code, price_amount, token_amount, is_active, sort_order)
VALUES ('c977673c-910c-4217-b451-b98724558016', 'image_token', 'IDR', 1000, 833, TRUE, 1)
ON CONFLICT (id) DO NOTHING;
-- for the first time, no need to insert to changes table

-- AUDIT / CHANGES (biasanya INSERT-only)
CREATE TABLE IF NOT EXISTS app_token_product_changes (
  id BIGSERIAL PRIMARY KEY,

  -- who made the change
  profile_id UUID NOT NULL,
  FOREIGN KEY (profile_id) REFERENCES profiles(id) ON DELETE RESTRICT,

  -- snapshot of product
  token_type token_type NOT NULL,
  currency_code VARCHAR(3) NOT NULL,
  price_amount BIGINT NOT NULL,
  token_amount BIGINT NOT NULL,
  is_active BOOLEAN NOT NULL,
  sort_order INT NOT NULL,

  created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER trigger_app_token_product_changes_updated_at
BEFORE UPDATE ON app_token_product_changes
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TRIGGER IF EXISTS trigger_app_token_product_changes_updated_at ON app_token_product_changes;
DROP TABLE IF EXISTS app_token_product_changes;

DROP TRIGGER IF EXISTS trigger_app_token_products_updated_at ON app_token_products;
DROP TABLE IF EXISTS app_token_products;

DROP TYPE IF EXISTS token_type;

-- +goose StatementEnd
