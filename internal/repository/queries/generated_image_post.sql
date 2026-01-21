-- internal/repository/queries/generated_image_post.sql

-- name: CreateGeneratedImagePost :one
INSERT INTO generated_image_posts (
    status,
    business_root_id,
    business_product_id,
    app_generative_image_model_id,
    app_generative_text_model_id,
    recorded_model_name,
    recorded_model_provider,
    mode,
    num_of_images,
    ratio,
    additional_prompt,
    design_style,
    category,
    reference_image_url,
    mask_image_url,
    image_size,
    current_caption,
    adv_bk_name,
    adv_bk_category,
    adv_bk_description,
    adv_bk_location,
    adv_bk_logo,
    adv_bk_unique_selling_point,
    adv_bk_website,
    adv_bk_vision_mission,
    adv_bk_color_tone,
    adv_pd_name,
    adv_pd_category,
    adv_pd_description,
    adv_pd_price,
    adv_rl_hashtags,
    rss_title,
    rss_url,
    rss_published_at,
    rss_image_url,
    rss_summary,
    rss_publisher
) VALUES (
    @status,
    @business_root_id,
    @business_product_id,
    @app_generative_image_model_id,
    sqlc.narg('app_generative_text_model_id'),
    @recorded_model_name,
    @recorded_model_provider,
    @mode,
    @num_of_images,
    @ratio,
    sqlc.narg('additional_prompt'),
    sqlc.narg('design_style'),
    sqlc.narg('category'),
    sqlc.narg('reference_image_url'),
    sqlc.narg('mask_image_url'),
    sqlc.narg('image_size'),
    sqlc.narg('current_caption'),
    @adv_bk_name,
    @adv_bk_category,
    @adv_bk_description,
    @adv_bk_location,
    @adv_bk_logo,
    @adv_bk_unique_selling_point,
    @adv_bk_website,
    @adv_bk_vision_mission,
    @adv_bk_color_tone,
    @adv_pd_name,
    @adv_pd_category,
    @adv_pd_description,
    @adv_pd_price,
    @adv_rl_hashtags,
    sqlc.narg('rss_title'),
    sqlc.narg('rss_url'),
    sqlc.narg('rss_published_at'),
    sqlc.narg('rss_image_url'),
    sqlc.narg('rss_summary'),
    sqlc.narg('rss_publisher')
) RETURNING *;

-- name: GetGeneratedImagePostById :one
SELECT * FROM generated_image_posts
WHERE id = @id AND deleted_at IS NULL;

-- name: GetAllGeneratedImagePostsByBusinessId :many
SELECT
    gip.*,
    -- items as JSON array
    COALESCE(
        (SELECT json_agg(json_build_object(
            'id', i.id,
            'image_url', i.image_url,
            'token_used', i.token_used,
            'created_at', i.created_at
        ) ORDER BY i.id)
         FROM generated_image_post_items i
         WHERE i.generated_image_post_id = gip.id AND i.deleted_at IS NULL),
        '[]'::json
    ) AS items,
    -- caption (COALESCE to handle NULL when no caption exists)
    COALESCE(
        (SELECT json_build_object(
            'id', c.id,
            'caption_text', c.caption_text,
            'token_used', c.token_used,
            'created_at', c.created_at
        )
         FROM generated_image_post_captions c
         WHERE c.generated_image_post_id = gip.id AND c.deleted_at IS NULL
         LIMIT 1),
        '{}'::json
    ) AS caption
FROM generated_image_posts gip
WHERE gip.business_root_id = @business_root_id
  AND gip.deleted_at IS NULL
  -- filter by status
  AND (sqlc.narg('status')::generated_image_post_status IS NULL OR gip.status = sqlc.narg('status')::generated_image_post_status)
  -- filter by mode
  AND (sqlc.narg('mode')::generated_image_post_mode IS NULL OR gip.mode = sqlc.narg('mode')::generated_image_post_mode)
  -- filter by product
  AND (sqlc.narg('business_product_id')::BIGINT IS NULL OR gip.business_product_id = sqlc.narg('business_product_id')::BIGINT)
  -- filter by date range
  AND (sqlc.narg('date_start')::TIMESTAMPTZ IS NULL OR gip.created_at >= sqlc.narg('date_start')::TIMESTAMPTZ)
  AND (sqlc.narg('date_end')::TIMESTAMPTZ IS NULL OR gip.created_at <= sqlc.narg('date_end')::TIMESTAMPTZ)
ORDER BY
    CASE WHEN sqlc.narg('sort_by')::TEXT = 'id' AND sqlc.narg('sort_dir')::TEXT = 'asc' THEN gip.id END ASC,
    CASE WHEN sqlc.narg('sort_by')::TEXT = 'id' AND sqlc.narg('sort_dir')::TEXT = 'desc' THEN gip.id END DESC,
    CASE WHEN sqlc.narg('sort_by')::TEXT = 'created_at' AND sqlc.narg('sort_dir')::TEXT = 'asc' THEN gip.created_at END ASC,
    CASE WHEN sqlc.narg('sort_by')::TEXT = 'created_at' AND sqlc.narg('sort_dir')::TEXT = 'desc' THEN gip.created_at END DESC,
    CASE WHEN sqlc.narg('sort_by')::TEXT = 'updated_at' AND sqlc.narg('sort_dir')::TEXT = 'asc' THEN gip.updated_at END ASC,
    CASE WHEN sqlc.narg('sort_by')::TEXT = 'updated_at' AND sqlc.narg('sort_dir')::TEXT = 'desc' THEN gip.updated_at END DESC,
    gip.id DESC -- default sort
LIMIT @page_limit OFFSET @page_offset;

-- name: CountGeneratedImagePostsByBusinessId :one
SELECT COUNT(*)
FROM generated_image_posts gip
WHERE gip.business_root_id = @business_root_id
  AND gip.deleted_at IS NULL
  AND (sqlc.narg('status')::generated_image_post_status IS NULL OR gip.status = sqlc.narg('status')::generated_image_post_status)
  AND (sqlc.narg('mode')::generated_image_post_mode IS NULL OR gip.mode = sqlc.narg('mode')::generated_image_post_mode)
  AND (sqlc.narg('business_product_id')::BIGINT IS NULL OR gip.business_product_id = sqlc.narg('business_product_id')::BIGINT)
  AND (sqlc.narg('date_start')::TIMESTAMPTZ IS NULL OR gip.created_at >= sqlc.narg('date_start')::TIMESTAMPTZ)
  AND (sqlc.narg('date_end')::TIMESTAMPTZ IS NULL OR gip.created_at <= sqlc.narg('date_end')::TIMESTAMPTZ);

-- name: UpdateGeneratedImagePostStatus :exec
UPDATE generated_image_posts
SET status = @status
WHERE id = @id;

-- name: UpdateGeneratedImagePostPrompts :exec
UPDATE generated_image_posts
SET 
    generated_image_prompt = sqlc.narg('generated_image_prompt'),
    generated_caption_prompt = sqlc.narg('generated_caption_prompt')
WHERE id = @id;

-- name: UpdateGeneratedImagePostError :exec
UPDATE generated_image_posts
SET 
    status = 'failed',
    error_log = @error_log
WHERE id = @id;
