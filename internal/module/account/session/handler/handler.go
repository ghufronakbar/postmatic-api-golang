// internal/module/account/session/handler/handler.go
package session_handler

import (
	"net/http"
	"postmatic-api/internal/internal_middleware"
	session_service "postmatic-api/internal/module/account/session/service"
	"postmatic-api/pkg/errs"
	"postmatic-api/pkg/response"
	"postmatic-api/pkg/utils"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	sessSvc *session_service.SessionService
}

func NewHandler(sessSvc *session_service.SessionService) *Handler {
	return &Handler{sessSvc: sessSvc}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.GetSession)
	r.Get("/all", h.GetAllSession)
	r.Post("/logout", h.Logout)
	r.Post("/logout-all", h.LogoutAll)

	return r
}

func (h *Handler) LogoutAll(w http.ResponseWriter, r *http.Request) {
	user, err := internal_middleware.GetProfileFromContext(r.Context())
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}
	profileId := user.ID

	input := session_service.LogoutAllInput{
		ProfileID: profileId,
	}
	// 4. Panggil Service
	// Tidak perlu mapping manual lagi! (req sudah bertipe DTO)
	err = h.sessSvc.LogoutAll(r.Context(), input)

	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.OK(w, r, "LOGOUT_ALL_SUCCESS", nil)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	var req session_service.LogoutInput

	if appErr := utils.ValidateStruct(r.Body, &req); appErr != nil {
		response.ValidationFailed(w, r, appErr.ValidationErrors)
		return
	}

	profile, err := internal_middleware.GetProfileFromContext(r.Context())
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	profileId := profile.ID

	// Tidak perlu mapping manual lagi! (req sudah bertipe DTO)
	res, err := h.sessSvc.Logout(r.Context(), req, profileId)

	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.OK(w, r, "LOGOUT_SUCCESS", res)
}

func (h *Handler) GetSession(w http.ResponseWriter, r *http.Request) {
	// 1. Gunakan Struct dari DTO
	user, err := internal_middleware.GetProfileFromContext(r.Context())
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	refreshToken := r.Header.Get("X-Postmatic-RefreshToken")

	if refreshToken == "" {
		response.Error(w, r, errs.NewUnauthorized("INVALID_REFRESH_TOKEN"), nil)
		return
	}

	res, err := h.sessSvc.GetSession(r.Context(), user, refreshToken)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.OK(w, r, "GET_SESSION_SUCCESS", res)
}

func (h *Handler) GetAllSession(w http.ResponseWriter, r *http.Request) {
	user, err := internal_middleware.GetProfileFromContext(r.Context())
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	profileId := user.ID

	res, err := h.sessSvc.GetAllSession(r.Context(), profileId)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.OK(w, r, "GET_SESSION_LIST_SUCCESS", res)
}
