-- name: CreateBusinessMemberStatusHistory :one
INSERT INTO business_member_status_histories (
	member_id,
	status,
	role
)
VALUES (
	$1,
	$2,
	$3
)
RETURNING *;