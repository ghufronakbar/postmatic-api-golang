// internal/http/handler/app_handler/timezone_handler.go
package app_handler

import (
	"net/http"
	"postmatic-api/internal/module/app/timezone"

	"postmatic-api/pkg/response"

	"github.com/go-chi/chi/v5"
)

type TimezoneHandler struct {
	tzSvc *timezone.TimezoneService
}

func NewTimezoneHandler(tzSvc *timezone.TimezoneService) *TimezoneHandler {
	return &TimezoneHandler{tzSvc: tzSvc}
}

func (h *TimezoneHandler) TimezoneRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.GetAllTimezone)

	return r
}

func (h *TimezoneHandler) GetAllTimezone(w http.ResponseWriter, r *http.Request) {

	res, err := h.tzSvc.GetAllTimezone(r.Context())

	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.LIST(w, r, "SUCCESS_GET_TIMEZONE", res, nil, nil)
}
