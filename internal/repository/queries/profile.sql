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

-- name: UpdateProfile :one
UPDATE profiles
SET name = $2, image_url = $3, country_code = $4, phone = $5, description = $6
WHERE id = $1
RETURNING *;
