// internal/http/handler/business_handler/business_rss_subscription_handler.go
package business_handler

import (
	"net/http"
	"postmatic-api/internal/http/middleware"
	"postmatic-api/internal/module/business/business_rss_subscription"
	"strconv"

	"postmatic-api/pkg/errs"
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

func (h *BusinessRssSubscriptionHandler) CreateBusinessRssSubscriptionByBusinessRootID(w http.ResponseWriter, r *http.Request) {
	var req business_rss_subscription.CreateBusinessRSSSubscriptionInput

	business, _ := middleware.OwnedBusinessFromContext(r.Context())

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

func (h *BusinessRssSubscriptionHandler) UpdateBusinessRssSubscriptionByBusinessRootID(w http.ResponseWriter, r *http.Request) {
	var req business_rss_subscription.UpdateBusinessRSSSubscriptionInput

	businessRssSubscriptionId := chi.URLParam(r, "businessRssSubscriptionId")

	intBusinessRssSubscriptionId, err := strconv.ParseInt(businessRssSubscriptionId, 10, 64)
	if err != nil {
		response.Error(w, r, errs.NewValidationFailed(map[string]string{
			"businessRssSubscriptionId": "businessRssSubscriptionId must be an integer64",
		}), nil)
		return
	}

	business, _ := middleware.OwnedBusinessFromContext(r.Context())
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

func (h *BusinessRssSubscriptionHandler) HardDeleteBusinessRssSubscriptionByBusinessRootID(w http.ResponseWriter, r *http.Request) {
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
