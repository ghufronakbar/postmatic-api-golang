-- name: GetAllPaymentMethods :many
SELECT
  p.*
FROM app_payment_methods p
WHERE
  p.deleted_at IS NULL
  AND (
    COALESCE(sqlc.narg(search), '') = ''
    OR p.name ILIKE ('%' || sqlc.narg(search) || '%')
    OR p.code ILIKE ('%' || sqlc.narg(search) || '%')
  )
  AND (
    sqlc.arg(is_admin)::boolean = TRUE
    OR p.is_active = TRUE
  )
ORDER BY
  -- name
  CASE WHEN sqlc.arg(sort_by) = 'name' AND sqlc.arg(sort_dir) = 'asc'  THEN p.name END ASC,
  CASE WHEN sqlc.arg(sort_by) = 'name' AND sqlc.arg(sort_dir) = 'desc' THEN p.name END DESC,

  -- code
  CASE WHEN sqlc.arg(sort_by) = 'code' AND sqlc.arg(sort_dir) = 'asc'  THEN p.code END ASC,
  CASE WHEN sqlc.arg(sort_by) = 'code' AND sqlc.arg(sort_dir) = 'desc' THEN p.code END DESC,

  -- created_at
  CASE WHEN sqlc.arg(sort_by) = 'created_at' AND sqlc.arg(sort_dir) = 'asc'  THEN p.created_at END ASC,
  CASE WHEN sqlc.arg(sort_by) = 'created_at' AND sqlc.arg(sort_dir) = 'desc' THEN p.created_at END DESC,

  -- updated_at
  CASE WHEN sqlc.arg(sort_by) = 'updated_at' AND sqlc.arg(sort_dir) = 'asc'  THEN p.updated_at END ASC,
  CASE WHEN sqlc.arg(sort_by) = 'updated_at' AND sqlc.arg(sort_dir) = 'desc' THEN p.updated_at END DESC,

  -- id
  CASE WHEN sqlc.arg(sort_by) = 'id' AND sqlc.arg(sort_dir) = 'asc'  THEN p.id END ASC,
  CASE WHEN sqlc.arg(sort_by) = 'id' AND sqlc.arg(sort_dir) = 'desc' THEN p.id END DESC,

  -- fallback stable order
  p.id DESC
LIMIT sqlc.arg(page_limit)
OFFSET sqlc.arg(page_offset);

-- name: CountAllPaymentMethods :one
SELECT COUNT(*)::bigint AS total
FROM app_payment_methods p
WHERE
  p.deleted_at IS NULL
  AND (
    COALESCE(sqlc.narg(search), '') = ''
    OR p.name ILIKE ('%' || sqlc.narg(search) || '%')
    OR p.code ILIKE ('%' || sqlc.narg(search) || '%')
  )
  AND (
    sqlc.arg(is_admin)::boolean = TRUE
    OR p.is_active = TRUE
  );

-- name: GetPaymentMethodById :one
SELECT * FROM app_payment_methods
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetPaymentMethodByIdAdmin :one
SELECT * FROM app_payment_methods
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetPaymentMethodByIdUser :one
SELECT * FROM app_payment_methods
WHERE id = $1 AND deleted_at IS NULL AND is_active = TRUE;

-- name: GetPaymentMethodByCode :one
SELECT * FROM app_payment_methods
WHERE code = $1 AND deleted_at IS NULL;

-- name: GetPaymentMethodByCodeAdmin :one
SELECT * FROM app_payment_methods
WHERE code = $1 AND deleted_at IS NULL;

-- name: GetPaymentMethodByCodeUser :one
SELECT * FROM app_payment_methods
WHERE code = $1 AND deleted_at IS NULL AND is_active = TRUE;

-- name: CreatePaymentMethod :one
INSERT INTO app_payment_methods (
  code,
  name,
  type,
  image,
  tax_fee,
  admin_type,
  admin_fee,
  is_active
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: UpdatePaymentMethod :one
UPDATE app_payment_methods
SET
  code = $2,
  name = $3,
  type = $4,
  image = $5,
  tax_fee = $6,
  admin_type = $7,
  admin_fee = $8,
  is_active = $9
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeletePaymentMethod :one
UPDATE app_payment_methods
SET deleted_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: CreatePaymentMethodChange :one
INSERT INTO app_payment_method_changes (
  action,
  profile_id,
  payment_method_id,
  before_code,
  before_name,
  before_type,
  before_image,
  before_admin_type,
  before_admin_fee,
  before_tax_fee,
  before_is_active,
  after_code,
  after_name,
  after_type,
  after_image,
  after_admin_type,
  after_admin_fee,
  after_tax_fee,
  after_is_active
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19
) RETURNING *;
