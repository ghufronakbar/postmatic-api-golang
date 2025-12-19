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

WHERE bm.business_root_id = ANY(sqlc.arg(business_root_ids)::uuid[])

ORDER BY bm.business_root_id, bm.created_at ASC;

-- name: GetMembersByBusinessRootID :many
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
AND business_root_id = sqlc.arg(business_root_id);
