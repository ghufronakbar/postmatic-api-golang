-- name: GetAllRSSCategory :many
SELECT
  c.*
FROM app_rss_categories c
WHERE
  c.deleted_at IS NULL
  AND (
    COALESCE(sqlc.narg(search), '') = ''
    OR c.name ILIKE ('%' || sqlc.narg(search) || '%')
  )
ORDER BY
  -- name
  CASE WHEN sqlc.arg(sort_by) = 'name' AND sqlc.arg(sort_dir) = 'asc'  THEN c.name END ASC,
  CASE WHEN sqlc.arg(sort_by) = 'name' AND sqlc.arg(sort_dir) = 'desc' THEN c.name END DESC,

  -- created_at
  CASE WHEN sqlc.arg(sort_by) = 'created_at' AND sqlc.arg(sort_dir) = 'asc'  THEN c.created_at END ASC,
  CASE WHEN sqlc.arg(sort_by) = 'created_at' AND sqlc.arg(sort_dir) = 'desc' THEN c.created_at END DESC,

  -- updated_at
  CASE WHEN sqlc.arg(sort_by) = 'updated_at' AND sqlc.arg(sort_dir) = 'asc'  THEN c.updated_at END ASC,
  CASE WHEN sqlc.arg(sort_by) = 'updated_at' AND sqlc.arg(sort_dir) = 'desc' THEN c.updated_at END DESC,

  -- fallback stable order
  c.created_at DESC,
  c.id DESC
LIMIT sqlc.arg(page_limit)
OFFSET sqlc.arg(page_offset);


-- name: CountAllRSSCategory :one
SELECT COUNT(*)::bigint AS total
FROM app_rss_categories c
WHERE
  c.deleted_at IS NULL
  AND (
    COALESCE(sqlc.narg(search), '') = ''
    OR c.name ILIKE ('%' || sqlc.narg(search) || '%')
  );
