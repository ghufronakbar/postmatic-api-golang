-- name: GetAllCreatorImage :many
WITH p0 AS (
  SELECT
    lower(COALESCE(NULLIF(sqlc.narg(sort_by),  ''), '')) AS sb_in,
    lower(COALESCE(NULLIF(sqlc.narg(sort_dir), ''), '')) AS sd_in
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
    ci.id,
    ci.name,
    ci.image_url,
    ci.is_published,
    ci.price,
    ci.profile_id,
    ci.created_at,
    ci.updated_at,

    pub.id        AS publisher_id,
    pub.name      AS publisher_name,
    pub.image_url AS publisher_image,

    COALESCE(
      jsonb_agg(DISTINCT jsonb_build_object('id', tc.id, 'name', tc.name))
        FILTER (WHERE tc.id IS NOT NULL),
      '[]'::jsonb
    ) AS type_category_subs,

    COALESCE(
      jsonb_agg(DISTINCT jsonb_build_object('id', pc.id, 'name', pc.indonesian_name))
        FILTER (WHERE pc.id IS NOT NULL),
      '[]'::jsonb
    ) AS product_category_subs

  FROM creator_images ci
  LEFT JOIN profiles pub ON pub.id = ci.profile_id

  LEFT JOIN creator_image_type_categories citc
    ON citc.creator_image_id = ci.id
  LEFT JOIN app_creator_image_type_categories tc
    ON tc.id = citc.type_category_id

  LEFT JOIN creator_image_product_categories cipc
    ON cipc.creator_image_id = ci.id
  LEFT JOIN app_creator_image_product_categories pc
    ON pc.id = cipc.product_category_id

  WHERE
    ci.deleted_at IS NULL
    AND ci.is_banned = FALSE

    -- ✅ published filter (nullable)
    AND (
      sqlc.narg(published)::boolean IS NULL
      OR ci.is_published = sqlc.narg(published)
    )

    -- profile filter:
    AND (
      (sqlc.narg(profile_id)::uuid IS NULL AND ci.profile_id IS NULL)
      OR ci.profile_id = sqlc.narg(profile_id)
    )

    -- search
    AND (
      COALESCE(sqlc.narg(search), '') = ''
      OR ci.name ILIKE ('%' || sqlc.narg(search) || '%')
    )

    -- date range (created_at)
    AND (
      sqlc.narg(date_start)::date IS NULL
      OR ci.created_at::date >= sqlc.narg(date_start)::date
    )
    AND (
      sqlc.narg(date_end)::date IS NULL
      OR ci.created_at::date <= sqlc.narg(date_end)::date
    )

    -- filter by type category (EXISTS)
    AND (
      sqlc.narg(type_category_id)::bigint IS NULL
      OR EXISTS (
        SELECT 1
        FROM creator_image_type_categories f
        WHERE f.creator_image_id = ci.id
          AND f.type_category_id = sqlc.narg(type_category_id)
      )
    )

    -- filter by product category (EXISTS)
    AND (
      sqlc.narg(product_category_id)::bigint IS NULL
      OR EXISTS (
        SELECT 1
        FROM creator_image_product_categories f
        WHERE f.creator_image_id = ci.id
          AND f.product_category_id = sqlc.narg(product_category_id)
      )
    )

  GROUP BY
    ci.id, ci.name, ci.image_url, ci.is_published, ci.price, ci.profile_id, ci.created_at, ci.updated_at,
    pub.id, pub.name, pub.image_url
)
SELECT
  q.*
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

  -- id
  CASE WHEN p.sort_by = 'id' AND p.sort_dir = 'asc'  THEN q.id END ASC,
  CASE WHEN p.sort_by = 'id' AND p.sort_dir = 'desc' THEN q.id END DESC,

  -- fallback stable order
  q.id DESC
LIMIT sqlc.arg(page_limit)
OFFSET sqlc.arg(page_offset);

-- name: CountAllCreatorImage :one
SELECT COUNT(*)::bigint AS total
FROM creator_images ci
WHERE
  ci.deleted_at IS NULL
  AND ci.is_banned = FALSE

  -- ✅ published filter (nullable)
  AND (
    sqlc.narg(published)::boolean IS NULL
    OR ci.is_published = sqlc.narg(published)
  )

  AND (
    (sqlc.narg(profile_id)::uuid IS NULL AND ci.profile_id IS NULL)
    OR ci.profile_id = sqlc.narg(profile_id)
  )

  AND (
    COALESCE(sqlc.narg(search), '') = ''
    OR ci.name ILIKE ('%' || sqlc.narg(search) || '%')
  )

  AND (
    sqlc.narg(date_start)::date IS NULL
    OR ci.created_at::date >= sqlc.narg(date_start)::date
  )
  AND (
    sqlc.narg(date_end)::date IS NULL
    OR ci.created_at::date <= sqlc.narg(date_end)::date
  )

  AND (
    sqlc.narg(type_category_id)::bigint IS NULL
    OR EXISTS (
      SELECT 1
      FROM creator_image_type_categories f
      WHERE f.creator_image_id = ci.id
        AND f.type_category_id = sqlc.narg(type_category_id)
    )
  )

  AND (
    sqlc.narg(product_category_id)::bigint IS NULL
    OR EXISTS (
      SELECT 1
      FROM creator_image_product_categories f
      WHERE f.creator_image_id = ci.id
        AND f.product_category_id = sqlc.narg(product_category_id)
    )
  );


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
