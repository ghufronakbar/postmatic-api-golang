package image_uploader

type ImageUploaderResponse struct {
	ID          int64  `json:"id"`
	Hashkey     string `json:"hashkey"`
	IsDuplicate bool   `json:"isDuplicate"`
	PublicId    string `json:"publicId"`
	Size        int64  `json:"size"`
	ImageUrl    string `json:"imageUrl"`
	Provider    string `json:"provider"`
	Format      string `json:"format"`
}
