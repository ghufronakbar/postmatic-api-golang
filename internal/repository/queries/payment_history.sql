-- name: CreatePaymentHistory :one
INSERT INTO payment_histories (
    profile_id,
    business_root_id,
    product_amount,
    status,
    currency,
    payment_method,
    payment_method_type,
    record_product_name,
    record_product_type,
    record_product_price,
    record_product_image_url,
    reference_product_id,
    subtotal_item_amount,
    discount_amount,
    discount_percentage,
    discount_type,
    admin_fee_amount,
    admin_fee_percentage,
    admin_fee_type,
    tax_amount,
    tax_percentage,
    referral_record_id,
    midtrans_transaction_id,
    midtrans_expired_at,
    payment_pending_at,
    total_amount
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26
) RETURNING *;

-- name: GetPaymentHistoryById :one
SELECT * FROM payment_histories
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetPaymentHistoryByIdAndProfile :one
SELECT * FROM payment_histories
WHERE id = $1 AND profile_id = $2 AND deleted_at IS NULL;

-- name: GetPaymentHistoryByMidtransTransactionId :one
SELECT * FROM payment_histories
WHERE midtrans_transaction_id = $1 AND deleted_at IS NULL;

-- name: GetAllPaymentHistories :many
SELECT p.*
FROM payment_histories p
WHERE
    p.deleted_at IS NULL
    AND p.profile_id = sqlc.arg(profile_id)
    AND (
        COALESCE(sqlc.narg(search), '') = ''
        OR p.record_product_name ILIKE ('%' || sqlc.narg(search) || '%')
        OR p.payment_method ILIKE ('%' || sqlc.narg(search) || '%')
    )
    AND (
        sqlc.narg(status)::payment_status IS NULL
        OR p.status = sqlc.narg(status)::payment_status
    )
ORDER BY
    CASE WHEN sqlc.arg(sort_by) = 'created_at' AND sqlc.arg(sort_dir) = 'asc' THEN p.created_at END ASC,
    CASE WHEN sqlc.arg(sort_by) = 'created_at' AND sqlc.arg(sort_dir) = 'desc' THEN p.created_at END DESC,
    CASE WHEN sqlc.arg(sort_by) = 'total_amount' AND sqlc.arg(sort_dir) = 'asc' THEN p.total_amount END ASC,
    CASE WHEN sqlc.arg(sort_by) = 'total_amount' AND sqlc.arg(sort_dir) = 'desc' THEN p.total_amount END DESC,
    p.created_at DESC
LIMIT sqlc.arg(page_limit)
OFFSET sqlc.arg(page_offset);

-- name: CountAllPaymentHistories :one
SELECT COUNT(*)::bigint AS total
FROM payment_histories p
WHERE
    p.deleted_at IS NULL
    AND p.profile_id = sqlc.arg(profile_id)
    AND (
        COALESCE(sqlc.narg(search), '') = ''
        OR p.record_product_name ILIKE ('%' || sqlc.narg(search) || '%')
        OR p.payment_method ILIKE ('%' || sqlc.narg(search) || '%')
    )
    AND (
        sqlc.narg(status)::payment_status IS NULL
        OR p.status = sqlc.narg(status)::payment_status
    );

-- name: UpdatePaymentHistoryStatus :one
UPDATE payment_histories
SET 
    status = @status::payment_status,
    payment_success_at = CASE WHEN @status::payment_status = 'success'::payment_status THEN NOW() ELSE payment_success_at END,
    payment_failed_at = CASE WHEN @status::payment_status = 'failed'::payment_status THEN NOW() ELSE payment_failed_at END,
    payment_canceled_at = CASE WHEN @status::payment_status = 'canceled'::payment_status THEN NOW() ELSE payment_canceled_at END,
    payment_expired_at = CASE WHEN @status::payment_status = 'expired'::payment_status THEN NOW() ELSE payment_expired_at END,
    payment_refunded_at = CASE WHEN @status::payment_status = 'refunded'::payment_status THEN NOW() ELSE payment_refunded_at END
WHERE id = @id AND deleted_at IS NULL
RETURNING *;

-- name: UpdatePaymentHistoryMidtransId :one
UPDATE payment_histories
SET midtrans_transaction_id = $2
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: GetPaymentHistoryByIdAndBusiness :one
SELECT * FROM payment_histories
WHERE id = $1 AND business_root_id = $2 AND deleted_at IS NULL;

-- name: GetAllPaymentHistoriesByBusiness :many
SELECT p.*
FROM payment_histories p
WHERE
    p.deleted_at IS NULL
    AND p.business_root_id = sqlc.arg(business_root_id)
    AND (
        COALESCE(sqlc.narg(search), '') = ''
        OR p.record_product_name ILIKE ('%' || sqlc.narg(search) || '%')
        OR p.payment_method ILIKE ('%' || sqlc.narg(search) || '%')
    )
    AND (
        sqlc.narg(status)::payment_status IS NULL
        OR p.status = sqlc.narg(status)::payment_status
    )
ORDER BY
    CASE WHEN sqlc.arg(sort_by) = 'created_at' AND sqlc.arg(sort_dir) = 'asc' THEN p.created_at END ASC,
    CASE WHEN sqlc.arg(sort_by) = 'created_at' AND sqlc.arg(sort_dir) = 'desc' THEN p.created_at END DESC,
    CASE WHEN sqlc.arg(sort_by) = 'total_amount' AND sqlc.arg(sort_dir) = 'asc' THEN p.total_amount END ASC,
    CASE WHEN sqlc.arg(sort_by) = 'total_amount' AND sqlc.arg(sort_dir) = 'desc' THEN p.total_amount END DESC,
    p.created_at DESC
LIMIT sqlc.arg(page_limit)
OFFSET sqlc.arg(page_offset);

-- name: CountAllPaymentHistoriesByBusiness :one
SELECT COUNT(*)::bigint AS total
FROM payment_histories p
WHERE
    p.deleted_at IS NULL
    AND p.business_root_id = sqlc.arg(business_root_id)
    AND (
        COALESCE(sqlc.narg(search), '') = ''
        OR p.record_product_name ILIKE ('%' || sqlc.narg(search) || '%')
        OR p.payment_method ILIKE ('%' || sqlc.narg(search) || '%')
    )
    AND (
        sqlc.narg(status)::payment_status IS NULL
        OR p.status = sqlc.narg(status)::payment_status
    );
