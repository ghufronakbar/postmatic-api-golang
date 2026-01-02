-- name: GetProfileReferralCodeByProfileIdBasic :one
SELECT *
FROM profile_referral_codes
WHERE profile_id = $1
  AND type = 'basic'
  AND deleted_at IS NULL;


-- name: CreateProfileReferralCode :one
INSERT INTO profile_referral_codes (profile_id, code, type, is_active, total_discount, discount_type, expired_days, max_discount, max_usage, reward_per_referral)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;


-- name: GetProfileReferralCodeByCode :one
SELECT *
FROM profile_referral_codes
WHERE code = $1
  AND deleted_at IS NULL;

