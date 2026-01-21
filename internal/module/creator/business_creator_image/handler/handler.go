// internal/module/creator/business_creator_image/handler/handler.go
package business_creator_image_handler

import (
	"net/http"
	"strconv"

	"postmatic-api/internal/internal_middleware"
	business_creator_image_service "postmatic-api/internal/module/creator/business_creator_image/service"
	"postmatic-api/pkg/response"
	"postmatic-api/pkg/utils"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	svc        *business_creator_image_service.BusinessCreatorImageService
	middleware *internal_middleware.OwnedBusiness
}

func NewHandler(svc *business_creator_image_service.BusinessCreatorImageService, ownedMw *internal_middleware.OwnedBusiness) *Handler {
	return &Handler{svc: svc, middleware: ownedMw}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	// owned business middleware
	r.Route("/{businessId}", func(r chi.Router) {
		r.Use(h.middleware.OwnedBusinessMiddleware)
		r.Use(func(next http.Handler) http.Handler {
			return internal_middleware.ReqFilterMiddleware(next, business_creator_image_service.SORT_BY)
		})
		r.Get("/", h.GetSavedCreatorImages)
		r.Post("/", h.CreateSavedCreatorImage)
		r.Delete("/{creatorImageId}", h.DeleteSavedCreatorImage)
	})

	return r
}

func (h *Handler) GetSavedCreatorImages(w http.ResponseWriter, r *http.Request) {
	bus, err := internal_middleware.OwnedBusinessFromContext(r.Context())
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	filter := internal_middleware.GetFilterFromContext(r.Context())

	q := r.URL.Query()
	typeCategoryIdQuery := q.Get("typeCategoryId")

	var typeCategoryId *int64
	var productCategoryId *int64

	if typeCategoryIdQuery != "" {
		typeCategoryIdInt, err := strconv.ParseInt(typeCategoryIdQuery, 10, 64)
		if err != nil {
			response.ValidationFailed(w, r, map[string]string{"typeCategoryId": "TYPE_CATEGORY_ID_MUST_BE_INTEGER"})
			return
		}
		typeCategoryId = &typeCategoryIdInt
	}

	if filter.Category != "" {
		productCategoryIdInt, err := strconv.ParseInt(filter.Category, 10, 64)
		if err != nil {
			response.ValidationFailed(w, r, map[string]string{"productCategoryId": "PRODUCT_CATEGORY_ID_MUST_BE_INTEGER"})
			return
		}
		productCategoryId = &productCategoryIdInt
	}

	filterData := business_creator_image_service.GetSavedCreatorImageFilter{
		BusinessRootID:    bus.BusinessRootID,
		Search:            filter.Search,
		SortBy:            filter.SortBy,
		PageOffset:        filter.Offset(),
		PageLimit:         filter.Limit,
		SortDir:           filter.Sort,
		Page:              filter.Page,
		DateStart:         filter.DateStart,
		DateEnd:           filter.DateEnd,
		TypeCategoryID:    typeCategoryId,
		ProductCategoryID: productCategoryId,
	}

	res, pag, err := h.svc.GetSavedCreatorImages(r.Context(), filterData)
	if err != nil {
		response.Error(w, r, err, res)
		return
	}

	response.LIST(w, r, "GET_SAVED_CREATOR_IMAGES_SUCCESS", res, &filter, pag)
}

func (h *Handler) CreateSavedCreatorImage(w http.ResponseWriter, r *http.Request) {
	bus, err := internal_middleware.OwnedBusinessFromContext(r.Context())
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	var req business_creator_image_service.CreateSavedInput
	req.BusinessRootID = bus.BusinessRootID

	if appErr := utils.ValidateStruct(r.Body, &req); appErr != nil {
		response.ValidationFailed(w, r, appErr.ValidationErrors)
		return
	}

	res, err := h.svc.CreateSavedCreatorImage(r.Context(), req)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.OK(w, r, "CREATE_SAVED_CREATOR_IMAGE_SUCCESS", res)
}

func (h *Handler) DeleteSavedCreatorImage(w http.ResponseWriter, r *http.Request) {
	bus, err := internal_middleware.OwnedBusinessFromContext(r.Context())
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	creatorImageIdStr := chi.URLParam(r, "creatorImageId")
	creatorImageId, err := strconv.ParseInt(creatorImageIdStr, 10, 64)
	if err != nil {
		response.ValidationFailed(w, r, map[string]string{"creatorImageId": "CREATOR_IMAGE_ID_MUST_BE_INTEGER"})
		return
	}

	input := business_creator_image_service.DeleteSavedInput{
		BusinessRootID: bus.BusinessRootID,
		CreatorImageID: creatorImageId,
	}

	res, err := h.svc.DeleteSavedCreatorImage(r.Context(), input)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.OK(w, r, "DELETE_SAVED_CREATOR_IMAGE_SUCCESS", res)
}
