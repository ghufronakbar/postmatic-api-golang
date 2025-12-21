// internal/module/business/business_product/dto.go
package business_product

type CreateUpdateBusinessProductInput struct {
	Name        string   `json:"name" validate:"required"`
	Category    string   `json:"category" validate:"required"`
	Description string   `json:"description" validate:"required"`
	Price       int64    `json:"price" validate:"required"`
	Currency    string   `json:"currency" validate:"required"`
	ImageUrls   []string `json:"imageUrls" validate:"required,min=1"`
}
