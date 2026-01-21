// internal/module/creator/business_creator_image/service/viewmodel.go
package business_creator_image_service

import "time"

type SavedCreatorImageResponse struct {
	ID                  int64                `json:"id"`
	Name                string               `json:"name"`
	ImageURL            *string              `json:"imageUrl"` // nil jika banned/not published/deleted
	IsPublished         bool                 `json:"isPublished"`
	Price               int64                `json:"price"`
	Publisher           *PublisherSub        `json:"publisher"`
	TypeCategorySubs    []TypeCategorySub    `json:"typeCategories"`
	ProductCategorySubs []ProductCategorySub `json:"productCategories"`
	NotShowingReason    *string              `json:"notShowingReason"` // CONTENT_IMAGE_BANNED, CONTENT_IMAGE_CURRENTLY_NOT_PUBLISHED, CONTENT_IMAGE_DELETED
	SavedAt             time.Time            `json:"savedAt"`
	CreatedAt           time.Time            `json:"createdAt"`
	UpdatedAt           time.Time            `json:"updatedAt"`
}

type SavedCreatorImageActionResponse struct {
	ID             int64     `json:"id"`
	BusinessRootID int64     `json:"businessRootId"`
	CreatorImageID int64     `json:"creatorImageId"`
	CreatedAt      time.Time `json:"createdAt"`
}

// Sub-structs for response
type TypeCategorySub struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type ProductCategorySub struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type PublisherSub struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Image *string `json:"image"`
}
