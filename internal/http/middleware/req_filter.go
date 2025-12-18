// internal/http/middleware/req_filter.go
package middleware

import (
	"context"
	"net/http"
	"postmatic-api/pkg/filter"
	"slices"
	"strconv"
	"strings"
)

type contextReqFilterKey string

const ReqFilterContextKey contextReqFilterKey = "reqFilter"

var allowedSort = []string{"asc", "desc"}
var allowedSortByDefault = []string{"createdAt", "updatedAt", "name"}

func ReqFilterMiddleware(next http.Handler, allowedSortBy []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()

		search := strings.TrimSpace(q.Get("search"))
		sortBy := q.Get("sortBy")
		sort := q.Get("sort")
		category := q.Get("category")

		page, err := strconv.Atoi(q.Get("page"))
		if err != nil || page < 1 {
			page = 1
		}

		limit, err := strconv.Atoi(q.Get("limit"))
		if err != nil || limit < 1 {
			limit = 10
		}
		if limit > 50 { // clamp internal
			limit = 50
		}

		// whitelist sort
		if !slices.Contains(allowedSort, sort) {
			sort = "asc"
		}

		// whitelist sortBy
		allSortBy := append([]string{}, allowedSortByDefault...)
		allSortBy = append(allSortBy, allowedSortBy...)
		if !slices.Contains(allSortBy, sortBy) {
			sortBy = "createdAt"
		}

		ctx := context.WithValue(r.Context(), ReqFilterContextKey, filter.ReqFilter{
			Search:   search,
			Page:     page,
			Limit:    limit,
			Sort:     sort,
			SortBy:   sortBy,
			Category: category,
		})

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetFilterFromContext(ctx context.Context) filter.ReqFilter {
	fil, ok := ctx.Value(ReqFilterContextKey).(filter.ReqFilter)
	if !ok {
		return filter.ReqFilter{Search: "", SortBy: "createdAt", Page: 1, Limit: 10, Sort: "asc", Category: ""}
	}
	return fil
}
