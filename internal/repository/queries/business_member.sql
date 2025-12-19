-- name: CreateBusinessMember :one
INSERT INTO business_members (
	status,
	role,
	answered_at,
	business_root_id,
	profile_id
)
VALUES (
	$1,
	$2,
	$3,
	$4,
	$5
)
RETURNING *;
