// internal/http/handler/business_handler/business_information_handler.go
package business_handler

import (
	"errors"
	"net/http"
	"postmatic-api/internal/http/middleware"
	"postmatic-api/internal/module/business/business_information"

	"postmatic-api/pkg/response"
	"postmatic-api/pkg/utils"

	"github.com/go-chi/chi/v5"
)

type BusinessInformationHandler struct {
	busInSvc *business_information.BusinessInformationService
}

func NewBusinessInformationHandler(busInSvc *business_information.BusinessInformationService) *BusinessInformationHandler {
	return &BusinessInformationHandler{busInSvc: busInSvc}
}

func (h *BusinessInformationHandler) BusinessInformationRoutes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.GetJoinedBusinessesByProfileID)
	r.Post("/", h.SetupBusinessRootFirstTime)

	return r
}

func (h *BusinessInformationHandler) GetJoinedBusinessesByProfileID(w http.ResponseWriter, r *http.Request) {
	prof, err := middleware.GetUserFromContext(r.Context())
	if err != nil {
		response.Error(w, err, nil)
		return
	}
	if prof == nil {
		response.Error(w, errors.New("USER_NOT_FOUND"), nil)
		return
	}

	filter := middleware.GetFilterFromContext(r.Context())

	filterQuery := business_information.GetJoinedBusinessesByProfileIDFilter{
		Search:     filter.Search,
		SortBy:     filter.SortByDB(),
		PageOffset: filter.Offset(),
		PageLimit:  filter.Limit,
		SortDir:    filter.Sort,
		Page:       filter.Page,
	}

	res, pagination, err := h.busInSvc.GetJoinedBusinessesByProfileID(r.Context(), prof.ID, filterQuery)
	if err != nil {
		response.Error(w, err, nil)
		return
	}

	response.LIST(w, "SUCCESS_GET_JOINED_BUSINESS", res, &filter, pagination)
}

func (h *BusinessInformationHandler) SetupBusinessRootFirstTime(w http.ResponseWriter, r *http.Request) {
	var req business_information.BusinessSetupInput

	if appErr := utils.ValidateStruct(r.Body, &req); appErr != nil {
		response.ValidationFailed(w, appErr.ValidationErrors)
		return
	}

	// 1) Ambil user dari context
	prof, err := middleware.GetUserFromContext(r.Context())
	if err != nil {
		response.Error(w, err, nil)
		return
	}
	if prof == nil {
		response.Error(w, errors.New("USER_NOT_FOUND"), nil)
		return
	}

	//  Jalankan service
	res, err := h.busInSvc.SetupBusinessRootFirstTime(r.Context(), prof.ID, req)
	if err != nil {
		response.Error(w, err, res)
		return
	}

	response.OK(w, "SUCCESS_CREATE_BUSINESS", res)
}
