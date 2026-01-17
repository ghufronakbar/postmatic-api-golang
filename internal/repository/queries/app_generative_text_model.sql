-- name: GetAllGenerativeTextModels :many
SELECT
  g.*
FROM app_generative_text_models g
WHERE
  g.deleted_at IS NULL
  AND (
    COALESCE(sqlc.narg(search), '') = ''
    OR g.model ILIKE ('%' || sqlc.narg(search) || '%')
    OR g.label ILIKE ('%' || sqlc.narg(search) || '%')
  )
  AND (
    sqlc.arg(is_admin)::boolean = TRUE
    OR g.is_active = TRUE
  )
ORDER BY
  -- label
  CASE WHEN sqlc.arg(sort_by) = 'label' AND sqlc.arg(sort_dir) = 'asc'  THEN g.label END ASC,
  CASE WHEN sqlc.arg(sort_by) = 'label' AND sqlc.arg(sort_dir) = 'desc' THEN g.label END DESC,

  -- model
  CASE WHEN sqlc.arg(sort_by) = 'model' AND sqlc.arg(sort_dir) = 'asc'  THEN g.model END ASC,
  CASE WHEN sqlc.arg(sort_by) = 'model' AND sqlc.arg(sort_dir) = 'desc' THEN g.model END DESC,

  -- created_at
  CASE WHEN sqlc.arg(sort_by) = 'created_at' AND sqlc.arg(sort_dir) = 'asc'  THEN g.created_at END ASC,
  CASE WHEN sqlc.arg(sort_by) = 'created_at' AND sqlc.arg(sort_dir) = 'desc' THEN g.created_at END DESC,

  -- updated_at
  CASE WHEN sqlc.arg(sort_by) = 'updated_at' AND sqlc.arg(sort_dir) = 'asc'  THEN g.updated_at END ASC,
  CASE WHEN sqlc.arg(sort_by) = 'updated_at' AND sqlc.arg(sort_dir) = 'desc' THEN g.updated_at END DESC,

  -- id
  CASE WHEN sqlc.arg(sort_by) = 'id' AND sqlc.arg(sort_dir) = 'asc'  THEN g.id END ASC,
  CASE WHEN sqlc.arg(sort_by) = 'id' AND sqlc.arg(sort_dir) = 'desc' THEN g.id END DESC,

  -- fallback stable order
  g.id DESC
LIMIT sqlc.arg(page_limit)
OFFSET sqlc.arg(page_offset);

-- name: CountAllGenerativeTextModels :one
SELECT COUNT(*)::bigint AS total
FROM app_generative_text_models g
WHERE
  g.deleted_at IS NULL
  AND (
    COALESCE(sqlc.narg(search), '') = ''
    OR g.model ILIKE ('%' || sqlc.narg(search) || '%')
    OR g.label ILIKE ('%' || sqlc.narg(search) || '%')
  )
  AND (
    sqlc.arg(is_admin)::boolean = TRUE
    OR g.is_active = TRUE
  );

-- name: GetGenerativeTextModelById :one
SELECT * FROM app_generative_text_models
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetGenerativeTextModelByIdAdmin :one
SELECT * FROM app_generative_text_models
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetGenerativeTextModelByIdUser :one
SELECT * FROM app_generative_text_models
WHERE id = $1 AND deleted_at IS NULL AND is_active = TRUE;

-- name: GetGenerativeTextModelByModel :one
SELECT * FROM app_generative_text_models
WHERE model = $1 AND deleted_at IS NULL;

-- name: GetGenerativeTextModelByModelAdmin :one
SELECT * FROM app_generative_text_models
WHERE model = $1 AND deleted_at IS NULL;

-- name: GetGenerativeTextModelByModelUser :one
SELECT * FROM app_generative_text_models
WHERE model = $1 AND deleted_at IS NULL AND is_active = TRUE;

-- name: CreateGenerativeTextModel :one
INSERT INTO app_generative_text_models (
  model,
  label,
  image,
  provider,
  is_active
) VALUES (
  $1, $2, $3, $4, $5
) RETURNING *;

-- name: UpdateGenerativeTextModel :one
UPDATE app_generative_text_models
SET
  model = $2,
  label = $3,
  image = $4,
  provider = $5,
  is_active = $6
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteGenerativeTextModel :one
UPDATE app_generative_text_models
SET deleted_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: CreateGenerativeTextModelChange :one
INSERT INTO app_generative_text_model_changes (
  action,
  profile_id,
  generative_text_model_id,
  before_model,
  before_label,
  before_image,
  before_provider,
  before_is_active,
  after_model,
  after_label,
  after_image,
  after_provider,
  after_is_active
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
) RETURNING *;
