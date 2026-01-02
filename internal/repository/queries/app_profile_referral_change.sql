-- name: InsertAppProfileReferralChange :one
INSERT INTO app_profile_referral_changes (
  profile_id,
  total_discount,
  discount_type,
  expired_days,
  max_discount,
  max_usage,
  reward_per_referral
)
VALUES (
  sqlc.arg(profile_id),
  sqlc.arg(total_discount),
  sqlc.arg(discount_type),
  sqlc.arg(expired_days),
  sqlc.arg(max_discount),
  sqlc.arg(max_usage),
  sqlc.arg(reward_per_referral)
)
RETURNING *;
