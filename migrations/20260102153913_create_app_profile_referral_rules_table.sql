-- +goose Up
-- +goose StatementBegin

-- ENUM (idempotent)
CREATE TYPE discount_type AS ENUM ('fixed', 'percentage');


-- EXPECTED ONLY 1 ROW (APP RULES)
CREATE TABLE IF NOT EXISTS app_profile_referral_rules (
  id smallint PRIMARY KEY DEFAULT 1 CHECK (id = 1),

  -- CONSUMER
  -- ex: Rp 10.000 or 10%
  total_discount BIGINT NOT NULL DEFAULT 0,
  discount_type discount_type NOT NULL DEFAULT 'fixed',
  -- if null there is no expiration
  expired_days INT,
  -- max discount for consumer
  max_discount BIGINT NOT NULL DEFAULT 0,
  -- max usage for consumer, if null there is no limit
  max_usage INT,

  -- PRODUCER
  -- ex: Rp 10.000
  reward_per_referral BIGINT NOT NULL DEFAULT 0,

  created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

  -- optional sanity checks
  CONSTRAINT app_profile_referral_rules_total_discount_nonneg CHECK (total_discount >= 0),
  CONSTRAINT app_profile_referral_rules_max_discount_nonneg  CHECK (max_discount >= 0),
  CONSTRAINT app_profile_referral_rules_reward_nonneg        CHECK (reward_per_referral >= 0),
  CONSTRAINT app_profile_referral_rules_expired_days_pos     CHECK (expired_days IS NULL OR expired_days > 0),
  CONSTRAINT app_profile_referral_rules_max_usage_pos        CHECK (max_usage IS NULL OR max_usage > 0)
);

CREATE TRIGGER trigger_app_profile_referral_rules_updated_at
BEFORE UPDATE ON app_profile_referral_rules
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

-- seed singleton row
INSERT INTO app_profile_referral_rules (id)
VALUES (1)
ON CONFLICT (id) DO NOTHING;

-- AUDIT / CHANGES (biasanya INSERT-only)
CREATE TABLE IF NOT EXISTS app_profile_referral_changes (
  id BIGSERIAL PRIMARY KEY,

  -- who made the change
  profile_id UUID NOT NULL,
  FOREIGN KEY (profile_id) REFERENCES profiles(id) ON DELETE RESTRICT,

  -- snapshot of rules
  total_discount BIGINT NOT NULL,
  discount_type discount_type NOT NULL,
  expired_days INT,
  max_discount BIGINT NOT NULL,
  max_usage INT,
  reward_per_referral BIGINT NOT NULL,

  created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER trigger_app_profile_referral_changes_updated_at
BEFORE UPDATE ON app_profile_referral_changes
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

-- trigger: when changes inserted/updated, touch rules.updated_at
CREATE OR REPLACE FUNCTION touch_app_profile_referral_rules_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  UPDATE app_profile_referral_rules
  SET updated_at = now()
  WHERE id = 1;

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_touch_app_profile_referral_rules_updated_at
AFTER INSERT OR UPDATE ON app_profile_referral_changes
FOR EACH ROW
EXECUTE FUNCTION touch_app_profile_referral_rules_updated_at();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TRIGGER IF EXISTS trigger_touch_app_profile_referral_rules_updated_at ON app_profile_referral_changes;
DROP FUNCTION IF EXISTS touch_app_profile_referral_rules_updated_at;

DROP TRIGGER IF EXISTS trigger_app_profile_referral_changes_updated_at ON app_profile_referral_changes;
DROP TABLE IF EXISTS app_profile_referral_changes;

DROP TRIGGER IF EXISTS trigger_app_profile_referral_rules_updated_at ON app_profile_referral_rules;
DROP TABLE IF EXISTS app_profile_referral_rules;

DROP TYPE IF EXISTS discount_type;

-- +goose StatementEnd
