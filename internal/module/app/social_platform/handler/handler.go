// internal/module/app/social_platform/handler/handler.go
package social_platform_handler

import (
	"net/http"
	"strconv"

	"postmatic-api/internal/internal_middleware"
	social_platform_service "postmatic-api/internal/module/app/social_platform/service"
	"postmatic-api/pkg/filter"
	"postmatic-api/pkg/response"
	"postmatic-api/pkg/utils"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Handler handles social platform HTTP requests
type Handler struct {
	svc *social_platform_service.SocialPlatformService
}

// NewHandler creates a new Handler
func NewHandler(svc *social_platform_service.SocialPlatformService) *Handler {
	return &Handler{svc: svc}
}

// Routes returns the routes for social platform
// adminOnly middleware should be applied for POST, PUT, DELETE
func (h *Handler) Routes(allAllowed, adminOnly func(http.Handler) http.Handler) chi.Router {
	r := chi.NewRouter()

	// Public routes (all users can access)
	r.Group(func(r chi.Router) {
		r.Use(allAllowed)
		r.Use(func(next http.Handler) http.Handler {
			return internal_middleware.ReqFilterMiddleware(next, social_platform_service.SORT_BY)
		})
		r.Get("/", h.GetAll)
		r.Get("/platform-code", h.GetPlatformCodes)
	})

	// Admin only routes
	r.Group(func(r chi.Router) {
		r.Use(adminOnly)
		r.Post("/", h.Create)
		r.Put("/{id}", h.Update)
		r.Delete("/{id}", h.Delete)
	})

	return r
}

// GetAll handles GET /api/app/social-platform
// @Summary Get all social platforms
// @Tags SocialPlatform
// @Accept json
// @Produce json
// @Param search query string false "Search by name or platform code"
// @Param sortBy query string false "Sort by field (id, name, platform_code, is_active)"
// @Param sort query string false "Sort direction (asc, desc)"
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Success 200 {object} response.Response{data=[]social_platform_service.SocialPlatformResponse}
// @Router /api/app/social-platform [get]
func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get profile to check if admin
	claims, _ := internal_middleware.GetProfileFromContext(ctx)
	isAdmin := claims != nil && claims.Role == "admin"

	// Get filter from middleware
	reqFilter := internal_middleware.GetFilterFromContext(ctx)

	var searchPtr *string
	if reqFilter.Search != "" {
		searchPtr = &reqFilter.Search
	}

	svcFilter := social_platform_service.GetSocialPlatformsFilter{
		IncludeInactive: isAdmin, // Admin can see inactive
		Search:          searchPtr,
		SortBy:          reqFilter.SortBy,
		SortDir:         reqFilter.Sort,
		Page:            reqFilter.Page,
		Limit:           reqFilter.Limit,
	}

	data, pag, err := h.svc.GetAll(ctx, svcFilter)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.LIST(w, r, "SOCIAL_PLATFORMS_RETRIEVED", data, &filter.ReqFilter{
		Search: reqFilter.Search,
		Page:   reqFilter.Page,
		Limit:  reqFilter.Limit,
		SortBy: reqFilter.SortBy,
		Sort:   reqFilter.Sort,
	}, pag)
}

// GetPlatformCodes handles GET /api/app/social-platform/platform-code
// @Summary Get all platform codes
// @Tags SocialPlatform
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=[]string}
// @Router /api/app/social-platform/platform-code [get]
func (h *Handler) GetPlatformCodes(w http.ResponseWriter, r *http.Request) {
	codes := h.svc.GetPlatformCodes()
	response.OK(w, r, "PLATFORM_CODES_RETRIEVED", codes)
}

// Create handles POST /api/app/social-platform
// @Summary Create social platform
// @Tags SocialPlatform
// @Accept json
// @Produce json
// @Param body body social_platform_service.CreateSocialPlatformInput true "Social platform input"
// @Success 201 {object} response.Response{data=social_platform_service.SocialPlatformResponse}
// @Router /api/app/social-platform [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var input social_platform_service.CreateSocialPlatformInput
	if appErr := utils.ValidateStruct(r.Body, &input); appErr != nil {
		response.ValidationFailed(w, r, appErr.ValidationErrors)
		return
	}

	// Get profile ID
	claims, _ := internal_middleware.GetProfileFromContext(ctx)
	if claims == nil {
		response.ValidationFailed(w, r, map[string]string{"authorization": "PROFILE_ID_REQUIRED"})
		return
	}
	input.ProfileID = claims.ID

	data, err := h.svc.Create(ctx, input)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.OK(w, r, "SOCIAL_PLATFORM_CREATED", data)
}

// Update handles PUT /api/app/social-platform/{id}
// @Summary Update social platform
// @Tags SocialPlatform
// @Accept json
// @Produce json
// @Param id path int true "Social platform ID"
// @Param body body social_platform_service.UpdateSocialPlatformInput true "Social platform input"
// @Success 200 {object} response.Response{data=social_platform_service.SocialPlatformResponse}
// @Router /api/app/social-platform/{id} [put]
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.ValidationFailed(w, r, map[string]string{"id": "INVALID_ID"})
		return
	}

	var input social_platform_service.UpdateSocialPlatformInput
	if appErr := utils.ValidateStruct(r.Body, &input); appErr != nil {
		response.ValidationFailed(w, r, appErr.ValidationErrors)
		return
	}

	// Get profile ID
	claims, _ := internal_middleware.GetProfileFromContext(ctx)
	if claims == nil {
		response.ValidationFailed(w, r, map[string]string{"authorization": "PROFILE_ID_REQUIRED"})
		return
	}
	input.ProfileID = claims.ID

	data, err := h.svc.Update(ctx, id, input)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.OK(w, r, "SOCIAL_PLATFORM_UPDATED", data)
}

// Delete handles DELETE /api/app/social-platform/{id}
// @Summary Delete social platform
// @Tags SocialPlatform
// @Accept json
// @Produce json
// @Param id path int true "Social platform ID"
// @Success 200 {object} response.Response{data=social_platform_service.SocialPlatformResponse}
// @Router /api/app/social-platform/{id} [delete]
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.ValidationFailed(w, r, map[string]string{"id": "INVALID_ID"})
		return
	}

	// Get profile ID
	claims, _ := internal_middleware.GetProfileFromContext(ctx)
	if claims == nil {
		response.ValidationFailed(w, r, map[string]string{"authorization": "PROFILE_ID_REQUIRED"})
		return
	}

	data, err := h.svc.Delete(ctx, id, uuid.UUID(claims.ID))
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.OK(w, r, "SOCIAL_PLATFORM_DELETED", data)
}
