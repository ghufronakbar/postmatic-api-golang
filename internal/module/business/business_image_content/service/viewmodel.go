// internal/module/business/business_image_content/viewmodel.go
package business_image_content_service

import "time"

type BusinessImageContentResponse struct {
	ID                int64     `json:"id"`
	BusinessRootID    int64     `json:"businessRootId"`
	BusinessProductID *int64    `json:"businessProductId"`
	Caption           string    `json:"caption"`
	Type              string    `json:"type"`
	ReadyToPost       bool      `json:"readyToPost"`
	Category          string    `json:"category"`
	ImageUrls         []string  `json:"imageUrls"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

type SoftDeleteBusinessImageContentResponse struct {
	ID int64 `json:"id"`
}
