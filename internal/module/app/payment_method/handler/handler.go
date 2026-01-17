package payment_method_handler

import (
	"net/http"
	"postmatic-api/internal/internal_middleware"
	payment_method_service "postmatic-api/internal/module/app/payment_method/service"
	"postmatic-api/internal/repository/entity"
	"strconv"

	"postmatic-api/pkg/response"
	"postmatic-api/pkg/utils"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	svc *payment_method_service.PaymentMethodService
}

func NewHandler(svc *payment_method_service.PaymentMethodService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Routes(allAllowed, adminOnly func(http.Handler) http.Handler) chi.Router {
	r := chi.NewRouter()

	// Route without filter middleware
	r.Group(func(r chi.Router) {
		r.Use(allAllowed)
		r.Get("/type", h.GetPaymentMethodTypes)
	})

	// All allowed routes with filter
	r.Group(func(r chi.Router) {
		r.Use(allAllowed)
		r.Use(func(next http.Handler) http.Handler {
			return internal_middleware.ReqFilterMiddleware(next, payment_method_service.SORT_BY)
		})
		r.Get("/", h.GetAllPaymentMethods)
		r.Get("/{id}", h.GetPaymentMethodById)
		r.Get("/{code}/code", h.GetPaymentMethodByCode)
	})

	// Admin only routes
	r.Group(func(r chi.Router) {
		r.Use(adminOnly)
		r.Post("/", h.CreatePaymentMethod)
		r.Put("/{id}", h.UpdatePaymentMethod)
		r.Delete("/{id}", h.DeletePaymentMethod)
	})

	return r
}

func (h *Handler) GetAllPaymentMethods(w http.ResponseWriter, r *http.Request) {
	filter := internal_middleware.GetFilterFromContext(r.Context())
	prof, _ := internal_middleware.GetProfileFromContext(r.Context())

	isAdmin := prof.Role == entity.AppRoleAdmin

	filterData := payment_method_service.GetPaymentMethodsFilter{
		Search:     filter.Search,
		SortBy:     filter.SortByDB(),
		SortDir:    filter.Sort,
		PageOffset: filter.Offset(),
		PageLimit:  filter.Limit,
		Page:       filter.Page,
		IsAdmin:    isAdmin,
	}

	res, pag, err := h.svc.GetAllPaymentMethods(r.Context(), filterData)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.LIST(w, r, "GET_PAYMENT_METHODS_SUCCESS", res, &filter, pag)
}

func (h *Handler) GetPaymentMethodById(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		response.ValidationFailed(w, r, map[string]string{"id": "ID_MUST_BE_INTEGER"})
		return
	}

	prof, _ := internal_middleware.GetProfileFromContext(r.Context())
	isAdmin := prof.Role == entity.AppRoleAdmin

	res, err := h.svc.GetPaymentMethodById(r.Context(), id, isAdmin)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.OK(w, r, "GET_PAYMENT_METHOD_SUCCESS", res)
}

func (h *Handler) GetPaymentMethodByCode(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	if code == "" {
		response.ValidationFailed(w, r, map[string]string{"code": "CODE_REQUIRED"})
		return
	}

	prof, _ := internal_middleware.GetProfileFromContext(r.Context())
	isAdmin := prof.Role == entity.AppRoleAdmin

	res, err := h.svc.GetPaymentMethodByCode(r.Context(), code, isAdmin)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.OK(w, r, "GET_PAYMENT_METHOD_SUCCESS", res)
}

func (h *Handler) CreatePaymentMethod(w http.ResponseWriter, r *http.Request) {
	var req payment_method_service.CreatePaymentMethodInput

	if appErr := utils.ValidateStruct(r.Body, &req); appErr != nil {
		response.ValidationFailed(w, r, appErr.ValidationErrors)
		return
	}

	prof, _ := internal_middleware.GetProfileFromContext(r.Context())
	req.ProfileID = prof.ID.String()

	res, err := h.svc.CreatePaymentMethod(r.Context(), req)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.OK(w, r, "CREATE_PAYMENT_METHOD_SUCCESS", res)
}

func (h *Handler) UpdatePaymentMethod(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		response.ValidationFailed(w, r, map[string]string{"id": "ID_MUST_BE_INTEGER"})
		return
	}

	var req payment_method_service.UpdatePaymentMethodInput

	if appErr := utils.ValidateStruct(r.Body, &req); appErr != nil {
		response.ValidationFailed(w, r, appErr.ValidationErrors)
		return
	}

	prof, _ := internal_middleware.GetProfileFromContext(r.Context())
	req.ID = id
	req.ProfileID = prof.ID.String()

	res, err := h.svc.UpdatePaymentMethod(r.Context(), req)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.OK(w, r, "UPDATE_PAYMENT_METHOD_SUCCESS", res)
}

func (h *Handler) DeletePaymentMethod(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		response.ValidationFailed(w, r, map[string]string{"id": "ID_MUST_BE_INTEGER"})
		return
	}

	prof, _ := internal_middleware.GetProfileFromContext(r.Context())

	res, err := h.svc.DeletePaymentMethod(r.Context(), id, prof.ID.String())
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.OK(w, r, "DELETE_PAYMENT_METHOD_SUCCESS", res)
}

func (h *Handler) GetPaymentMethodTypes(w http.ResponseWriter, r *http.Request) {
	types := []string{
		string(entity.AppPaymentMethodTypeBank),
		string(entity.AppPaymentMethodTypeEwallet),
	}
	response.OK(w, r, "GET_PAYMENT_METHOD_TYPES_SUCCESS", types)
}
