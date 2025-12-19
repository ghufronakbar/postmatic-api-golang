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
