// internal/module/creator/creator_image/viewmodel.go
package creator_image

import "time"

type CreatorImageCreateUpdateDeleteResponse struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	ImageURL    string    `json:"imageUrl"`
	IsPublished bool      `json:"isPublished"`
	Price       int64     `json:"price"`
	ProfileId   *string   `json:"profileId"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// GET RESPONSE
type CreatorImageResponse struct {
	ID                  int64                `json:"id"`
	Name                string               `json:"name"`
	ImageURL            string               `json:"imageUrl"`
	IsPublished         bool                 `json:"isPublished"`
	Price               int64                `json:"price"`
	Publisher           *PublisherSub        `json:"publisher"`
	TypeCategorySubs    []TypeCategorySub    `json:"typeCategories"`
	ProductCategorySubs []ProductCategorySub `json:"productCategories"`
	CreatedAt           time.Time            `json:"createdAt"`
	UpdatedAt           time.Time            `json:"updatedAt"`
}

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
