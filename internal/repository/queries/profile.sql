-- name: GetProfileByEmail :one
SELECT * FROM profiles
WHERE email = $1 LIMIT 1;

-- name: GetProfileById :one
SELECT * FROM profiles
WHERE id = $1 LIMIT 1;

-- name: CreateProfile :one
INSERT INTO profiles (name, email, image_url, country_code, phone, description)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;