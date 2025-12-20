// internal/http/handler/business_handler/business_information_handler.go
package business_handler

import (
	"fmt"
	"net/http"
	"postmatic-api/internal/http/middleware"
	"postmatic-api/internal/module/business/business_information"

	"postmatic-api/pkg/response"
	"postmatic-api/pkg/utils"

	"github.com/go-chi/chi/v5"
)

type BusinessInformationHandler struct {
	busInSvc   *business_information.BusinessInformationService
	middleware *middleware.OwnedBusiness
}

func NewBusinessInformationHandler(busInSvc *business_information.BusinessInformationService, ownedMw *middleware.OwnedBusiness) *BusinessInformationHandler {
	return &BusinessInformationHandler{busInSvc: busInSvc, middleware: ownedMw}
}

func (h *BusinessInformationHandler) BusinessInformationRoutes() chi.Router {
	r := chi.NewRouter()

	// Route yang tidak butuh businessId
	r.Get("/", h.GetJoinedBusinessesByProfileID)
	r.Post("/", h.SetupBusinessRootFirstTime)

	// Route yang butuh businessId -> middleware dipasang DI SINI
	r.Route("/{businessId}", func(r chi.Router) {
		r.Use(h.middleware.OwnedBusinessMiddleware)

		r.Get("/", h.GetBusinessById)
		r.Delete("/", h.DeleteBusinessById)
	})

	return r
}

func (h *BusinessInformationHandler) GetJoinedBusinessesByProfileID(w http.ResponseWriter, r *http.Request) {
	prof, _ := middleware.GetUserFromContext(r.Context())

	filter := middleware.GetFilterFromContext(r.Context())

	filterQuery := business_information.GetJoinedBusinessesByProfileIDFilter{
		Search:     filter.Search,
		SortBy:     filter.SortByDB(),
		PageOffset: filter.Offset(),
		PageLimit:  filter.Limit,
		SortDir:    filter.Sort,
		Page:       filter.Page,
		DateStart:  filter.DateStart,
		DateEnd:    filter.DateEnd,
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
	prof, _ := middleware.GetUserFromContext(r.Context())

	//  Jalankan service
	res, err := h.busInSvc.SetupBusinessRootFirstTime(r.Context(), prof.ID, req)
	if err != nil {
		response.Error(w, err, res)
		return
	}

	response.OK(w, "SUCCESS_CREATE_BUSINESS", res)
}

func (h *BusinessInformationHandler) GetBusinessById(w http.ResponseWriter, r *http.Request) {
	businessId := chi.URLParam(r, "businessId")

	prof, _ := middleware.GetUserFromContext(r.Context())

	res, err := h.busInSvc.GetBusinessById(r.Context(), businessId, prof.ID)
	if err != nil {
		response.Error(w, err, nil)
		return
	}

	response.OK(w, "SUCCESS_GET_BUSINESS", res)
}

func (h *BusinessInformationHandler) DeleteBusinessById(w http.ResponseWriter, r *http.Request) {
	businessId := chi.URLParam(r, "businessId")

	prof, _ := middleware.GetUserFromContext(r.Context())

	res, err := h.busInSvc.DeleteBusinessById(r.Context(), businessId, prof.ID)
	if err != nil {
		fmt.Println(err)
		response.Error(w, err, nil)
		return
	}

	response.OK(w, "DELETE_BUSINESS_SUCCESS", res)
}
