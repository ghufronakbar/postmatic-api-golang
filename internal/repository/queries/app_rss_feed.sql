-- name: GetAllRSSFeed :many
SELECT
  f.*
FROM app_rss_feeds f
WHERE
  f.deleted_at IS NULL
  AND (
    COALESCE(sqlc.narg(search), '') = ''
    OR f.title ILIKE ('%' || sqlc.narg(search) || '%')
    OR f.publisher ILIKE ('%' || sqlc.narg(search) || '%')
    OR f.url ILIKE ('%' || sqlc.narg(search) || '%')
  )
  AND (
    sqlc.narg(category)::uuid IS NULL
    OR f.app_rss_category_id = sqlc.narg(category)::uuid
  )
ORDER BY
  -- title
  CASE WHEN sqlc.arg(sort_by) = 'title' AND sqlc.arg(sort_dir) = 'asc'  THEN f.title END ASC,
  CASE WHEN sqlc.arg(sort_by) = 'title' AND sqlc.arg(sort_dir) = 'desc' THEN f.title END DESC,

  -- created_at
  CASE WHEN sqlc.arg(sort_by) = 'created_at' AND sqlc.arg(sort_dir) = 'asc'  THEN f.created_at END ASC,
  CASE WHEN sqlc.arg(sort_by) = 'created_at' AND sqlc.arg(sort_dir) = 'desc' THEN f.created_at END DESC,

  -- updated_at
  CASE WHEN sqlc.arg(sort_by) = 'updated_at' AND sqlc.arg(sort_dir) = 'asc'  THEN f.updated_at END ASC,
  CASE WHEN sqlc.arg(sort_by) = 'updated_at' AND sqlc.arg(sort_dir) = 'desc' THEN f.updated_at END DESC,

  -- fallback stable order
  f.created_at DESC,
  f.id DESC
LIMIT sqlc.arg(page_limit)
OFFSET sqlc.arg(page_offset);


-- name: CountAllRSSFeed :one
SELECT COUNT(*)::bigint AS total
FROM app_rss_feeds f
WHERE
  f.deleted_at IS NULL
  AND (
    COALESCE(sqlc.narg(search), '') = ''
    OR f.title ILIKE ('%' || sqlc.narg(search) || '%')
    OR f.publisher ILIKE ('%' || sqlc.narg(search) || '%')
    OR f.url ILIKE ('%' || sqlc.narg(search) || '%')
  )
  AND (
    sqlc.narg(category)::uuid IS NULL
    OR f.app_rss_category_id = sqlc.narg(category)::uuid
  );

-- name: GetRssFeedById :one
SELECT * FROM app_rss_feeds
WHERE id = $1 AND deleted_at IS NULL;