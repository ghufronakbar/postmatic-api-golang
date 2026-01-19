-- name: CreateGenerativeTokenImageTransaction :one
INSERT INTO generative_token_image_transactions (
    type,
    amount,
    profile_id,
    business_root_id,
    payment_history_id
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetGenerativeTokenImageTransactionByPaymentHistoryId :one
SELECT * FROM generative_token_image_transactions
WHERE payment_history_id = $1 AND deleted_at IS NULL;

-- name: GetSuccessPaymentIdsWithoutTokenTransaction :many
SELECT ph.id, ph.profile_id, ph.business_root_id, ph.product_amount
FROM payment_histories ph
LEFT JOIN generative_token_image_transactions gt 
    ON gt.payment_history_id = ph.id AND gt.deleted_at IS NULL
WHERE 
    ph.id = ANY(sqlc.arg(payment_ids)::uuid[])
    AND ph.status = 'success'
    AND ph.deleted_at IS NULL
    AND gt.id IS NULL;

-- name: SumTokenByBusinessAndType :one
SELECT 
    COALESCE(SUM(amount), 0)::bigint AS total
FROM generative_token_image_transactions
WHERE 
    business_root_id = $1 
    AND type = $2
    AND deleted_at IS NULL;
