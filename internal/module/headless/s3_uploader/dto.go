// internal/module/headless/s3_uploader/dto.go
package s3_uploader

type PresignUploadImageInput struct {
	Hash        string `json:"hash" validate:"required"`
	Format      string `json:"format" validate:"required"`      // contoh: png, jpg, webp
	ContentType string `json:"contentType" validate:"required"` // contoh: image/png
	Size        int64  `json:"size" validate:"required"`        // opsional (buat validasi tambahan nanti)
}
