-- +goose Up
-- +goose StatementBegin
-- record referral usage
CREATE TYPE referral_record_status AS ENUM ('pending', 'success', 'failed', 'canceled');

CREATE TABLE IF NOT EXISTS referral_records (
    id BIGSERIAL PRIMARY KEY,

    -- user that consume the referral code
    consumer_profile_id UUID NOT NULL,
    FOREIGN KEY (consumer_profile_id) REFERENCES profiles (id),

    -- business root that consume the referral code
    business_root_id BIGINT NOT NULL,
    FOREIGN KEY (business_root_id) REFERENCES business_roots (id),

    -- relation into referral code that used
    profile_referral_code_id BIGINT NOT NULL,
    FOREIGN KEY (profile_referral_code_id) REFERENCES profile_referral_codes (id),

    -- DENORMALIZED FOR RECORD
    -- ex: basic or special
    record_type referral_type NOT NULL,
    -- ex: Rp 10.000 or 10%
    record_total_discount BIGINT NOT NULL DEFAULT 0,
    record_discount_type discount_type NOT NULL DEFAULT 'fixed',
    -- if null there is no expiration
    record_expired_days INT,
    -- max discount for consumer
    record_max_discount BIGINT NOT NULL DEFAULT 0,
    -- max usage for consumer, if null there is no limit
    record_max_usage INT,
    -- ex: Rp 10.000
    record_reward_per_referral BIGINT NOT NULL DEFAULT 0,

    -- CONSUMER (hal yang didapat oleh consumer referral)
    -- discount amount granted to consumer (harus dalam satuan uang)
    discount_amount_granted BIGINT NOT NULL DEFAULT 0,
    -- currency untuk discount amount (saat ini hanya IDR)
    discount_currency VARCHAR(3) NOT NULL DEFAULT 'IDR',

    -- PRODUCER (hal yang didapat oleh producer referral)
    -- reward amount granted to producer (harus dalam satuan uang)
    reward_amount_granted BIGINT NOT NULL DEFAULT 0,
    -- currency untuk reward amount (saat ini hanya IDR)
    reward_currency VARCHAR(3) NOT NULL DEFAULT 'IDR',

    -- status for tracking usage (pending -> success/failed/canceled)
    status referral_record_status NOT NULL DEFAULT 'pending',
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);
CREATE TRIGGER trigger_referral_records_updated_at
BEFORE UPDATE ON referral_records
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trigger_referral_records_updated_at ON referral_records;
DROP TABLE IF EXISTS referral_records;
DROP TYPE IF EXISTS referral_record_status;
-- +goose StatementEnd
