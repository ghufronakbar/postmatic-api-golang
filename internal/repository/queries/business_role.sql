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


-- name: GetBusinessRoleByBusinessRootID :one
SELECT br.*
FROM business_roles br
WHERE br.business_root_id = $1
  AND br.deleted_at IS NULL;


-- name: UpsertBusinessRoleByBusinessRootID :one
INSERT INTO business_roles (
  target_audience,
  tone,
  audience_persona,
  hashtags,
  call_to_action,
  goals,
  business_root_id
) VALUES (
  sqlc.arg(target_audience),
  sqlc.arg(tone),
  sqlc.arg(audience_persona),
  sqlc.arg(hashtags),
  sqlc.arg(call_to_action),
  sqlc.narg(goals),
  sqlc.arg(business_root_id)
)
ON CONFLICT (business_root_id) DO UPDATE SET
  target_audience  = EXCLUDED.target_audience,
  tone             = EXCLUDED.tone,
  audience_persona = EXCLUDED.audience_persona,
  hashtags         = EXCLUDED.hashtags,
  call_to_action   = EXCLUDED.call_to_action,
  goals            = EXCLUDED.goals,
  deleted_at       = NULL
RETURNING *;