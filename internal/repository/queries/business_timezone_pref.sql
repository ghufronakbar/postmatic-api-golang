-- name: GetBusinessTimezonePrefByBusinessRootId :one
SELECT * FROM business_timezone_prefs WHERE business_root_id = $1 LIMIT 1;

-- name: UpsertBusinessTimezonePref :one
INSERT INTO business_timezone_prefs (business_root_id, timezone)
VALUES ($1, $2)
ON CONFLICT (business_root_id) DO UPDATE
SET timezone = EXCLUDED.timezone
RETURNING *;
