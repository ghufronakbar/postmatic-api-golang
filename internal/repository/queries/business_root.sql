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
  AND br.deleted_at IS NULL
  AND bm.status = 'accepted'
  AND bk.deleted_at IS NULL
  AND (
    COALESCE(sqlc.narg(search), '') = ''
    OR bk.name ILIKE ('%' || sqlc.narg(search) || '%')
  )
  AND (
    sqlc.narg(date_start)::date IS NULL
    OR COALESCE(bm.answered_at)::date >= sqlc.narg(date_start)::date
  )
  AND (
    sqlc.narg(date_end)::date IS NULL
    OR COALESCE(bm.answered_at)::date <= sqlc.narg(date_end)::date
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

  -- id
  CASE WHEN sqlc.arg(sort_by) = 'id' AND sqlc.arg(sort_dir) = 'asc'  THEN bm.id END ASC,
  CASE WHEN sqlc.arg(sort_by) = 'id' AND sqlc.arg(sort_dir) = 'desc' THEN bm.id END DESC,

  -- fallback stable order
  bm.id DESC

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
  )
  AND (
    sqlc.narg(date_start)::date IS NULL
    OR COALESCE(bm.answered_at)::date >= sqlc.narg(date_start)::date
  )
  AND (
    sqlc.narg(date_end)::date IS NULL
    OR COALESCE(bm.answered_at)::date <= sqlc.narg(date_end)::date
  );

-- name: CreateBusinessRoot :one
INSERT INTO business_roots 
DEFAULT VALUES
RETURNING id;

-- name: SoftDeleteBusinessRoot :one
UPDATE business_roots
SET deleted_at = NOW()
WHERE id = sqlc.arg(id)
RETURNING id;

-- name: GetBusinessRootById :one
SELECT id, deleted_at
FROM business_roots
WHERE id = sqlc.arg(id);
