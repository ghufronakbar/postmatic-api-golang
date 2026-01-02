-- name: GetAppProfileReferralRules :one
SELECT
  id,
  total_discount,
  discount_type,
  expired_days,
  max_discount,
  max_usage,
  reward_per_referral,
  created_at,
  updated_at
FROM app_profile_referral_rules
WHERE id = 1;

-- name: UpsertAppProfileReferralRules :one
INSERT INTO app_profile_referral_rules (
  id,
  total_discount,
  discount_type,
  expired_days,
  max_discount,
  max_usage,
  reward_per_referral
)
VALUES (
  1,
  sqlc.arg(total_discount),
  sqlc.arg(discount_type),
  sqlc.arg(expired_days),
  sqlc.arg(max_discount),
  sqlc.arg(max_usage),
  sqlc.arg(reward_per_referral)
)
ON CONFLICT (id) DO UPDATE
SET
  total_discount       = EXCLUDED.total_discount,
  discount_type        = EXCLUDED.discount_type,
  expired_days         = EXCLUDED.expired_days,
  max_discount         = EXCLUDED.max_discount,
  max_usage            = EXCLUDED.max_usage,
  reward_per_referral  = EXCLUDED.reward_per_referral

RETURNING
  id,
  total_discount,
  discount_type,
  expired_days,
  max_discount,
  max_usage,
  reward_per_referral,
  created_at,
  updated_at;
