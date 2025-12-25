// internal/http/handler/app_handler/rss_handler.go
package app_handler

import (
	"net/http"
	"postmatic-api/internal/http/middleware"
	"postmatic-api/internal/module/app/rss"

	"postmatic-api/pkg/response"

	"github.com/go-chi/chi/v5"
)

type RSSHandler struct {
	rssSvc *rss.RSSService
}

func NewRSSHandler(rssSvc *rss.RSSService) *RSSHandler {
	return &RSSHandler{rssSvc: rssSvc}
}

func (h *RSSHandler) RSSRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.GetRSSFeed)
	r.Get("/category", h.GetRSSCategory)

	return r
}

func (h *RSSHandler) GetRSSFeed(w http.ResponseWriter, r *http.Request) {

	filter := middleware.GetFilterFromContext(r.Context())

	filterQuery := rss.GetRSSFeedFilter{
		Search:     filter.Search,
		SortBy:     filter.SortByDB(),
		PageOffset: filter.Offset(),
		PageLimit:  filter.Limit,
		SortDir:    filter.Sort,
		Page:       filter.Page,
		Category:   filter.Category,
	}

	res, pagination, err := h.rssSvc.GetRSSFeed(r.Context(), filterQuery)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.LIST(w, r, "SUCCESS_GET_RSS_FEED", res, &filter, pagination)
}

func (h *RSSHandler) GetRSSCategory(w http.ResponseWriter, r *http.Request) {

	filter := middleware.GetFilterFromContext(r.Context())

	filterQuery := rss.GetRSSCategoryFilter{
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
