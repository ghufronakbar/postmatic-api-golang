-- +goose Up
-- +goose StatementBegin
-- every user has 1 basic referral code (created by system)
-- special referral code can be created by admin
CREATE TYPE referral_type AS ENUM ('basic', 'special');
CREATE TABLE IF NOT EXISTS profile_referral_codes (
    id BIGSERIAL PRIMARY KEY,
    -- user that has created the referral code
    profile_id UUID NOT NULL,
    FOREIGN KEY (profile_id) REFERENCES profiles (id),
    -- unique code to consumer use
    code VARCHAR(255) NOT NULL UNIQUE,
    type referral_type NOT NULL,
    is_active BOOLEAN NOT NULL,

    -- DENORMALIZED FOR RECORD
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
    deleted_at TIMESTAMPTZ
);
CREATE TRIGGER trigger_profile_referral_codes_updated_at
BEFORE UPDATE ON profile_referral_codes
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE UNIQUE INDEX IF NOT EXISTS uq_profile_one_basic
ON profile_referral_codes (profile_id)
WHERE type = 'basic' AND deleted_at IS NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS uq_profile_one_basic;
DROP TRIGGER IF EXISTS trigger_profile_referral_codes_updated_at ON profile_referral_codes;
DROP TABLE IF EXISTS profile_referral_codes;
DROP TYPE IF EXISTS referral_type;
-- +goose StatementEnd
