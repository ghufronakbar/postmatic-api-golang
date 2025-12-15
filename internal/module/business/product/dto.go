// internal/module/business/product/dto.go
package product

// CreateProductInput sekarang bertindak sebagai:
// 1. Service Input (Business Layer)
// 2. HTTP Request Body (Transport Layer)
// 3. Validation Schema
type CreateProductInput struct {
	// Gunakan tags lengkap disini
	Name  string `json:"name" validate:"required,min=3"`
	Price int    `json:"price" validate:"required,gte=100"`
}
