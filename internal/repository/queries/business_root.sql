-- name: GetJoinedBusinessesByProfileID :many
SELECT
  bm.id                AS member_id,
  bm.status            AS member_status,
  bm.role              AS member_role,
  bm.answered_at       AS member_answered_at,

  br.id                AS business_root_id,
  br.created_at        AS business_root_created_at,
  br.updated_at        AS business_root_updated_at,

  bk.name              AS business_name,
  bk.description       AS business_description,
  bk.primary_logo_url  AS business_logo_url

FROM business_members bm
JOIN business_roots br
  ON br.id = bm.business_root_id
JOIN business_knowledges bk
  ON bk.business_root_id = br.id

WHERE
  bm.profile_id = sqlc.arg(profile_id)
  AND bm.status = 'accepted'
  AND bk.deleted_at IS NULL
  AND (
    COALESCE(sqlc.narg(search), '') = ''
    OR bk.name ILIKE ('%' || sqlc.narg(search) || '%')
  )

ORDER BY
  -- name
  CASE WHEN sqlc.arg(sort_by) = 'name' AND sqlc.arg(sort_dir) = 'asc'  THEN bk.name END ASC,
  CASE WHEN sqlc.arg(sort_by) = 'name' AND sqlc.arg(sort_dir) = 'desc' THEN bk.name END DESC,

  -- created_at
  CASE WHEN sqlc.arg(sort_by) = 'created_at' AND sqlc.arg(sort_dir) = 'asc'  THEN br.created_at END ASC,
  CASE WHEN sqlc.arg(sort_by) = 'created_at' AND sqlc.arg(sort_dir) = 'desc' THEN br.created_at END DESC,

  -- updated_at
  CASE WHEN sqlc.arg(sort_by) = 'updated_at' AND sqlc.arg(sort_dir) = 'asc'  THEN br.updated_at END ASC,
  CASE WHEN sqlc.arg(sort_by) = 'updated_at' AND sqlc.arg(sort_dir) = 'desc' THEN br.updated_at END DESC,

  -- fallback stable order
  br.created_at DESC,
  br.id DESC

LIMIT sqlc.arg(page_limit)
OFFSET sqlc.arg(page_offset);


-- name: CountJoinedBusinessesByProfileID :one
SELECT COUNT(*)::bigint AS total
FROM business_members bm
JOIN business_roots br
  ON br.id = bm.business_root_id
JOIN business_knowledges bk
  ON bk.business_root_id = br.id
WHERE
  bm.profile_id = sqlc.arg(profile_id)
  AND bm.status = 'accepted'
  AND bk.deleted_at IS NULL
  AND (
    COALESCE(sqlc.narg(search), '') = ''
    OR bk.name ILIKE ('%' || sqlc.narg(search) || '%')
  );


-- name: GetBusinessMembersByBusinessRootIDs :many
SELECT
  bm.business_root_id,
  bm.status,
  bm.role,

  p.id        AS profile_id,
  p.name      AS profile_name,
  p.image_url AS profile_image_url,
  p.email     AS profile_email

FROM business_members bm
JOIN profiles p ON p.id = bm.profile_id
WHERE
  bm.business_root_id = ANY(sqlc.arg(business_root_ids)::uuid[])
ORDER BY bm.business_root_id, bm.created_at ASC;
