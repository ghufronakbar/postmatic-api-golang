-- internal/repository/queries/business_saved_template_creator_image.sql

-- name: GetAllSavedCreatorImageByBusinessId :many
SELECT
    bstci.id,
    bstci.business_root_id,
    bstci.creator_image_id,
    bstci.created_at AS saved_at,
    ci.name,
    ci.image_url,
    ci.is_published,
    ci.is_banned,
    ci.banned_reason,
    ci.price,
    ci.profile_id AS publisher_id,
    p.name AS publisher_name,
    p.image_url AS publisher_image,
    ci.created_at,
    ci.updated_at,
    ci.deleted_at AS creator_image_deleted_at,
    -- type categories as JSON array
    COALESCE(
        (SELECT json_agg(json_build_object('id', tc.id, 'name', tc.name))
         FROM creator_image_type_categories citc
         JOIN app_creator_image_type_categories tc ON tc.id = citc.type_category_id
         WHERE citc.creator_image_id = ci.id),
        '[]'::json
    ) AS type_category_subs,
    -- product categories as JSON array
    COALESCE(
        (SELECT json_agg(json_build_object('id', pc.id, 'name', pc.indonesian_name))
         FROM creator_image_product_categories cipc
         JOIN app_creator_image_product_categories pc ON pc.id = cipc.product_category_id
         WHERE cipc.creator_image_id = ci.id),
        '[]'::json
    ) AS product_category_subs
FROM business_saved_template_creator_images bstci
JOIN creator_images ci ON ci.id = bstci.creator_image_id
LEFT JOIN profiles p ON p.id = ci.profile_id
WHERE bstci.business_root_id = @business_root_id
  AND bstci.deleted_at IS NULL
  -- search by name
  AND (sqlc.narg('search')::TEXT IS NULL OR ci.name ILIKE '%' || sqlc.narg('search')::TEXT || '%')
  -- filter by date range (on saved_at)
  AND (sqlc.narg('date_start')::TIMESTAMPTZ IS NULL OR bstci.created_at >= sqlc.narg('date_start')::TIMESTAMPTZ)
  AND (sqlc.narg('date_end')::TIMESTAMPTZ IS NULL OR bstci.created_at <= sqlc.narg('date_end')::TIMESTAMPTZ)
  -- filter by type category
  AND (sqlc.narg('type_category_id')::BIGINT IS NULL OR EXISTS (
      SELECT 1 FROM creator_image_type_categories citc
      WHERE citc.creator_image_id = ci.id AND citc.type_category_id = sqlc.narg('type_category_id')::BIGINT
  ))
  -- filter by product category
  AND (sqlc.narg('product_category_id')::BIGINT IS NULL OR EXISTS (
      SELECT 1 FROM creator_image_product_categories cipc
      WHERE cipc.creator_image_id = ci.id AND cipc.product_category_id = sqlc.narg('product_category_id')::BIGINT
  ))
ORDER BY
    CASE WHEN sqlc.narg('sort_by')::TEXT = 'id' AND sqlc.narg('sort_dir')::TEXT = 'asc' THEN bstci.id END ASC,
    CASE WHEN sqlc.narg('sort_by')::TEXT = 'id' AND sqlc.narg('sort_dir')::TEXT = 'desc' THEN bstci.id END DESC,
    CASE WHEN sqlc.narg('sort_by')::TEXT = 'created_at' AND sqlc.narg('sort_dir')::TEXT = 'asc' THEN bstci.created_at END ASC,
    CASE WHEN sqlc.narg('sort_by')::TEXT = 'created_at' AND sqlc.narg('sort_dir')::TEXT = 'desc' THEN bstci.created_at END DESC,
    CASE WHEN sqlc.narg('sort_by')::TEXT = 'updated_at' AND sqlc.narg('sort_dir')::TEXT = 'asc' THEN bstci.updated_at END ASC,
    CASE WHEN sqlc.narg('sort_by')::TEXT = 'updated_at' AND sqlc.narg('sort_dir')::TEXT = 'desc' THEN bstci.updated_at END DESC,
    CASE WHEN sqlc.narg('sort_by')::TEXT = 'name' AND sqlc.narg('sort_dir')::TEXT = 'asc' THEN ci.name END ASC,
    CASE WHEN sqlc.narg('sort_by')::TEXT = 'name' AND sqlc.narg('sort_dir')::TEXT = 'desc' THEN ci.name END DESC,
    bstci.id DESC -- default sort
LIMIT @page_limit OFFSET @page_offset;

-- name: CountSavedCreatorImageByBusinessId :one
SELECT COUNT(*)
FROM business_saved_template_creator_images bstci
JOIN creator_images ci ON ci.id = bstci.creator_image_id
WHERE bstci.business_root_id = @business_root_id
  AND bstci.deleted_at IS NULL
  AND (sqlc.narg('search')::TEXT IS NULL OR ci.name ILIKE '%' || sqlc.narg('search')::TEXT || '%')
  AND (sqlc.narg('date_start')::TIMESTAMPTZ IS NULL OR bstci.created_at >= sqlc.narg('date_start')::TIMESTAMPTZ)
  AND (sqlc.narg('date_end')::TIMESTAMPTZ IS NULL OR bstci.created_at <= sqlc.narg('date_end')::TIMESTAMPTZ)
  AND (sqlc.narg('type_category_id')::BIGINT IS NULL OR EXISTS (
      SELECT 1 FROM creator_image_type_categories citc
      WHERE citc.creator_image_id = ci.id AND citc.type_category_id = sqlc.narg('type_category_id')::BIGINT
  ))
  AND (sqlc.narg('product_category_id')::BIGINT IS NULL OR EXISTS (
      SELECT 1 FROM creator_image_product_categories cipc
      WHERE cipc.creator_image_id = ci.id AND cipc.product_category_id = sqlc.narg('product_category_id')::BIGINT
  ));

-- name: CheckSavedCreatorImageExists :one
SELECT EXISTS (
    SELECT 1 FROM business_saved_template_creator_images
    WHERE business_root_id = @business_root_id
      AND creator_image_id = @creator_image_id
      AND deleted_at IS NULL
) AS exists;

-- name: CreateSavedCreatorImage :one
INSERT INTO business_saved_template_creator_images (
    business_root_id,
    creator_image_id
) VALUES (
    @business_root_id,
    @creator_image_id
) RETURNING *;

-- name: GetSavedCreatorImageByBusinessAndCreatorImage :one
SELECT * FROM business_saved_template_creator_images
WHERE business_root_id = @business_root_id
  AND creator_image_id = @creator_image_id
  AND deleted_at IS NULL;

-- name: SoftDeleteSavedCreatorImage :exec
UPDATE business_saved_template_creator_images
SET deleted_at = NOW()
WHERE business_root_id = @business_root_id
  AND creator_image_id = @creator_image_id
  AND deleted_at IS NULL;
