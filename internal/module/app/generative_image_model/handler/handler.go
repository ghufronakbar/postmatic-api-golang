package generative_image_model_handler

import (
	"net/http"
	"postmatic-api/internal/internal_middleware"
	generative_image_model_service "postmatic-api/internal/module/app/generative_image_model/service"
	"postmatic-api/internal/repository/entity"
	"strconv"

	"postmatic-api/pkg/response"
	"postmatic-api/pkg/utils"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	svc *generative_image_model_service.GenerativeImageModelService
}

func NewHandler(svc *generative_image_model_service.GenerativeImageModelService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Routes(allAllowed, adminOnly func(http.Handler) http.Handler) chi.Router {
	r := chi.NewRouter()

	// Route without filter middleware
	r.Group(func(r chi.Router) {
		r.Use(allAllowed)
		r.Get("/provider", h.GetProviderTypes)
	})

	// All allowed routes with filter
	r.Group(func(r chi.Router) {
		r.Use(allAllowed)
		r.Use(func(next http.Handler) http.Handler {
			return internal_middleware.ReqFilterMiddleware(next, generative_image_model_service.SORT_BY)
		})
		r.Get("/", h.GetAllGenerativeImageModels)
		r.Get("/{id}", h.GetGenerativeImageModelById)
		r.Get("/{model}/model", h.GetGenerativeImageModelByModel)
	})

	// Admin only routes
	r.Group(func(r chi.Router) {
		r.Use(adminOnly)
		r.Post("/", h.CreateGenerativeImageModel)
		r.Put("/{id}", h.UpdateGenerativeImageModel)
		r.Delete("/{id}", h.DeleteGenerativeImageModel)
	})

	return r
}

func (h *Handler) GetProviderTypes(w http.ResponseWriter, r *http.Request) {
	types := []string{
		string(entity.AppGenerativeImageModelProviderTypeOpenai),
		string(entity.AppGenerativeImageModelProviderTypeGoogle),
	}
	response.OK(w, r, "GET_GENERATIVE_IMAGE_MODEL_PROVIDERS_SUCCESS", types)
}

func (h *Handler) GetAllGenerativeImageModels(w http.ResponseWriter, r *http.Request) {
	filter := internal_middleware.GetFilterFromContext(r.Context())
	prof, _ := internal_middleware.GetProfileFromContext(r.Context())

	isAdmin := prof.Role == entity.AppRoleAdmin

	filterData := generative_image_model_service.GetGenerativeImageModelsFilter{
		Search:     filter.Search,
		SortBy:     filter.SortByDB(),
		SortDir:    filter.Sort,
		PageOffset: filter.Offset(),
		PageLimit:  filter.Limit,
		Page:       filter.Page,
		IsAdmin:    isAdmin,
	}

	res, pag, err := h.svc.GetAllGenerativeImageModels(r.Context(), filterData)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.LIST(w, r, "GET_GENERATIVE_IMAGE_MODELS_SUCCESS", res, &filter, pag)
}

func (h *Handler) GetGenerativeImageModelById(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		response.ValidationFailed(w, r, map[string]string{"id": "ID_MUST_BE_INTEGER"})
		return
	}

	prof, _ := internal_middleware.GetProfileFromContext(r.Context())
	isAdmin := prof.Role == entity.AppRoleAdmin

	res, err := h.svc.GetGenerativeImageModelById(r.Context(), id, isAdmin)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.OK(w, r, "GET_GENERATIVE_IMAGE_MODEL_SUCCESS", res)
}

func (h *Handler) GetGenerativeImageModelByModel(w http.ResponseWriter, r *http.Request) {
	modelName := chi.URLParam(r, "model")
	if modelName == "" {
		response.ValidationFailed(w, r, map[string]string{"model": "MODEL_REQUIRED"})
		return
	}

	prof, _ := internal_middleware.GetProfileFromContext(r.Context())
	isAdmin := prof.Role == entity.AppRoleAdmin

	res, err := h.svc.GetGenerativeImageModelByModel(r.Context(), modelName, isAdmin)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.OK(w, r, "GET_GENERATIVE_IMAGE_MODEL_SUCCESS", res)
}

func (h *Handler) CreateGenerativeImageModel(w http.ResponseWriter, r *http.Request) {
	var req generative_image_model_service.CreateGenerativeImageModelInput

	if appErr := utils.ValidateStruct(r.Body, &req); appErr != nil {
		response.ValidationFailed(w, r, appErr.ValidationErrors)
		return
	}

	prof, _ := internal_middleware.GetProfileFromContext(r.Context())
	req.ProfileID = prof.ID.String()

	res, err := h.svc.CreateGenerativeImageModel(r.Context(), req)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.OK(w, r, "CREATE_GENERATIVE_IMAGE_MODEL_SUCCESS", res)
}

func (h *Handler) UpdateGenerativeImageModel(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		response.ValidationFailed(w, r, map[string]string{"id": "ID_MUST_BE_INTEGER"})
		return
	}

	var req generative_image_model_service.UpdateGenerativeImageModelInput

	if appErr := utils.ValidateStruct(r.Body, &req); appErr != nil {
		response.ValidationFailed(w, r, appErr.ValidationErrors)
		return
	}

	prof, _ := internal_middleware.GetProfileFromContext(r.Context())
	req.ID = id
	req.ProfileID = prof.ID.String()

	res, err := h.svc.UpdateGenerativeImageModel(r.Context(), req)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.OK(w, r, "UPDATE_GENERATIVE_IMAGE_MODEL_SUCCESS", res)
}

func (h *Handler) DeleteGenerativeImageModel(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		response.ValidationFailed(w, r, map[string]string{"id": "ID_MUST_BE_INTEGER"})
		return
	}

	prof, _ := internal_middleware.GetProfileFromContext(r.Context())

	res, err := h.svc.DeleteGenerativeImageModel(r.Context(), id, prof.ID.String())
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.OK(w, r, "DELETE_GENERATIVE_IMAGE_MODEL_SUCCESS", res)
}
