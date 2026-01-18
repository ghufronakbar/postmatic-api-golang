-- name: CreateReferralRecord :one
INSERT INTO referral_records (
    consumer_profile_id,
    business_root_id,
    profile_referral_code_id,
    record_type,
    record_total_discount,
    record_discount_type,
    record_expired_days,
    record_max_discount,
    record_max_usage,
    record_reward_per_referral,
    discount_amount_granted,
    discount_currency,
    reward_amount_granted,
    reward_currency,
    status
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
) RETURNING *;

-- name: GetReferralRecordById :one
SELECT * FROM referral_records
WHERE id = $1 AND deleted_at IS NULL;

-- name: UpdateReferralRecordStatus :one
UPDATE referral_records
SET status = $2
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: CheckProfileUsedReferralCode :one
SELECT EXISTS(
    SELECT 1 FROM referral_records
    WHERE consumer_profile_id = $1 
    AND profile_referral_code_id = $2 
    AND status IN ('pending', 'success')
    AND deleted_at IS NULL
) AS used;

-- name: CheckBusinessUsedReferralCode :one
SELECT EXISTS(
    SELECT 1 FROM referral_records
    WHERE business_root_id = $1 
    AND profile_referral_code_id = $2 
    AND status IN ('pending', 'success')
    AND deleted_at IS NULL
) AS used;

-- name: CountReferralCodeUsage :one
SELECT COUNT(*)::int AS count FROM referral_records
WHERE profile_referral_code_id = $1 
AND status IN ('pending', 'success')
AND deleted_at IS NULL;
