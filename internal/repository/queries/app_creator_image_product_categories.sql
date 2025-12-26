-- name: GetAllAppCreatorImageProductCategories :many
WITH p0 AS (
  SELECT
    lower(COALESCE(NULLIF(sqlc.arg(sort_by),  ''), '')) AS sb_in,
    lower(COALESCE(NULLIF(sqlc.arg(sort_dir), ''), '')) AS sd_in,
    lower(COALESCE(NULLIF(sqlc.arg(locale),   ''), '')) AS loc_in
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
    END AS sort_dir,
    CASE
      WHEN loc_in IN ('id', 'en') THEN loc_in
      ELSE 'id'
    END AS locale
  FROM p0
),
q AS (
  SELECT
    c.id,
    c.indonesian_name,
    c.english_name,
    c.created_at,
    c.updated_at,
    COUNT(DISTINCT cipc.creator_image_id)::bigint AS total_data
  FROM app_creator_image_product_categories c
  LEFT JOIN creator_image_product_categories cipc
    ON cipc.product_category_id = c.id
  LEFT JOIN creator_images ci
    ON ci.id = cipc.creator_image_id
   AND ci.deleted_at IS NULL
  WHERE
    sqlc.arg(search) = ''
    OR c.indonesian_name ILIKE ('%' || sqlc.arg(search) || '%')
    OR c.english_name ILIKE ('%' || sqlc.arg(search) || '%')
  GROUP BY c.id, c.indonesian_name, c.english_name, c.created_at, c.updated_at
)
SELECT
  q.id,
  q.indonesian_name,
  q.english_name,
  q.total_data
FROM q
CROSS JOIN p
ORDER BY
  -- name (depends on locale)
  CASE
    WHEN p.sort_by = 'name' AND p.sort_dir = 'asc'
    THEN (CASE WHEN p.locale = 'en' THEN q.english_name ELSE q.indonesian_name END)
  END ASC,
  CASE
    WHEN p.sort_by = 'name' AND p.sort_dir = 'desc'
    THEN (CASE WHEN p.locale = 'en' THEN q.english_name ELSE q.indonesian_name END)
  END DESC,

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


-- name: CountAllAppCreatorImageProductCategories :one
SELECT COUNT(*)::bigint AS total
FROM app_creator_image_product_categories c
WHERE
  sqlc.arg(search) = ''
  OR c.indonesian_name ILIKE ('%' || sqlc.arg(search) || '%')
  OR c.english_name ILIKE ('%' || sqlc.arg(search) || '%');
