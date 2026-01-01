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

-- name: GetBusinessMemberStatusHistoryByMemberID :one
WITH latest_history AS (
  SELECT
    h.id,
    h.status,
    h.role,
    h.member_id,
    h.created_at,
    h.updated_at,
    h.deleted_at
  FROM business_member_status_histories AS h
  WHERE
    h.member_id = sqlc.arg(member_id)
    AND h.deleted_at IS NULL
  ORDER BY h.id DESC
  LIMIT 1
)
SELECT
  -- business_member_status_histories (latest row)
  h.id         AS history_id,
  h.status     AS history_status,
  h.role       AS history_role,
  h.member_id  AS history_member_id,
  h.created_at AS history_created_at,
  h.updated_at AS history_updated_at,
  h.deleted_at AS history_deleted_at,

  -- business_members
  m.id               AS member_id,
  m.status           AS member_status,
  m.role             AS member_role,
  m.answered_at      AS member_answered_at,
  m.business_root_id AS member_business_root_id,
  m.profile_id       AS member_profile_id,
  m.created_at       AS member_created_at,
  m.updated_at       AS member_updated_at,
  m.deleted_at       AS member_deleted_at,

  -- profiles
  p.name      AS profile_name,
  p.image_url AS profile_image_url,
  p.email     AS profile_email,

  -- business_root
  bk.name AS business_root_name,
  br.id AS business_root_id
FROM latest_history AS h
JOIN business_members AS m
  ON m.id = h.member_id
JOIN profiles AS p
  ON p.id = m.profile_id
JOIN business_roots AS br
  ON br.id = m.business_root_id
JOIN business_knowledges AS bk
  ON bk.business_root_id = br.id
WHERE
  m.deleted_at IS NULL;
