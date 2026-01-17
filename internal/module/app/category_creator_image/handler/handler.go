// internal/module/app/category_creator_image/handler/handler.go
package category_creator_image_handler

import (
	"net/http"
	"postmatic-api/internal/internal_middleware"
	category_creator_image_service "postmatic-api/internal/module/app/category_creator_image/service"

	"postmatic-api/pkg/response"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	cat *category_creator_image_service.CategoryCreatorImageService
}

func NewHandler(cat *category_creator_image_service.CategoryCreatorImageService) *Handler {
	return &Handler{cat: cat}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/type", h.GetCategoryCreatorImageType)
	r.Get("/product", h.GetCategoryCreatorImageProduct)

	return r
}

func (h *Handler) GetCategoryCreatorImageType(w http.ResponseWriter, r *http.Request) {

	filter := internal_middleware.GetFilterFromContext(r.Context())
	filterData := category_creator_image_service.GetCategoryCreatorImageTypeFilter{
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

func (h *Handler) GetCategoryCreatorImageProduct(w http.ResponseWriter, r *http.Request) {

	filter := internal_middleware.GetFilterFromContext(r.Context())
	filterData := category_creator_image_service.GetCategoryCreatorImageProductFilter{
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
