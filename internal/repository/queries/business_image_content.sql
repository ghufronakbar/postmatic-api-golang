-- name: CreateBusinessImageContent :one
INSERT INTO business_image_contents (
    caption,
    type,
    ready_to_post,
    category,
    image_urls,
    business_root_id,
    business_product_id
)
VALUES (
    sqlc.arg(caption),
    sqlc.arg(type),
    sqlc.arg(ready_to_post),
    sqlc.arg(category),
    sqlc.arg(image_urls),
    sqlc.arg(business_root_id),
    sqlc.arg(business_product_id)
)
RETURNING *;

-- name: UpdateBusinessImageContent :one
UPDATE business_image_contents
SET
    caption = sqlc.arg(caption),
    type = sqlc.arg(type),
    ready_to_post = sqlc.arg(ready_to_post),
    category = sqlc.arg(category),
    image_urls = sqlc.arg(image_urls)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: SoftDeleteBusinessImageContentByBusinessImageContentId :one
UPDATE business_image_contents
SET deleted_at = NOW()
WHERE id = sqlc.arg(id) AND deleted_at IS NULL
RETURNING *;

-- name: GetBusinessImageContentsByBusinessRootId :many
WITH p AS (
  SELECT
    COALESCE(NULLIF(sqlc.narg(sort_by),  ''), 'created_at') AS sort_by,
    COALESCE(NULLIF(sqlc.narg(sort_dir), ''), 'desc')       AS sort_dir
)
SELECT b.*
FROM business_image_contents b
CROSS JOIN p
WHERE
  b.business_root_id = sqlc.arg(business_root_id)
  AND b.deleted_at IS NULL

  -- search (name + description)
  AND (
    COALESCE(sqlc.narg(search), '') = ''
    OR b.caption ILIKE ('%' || sqlc.narg(search) || '%')
    OR COALESCE(b.category, '') ILIKE ('%' || sqlc.narg(search) || '%')
  )

  -- ready_to_post
  AND (
    sqlc.narg(ready_to_post)::boolean IS NULL
    OR b.ready_to_post = sqlc.narg(ready_to_post)
  )

  -- date range (berdasarkan created_at)
  AND (
    sqlc.narg(date_start)::date IS NULL
    OR b.created_at::date >= sqlc.narg(date_start)::date
  )
  AND (
    sqlc.narg(date_end)::date IS NULL
    OR b.created_at::date <= sqlc.narg(date_end)::date
  )

ORDER BY
  -- caption
  CASE WHEN p.sort_by = 'caption' AND p.sort_dir = 'asc'  THEN b.caption END ASC,
  CASE WHEN p.sort_by = 'caption' AND p.sort_dir = 'desc' THEN b.caption END DESC,

  -- created_at
  CASE WHEN p.sort_by = 'created_at' AND p.sort_dir = 'asc'  THEN b.created_at END ASC,
  CASE WHEN p.sort_by = 'created_at' AND p.sort_dir = 'desc' THEN b.created_at END DESC,

  -- updated_at
  CASE WHEN p.sort_by = 'updated_at' AND p.sort_dir = 'asc'  THEN b.updated_at END ASC,
  CASE WHEN p.sort_by = 'updated_at' AND p.sort_dir = 'desc' THEN b.updated_at END DESC,

  -- id
  CASE WHEN p.sort_by = 'id' AND p.sort_dir = 'asc'  THEN b.id END ASC,
  CASE WHEN p.sort_by = 'id' AND p.sort_dir = 'desc' THEN b.id END DESC,

  -- fallback stable order
  b.id DESC

LIMIT sqlc.arg(page_limit)
OFFSET sqlc.arg(page_offset);


-- name: CountBusinessImageContentsByBusinessRootId :one
SELECT COUNT(*)::bigint AS total
FROM business_image_contents b
WHERE
  b.business_root_id = sqlc.arg(business_root_id)
  AND b.deleted_at IS NULL

  -- search (name + description)
  AND (
    COALESCE(sqlc.narg(search), '') = ''
    OR b.caption ILIKE ('%' || sqlc.narg(search) || '%')
    OR COALESCE(b.category, '') ILIKE ('%' || sqlc.narg(search) || '%')
  )

  -- ready_to_post
  AND (
    sqlc.narg(ready_to_post)::boolean IS NULL
    OR b.ready_to_post = sqlc.narg(ready_to_post)
  )

  -- date range (berdasarkan created_at)
  AND (
    sqlc.narg(date_start)::date IS NULL
    OR b.created_at::date >= sqlc.narg(date_start)::date
  )
  AND (
    sqlc.narg(date_end)::date IS NULL
    OR b.created_at::date <= sqlc.narg(date_end)::date
  );
