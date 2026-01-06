-- name: GetAppTokenProductByTypeCurrency :one
SELECT *
FROM app_token_products
WHERE token_type = $1
  AND currency_code = $2
  AND is_active = TRUE
  AND deleted_at IS NULL
ORDER BY sort_order ASC, created_at DESC
LIMIT 1;