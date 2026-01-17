// internal/module/business/business_product/dto.go
package business_product_service

type CreateBusinessProductInput struct {
	Name           string   `json:"name" validate:"required"`
	Category       string   `json:"category" validate:"required"`
	Description    string   `json:"description" validate:"required"`
	Price          int64    `json:"price" validate:"required"`
	Currency       string   `json:"currency" validate:"required"`
	ImageUrls      []string `json:"imageUrls" validate:"required,min=1"`
	BusinessRootID int64    `json:"businessRootId" validate:"required"`
}

type UpdateBusinessProductInput struct {
	ID          int64    `json:"id" validate:"required"`
	Name        string   `json:"name" validate:"required"`
	Category    string   `json:"category" validate:"required"`
	Description string   `json:"description" validate:"required"`
	Price       int64    `json:"price" validate:"required"`
	Currency    string   `json:"currency" validate:"required"`
	ImageUrls   []string `json:"imageUrls" validate:"required,min=1"`
}
