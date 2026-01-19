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

-- name: GetAllTokenTransactionsByBusiness :many
SELECT t.*
FROM generative_token_image_transactions t
WHERE
    t.deleted_at IS NULL
    AND t.business_root_id = sqlc.arg(business_root_id)
    AND (
        sqlc.narg(type)::token_transaction_type IS NULL
        OR t.type = sqlc.narg(type)::token_transaction_type
    )
    AND (
        sqlc.narg(date_start)::date IS NULL
        OR t.created_at::date >= sqlc.narg(date_start)::date
    )
    AND (
        sqlc.narg(date_end)::date IS NULL
        OR t.created_at::date <= sqlc.narg(date_end)::date
    )
ORDER BY
    CASE WHEN sqlc.arg(sort_by) = 'id' AND sqlc.arg(sort_dir) = 'asc' THEN t.id END ASC,
    CASE WHEN sqlc.arg(sort_by) = 'id' AND sqlc.arg(sort_dir) = 'desc' THEN t.id END DESC,
    CASE WHEN sqlc.arg(sort_by) = 'created_at' AND sqlc.arg(sort_dir) = 'asc' THEN t.created_at END ASC,
    CASE WHEN sqlc.arg(sort_by) = 'created_at' AND sqlc.arg(sort_dir) = 'desc' THEN t.created_at END DESC,
    CASE WHEN sqlc.arg(sort_by) = 'amount' AND sqlc.arg(sort_dir) = 'asc' THEN t.amount END ASC,
    CASE WHEN sqlc.arg(sort_by) = 'amount' AND sqlc.arg(sort_dir) = 'desc' THEN t.amount END DESC,
    t.id DESC
LIMIT sqlc.arg(page_limit)
OFFSET sqlc.arg(page_offset);

-- name: CountAllTokenTransactionsByBusiness :one
SELECT COUNT(*)::bigint AS total
FROM generative_token_image_transactions t
WHERE
    t.deleted_at IS NULL
    AND t.business_root_id = sqlc.arg(business_root_id)
    AND (
        sqlc.narg(type)::token_transaction_type IS NULL
        OR t.type = sqlc.narg(type)::token_transaction_type
    )
    AND (
        sqlc.narg(date_start)::date IS NULL
        OR t.created_at::date >= sqlc.narg(date_start)::date
    )
    AND (
        sqlc.narg(date_end)::date IS NULL
        OR t.created_at::date <= sqlc.narg(date_end)::date
    );

