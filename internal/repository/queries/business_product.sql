-- name: CreateBusinessProduct :one
INSERT INTO business_products (
    name,
    category,
    description,
    currency,
    price,
    image_urls,
    business_root_id
)
VALUES (
    sqlc.arg(name),
    sqlc.arg(category),
    sqlc.arg(description),
    sqlc.arg(currency),
    sqlc.arg(price),
    sqlc.arg(image_urls),
    sqlc.arg(business_root_id)
)
RETURNING *;

-- name: UpdateBusinessProduct :one
UPDATE business_products
SET
    name = sqlc.arg(name),
    category = sqlc.arg(category),
    description = sqlc.arg(description),
    currency = sqlc.arg(currency),
    price = sqlc.arg(price),
    image_urls = sqlc.arg(image_urls)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: SoftDeleteBusinessProductByBusinessRootID :one
UPDATE business_products
SET deleted_at = NOW()
WHERE business_root_id = sqlc.arg(business_root_id)
RETURNING id;


-- name: SoftDeleteBusinessProductByBusinessProductId :one
UPDATE business_products
SET deleted_at = NOW()
WHERE id = sqlc.arg(id)
RETURNING id;

-- name: GetBusinessProductsByBusinessRootId :many
WITH p AS (
  SELECT
    COALESCE(NULLIF(sqlc.narg(sort_by),  ''), 'created_at') AS sort_by,
    COALESCE(NULLIF(sqlc.narg(sort_dir), ''), 'desc')       AS sort_dir
)
SELECT bp.*
FROM business_products bp
CROSS JOIN p
WHERE
  bp.business_root_id = sqlc.arg(business_root_id)
  AND bp.deleted_at IS NULL

  -- search (name + description)
  AND (
    COALESCE(sqlc.narg(search), '') = ''
    OR bp.name ILIKE ('%' || sqlc.narg(search) || '%')
    OR COALESCE(bp.description, '') ILIKE ('%' || sqlc.narg(search) || '%')
  )

  -- category
  AND (
    sqlc.narg(category)::text IS NULL
    OR bp.category = sqlc.narg(category)
  )

  -- date range (berdasarkan created_at)
  AND (
    sqlc.narg(date_start)::date IS NULL
    OR bp.created_at::date >= sqlc.narg(date_start)::date
  )
  AND (
    sqlc.narg(date_end)::date IS NULL
    OR bp.created_at::date <= sqlc.narg(date_end)::date
  )

ORDER BY
  -- name
  CASE WHEN p.sort_by = 'name' AND p.sort_dir = 'asc'  THEN bp.name END ASC,
  CASE WHEN p.sort_by = 'name' AND p.sort_dir = 'desc' THEN bp.name END DESC,

  -- created_at (default)
  CASE WHEN p.sort_by = 'created_at' AND p.sort_dir = 'asc'  THEN bp.created_at END ASC,
  CASE WHEN p.sort_by = 'created_at' AND p.sort_dir = 'desc' THEN bp.created_at END DESC,

  -- updated_at
  CASE WHEN p.sort_by = 'updated_at' AND p.sort_dir = 'asc'  THEN bp.updated_at END ASC,
  CASE WHEN p.sort_by = 'updated_at' AND p.sort_dir = 'desc' THEN bp.updated_at END DESC,

  -- price
  CASE WHEN p.sort_by = 'price' AND p.sort_dir = 'asc'  THEN bp.price END ASC,
  CASE WHEN p.sort_by = 'price' AND p.sort_dir = 'desc' THEN bp.price END DESC,

  CASE WHEN p.sort_by = 'id' AND p.sort_dir = 'asc'  THEN bp.id END ASC,
  CASE WHEN p.sort_by = 'id' AND p.sort_dir = 'desc' THEN bp.id END DESC,

  -- fallback stable order
  bp.id DESC

LIMIT sqlc.arg(page_limit)
OFFSET sqlc.arg(page_offset);


-- name: CountBusinessProductsByBusinessRootId :one
SELECT COUNT(*)::bigint AS total
FROM business_products bp
WHERE
  bp.business_root_id = sqlc.arg(business_root_id)
  AND bp.deleted_at IS NULL

  -- search (name + description)
  AND (
    COALESCE(sqlc.narg(search), '') = ''
    OR bp.name ILIKE ('%' || sqlc.narg(search) || '%')
    OR COALESCE(bp.description, '') ILIKE ('%' || sqlc.narg(search) || '%')
  )

  -- category
  AND (
    sqlc.narg(category)::text IS NULL
    OR bp.category = sqlc.narg(category)
  )

  -- date range (berdasarkan created_at)
  AND (
    sqlc.narg(date_start)::date IS NULL
    OR bp.created_at::date >= sqlc.narg(date_start)::date
  )
  AND (
    sqlc.narg(date_end)::date IS NULL
    OR bp.created_at::date <= sqlc.narg(date_end)::date
  );

-- name: GetBusinessProductByBusinessProductId :one
SELECT * FROM business_products WHERE id = sqlc.arg(id) AND deleted_at IS NULL;
