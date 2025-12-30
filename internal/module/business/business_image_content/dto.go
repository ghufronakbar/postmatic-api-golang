// internal/module/business/business_image_content/dto.go
package business_image_content

type CreateUpdateBusinessImageContentInput struct {
	BusinessRootID    int64
	ImageUrls         []string `json:"imageUrls" validate:"required,min=1"`
	Caption           string   `json:"caption"`
	Type              string   `json:"type" validate:"required,oneof=personal generated"`
	ReadyToPost       bool     `json:"readyToPost"`
	Category          string   `json:"category"`
	BusinessProductID *int64   `json:"businessProductId"`
}
