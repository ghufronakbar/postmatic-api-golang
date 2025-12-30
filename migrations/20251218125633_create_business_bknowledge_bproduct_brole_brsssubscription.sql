-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS business_roots ( 
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 1 to 1 with business_root
CREATE TABLE IF NOT EXISTS business_knowledges ( 
    id BIGSERIAL PRIMARY KEY,

    name VARCHAR(255) NOT NULL,
    primary_logo_url VARCHAR(255),
    category VARCHAR(255) NOT NULL,
    description TEXT,

    unique_selling_point TEXT,
    website_url VARCHAR(255),
    vision_mission TEXT,
    location VARCHAR(255),
    color_tone VARCHAR(6),

    business_root_id BIGINT NOT NULL UNIQUE,
    FOREIGN KEY (business_root_id) REFERENCES business_roots(id) ON DELETE CASCADE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- 1 to many with business_root
CREATE TABLE IF NOT EXISTS business_products ( 
    id BIGSERIAL PRIMARY KEY,

    name VARCHAR(255) NOT NULL,
    category VARCHAR(255) NOT NULL,
    description TEXT,

    currency CHAR(3) NOT NULL,
    price BIGINT NOT NULL,

    image_urls VARCHAR(255)[] NOT NULL,

    business_root_id BIGINT NOT NULL,
    FOREIGN KEY (business_root_id) REFERENCES business_roots(id) ON DELETE CASCADE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- 1 to 1 with business_root
CREATE TABLE IF NOT EXISTS business_roles ( 
    id BIGSERIAL PRIMARY KEY,

    target_audience VARCHAR(255) NOT NULL,
    tone VARCHAR(255) NOT NULL,
    audience_persona VARCHAR(255) NOT NULL,
    hashtags VARCHAR(255)[] NOT NULL,
    call_to_action VARCHAR(255) NOT NULL,
    goals TEXT,

    business_root_id BIGINT NOT NULL UNIQUE,
    FOREIGN KEY (business_root_id) REFERENCES business_roots(id) ON DELETE CASCADE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- 1 to many with business_root
CREATE TABLE IF NOT EXISTS business_rss_subscriptions ( 
    id BIGSERIAL PRIMARY KEY,

    title VARCHAR(255) NOT NULL,
    is_active BOOLEAN NOT NULL,

    business_root_id BIGINT NOT NULL,
    FOREIGN KEY (business_root_id) REFERENCES business_roots(id) ON DELETE CASCADE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- Func for trigger root
CREATE OR REPLACE FUNCTION touch_business_root_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  IF TG_OP = 'INSERT' THEN
    UPDATE business_roots
    SET updated_at = CURRENT_TIMESTAMP
    WHERE id = NEW.business_root_id;

    RETURN NEW;

  ELSIF TG_OP = 'UPDATE' THEN
    -- Kalau pindah root, touch root lama juga
    IF OLD.business_root_id IS DISTINCT FROM NEW.business_root_id THEN
      UPDATE business_roots
      SET updated_at = CURRENT_TIMESTAMP
      WHERE id = OLD.business_root_id;
    END IF;

    UPDATE business_roots
    SET updated_at = CURRENT_TIMESTAMP
    WHERE id = NEW.business_root_id;

    RETURN NEW;

  ELSIF TG_OP = 'DELETE' THEN
    UPDATE business_roots
    SET updated_at = CURRENT_TIMESTAMP
    WHERE id = OLD.business_root_id;

    RETURN OLD;
  END IF;

  RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Trigger Each Table

CREATE TRIGGER trg_business_roots_set_updated_at
BEFORE UPDATE ON business_roots
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_business_knowledges_set_updated_at
BEFORE UPDATE ON business_knowledges
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_business_products_set_updated_at
BEFORE UPDATE ON business_products
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_business_roles_set_updated_at
BEFORE UPDATE ON business_roles
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_business_rss_subscriptions_set_updated_at
BEFORE UPDATE ON business_rss_subscriptions
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

-- Trigger For Root
CREATE TRIGGER trg_touch_root_from_knowledges
AFTER INSERT OR UPDATE OR DELETE ON business_knowledges
FOR EACH ROW
EXECUTE FUNCTION touch_business_root_updated_at();

CREATE TRIGGER trg_touch_root_from_products
AFTER INSERT OR UPDATE OR DELETE ON business_products
FOR EACH ROW
EXECUTE FUNCTION touch_business_root_updated_at();

CREATE TRIGGER trg_touch_root_from_roles
AFTER INSERT OR UPDATE OR DELETE ON business_roles
FOR EACH ROW
EXECUTE FUNCTION touch_business_root_updated_at();

CREATE TRIGGER trg_touch_root_from_rss_subscriptions
AFTER INSERT OR UPDATE OR DELETE ON business_rss_subscriptions
FOR EACH ROW
EXECUTE FUNCTION touch_business_root_updated_at();


-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trg_touch_root_from_rss_subscriptions ON business_rss_subscriptions;
DROP TRIGGER IF EXISTS trg_touch_root_from_roles ON business_roles;
DROP TRIGGER IF EXISTS trg_touch_root_from_products ON business_products;
DROP TRIGGER IF EXISTS trg_touch_root_from_knowledges ON business_knowledges;

DROP TRIGGER IF EXISTS trg_business_rss_subscriptions_set_updated_at ON business_rss_subscriptions;
DROP TRIGGER IF EXISTS trg_business_roles_set_updated_at ON business_roles;
DROP TRIGGER IF EXISTS trg_business_products_set_updated_at ON business_products;
DROP TRIGGER IF EXISTS trg_business_knowledges_set_updated_at ON business_knowledges;
DROP TRIGGER IF EXISTS trg_business_roots_set_updated_at ON business_roots;

-- kalau function ini hanya dipakai di migration ini:
DROP FUNCTION IF EXISTS touch_business_root_updated_at();
-- set_updated_at() drop hanya kalau memang kamu mau hilangkan juga

DROP TABLE IF EXISTS business_rss_subscriptions;
DROP TABLE IF EXISTS business_roles;
DROP TABLE IF EXISTS business_products;
DROP TABLE IF EXISTS business_knowledges;
DROP TABLE IF EXISTS business_roots;
-- +goose StatementEnd
