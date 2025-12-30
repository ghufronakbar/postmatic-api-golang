// internal/http/handler/business_handler/business_image_content_handler.go
package business_handler

import (
	"net/http"
	"postmatic-api/internal/http/middleware"
	"postmatic-api/internal/module/business/business_image_content"
	"strconv"

	"postmatic-api/pkg/errs"
	"postmatic-api/pkg/response"
	"postmatic-api/pkg/utils"

	"github.com/go-chi/chi/v5"
)

type BusinessImageContentHandler struct {
	busInSvc   *business_image_content.BusinessImageContentService
	middleware *middleware.OwnedBusiness
}

func NewBusinessImageContentHandler(busInSvc *business_image_content.BusinessImageContentService, ownedMw *middleware.OwnedBusiness) *BusinessImageContentHandler {
	return &BusinessImageContentHandler{busInSvc: busInSvc, middleware: ownedMw}
}

func (h *BusinessImageContentHandler) BusinessImageContentRoutes() chi.Router {
	r := chi.NewRouter()

	// owned business middleware
	r.Route("/{businessId}", func(r chi.Router) {
		r.Use(h.middleware.OwnedBusinessMiddleware)
		r.Get("/", h.GetBusinessImageContentsByBusinessRootID)
		r.Post("/", h.CreateBusinessImageContent)
		r.Put("/{businessImageContentId}", h.UpdateBusinessImageContent)
		r.Delete("/{businessImageContentId}", h.DeleteBusinessImageContent)
	})

	return r
}

func (h *BusinessImageContentHandler) GetBusinessImageContentsByBusinessRootID(w http.ResponseWriter, r *http.Request) {
	business, _ := middleware.OwnedBusinessFromContext(r.Context())

	filter := middleware.GetFilterFromContext(r.Context())

	var readyToPost *bool
	if filter.Category != "" {
		if filter.Category == "readyToPosts" {
			readyToPost = utils.ParseBoolPtr("true")
		} else {
			readyToPost = utils.ParseBoolPtr("false")
		}
	}

	filterQuery := business_image_content.GetBusinessImageContentsByBusinessRootIDFilter{
		Search:      filter.Search,
		SortBy:      filter.SortByDB(),
		PageOffset:  filter.Offset(),
		PageLimit:   filter.Limit,
		SortDir:     filter.Sort,
		Page:        filter.Page,
		DateStart:   filter.DateStart,
		DateEnd:     filter.DateEnd,
		ReadyToPost: readyToPost,
	}

	filterQuery.BusinessRootID = business.BusinessRootID
	res, pagination, err := h.busInSvc.GetBusinessImageContentsByBusinessRootID(r.Context(), filterQuery)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.LIST(w, r, "SUCCESS_GET_BUSINESS_IMAGE_CONTENTS", res, &filter, &pagination)
}

func (h *BusinessImageContentHandler) CreateBusinessImageContent(w http.ResponseWriter, r *http.Request) {
	business, _ := middleware.OwnedBusinessFromContext(r.Context())

	var req business_image_content.CreateUpdateBusinessImageContentInput

	req.Type = "personal"
	req.BusinessRootID = business.BusinessRootID
	if appErr := utils.ValidateStruct(r.Body, &req); appErr != nil {
		response.ValidationFailed(w, r, appErr.ValidationErrors)
		return
	}

	res, err := h.busInSvc.CreateBusinessImageContent(r.Context(), req)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.OK(w, r, "SUCCESS_CREATE_BUSINESS_IMAGE_CONTENT", res)
}

func (h *BusinessImageContentHandler) UpdateBusinessImageContent(w http.ResponseWriter, r *http.Request) {
	business, _ := middleware.OwnedBusinessFromContext(r.Context())
	id := chi.URLParam(r, "businessImageContentId")
	intId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		response.Error(w, r, errs.NewValidationFailed(map[string]string{"businessImageContentId": "must be int64"}), nil)
		return
	}

	var req business_image_content.CreateUpdateBusinessImageContentInput
	req.BusinessRootID = business.BusinessRootID
	if appErr := utils.ValidateStruct(r.Body, &req); appErr != nil {
		response.ValidationFailed(w, r, appErr.ValidationErrors)
		return
	}

	res, err := h.busInSvc.UpdateBusinessImageContent(r.Context(), req, intId)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.OK(w, r, "SUCCESS_UPDATE_BUSINESS_IMAGE_CONTENT", res)
}

func (h *BusinessImageContentHandler) DeleteBusinessImageContent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "businessImageContentId")
	intId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		response.Error(w, r, errs.NewValidationFailed(map[string]string{"businessImageContentId": "must be int64"}), nil)
		return
	}

	res, err := h.busInSvc.DeleteBusinessImageContent(r.Context(), intId)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.OK(w, r, "SUCCESS_DELETE_BUSINESS_IMAGE_CONTENT", res)
}
