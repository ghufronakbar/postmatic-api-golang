// internal/module/headless/s3_uploader/viewmodel.go
package s3_uploader

type PresignUploadImageResponse struct {
	Provider         string            `json:"provider"`
	Bucket           string            `json:"bucket"`
	PublicId         string            `json:"publicId"` // object key
	UploadUrl        string            `json:"uploadUrl"`
	Headers          map[string]string `json:"headers"` // wajib dipakai saat PUT
	ExpiresInSeconds int64             `json:"expiresInSeconds"`
}
