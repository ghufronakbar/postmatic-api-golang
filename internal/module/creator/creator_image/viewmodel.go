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
