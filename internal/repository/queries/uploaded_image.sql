-- name: GetUploadedImageByHashkey :one
SELECT * FROM uploaded_images WHERE hashkey = $1;

-- name: InsertUploadedImage :one
INSERT INTO uploaded_images (hashkey, public_id, image_url, size, provider)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (hashkey)
DO UPDATE SET
  public_id = EXCLUDED.public_id,
  image_url = EXCLUDED.image_url,
  size      = EXCLUDED.size,
  provider  = EXCLUDED.provider
RETURNING id, hashkey, public_id, image_url, size, provider;
