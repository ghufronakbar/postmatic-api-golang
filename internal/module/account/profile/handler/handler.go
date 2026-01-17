// internal/module/account/profile/handler/handler.go
package profile_handler

import (
	"errors"
	"fmt"
	"net/http"
	"postmatic-api/internal/internal_middleware"
	profile_service "postmatic-api/internal/module/account/profile/service"

	"postmatic-api/pkg/response"
	"postmatic-api/pkg/utils"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	profSvc *profile_service.ProfileService
}

func NewHandler(profSvc *profile_service.ProfileService) *Handler {
	return &Handler{profSvc: profSvc}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.GetProfile)
	r.Put("/", h.UpdateProfile)
	r.Put("/password", h.UpdatePassword)
	r.Post("/password", h.SetupPassword)

	return r
}

func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	user, err := internal_middleware.GetProfileFromContext(r.Context())

	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	if user == nil {
		fmt.Println("USER_NOT_FOUND")
		response.Error(w, r, errors.New("USER_NOT_FOUND"), nil)
		return
	}

	res, err := h.profSvc.GetProfile(r.Context(), user.ID)

	if err != nil {
		fmt.Println(err)
		response.Error(w, r, err, res)
		return
	}

	response.OK(w, r, "GET_PROFILE_SUCCESS", res)
}

func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	// 1. Gunakan Struct dari DTO
	var req profile_service.UpdateProfileInput

	if appErr := utils.ValidateStruct(r.Body, &req); appErr != nil {
		response.ValidationFailed(w, r, appErr.ValidationErrors)
		return
	}
	user, err := internal_middleware.GetProfileFromContext(r.Context())

	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	if user == nil {
		response.Error(w, r, errors.New("USER_NOT_FOUND"), nil)
		return
	}

	res, err := h.profSvc.UpdateProfile(r.Context(), user.ID, req)

	if err != nil {
		response.Error(w, r, err, res)
		return
	}

	response.OK(w, r, "UPDATE_PROFILE_SUCCESS", res)
}

func (h *Handler) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	// 1. Gunakan Struct dari DTO
	var req profile_service.UpdatePasswordInput

	if appErr := utils.ValidateStruct(r.Body, &req); appErr != nil {
		response.ValidationFailed(w, r, appErr.ValidationErrors)
		return
	}
	user, err := internal_middleware.GetProfileFromContext(r.Context())

	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	if user == nil {
		response.Error(w, r, errors.New("USER_NOT_FOUND"), nil)
		return
	}

	err = h.profSvc.UpdatePassword(r.Context(), user.ID, req)

	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.OK(w, r, "UPDATE_PASSWORD_SUCCESS", nil)
}

func (h *Handler) SetupPassword(w http.ResponseWriter, r *http.Request) {
	// 1. Gunakan Struct yang BENAR (SetupPasswordInput)
	var req profile_service.SetupPasswordInput

	if appErr := utils.ValidateStruct(r.Body, &req); appErr != nil {
		response.ValidationFailed(w, r, appErr.ValidationErrors)
		return
	}
	user, err := internal_middleware.GetProfileFromContext(r.Context())
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}
	if user == nil {
		response.Error(w, r, errors.New("USER_NOT_FOUND"), nil)
		return
	}

	// 2. Panggil Service yang BENAR (SetupPassword)
	res, err := h.profSvc.SetupPassword(r.Context(), user.ID, req)

	if err != nil {
		// Jika error karena rate limit, kita bisa sertakan data retryAfter
		response.Error(w, r, err, res)
		return
	}

	// 3. Response setup password biasanya return sukses info
	response.OK(w, r, "SETUP_PASSWORD_SUCCESS_CHECK_EMAIL", res)
}
