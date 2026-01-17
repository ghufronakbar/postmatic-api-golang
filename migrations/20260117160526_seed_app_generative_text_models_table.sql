-- +goose Up
-- +goose StatementBegin

INSERT INTO app_generative_text_models (
    model, 
    label, 
    provider, 
    is_active
) VALUES 

-- ==========================================================================================
-- OPENAI MODELS
-- Input Image: Mendukung URL (Public) dan Base64.
-- ==========================================================================================

-- 1. GPT-4o (Omni)
-- Tipe: Flagship / High Intelligence.
-- Biaya: Menengah (Lebih hemat dari GPT-4 Turbo, tapi lebih mahal dari Mini).
-- Use Case: Analisis gambar yang kompleks, tulisan tangan sulit, atau butuh nuansa detail.
-- Input: URL & Base64.
(
    'gpt-4o',
    'GPT-4o (Omni)',
    'openai',
    true
),

-- 2. GPT-4o Mini
-- Tipe: Cost Efficient / Hemat.
-- Biaya: SANGAT MURAH (Sekitar 1/30 harga GPT-4o).
-- Use Case: Captioning massal, deskripsi produk standar, validasi gambar sederhana.
-- Input: URL & Base64.
(
    'gpt-4o-mini',
    'GPT-4o Mini',
    'openai',
    true
),

-- 3. GPT-4 Turbo (Legacy Vision)
-- Tipe: High End Legacy.
-- Biaya: MAHAL (Boros).
-- Note: Sebaiknya gunakan GPT-4o, tapi ini dimasukkan untuk backward compatibility.
-- Input: URL & Base64.
(
    'gpt-4-turbo',
    'GPT-4 Turbo',
    'openai',
    true
),

-- ==========================================================================================
-- GOOGLE GEMINI MODELS
-- Input Image: Utamanya Base64 (Inline Data) atau File API (Upload). 
-- Google API standar tidak langsung mendownload dari URL publik di payload request.
-- ==========================================================================================

-- 4. Gemini 1.5 Pro
-- Tipe: High Reasoning / Long Context.
-- Biaya: Menengah ke Atas (Tergantung panjang prompt/context).
-- Use Case: Analisis dokumen visual, chart, infografis, atau video panjang.
-- Input: Base64 (limitasi size) atau File API URI.
(
    'gemini-1.5-pro',
    'Gemini 1.5 Pro',
    'google',
    true
),

-- 5. Gemini 1.5 Flash
-- Tipe: High Speed / Cost Efficient.
-- Biaya: MURAH (Hemat).
-- Use Case: Alternatif murah untuk GPT-4o-mini, captioning cepat, OCR dokumen sederhana.
-- Input: Base64 (limitasi size) atau File API URI.
(
    'gemini-1.5-flash',
    'Gemini 1.5 Flash',
    'google',
    true
),

-- 6. Gemini 1.0 Pro Vision (Legacy)
-- Tipe: Legacy.
-- Biaya: Hemat.
-- Note: Model lama, kemampuan vision-nya di bawah 1.5 Flash. Hanya untuk legacy support.
-- Input: Base64.
(
    'gemini-pro-vision',
    'Gemini Pro Vision (Legacy)',
    'google',
    false -- Diset false (non-aktif) by default karena sudah ada 1.5 Flash
)

ON CONFLICT (model) WHERE deleted_at IS NULL 
DO UPDATE SET
    label = EXCLUDED.label,
    provider = EXCLUDED.provider,
    is_active = EXCLUDED.is_active,
    updated_at = NOW();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Cleanup data seeding
DELETE FROM app_generative_text_models 
WHERE model IN (
    'gpt-4o', 
    'gpt-4o-mini', 
    'gpt-4-turbo',
    'gemini-1.5-pro',
    'gemini-1.5-flash',
    'gemini-pro-vision'
);

-- +goose StatementEnd