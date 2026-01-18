-- name: CreatePaymentHistoryAction :one
INSERT INTO payment_history_actions (
    payment_history_id,
    name,
    label,
    value,
    value_type,
    payment_type,
    action_method,
    is_public
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: GetPaymentHistoryActionsByPaymentId :many
SELECT * FROM payment_history_actions
WHERE payment_history_id = $1
ORDER BY id ASC;

-- name: GetPublicPaymentHistoryActionsByPaymentId :many
SELECT * FROM payment_history_actions
WHERE payment_history_id = $1
AND is_public = TRUE
ORDER BY id ASC;

-- name: DeletePaymentHistoryActionsByPaymentId :exec
DELETE FROM payment_history_actions
WHERE payment_history_id = $1;
