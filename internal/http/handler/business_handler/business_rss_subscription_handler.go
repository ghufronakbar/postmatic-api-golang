// internal/http/handler/business_handler/business_rss_subscription_handler.go
package business_handler

import (
	"net/http"
	"postmatic-api/internal/http/middleware"
	"postmatic-api/internal/module/business/business_rss_subscription"

	"postmatic-api/pkg/response"
	"postmatic-api/pkg/utils"

	"github.com/go-chi/chi/v5"
)

type BusinessRssSubscriptionHandler struct {
	rssSvc     *business_rss_subscription.BusinessRssSubscriptionService
	middleware *middleware.OwnedBusiness
}

func NewBusinessRssSubscriptionHandler(busInSvc *business_rss_subscription.BusinessRssSubscriptionService, ownedMw *middleware.OwnedBusiness) *BusinessRssSubscriptionHandler {
	return &BusinessRssSubscriptionHandler{rssSvc: busInSvc, middleware: ownedMw}
}

func (h *BusinessRssSubscriptionHandler) BusinessRssSubscriptionRoutes() chi.Router {
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

func (h *BusinessRssSubscriptionHandler) GetAllRssSubscriptionBusinessRootId(w http.ResponseWriter, r *http.Request) {
	business, _ := middleware.OwnedBusinessFromContext(r.Context())

	filter := middleware.GetFilterFromContext(r.Context())

	filterQuery := business_rss_subscription.GetBusinessRssSubscriptionByBusinessRootIdFilter{
		Search:     filter.Search,
		SortBy:     filter.SortByDB(),
		PageOffset: filter.Offset(),
		PageLimit:  filter.Limit,
		SortDir:    filter.Sort,
		Page:       filter.Page,
	}

	res, pagination, err := h.rssSvc.GetBusinessRssSubscriptionByBusinessRootID(r.Context(), business.BusinessRootID, filterQuery)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.LIST(w, r, "SUCCESS_GET_BUSINESS_RSS_SUBSCRIPTION", res, &filter, &pagination)
}

func (h *BusinessRssSubscriptionHandler) CreateBusinessRssSubscriptionByBusinessRootID(w http.ResponseWriter, r *http.Request) {
	var req business_rss_subscription.CreateUpdateBusinessRSSSubscriptionInput

	business, _ := middleware.OwnedBusinessFromContext(r.Context())

	if appErr := utils.ValidateStruct(r.Body, &req); appErr != nil {
		response.ValidationFailed(w, r, appErr.ValidationErrors)
		return
	}

	res, err := h.rssSvc.CreateBusinessRssSubscription(r.Context(), business.BusinessRootID, req)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.OK(w, r, "SUCCESS_CREATE_BUSINESS_RSS_SUBSCRIPTION", res)
}

func (h *BusinessRssSubscriptionHandler) UpdateBusinessRssSubscriptionByBusinessRootID(w http.ResponseWriter, r *http.Request) {
	var req business_rss_subscription.CreateUpdateBusinessRSSSubscriptionInput

	if appErr := utils.ValidateStruct(r.Body, &req); appErr != nil {
		response.ValidationFailed(w, r, appErr.ValidationErrors)
		return
	}

	businessRssSubscriptionId := chi.URLParam(r, "businessRssSubscriptionId")
	business, _ := middleware.OwnedBusinessFromContext(r.Context())

	res, err := h.rssSvc.UpdateBusinessRssSubscription(r.Context(), business.BusinessRootID, businessRssSubscriptionId, req)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.OK(w, r, "SUCCESS_UPDATE_BUSINESS_RSS_SUBSCRIPTION", res)
}

func (h *BusinessRssSubscriptionHandler) HardDeleteBusinessRssSubscriptionByBusinessRootID(w http.ResponseWriter, r *http.Request) {
	businessRssSubscriptionId := chi.URLParam(r, "businessRssSubscriptionId")

	res, err := h.rssSvc.DeleteBusinessRssSubscription(r.Context(), businessRssSubscriptionId)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.OK(w, r, "SUCCESS_DELETE_BUSINESS_RSS_SUBSCRIPTION", res)
}
