// internal/module/app/rss/handler/handler.go
package rss_handler

import (
	"net/http"
	"postmatic-api/internal/internal_middleware"
	rss_service "postmatic-api/internal/module/app/rss/service"
	"strconv"

	"postmatic-api/pkg/errs"
	"postmatic-api/pkg/response"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	rssSvc *rss_service.RSSService
}

func NewHandler(rssSvc *rss_service.RSSService) *Handler {
	return &Handler{rssSvc: rssSvc}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.GetRSSFeed)
	r.Get("/category", h.GetRSSCategory)

	return r
}

func (h *Handler) GetRSSFeed(w http.ResponseWriter, r *http.Request) {

	filter := internal_middleware.GetFilterFromContext(r.Context())

	filterQuery := rss_service.GetRSSFeedFilter{
		Search:     filter.Search,
		SortBy:     filter.SortByDB(),
		PageOffset: filter.Offset(),
		PageLimit:  filter.Limit,
		SortDir:    filter.Sort,
		Page:       filter.Page,
	}

	if filter.Category != "" {
		intCatId, err := strconv.ParseInt(filter.Category, 10, 64)
		if err != nil {
			response.Error(w, r, errs.NewValidationFailed(map[string]string{
				"category": "category must be an integer64 or optional",
			}), nil)
			return
		}
		filterQuery.Category = intCatId
	}

	res, pagination, err := h.rssSvc.GetRSSFeed(r.Context(), filterQuery)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.LIST(w, r, "SUCCESS_GET_RSS_FEED", res, &filter, pagination)
}

func (h *Handler) GetRSSCategory(w http.ResponseWriter, r *http.Request) {

	filter := internal_middleware.GetFilterFromContext(r.Context())

	filterQuery := rss_service.GetRSSCategoryFilter{
		Search:     filter.Search,
		SortBy:     filter.SortByDB(),
		PageOffset: filter.Offset(),
		PageLimit:  filter.Limit,
		SortDir:    filter.Sort,
		Page:       filter.Page,
	}

	res, pagination, err := h.rssSvc.GetRSSCategory(r.Context(), filterQuery)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.LIST(w, r, "SUCCESS_GET_RSS_CATEGORY", res, &filter, pagination)
}
