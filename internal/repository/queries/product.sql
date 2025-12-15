-- name: GetProductById :one
SELECT * FROM products
WHERE id = $1 LIMIT 1;

-- name: ListProducts :many
SELECT * FROM products
ORDER BY created_at DESC;

-- name: CreateProduct :one
INSERT INTO products (name, price)
VALUES ($1, $2)
RETURNING *;

-- name: UpdateProduct :one
UPDATE products
SET name = $2, price = $3
WHERE id = $1
RETURNING *;

-- name: DeleteProduct :one
DELETE FROM products
WHERE id = $1
RETURNING *;

