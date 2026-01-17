// internal/module/business/business_information/handler/handler.go
package business_information_handler

import (
	"fmt"
	"net/http"
	"postmatic-api/internal/internal_middleware"
	business_information_service "postmatic-api/internal/module/business/business_information/service"

	"postmatic-api/pkg/response"
	"postmatic-api/pkg/utils"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	busInSvc   *business_information_service.BusinessInformationService
	middleware *internal_middleware.OwnedBusiness
}

func NewHandler(busInSvc *business_information_service.BusinessInformationService, ownedMw *internal_middleware.OwnedBusiness) *Handler {
	return &Handler{busInSvc: busInSvc, middleware: ownedMw}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	// owned business middleware
	r.Route("/{businessId}", func(r chi.Router) {
		r.Use(h.middleware.OwnedBusinessMiddleware)
		r.Get("/", h.GetBusinessById)
		r.Delete("/", h.DeleteBusinessById)
	})
	r.Route("/", func(r chi.Router) {
		r.Get("/", h.GetJoinedBusinessesByProfileID)
		r.Post("/", h.SetupBusinessRootFirstTime)
	})

	return r
}

func (h *Handler) GetJoinedBusinessesByProfileID(w http.ResponseWriter, r *http.Request) {
	prof, _ := internal_middleware.GetProfileFromContext(r.Context())

	filter := internal_middleware.GetFilterFromContext(r.Context())

	filterQuery := business_information_service.GetJoinedBusinessesByProfileIDFilter{
		Search:     filter.Search,
		SortBy:     filter.SortByDB(),
		PageOffset: filter.Offset(),
		PageLimit:  filter.Limit,
		SortDir:    filter.Sort,
		Page:       filter.Page,
		DateStart:  filter.DateStart,
		DateEnd:    filter.DateEnd,
		ProfileID:  prof.ID,
	}

	res, pagination, err := h.busInSvc.GetJoinedBusinessesByProfileID(r.Context(), filterQuery)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.LIST(w, r, "SUCCESS_GET_JOINED_BUSINESS", res, &filter, pagination)
}

func (h *Handler) SetupBusinessRootFirstTime(w http.ResponseWriter, r *http.Request) {
	var req business_information_service.BusinessSetupInput
	prof, _ := internal_middleware.GetProfileFromContext(r.Context())

	req.ProfileID = prof.ID

	if appErr := utils.ValidateStruct(r.Body, &req); appErr != nil {
		response.ValidationFailed(w, r, appErr.ValidationErrors)
		return
	}

	//  Jalankan service
	res, err := h.busInSvc.SetupBusinessRootFirstTime(r.Context(), req)
	if err != nil {
		response.Error(w, r, err, res)
		return
	}

	response.OK(w, r, "SUCCESS_CREATE_BUSINESS", res)
}

func (h *Handler) GetBusinessById(w http.ResponseWriter, r *http.Request) {
	business, _ := internal_middleware.OwnedBusinessFromContext(r.Context())

	prof, _ := internal_middleware.GetProfileFromContext(r.Context())

	res, err := h.busInSvc.GetBusinessById(r.Context(), business.BusinessRootID, prof.ID)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.OK(w, r, "SUCCESS_GET_BUSINESS", res)
}

func (h *Handler) DeleteBusinessById(w http.ResponseWriter, r *http.Request) {
	business, _ := internal_middleware.OwnedBusinessFromContext(r.Context())

	prof, _ := internal_middleware.GetProfileFromContext(r.Context())

	res, err := h.busInSvc.DeleteBusinessById(r.Context(), business.BusinessRootID, prof.ID)
	if err != nil {
		fmt.Println(err)
		response.Error(w, r, err, nil)
		return
	}

	response.OK(w, r, "DELETE_BUSINESS_SUCCESS", res)
}
