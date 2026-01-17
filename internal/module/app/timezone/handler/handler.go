// internal/module/app/timezone/handler/handler.go
package timezone_handler

import (
	"net/http"
	timezone_service "postmatic-api/internal/module/app/timezone/service"

	"postmatic-api/pkg/response"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	tzSvc *timezone_service.TimezoneService
}

func NewHandler(tzSvc *timezone_service.TimezoneService) *Handler {
	return &Handler{tzSvc: tzSvc}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.GetAllTimezone)

	return r
}

func (h *Handler) GetAllTimezone(w http.ResponseWriter, r *http.Request) {

	res, err := h.tzSvc.GetAllTimezone(r.Context())

	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.LIST(w, r, "SUCCESS_GET_TIMEZONE", res, nil, nil)
}
