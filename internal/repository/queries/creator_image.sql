-- name: GetAllPublishedCreatorImage :many
SELECT * FROM creator_images WHERE deleted_at IS NULL
AND is_banned = false AND is_published = true;

-- name: CountAllPublishedCreatorImage :one
SELECT COUNT(*)::bigint AS total FROM creator_images WHERE deleted_at IS NULL
AND is_banned = false AND is_published = true;


-- name: CreateCreatorImage :one
WITH ins AS (
  INSERT INTO creator_images (
    name,
    image_url,
    is_published,
    price,
    profile_id
  )
  VALUES (
    sqlc.arg(name),
    sqlc.arg(image_url),
    sqlc.arg(is_published),
    sqlc.arg(price),
    sqlc.narg(profile_id)
  )
  RETURNING *
),
ins_prod AS (
  INSERT INTO creator_image_product_categories (creator_image_id, product_category_id)
  SELECT
    ins.id,
    x
  FROM ins
  CROSS JOIN LATERAL unnest(
    COALESCE(sqlc.arg(product_category_ids)::bigint[], '{}'::bigint[])
  ) AS x
  ON CONFLICT DO NOTHING
),
ins_type AS (
  INSERT INTO creator_image_type_categories (creator_image_id, type_category_id)
  SELECT
    ins.id,
    x
  FROM ins
  CROSS JOIN LATERAL unnest(
    COALESCE(sqlc.arg(type_category_ids)::bigint[], '{}'::bigint[])
  ) AS x
  ON CONFLICT DO NOTHING
)
SELECT * FROM ins;

-- name: UpdateCreatorImage :one
WITH upd AS (
  UPDATE creator_images
  SET
    name = sqlc.arg(name),
    image_url = sqlc.arg(image_url),
    is_published = sqlc.arg(is_published),
    price = sqlc.arg(price),
    profile_id = sqlc.narg(profile_id)
  WHERE id = sqlc.arg(id)
  RETURNING *
),
del_prod AS (
  DELETE FROM creator_image_product_categories
  WHERE creator_image_id = (SELECT id FROM upd)
),
ins_prod AS (
  INSERT INTO creator_image_product_categories (creator_image_id, product_category_id)
  SELECT
    upd.id,
    x
  FROM upd
  CROSS JOIN LATERAL unnest(
    COALESCE(sqlc.arg(product_category_ids)::bigint[], '{}'::bigint[])
  ) AS x
  ON CONFLICT DO NOTHING
),
del_type AS (
  DELETE FROM creator_image_type_categories
  WHERE creator_image_id = (SELECT id FROM upd)
),
ins_type AS (
  INSERT INTO creator_image_type_categories (creator_image_id, type_category_id)
  SELECT
    upd.id,
    x
  FROM upd
  CROSS JOIN LATERAL unnest(
    COALESCE(sqlc.arg(type_category_ids)::bigint[], '{}'::bigint[])
  ) AS x
  ON CONFLICT DO NOTHING
)
SELECT * FROM upd;

-- name: SoftDeleteCreatorImage :exec
UPDATE creator_images SET deleted_at = NOW() WHERE id = $1;

-- name: GetCreatorImageById :one
SELECT * FROM creator_images WHERE id = $1;
