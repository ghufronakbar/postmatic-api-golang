// internal/module/app/image_uploader/viewmodel.go
package image_uploader_service

type ImageUploaderResponse struct {
	ID          int64  `json:"id"`
	Hashkey     string `json:"hashkey"`
	IsDuplicate bool   `json:"isDuplicate"`
	PublicId    string `json:"publicId"`
	ImageUrl    string `json:"imageUrl"`
	Size        int64  `json:"size"`
	Format      string `json:"format,omitempty"`
	Provider    string `json:"provider"`

	// Presign only
	Bucket           string            `json:"bucket,omitempty"`
	UploadUrl        string            `json:"uploadUrl,omitempty"`
	Headers          map[string]string `json:"headers,omitempty"`
	ExpiresInSeconds int64             `json:"expiresInSeconds,omitempty"`
}
