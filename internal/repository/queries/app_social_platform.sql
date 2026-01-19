-- name: CreateAppSocialPlatform :one
INSERT INTO app_social_platforms (
    platform_code,
    logo,
    name,
    hint,
    is_active
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetAppSocialPlatformById :one
SELECT * FROM app_social_platforms
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetAppSocialPlatformByPlatformCode :one
SELECT * FROM app_social_platforms
WHERE platform_code = $1 AND deleted_at IS NULL;

-- name: GetAllAppSocialPlatforms :many
SELECT * FROM app_social_platforms p
WHERE
    p.deleted_at IS NULL
    AND (
        sqlc.arg(include_inactive)::boolean = true
        OR p.is_active = true
    )
    AND (
        COALESCE(sqlc.narg(search), '') = ''
        OR p.name ILIKE ('%' || sqlc.narg(search) || '%')
        OR p.platform_code::text ILIKE ('%' || sqlc.narg(search) || '%')
    )
ORDER BY
    CASE WHEN sqlc.arg(sort_by) = 'id' AND sqlc.arg(sort_dir) = 'asc' THEN p.id END ASC,
    CASE WHEN sqlc.arg(sort_by) = 'id' AND sqlc.arg(sort_dir) = 'desc' THEN p.id END DESC,
    CASE WHEN sqlc.arg(sort_by) = 'name' AND sqlc.arg(sort_dir) = 'asc' THEN p.name END ASC,
    CASE WHEN sqlc.arg(sort_by) = 'name' AND sqlc.arg(sort_dir) = 'desc' THEN p.name END DESC,
    CASE WHEN sqlc.arg(sort_by) = 'platform_code' AND sqlc.arg(sort_dir) = 'asc' THEN p.platform_code END ASC,
    CASE WHEN sqlc.arg(sort_by) = 'platform_code' AND sqlc.arg(sort_dir) = 'desc' THEN p.platform_code END DESC,
    CASE WHEN sqlc.arg(sort_by) = 'is_active' AND sqlc.arg(sort_dir) = 'asc' THEN p.is_active END ASC,
    CASE WHEN sqlc.arg(sort_by) = 'is_active' AND sqlc.arg(sort_dir) = 'desc' THEN p.is_active END DESC,
    p.id DESC
LIMIT sqlc.arg(page_limit)
OFFSET sqlc.arg(page_offset);

-- name: CountAllAppSocialPlatforms :one
SELECT COUNT(*)::bigint AS total
FROM app_social_platforms p
WHERE
    p.deleted_at IS NULL
    AND (
        sqlc.arg(include_inactive)::boolean = true
        OR p.is_active = true
    )
    AND (
        COALESCE(sqlc.narg(search), '') = ''
        OR p.name ILIKE ('%' || sqlc.narg(search) || '%')
        OR p.platform_code::text ILIKE ('%' || sqlc.narg(search) || '%')
    );

-- name: UpdateAppSocialPlatform :one
UPDATE app_social_platforms
SET
    platform_code = $2,
    logo = $3,
    name = $4,
    hint = $5,
    is_active = $6
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: DeleteAppSocialPlatform :one
UPDATE app_social_platforms
SET deleted_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: CreateAppSocialPlatformChange :one
INSERT INTO app_social_platform_changes (
    action,
    profile_id,
    social_platform_id,
    before_platform_code,
    before_logo,
    before_name,
    before_hint,
    before_is_active,
    after_platform_code,
    after_logo,
    after_name,
    after_hint,
    after_is_active
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
) RETURNING *;
