// internal/module/creator/creator_image/handler/handler.go
package creator_image_handler

import (
	"fmt"
	"net/http"
	"postmatic-api/internal/internal_middleware"
	creator_image_service "postmatic-api/internal/module/creator/creator_image/service"
	"strconv"

	"postmatic-api/pkg/response"
	"postmatic-api/pkg/utils"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	creatorImageSvc *creator_image_service.CreatorImageService
}

func NewHandler(creatorImageSvc *creator_image_service.CreatorImageService) *Handler {
	return &Handler{creatorImageSvc: creatorImageSvc}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.GetCreatorImageByProfileId)
	r.Post("/", h.CreateCreatorImage)
	r.Put("/{creatorImageId}", h.UpdateCreatorImage)
	r.Delete("/{creatorImageId}", h.SoftDeleteCreatorImage)

	return r
}

func (h *Handler) GetCreatorImageByProfileId(w http.ResponseWriter, r *http.Request) {
	prof, _ := internal_middleware.GetProfileFromContext(r.Context())

	filter := internal_middleware.GetFilterFromContext(r.Context())

	q := r.URL.Query()
	typeCategoryIdQuery := q.Get("typeCategoryId")
	publishedQuery := q.Get("published")

	var typeCategoryId *int64
	var productCategoryId *int64
	var published *bool

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

	if publishedQuery != "" {
		published = utils.ParseBoolPtr(publishedQuery)
	}

	filterData := creator_image_service.GetCreatorImageFilter{
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
		Published:         published,
		ProfileID:         prof.ID.String(),
	}

	res, pag, err := h.creatorImageSvc.GetCreatorImageByProfileId(r.Context(), filterData)
	if err != nil {
		response.Error(w, r, err, res)
		return
	}

	response.LIST(w, r, "GET_CREATOR_IMAGE_SUCCESS", res, &filter, pag)
}

func (h *Handler) CreateCreatorImage(w http.ResponseWriter, r *http.Request) {
	prof, _ := internal_middleware.GetProfileFromContext(r.Context())
	var req creator_image_service.CreateCreatorImageInput
	req.ProfileID = prof.ID.String()

	if appErr := utils.ValidateStruct(r.Body, &req); appErr != nil {
		response.ValidationFailed(w, r, appErr.ValidationErrors)
		return
	}

	res, err := h.creatorImageSvc.CreateCreatorImage(r.Context(), req)

	if err != nil {
		fmt.Println(err)
		response.Error(w, r, err, res)
		return
	}

	response.OK(w, r, "CREATE_CREATOR_IMAGE_SUCCESS", res)
}

func (h *Handler) UpdateCreatorImage(w http.ResponseWriter, r *http.Request) {
	creatorImageId := chi.URLParam(r, "creatorImageId")
	intCreatorImageId, err := strconv.ParseInt(creatorImageId, 10, 64)
	if err != nil {
		response.ValidationFailed(w, r, map[string]string{"creatorImageId": "CREATOR_IMAGE_MUST_BE_INTEGER_64"})
		return
	}
	prof, _ := internal_middleware.GetProfileFromContext(r.Context())
	var req creator_image_service.UpdateCreatorImageInput
	req.ProfileID = prof.ID.String()
	req.CreatorImageId = intCreatorImageId

	if appErr := utils.ValidateStruct(r.Body, &req); appErr != nil {
		response.ValidationFailed(w, r, appErr.ValidationErrors)
		return
	}

	res, err := h.creatorImageSvc.UpdateCreatorImage(r.Context(), req)

	if err != nil {
		response.Error(w, r, err, res)
		return
	}

	response.OK(w, r, "UPDATE_CREATOR_IMAGE_SUCCESS", res)
}

func (h *Handler) SoftDeleteCreatorImage(w http.ResponseWriter, r *http.Request) {
	creatorImageId := chi.URLParam(r, "creatorImageId")
	intCreatorImageId, err := strconv.Atoi(creatorImageId)

	if err != nil {
		response.ValidationFailed(w, r, map[string]string{"creatorImageId": "CREATOR_IMAGE_MUST_BE_INTEGER"})
		return
	}

	prof, _ := internal_middleware.GetProfileFromContext(r.Context())

	res, err := h.creatorImageSvc.SoftDeleteCreatorImage(r.Context(), int64(intCreatorImageId), prof.ID.String())

	if err != nil {
		response.Error(w, r, err, res)
		return
	}

	response.OK(w, r, "DELETE_CREATOR_IMAGE_SUCCESS", res)
}
