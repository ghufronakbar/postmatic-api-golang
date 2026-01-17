// internal/module/business/business_product/handler/handler.go
package business_product_handler

import (
	"net/http"
	"postmatic-api/internal/internal_middleware"
	business_product_service "postmatic-api/internal/module/business/business_product/service"
	"strconv"

	"postmatic-api/pkg/errs"
	"postmatic-api/pkg/response"
	"postmatic-api/pkg/utils"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	busInSvc   *business_product_service.BusinessProductService
	middleware *internal_middleware.OwnedBusiness
}

func NewHandler(busInSvc *business_product_service.BusinessProductService, ownedMw *internal_middleware.OwnedBusiness) *Handler {
	return &Handler{busInSvc: busInSvc, middleware: ownedMw}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	// owned business middleware
	r.Route("/{businessId}", func(r chi.Router) {
		r.Use(h.middleware.OwnedBusinessMiddleware)
		r.Get("/", h.GetProductsByBusinessID)
		r.Post("/", h.CreateBusinessProductByBusinessRootID)
		r.Put("/{businessProductId}", h.UpdateBusinessProductByBusinessRootID)
		r.Delete("/{businessProductId}", h.SoftDeleteBusinessProductByBusinessRootID)
	})

	return r
}

func (h *Handler) GetProductsByBusinessID(w http.ResponseWriter, r *http.Request) {
	business, _ := internal_middleware.OwnedBusinessFromContext(r.Context())

	filter := internal_middleware.GetFilterFromContext(r.Context())

	filterQuery := business_product_service.GetBusinessProductsByBusinessRootIDFilter{
		Search:         filter.Search,
		SortBy:         filter.SortByDB(),
		PageOffset:     filter.Offset(),
		PageLimit:      filter.Limit,
		SortDir:        filter.Sort,
		Page:           filter.Page,
		DateStart:      filter.DateStart,
		DateEnd:        filter.DateEnd,
		Category:       filter.Category,
		BusinessRootID: business.BusinessRootID,
	}

	res, pagination, err := h.busInSvc.GetBusinessProductsByBusinessRootID(r.Context(), filterQuery)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.LIST(w, r, "SUCCESS_GET_BUSINESS_PRODUCTS", res, &filter, &pagination)
}

func (h *Handler) CreateBusinessProductByBusinessRootID(w http.ResponseWriter, r *http.Request) {
	var req business_product_service.CreateBusinessProductInput

	business, _ := internal_middleware.OwnedBusinessFromContext(r.Context())
	req.BusinessRootID = business.BusinessRootID

	if appErr := utils.ValidateStruct(r.Body, &req); appErr != nil {
		response.ValidationFailed(w, r, appErr.ValidationErrors)
		return
	}

	res, err := h.busInSvc.CreateBusinessProduct(r.Context(), req)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.OK(w, r, "SUCCESS_CREATE_BUSINESS_PRODUCT", res)
}

func (h *Handler) UpdateBusinessProductByBusinessRootID(w http.ResponseWriter, r *http.Request) {

	businessProductId := chi.URLParam(r, "businessProductId")

	intBusinessProductId, err := strconv.ParseInt(businessProductId, 10, 64)
	if err != nil {
		response.Error(w, r, errs.NewValidationFailed(map[string]string{
			"businessProductId": "businessProductId must be an integer64",
		}), nil)
		return
	}

	var req business_product_service.UpdateBusinessProductInput

	req.ID = intBusinessProductId
	if appErr := utils.ValidateStruct(r.Body, &req); appErr != nil {
		response.ValidationFailed(w, r, appErr.ValidationErrors)
		return
	}
	res, err := h.busInSvc.UpdateBusinessProduct(r.Context(), req)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.OK(w, r, "SUCCESS_UPDATE_BUSINESS_PRODUCT", res)
}

func (h *Handler) SoftDeleteBusinessProductByBusinessRootID(w http.ResponseWriter, r *http.Request) {
	businessProductId := chi.URLParam(r, "businessProductId")

	intBusinessProductId, err := strconv.ParseInt(businessProductId, 10, 64)
	if err != nil {
		response.Error(w, r, errs.NewValidationFailed(map[string]string{
			"businessProductId": "businessProductId must be an integer64",
		}), nil)
		return
	}

	res, err := h.busInSvc.SoftDeleteBusinessProductByBusinessRootID(r.Context(), intBusinessProductId)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.OK(w, r, "SUCCESS_DELETE_BUSINESS_PRODUCT", res)
}
