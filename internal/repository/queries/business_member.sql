-- name: GetMembersByBusinessRootIDs :many
SELECT
  bm.business_root_id,
  bm.status,
  bm.role,

  p.id        AS profile_id,
  p.name      AS profile_name,
  p.image_url AS profile_image_url,
  p.email     AS profile_email

FROM business_members bm
JOIN profiles p
  ON p.id = bm.profile_id

WHERE bm.business_root_id = ANY(sqlc.arg(business_root_ids)::bigint[])

ORDER BY bm.business_root_id, bm.created_at ASC;

-- name: GetMembersByBusinessRootID :many
SELECT
  bm.business_root_id,
  bm.id,
  bm.status,
  bm.role,
  bm.answered_at,
  bm.created_at,
  bm.updated_at,

  p.id        AS profile_id,
  p.name      AS profile_name,
  p.image_url AS profile_image_url,
  p.email     AS profile_email

FROM business_members bm
JOIN profiles p
  ON p.id = bm.profile_id

WHERE bm.business_root_id = sqlc.arg(business_root_id)

ORDER BY bm.business_root_id, bm.created_at ASC;

-- name: CreateBusinessMember :one
INSERT INTO business_members (
	status,
	role,
	answered_at,
	business_root_id,
	profile_id
)
VALUES (
	$1,
	$2,
	$3,
	$4,
	$5
)
RETURNING *;

-- name: SoftDeleteBusinessMemberByBusinessRootID :one
UPDATE business_members
SET deleted_at = NOW()
WHERE business_root_id = sqlc.arg(business_root_id)
RETURNING id;

-- name: GetMemberByProfileIdAndBusinessRootId :one
SELECT * FROM business_members
WHERE profile_id = sqlc.arg(profile_id)
AND business_root_id = sqlc.arg(business_root_id)
AND deleted_at IS NULL LIMIT 1;


-- name: GetMemberByEmailAndBusinessRootId :one
SELECT * FROM business_members bm
JOIN profiles p
  ON p.id = bm.profile_id
WHERE p.email = sqlc.arg(email)
AND bm.business_root_id = sqlc.arg(business_root_id)
AND bm.deleted_at IS NULL LIMIT 1;


-- name: GetMembersByBusinessRootIDWithStatus :many
SELECT
  bm.business_root_id,
  bm.id,
  bm.status,
  bm.role,
  bm.answered_at,
  bm.created_at,
  bm.updated_at,

  p.id        AS profile_id,
  p.name      AS profile_name,
  p.image_url AS profile_image_url,
  p.email     AS profile_email,

  (bm.profile_id = sqlc.arg(profile_id)) AS is_yourself
FROM business_members AS bm
JOIN profiles AS p
  ON p.id = bm.profile_id
WHERE
  bm.business_root_id = sqlc.arg(business_root_id)
  AND bm.deleted_at IS NULL
  AND (
    CAST(sqlc.narg(is_verified) AS boolean) IS NULL
    OR (bm.status = 'accepted'::business_member_status) = CAST(sqlc.narg(is_verified) AS boolean)
  )
ORDER BY bm.business_root_id, bm.created_at ASC;

-- name: UpdateManyBusinessMemberStatus :exec
UPDATE business_members
SET status = sqlc.arg(status)
WHERE id = ANY(sqlc.arg(ids)::bigint[]);

-- name: UpdateBusinessMemberStatus :one
UPDATE business_members
SET status = sqlc.arg(status)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: UpdateBusinessMemberRole :one
UPDATE business_members
SET role = sqlc.arg(role)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: SetBusinessMemberAnsweredAt :one
UPDATE business_members
SET answered_at = NOW()
WHERE id = sqlc.arg(id)
RETURNING *;