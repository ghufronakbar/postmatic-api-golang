// internal/module/business/business_rss_subscription/handler/handler.go
package business_rss_subscription_handler

import (
	"net/http"
	"postmatic-api/internal/internal_middleware"
	business_rss_subscription_service "postmatic-api/internal/module/business/business_rss_subscription/service"
	"strconv"

	"postmatic-api/pkg/errs"
	"postmatic-api/pkg/response"
	"postmatic-api/pkg/utils"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	rssSvc     *business_rss_subscription_service.BusinessRssSubscriptionService
	middleware *internal_middleware.OwnedBusiness
}

func NewHandler(busInSvc *business_rss_subscription_service.BusinessRssSubscriptionService, ownedMw *internal_middleware.OwnedBusiness) *Handler {
	return &Handler{rssSvc: busInSvc, middleware: ownedMw}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	// owned business middleware
	r.Route("/{businessId}", func(r chi.Router) {
		r.Use(h.middleware.OwnedBusinessMiddleware)
		r.Get("/", h.GetAllRssSubscriptionBusinessRootId)
		r.Post("/", h.CreateBusinessRssSubscriptionByBusinessRootID)
		r.Put("/{businessRssSubscriptionId}", h.UpdateBusinessRssSubscriptionByBusinessRootID)
		r.Delete("/{businessRssSubscriptionId}", h.HardDeleteBusinessRssSubscriptionByBusinessRootID)
	})

	return r
}

func (h *Handler) GetAllRssSubscriptionBusinessRootId(w http.ResponseWriter, r *http.Request) {
	business, _ := internal_middleware.OwnedBusinessFromContext(r.Context())

	filter := internal_middleware.GetFilterFromContext(r.Context())

	filterQuery := business_rss_subscription_service.GetBusinessRssSubscriptionByBusinessRootIdFilter{
		Search:         filter.Search,
		SortBy:         filter.SortByDB(),
		PageOffset:     filter.Offset(),
		PageLimit:      filter.Limit,
		SortDir:        filter.Sort,
		Page:           filter.Page,
		BusinessRootID: business.BusinessRootID,
	}

	res, pagination, err := h.rssSvc.GetBusinessRssSubscriptionByBusinessRootID(r.Context(), filterQuery)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.LIST(w, r, "SUCCESS_GET_BUSINESS_RSS_SUBSCRIPTION", res, &filter, &pagination)
}

func (h *Handler) CreateBusinessRssSubscriptionByBusinessRootID(w http.ResponseWriter, r *http.Request) {
	var req business_rss_subscription_service.CreateBusinessRSSSubscriptionInput

	business, _ := internal_middleware.OwnedBusinessFromContext(r.Context())

	req.BusinessRootID = business.BusinessRootID

	if appErr := utils.ValidateStruct(r.Body, &req); appErr != nil {
		response.ValidationFailed(w, r, appErr.ValidationErrors)
		return
	}

	res, err := h.rssSvc.CreateBusinessRssSubscription(r.Context(), req)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.OK(w, r, "SUCCESS_CREATE_BUSINESS_RSS_SUBSCRIPTION", res)
}

func (h *Handler) UpdateBusinessRssSubscriptionByBusinessRootID(w http.ResponseWriter, r *http.Request) {
	var req business_rss_subscription_service.UpdateBusinessRSSSubscriptionInput

	businessRssSubscriptionId := chi.URLParam(r, "businessRssSubscriptionId")

	intBusinessRssSubscriptionId, err := strconv.ParseInt(businessRssSubscriptionId, 10, 64)
	if err != nil {
		response.Error(w, r, errs.NewValidationFailed(map[string]string{
			"businessRssSubscriptionId": "businessRssSubscriptionId must be an integer64",
		}), nil)
		return
	}

	business, _ := internal_middleware.OwnedBusinessFromContext(r.Context())
	req.BusinessRootID = business.BusinessRootID
	req.ID = intBusinessRssSubscriptionId

	if appErr := utils.ValidateStruct(r.Body, &req); appErr != nil {
		response.ValidationFailed(w, r, appErr.ValidationErrors)
		return
	}

	res, err := h.rssSvc.UpdateBusinessRssSubscription(r.Context(), req)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.OK(w, r, "SUCCESS_UPDATE_BUSINESS_RSS_SUBSCRIPTION", res)
}

func (h *Handler) HardDeleteBusinessRssSubscriptionByBusinessRootID(w http.ResponseWriter, r *http.Request) {
	businessRssSubscriptionId := chi.URLParam(r, "businessRssSubscriptionId")

	intBusinessRssSubscriptionId, err := strconv.ParseInt(businessRssSubscriptionId, 10, 64)
	if err != nil {
		response.Error(w, r, errs.NewValidationFailed(map[string]string{
			"businessRssSubscriptionId": "businessRssSubscriptionId must be an integer64",
		}), nil)
		return
	}

	res, err := h.rssSvc.DeleteBusinessRssSubscription(r.Context(), intBusinessRssSubscriptionId)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.OK(w, r, "SUCCESS_DELETE_BUSINESS_RSS_SUBSCRIPTION", res)
}
