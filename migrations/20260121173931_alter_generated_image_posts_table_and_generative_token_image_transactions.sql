-- +goose Up
-- +goose StatementBegin

-- 1. Add 'attempts' column to generated_image_posts
ALTER TABLE generated_image_posts
ADD COLUMN IF NOT EXISTS attempts INT NOT NULL DEFAULT 1;

-- 2. Add 'generated_image_post_id' to generative_token_image_transactions
ALTER TABLE generative_token_image_transactions
ADD COLUMN IF NOT EXISTS generated_image_post_id BIGINT;

-- 3. Add Foreign Key Constraint
-- Menggunakan ON DELETE SET NULL agar jika post dihapus, 
-- history transaksi token TIDAK hilang (hanya referensinya jadi null).
ALTER TABLE generative_token_image_transactions
ADD CONSTRAINT fk_generative_token_img_trans_post
FOREIGN KEY (generated_image_post_id)
REFERENCES generated_image_posts(id)
ON DELETE SET NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- 1. Drop Foreign Key Constraint
ALTER TABLE generative_token_image_transactions
DROP CONSTRAINT IF EXISTS fk_generative_token_img_trans_post;

-- 2. Drop Column from transactions
ALTER TABLE generative_token_image_transactions
DROP COLUMN IF EXISTS generated_image_post_id;

-- 3. Drop Column from posts
ALTER TABLE generated_image_posts
DROP COLUMN IF EXISTS attempts;

-- +goose StatementEnd