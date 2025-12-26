// internal/module/creator/creator_image/dto.go
package creator_image

type CreateUpdateCreatorImageInput struct {
	Name               string  `json:"name" validate:"required"`
	ImageURL           string  `json:"imageUrl" validate:"required,url"`
	Price              int64   `json:"price" validate:"required,gte=0"`
	IsPublished        bool    `json:"isPublished"`
	TypeCategoryIds    []int64 `json:"typeCategoryIds" validate:"required,min=1,unique,gte=1"`
	ProductCategoryIds []int64 `json:"productCategoryIds" validate:"required,min=1,unique,gte=1"`
}
