-- name: GetAllAppCreatorImageTypeCategories :many
WITH p0 AS (
  SELECT
    lower(COALESCE(NULLIF(sqlc.arg(sort_by),  ''), '')) AS sb_in,
    lower(COALESCE(NULLIF(sqlc.arg(sort_dir), ''), '')) AS sd_in
),
p AS (
  SELECT
    CASE
      WHEN sb_in IN ('name', 'created_at', 'updated_at') THEN sb_in
      ELSE 'created_at'
    END AS sort_by,
    CASE
      WHEN sd_in IN ('asc', 'desc') THEN sd_in
      ELSE 'desc'
    END AS sort_dir
  FROM p0
),
q AS (
  SELECT
    t.id,
    t.name,
    t.created_at,
    t.updated_at,
    COUNT(DISTINCT citc.creator_image_id)::bigint AS total_data
  FROM app_creator_image_type_categories t
  LEFT JOIN creator_image_type_categories citc
    ON citc.type_category_id = t.id
  LEFT JOIN creator_images ci
    ON ci.id = citc.creator_image_id
   AND ci.deleted_at IS NULL
  WHERE
    sqlc.arg(search) = ''
    OR t.name ILIKE ('%' || sqlc.arg(search) || '%')
  GROUP BY t.id, t.name, t.created_at, t.updated_at
)
SELECT
  q.id,
  q.name,
  q.total_data
FROM q
CROSS JOIN p
ORDER BY
  -- name
  CASE WHEN p.sort_by = 'name' AND p.sort_dir = 'asc'  THEN q.name END ASC,
  CASE WHEN p.sort_by = 'name' AND p.sort_dir = 'desc' THEN q.name END DESC,

  -- created_at (default)
  CASE WHEN p.sort_by = 'created_at' AND p.sort_dir = 'asc'  THEN q.created_at END ASC,
  CASE WHEN p.sort_by = 'created_at' AND p.sort_dir = 'desc' THEN q.created_at END DESC,

  -- updated_at
  CASE WHEN p.sort_by = 'updated_at' AND p.sort_dir = 'asc'  THEN q.updated_at END ASC,
  CASE WHEN p.sort_by = 'updated_at' AND p.sort_dir = 'desc' THEN q.updated_at END DESC,

  -- fallback stable order
  q.created_at DESC,
  q.id DESC
LIMIT sqlc.arg(page_limit)
OFFSET sqlc.arg(page_offset);


-- name: CountAllAppCreatorImageTypeCategories :one
SELECT COUNT(*)::bigint AS total
FROM app_creator_image_type_categories t
WHERE
  sqlc.arg(search) = ''
  OR t.name ILIKE ('%' || sqlc.arg(search) || '%');

-- name: GetAppCreatorImageTypeCategoriesByIds :many
SELECT
  t.id
FROM app_creator_image_type_categories t
WHERE t.id = ANY(sqlc.arg(ids)::bigint[]);
