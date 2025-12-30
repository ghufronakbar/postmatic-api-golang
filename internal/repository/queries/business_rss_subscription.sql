-- name: GetBusinessRssSubscriptionsByBusinessRootID :many
SELECT
  -- business_rss_subscriptions (include created_at & updated_at)
  brs.id              AS subscription_id,
  brs.title           AS subscription_title,
  brs.is_active       AS subscription_is_active,
  brs.business_root_id AS subscription_business_root_id,
  brs.app_rss_feed_id AS subscription_app_rss_feed_id,
  brs.created_at      AS subscription_created_at,
  brs.updated_at      AS subscription_updated_at,
  brs.deleted_at      AS subscription_deleted_at,

  -- app_rss_feeds (exclude created_at & updated_at)
  arf.id                AS feed_id,
  arf.title             AS feed_title,
  arf.url               AS feed_url,
  arf.publisher         AS feed_publisher,
  arf.app_rss_category_id AS feed_app_rss_category_id,
  arf.deleted_at        AS feed_deleted_at,

  -- app_rss_categories (exclude created_at & updated_at)
  arc.id         AS category_id,
  arc.name       AS category_name,
  arc.deleted_at AS category_deleted_at

FROM business_rss_subscriptions brs
LEFT JOIN app_rss_feeds arf
  ON arf.id = brs.app_rss_feed_id
  AND arf.deleted_at IS NULL
LEFT JOIN app_rss_categories arc
  ON arc.id = arf.app_rss_category_id
  AND arc.deleted_at IS NULL

WHERE
  brs.business_root_id = sqlc.arg(business_root_id)
  AND brs.deleted_at IS NULL
  AND (
    COALESCE(sqlc.narg(search), '') = ''
    OR brs.title ILIKE ('%' || sqlc.narg(search) || '%')
    OR COALESCE(arf.title, '') ILIKE ('%' || sqlc.narg(search) || '%')
    OR COALESCE(arf.publisher, '') ILIKE ('%' || sqlc.narg(search) || '%')
    OR COALESCE(arf.url, '') ILIKE ('%' || sqlc.narg(search) || '%')
    OR COALESCE(arc.name, '') ILIKE ('%' || sqlc.narg(search) || '%')
  )

ORDER BY
  -- title
  CASE WHEN sqlc.arg(sort_by) = 'title' AND sqlc.arg(sort_dir) = 'asc'  THEN brs.title END ASC,
  CASE WHEN sqlc.arg(sort_by) = 'title' AND sqlc.arg(sort_dir) = 'desc' THEN brs.title END DESC,

  -- created_at
  CASE WHEN sqlc.arg(sort_by) = 'created_at' AND sqlc.arg(sort_dir) = 'asc'  THEN brs.created_at END ASC,
  CASE WHEN sqlc.arg(sort_by) = 'created_at' AND sqlc.arg(sort_dir) = 'desc' THEN brs.created_at END DESC,

  -- updated_at
  CASE WHEN sqlc.arg(sort_by) = 'updated_at' AND sqlc.arg(sort_dir) = 'asc'  THEN brs.updated_at END ASC,
  CASE WHEN sqlc.arg(sort_by) = 'updated_at' AND sqlc.arg(sort_dir) = 'desc' THEN brs.updated_at END DESC,

  CASE WHEN sqlc.arg(sort_by) = 'id' AND sqlc.arg(sort_dir) = 'asc'  THEN brs.id END ASC,
  CASE WHEN sqlc.arg(sort_by) = 'id' AND sqlc.arg(sort_dir) = 'desc' THEN brs.id END DESC,

  -- fallback stable order
  brs.id DESC

LIMIT sqlc.arg(page_limit)
OFFSET sqlc.arg(page_offset);


-- name: CountBusinessRssSubscriptionsByBusinessRootID :one
SELECT COUNT(*)::bigint AS total
FROM business_rss_subscriptions brs
LEFT JOIN app_rss_feeds arf
  ON arf.id = brs.app_rss_feed_id
  AND arf.deleted_at IS NULL
LEFT JOIN app_rss_categories arc
  ON arc.id = arf.app_rss_category_id
  AND arc.deleted_at IS NULL
WHERE
  brs.business_root_id = sqlc.arg(business_root_id)
  AND brs.deleted_at IS NULL
  AND (
    COALESCE(sqlc.narg(search), '') = ''
    OR brs.title ILIKE ('%' || sqlc.narg(search) || '%')
    OR COALESCE(arf.title, '') ILIKE ('%' || sqlc.narg(search) || '%')
    OR COALESCE(arf.publisher, '') ILIKE ('%' || sqlc.narg(search) || '%')
    OR COALESCE(arf.url, '') ILIKE ('%' || sqlc.narg(search) || '%')
    OR COALESCE(arc.name, '') ILIKE ('%' || sqlc.narg(search) || '%')
  );


-- name: HardDeleteBusinessRssSubscriptionByID :exec
DELETE FROM business_rss_subscriptions
WHERE id = sqlc.arg(id);

-- name: CreateBusinessRssSubscription :one
INSERT INTO business_rss_subscriptions (
  business_root_id,
  title,
  is_active,
  app_rss_feed_id
) VALUES (
  sqlc.arg(business_root_id),
  sqlc.arg(title),
  sqlc.arg(is_active),
  sqlc.arg(app_rss_feed_id)
) RETURNING *;

-- name: EditBusinessRssSubscription :one
UPDATE business_rss_subscriptions
SET
  title = sqlc.arg(title),
  is_active = sqlc.arg(is_active),
  app_rss_feed_id = sqlc.arg(app_rss_feed_id)
WHERE id = sqlc.arg(id)
RETURNING *;


-- name: GetBusinessRssSubscriptionByBusinessRootIdAndAppRssFeedId :one
SELECT * FROM business_rss_subscriptions
WHERE business_root_id = $1
AND app_rss_feed_id = $2
AND deleted_at IS NULL;

-- name: GetBusinessRssSubscriptionById :one
SELECT * FROM business_rss_subscriptions
WHERE id = $1;

-- name: GetBusinessRssSubscriptionByIDAndBusinessRootID :one
SELECT *
FROM business_rss_subscriptions
WHERE id = $1
  AND business_root_id = $2
  AND deleted_at IS NULL
LIMIT 1;

-- name: ExistsBusinessRssSubscriptionByBusinessRootIDAndFeedIDExceptID :one
SELECT EXISTS (
  SELECT 1
  FROM business_rss_subscriptions
  WHERE business_root_id = $1
    AND app_rss_feed_id = $2
    AND deleted_at IS NULL
    AND id <> $3
) AS exists;