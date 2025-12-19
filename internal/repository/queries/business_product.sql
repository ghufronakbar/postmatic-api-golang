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

-- name: SoftDeleteBusinessProductByBusinessRootID :one
UPDATE business_products
SET deleted_at = NOW()
WHERE business_root_id = sqlc.arg(business_root_id)
RETURNING id;