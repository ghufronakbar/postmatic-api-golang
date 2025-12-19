-- name: CreateBusinessRole :one
INSERT INTO business_roles (
    target_audience,
    tone,
    audience_persona,
    hashtags,
    call_to_action,
    goals,
    business_root_id
)
VALUES (
    sqlc.arg(target_audience),
    sqlc.arg(tone),
    sqlc.arg(audience_persona),
    sqlc.arg(hashtags),
    sqlc.arg(call_to_action),
    sqlc.arg(goals),
    sqlc.arg(business_root_id)
)
RETURNING *;

-- name: SoftDeleteBusinessRoleByBusinessRootID :one
UPDATE business_roles
SET deleted_at = NOW()
WHERE business_root_id = sqlc.arg(business_root_id)
RETURNING id;