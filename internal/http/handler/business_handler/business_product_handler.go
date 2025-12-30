// internal/http/handler/business_handler/business_product_handler.go
package business_handler

import (
	"net/http"
	"postmatic-api/internal/http/middleware"
	"postmatic-api/internal/module/business/business_product"
	"strconv"

	"postmatic-api/pkg/errs"
	"postmatic-api/pkg/response"
	"postmatic-api/pkg/utils"

	"github.com/go-chi/chi/v5"
)

type BusinessProductHandler struct {
	busInSvc   *business_product.BusinessProductService
	middleware *middleware.OwnedBusiness
}

func NewBusinessProductHandler(busInSvc *business_product.BusinessProductService, ownedMw *middleware.OwnedBusiness) *BusinessProductHandler {
	return &BusinessProductHandler{busInSvc: busInSvc, middleware: ownedMw}
}

func (h *BusinessProductHandler) BusinessProductRoutes() chi.Router {
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

func (h *BusinessProductHandler) GetProductsByBusinessID(w http.ResponseWriter, r *http.Request) {
	business, _ := middleware.OwnedBusinessFromContext(r.Context())

	filter := middleware.GetFilterFromContext(r.Context())

	filterQuery := business_product.GetBusinessProductsByBusinessRootIDFilter{
		Search:     filter.Search,
		SortBy:     filter.SortByDB(),
		PageOffset: filter.Offset(),
		PageLimit:  filter.Limit,
		SortDir:    filter.Sort,
		Page:       filter.Page,
		DateStart:  filter.DateStart,
		DateEnd:    filter.DateEnd,
		Category:   filter.Category,
	}

	res, pagination, err := h.busInSvc.GetBusinessProductsByBusinessRootID(r.Context(), business.BusinessRootID, filterQuery)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.LIST(w, r, "SUCCESS_GET_BUSINESS_PRODUCTS", res, &filter, &pagination)
}

func (h *BusinessProductHandler) CreateBusinessProductByBusinessRootID(w http.ResponseWriter, r *http.Request) {
	var req business_product.CreateUpdateBusinessProductInput

	business, _ := middleware.OwnedBusinessFromContext(r.Context())

	if appErr := utils.ValidateStruct(r.Body, &req); appErr != nil {
		response.ValidationFailed(w, r, appErr.ValidationErrors)
		return
	}

	res, err := h.busInSvc.CreateBusinessProduct(r.Context(), business.BusinessRootID, req)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.OK(w, r, "SUCCESS_CREATE_BUSINESS_PRODUCT", res)
}

func (h *BusinessProductHandler) UpdateBusinessProductByBusinessRootID(w http.ResponseWriter, r *http.Request) {
	var req business_product.CreateUpdateBusinessProductInput

	if appErr := utils.ValidateStruct(r.Body, &req); appErr != nil {
		response.ValidationFailed(w, r, appErr.ValidationErrors)
		return
	}

	businessProductId := chi.URLParam(r, "businessProductId")

	intBusinessProductId, err := strconv.ParseInt(businessProductId, 10, 64)
	if err != nil {
		response.Error(w, r, errs.NewValidationFailed(map[string]string{
			"businessProductId": "businessProductId must be an integer64",
		}), nil)
		return
	}

	res, err := h.busInSvc.UpdateBusinessProduct(r.Context(), intBusinessProductId, req)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.OK(w, r, "SUCCESS_UPDATE_BUSINESS_PRODUCT", res)
}

func (h *BusinessProductHandler) SoftDeleteBusinessProductByBusinessRootID(w http.ResponseWriter, r *http.Request) {
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
