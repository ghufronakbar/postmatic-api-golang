-- name: GetUserByEmailProfile :one
SELECT * FROM users
INNER JOIN profiles ON users.profile_id = profiles.id
WHERE profiles.email = $1 LIMIT 1;

-- name: GetUserById :one
SELECT * FROM users
WHERE id = $1 LIMIT 1;

-- name: ListUsersByProfileId :many
SELECT * FROM users
WHERE profile_id = $1
ORDER BY created_at DESC;

-- name: CreateUser :one
INSERT INTO users (profile_id, password, provider)
VALUES ($1, $2, $3)
RETURNING *;

-- name: VerifyUser :one
UPDATE users
SET verified_at = now()
WHERE id = $1 RETURNING *;