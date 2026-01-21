-- internal/repository/queries/generated_image_post_item.sql

-- name: CreateGeneratedImagePostItem :one
INSERT INTO generated_image_post_items (
    generated_image_post_id,
    image_url,
    token_used
) VALUES (
    @generated_image_post_id,
    @image_url,
    @token_used
) RETURNING *;

-- name: GetGeneratedImagePostItemsByPostId :many
SELECT * FROM generated_image_post_items
WHERE generated_image_post_id = @generated_image_post_id
  AND deleted_at IS NULL
ORDER BY id ASC;
