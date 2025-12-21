// internal/module/business/business_product/viewmodel.go
package business_product

import "time"

type BusinessProductResponse struct {
	ID             string    `json:"id"`
	BusinessRootID string    `json:"businessRootId"`
	Name           string    `json:"name" validate:"required"`
	Category       string    `json:"category" validate:"required"`
	Description    string    `json:"description" validate:"required"`
	Price          int64     `json:"price" validate:"required"`
	Currency       string    `json:"currency" validate:"required"`
	ImageUrls      []string  `json:"imageUrls" validate:"required,min=1"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type SoftDeleteBusinessProductResponse struct {
	ID string `json:"id"`
}
