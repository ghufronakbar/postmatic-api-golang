CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;


CREATE TYPE auth_provider AS ENUM ('google', 'credential');
CREATE TABLE IF NOT EXISTS profiles (
    -- Tambahkan DEFAULT gen_random_uuid()
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(), 
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,

    -- Gunakan sql.NullString di Go nantinya
    image_url VARCHAR(255),
    country_code VARCHAR(4),
    phone VARCHAR(20),
    description TEXT,

    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP

);

CREATE TRIGGER trigger_profiles_updated_at
BEFORE UPDATE ON profiles
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    password VARCHAR(255),
    provider auth_provider NOT NULL,
    verified_at TIMESTAMPTZ,

    -- Tambahkan NOT NULL agar di Go tipenya uuid.UUID (bukan NullUUID)
    profile_id UUID NOT NULL, 
    FOREIGN KEY (profile_id) REFERENCES profiles(id) ON DELETE CASCADE,

    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP

);

CREATE TRIGGER trigger_users_updated_at
BEFORE UPDATE ON users
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TABLE IF NOT EXISTS products ( 
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    price INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TRIGGER trigger_products_updated_at
BEFORE UPDATE ON products
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TABLE IF NOT EXISTS business_roots ( 
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 1 to 1 with business_root
CREATE TABLE IF NOT EXISTS business_knowledges ( 
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    name VARCHAR(255) NOT NULL,
    primary_logo_url VARCHAR(255),
    category VARCHAR(255) NOT NULL,
    description TEXT,

    unique_selling_point TEXT,
    website_url VARCHAR(255),
    vision_mission TEXT,
    location VARCHAR(255),
    color_tone VARCHAR(6),

    business_root_id UUID NOT NULL UNIQUE,
    FOREIGN KEY (business_root_id) REFERENCES business_roots(id) ON DELETE CASCADE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- 1 to many with business_root
CREATE TABLE IF NOT EXISTS business_products ( 
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    name VARCHAR(255) NOT NULL,
    category VARCHAR(255) NOT NULL,
    description TEXT,

    currency CHAR(3) NOT NULL,
    price DECIMAL(10, 2) NOT NULL,

    image_urls VARCHAR(255)[] NOT NULL,

    business_root_id UUID NOT NULL,
    FOREIGN KEY (business_root_id) REFERENCES business_roots(id) ON DELETE CASCADE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- 1 to 1 with business_root
CREATE TABLE IF NOT EXISTS business_roles ( 
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    target_audience VARCHAR(255) NOT NULL,
    tone VARCHAR(255) NOT NULL,
    audience_persona VARCHAR(255) NOT NULL,
    hashtags VARCHAR(255)[] NOT NULL,
    call_to_action VARCHAR(255) NOT NULL,
    goals TEXT,

    business_root_id UUID NOT NULL UNIQUE,
    FOREIGN KEY (business_root_id) REFERENCES business_roots(id) ON DELETE CASCADE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- 1 to many with business_root
CREATE TABLE IF NOT EXISTS business_rss_subscriptions ( 
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    title VARCHAR(255) NOT NULL,
    is_active BOOLEAN NOT NULL,

    business_root_id UUID NOT NULL,
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

CREATE TYPE business_member_status AS ENUM ('pending', 'accepted', 'rejected', 'left', 'kicked');
CREATE TYPE business_member_role   AS ENUM ('owner', 'admin', 'member');

CREATE TABLE IF NOT EXISTS business_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    status business_member_status NOT NULL DEFAULT 'pending',
    role   business_member_role   NOT NULL DEFAULT 'member',

    answered_at TIMESTAMPTZ,

    business_root_id UUID NOT NULL,
    profile_id       UUID NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT business_members_root_profile_unique UNIQUE (business_root_id, profile_id),

    FOREIGN KEY (business_root_id) REFERENCES business_roots(id) ON DELETE CASCADE,
    FOREIGN KEY (profile_id)       REFERENCES profiles(id)       ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_business_members_business_root_id ON business_members(business_root_id);
CREATE INDEX IF NOT EXISTS idx_business_members_profile_id       ON business_members(profile_id);

CREATE TRIGGER trigger_business_members_updated_at
BEFORE UPDATE ON business_members
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

ALTER TABLE business_roots
ADD COLUMN deleted_at TIMESTAMPTZ;

ALTER TABLE business_members ADD COLUMN deleted_at TIMESTAMPTZ;
