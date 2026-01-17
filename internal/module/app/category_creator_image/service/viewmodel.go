// internal/module/app/category_creator_image/viewmodel.go
package category_creator_image_service

type CategoryCreatorImageTypeResponse struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	TotalData int64  `json:"totalData"`
}

type CategoryCreatorImageProductResponse struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	TotalData int64  `json:"totalData"`
}
