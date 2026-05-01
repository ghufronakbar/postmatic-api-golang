-- internal/repository/queries/generated_image_post_caption.sql

-- name: CreateGeneratedImagePostCaption :one
INSERT INTO generated_image_post_captions (
    generated_image_post_id,
    caption_text,
    token_used
) VALUES (
    @generated_image_post_id,
    @caption_text,
    @token_used
) RETURNING *;

-- name: GetGeneratedImagePostCaptionByPostId :one
SELECT * FROM generated_image_post_captions
WHERE generated_image_post_id = @generated_image_post_id
  AND deleted_at IS NULL;

-- name: GetGeneratedImagePostCaptionsByPostIds :many
SELECT * FROM generated_image_post_captions
WHERE generated_image_post_id = ANY(sqlc.arg(post_ids)::bigint[])
  AND deleted_at IS NULL;
