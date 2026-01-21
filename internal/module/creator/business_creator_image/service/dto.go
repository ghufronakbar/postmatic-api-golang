// internal/module/creator/business_creator_image/service/dto.go
package business_creator_image_service

type GetSavedCreatorImageFilter struct {
	BusinessRootID    int64
	Search            string
	SortBy            string
	SortDir           string
	PageOffset        int
	PageLimit         int
	Page              int
	DateStart         *string
	DateEnd           *string
	TypeCategoryID    *int64
	ProductCategoryID *int64
}

type CreateSavedInput struct {
	BusinessRootID int64 `json:"-"`
	CreatorImageID int64 `json:"creatorImageId" validate:"required,gte=1"`
}

type DeleteSavedInput struct {
	BusinessRootID int64 `json:"-"`
	CreatorImageID int64 `json:"-"`
}
