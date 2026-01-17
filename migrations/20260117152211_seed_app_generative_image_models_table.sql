-- +goose Up
-- +goose StatementBegin

-- Menggunakan Insert dengan ON CONFLICT agar aman dijalankan berulang (Idempotent)
INSERT INTO app_generative_image_models (
    model, 
    label, 
    provider, 
    is_active, 
    valid_ratios, 
    image_sizes
) VALUES 
(
    'gpt-image-1',
    'GPT Image 1',
    'openai',
    true,
    ARRAY['1:1', '2:3', '3:2'],
    NULL
),
(
    'gemini-2.5-flash-image',
    'Gemini 2.5 Flash Image',
    'google',
    true,
    ARRAY['1:1', '2:3', '3:2', '3:4', '4:3', '4:5', '5:4', '9:16', '16:9', '21:9'],
    NULL
),
(
    'gemini-3-pro-image-preview',
    'Gemini 3 Pro Image Preview',
    'google',
    true,
    ARRAY['1:1', '2:3', '3:2', '3:4', '4:3', '4:5', '5:4', '9:16', '16:9', '21:9'],
    ARRAY['1K', '2K', '4K']
)
ON CONFLICT (model) WHERE deleted_at IS NULL 
DO UPDATE SET
    label = EXCLUDED.label,
    provider = EXCLUDED.provider,
    valid_ratios = EXCLUDED.valid_ratios,
    image_sizes = EXCLUDED.image_sizes,
    updated_at = NOW();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Menghapus data yang di-seed (Hard delete untuk cleanup)
DELETE FROM app_generative_image_models 
WHERE model IN (
    'gpt-image-1', 
    'gemini-2.5-flash-image', 
    'gemini-3-pro-image-preview'
);

-- +goose StatementEnd