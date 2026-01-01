// internal/module/creator/creator_image/dto.go
package creator_image

type CreateCreatorImageInput struct {
	ProfileID          string  `json:"profileId"`
	Name               string  `json:"name" validate:"required"`
	ImageURL           string  `json:"imageUrl" validate:"required,url"`
	Price              int64   `json:"price" validate:"gte=0"`
	IsPublished        bool    `json:"isPublished"`
	TypeCategoryIds    []int64 `json:"typeCategoryIds" validate:"required,min=1,unique,gte=1"`
	ProductCategoryIds []int64 `json:"productCategoryIds" validate:"required,min=1,unique,gte=1"`
}
type UpdateCreatorImageInput struct {
	CreatorImageId     int64   `json:"creatorImageId" validate:"required,gte=1"`
	ProfileID          string  `json:"profileId"`
	Name               string  `json:"name" validate:"required"`
	ImageURL           string  `json:"imageUrl" validate:"required,url"`
	Price              int64   `json:"price" validate:"gte=0"`
	IsPublished        bool    `json:"isPublished"`
	TypeCategoryIds    []int64 `json:"typeCategoryIds" validate:"required,min=1,unique,gte=1"`
	ProductCategoryIds []int64 `json:"productCategoryIds" validate:"required,min=1,unique,gte=1"`
}
