// internal/http/handler/app_handler/category_creator_image_handler.go
package app_handler

import (
	"net/http"
	"postmatic-api/internal/http/middleware"
	"postmatic-api/internal/module/app/category_creator_image"

	"postmatic-api/pkg/response"

	"github.com/go-chi/chi/v5"
)

type CategoryCreatorImageHandler struct {
	cat *category_creator_image.CategoryCreatorImageService
}

func NewCategoryCreatorImageHandler(cat *category_creator_image.CategoryCreatorImageService) *CategoryCreatorImageHandler {
	return &CategoryCreatorImageHandler{cat: cat}
}

func (h *CategoryCreatorImageHandler) CategoryCreatorImageRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/type", h.GetCategoryCreatorImageType)
	r.Get("/product", h.GetCategoryCreatorImageProduct)

	return r
}

func (h *CategoryCreatorImageHandler) GetCategoryCreatorImageType(w http.ResponseWriter, r *http.Request) {

	filter := middleware.GetFilterFromContext(r.Context())
	filterData := category_creator_image.GetCategoryCreatorImageTypeFilter{
		Search:     filter.Search,
		SortBy:     filter.SortByDB(),
		PageOffset: filter.Offset(),
		PageLimit:  filter.Limit,
		SortDir:    filter.Sort,
		Page:       filter.Page,
	}
	res, pagination, err := h.cat.GetCategoryCreatorImageType(r.Context(), filterData)

	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.LIST(w, r, "SUCCESS_GET_CATEGORY_CREATOR_IMAGE_TYPE", res, &filter, pagination)
}

func (h *CategoryCreatorImageHandler) GetCategoryCreatorImageProduct(w http.ResponseWriter, r *http.Request) {

	filter := middleware.GetFilterFromContext(r.Context())
	filterData := category_creator_image.GetCategoryCreatorImageProductFilter{
		Search:     filter.Search,
		SortBy:     filter.SortByDB(),
		PageOffset: filter.Offset(),
		PageLimit:  filter.Limit,
		SortDir:    filter.Sort,
		Page:       filter.Page,
		Locale:     filter.Category,
	}
	res, pagination, err := h.cat.GetCategoryCreatorImageProduct(r.Context(), filterData)

	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.LIST(w, r, "SUCCESS_GET_CATEGORY_CREATOR_IMAGE_PRODUCT", res, &filter, pagination)
}
