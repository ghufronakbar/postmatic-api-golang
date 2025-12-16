// internal/http/handler/account_handler/profile_handler.go
package account_handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"postmatic-api/internal/http/middleware"
	"postmatic-api/internal/module/account/profile"

	"postmatic-api/pkg/response"
	"postmatic-api/pkg/utils"

	"github.com/go-chi/chi/v5"
)

type ProfileHandler struct {
	profSvc *profile.ProfileService
}

func NewProfileHandler(profSvc *profile.ProfileService) *ProfileHandler {
	return &ProfileHandler{profSvc: profSvc}
}

func (h *ProfileHandler) ProfileRoutes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.GetProfile)
	r.Put("/", h.UpdateProfile)
	r.Put("/password", h.UpdatePassword)
	r.Post("/password", h.SetupPassword)

	return r
}

func (h *ProfileHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	user, err := middleware.GetUserFromContext(r.Context())

	if err != nil {
		response.Error(w, err, nil)
		return
	}

	if user == nil {
		fmt.Println("USER_NOT_FOUND")
		response.Error(w, errors.New("USER_NOT_FOUND"), nil)
		return
	}

	res, err := h.profSvc.GetProfile(r.Context(), user.ID)

	if err != nil {
		fmt.Println(err)
		response.Error(w, err, res)
		return
	}

	response.OK(w, "GET_PROFILE_SUCCESS", res)
}

func (h *ProfileHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	// 1. Gunakan Struct dari DTO
	var req profile.UpdateProfileInput
	user, err := middleware.GetUserFromContext(r.Context())

	if err != nil {
		response.Error(w, err, nil)
		return
	}

	if user == nil {
		response.Error(w, errors.New("USER_NOT_FOUND"), nil)
		return
	}

	// 2. Decode langsung ke struct tersebut
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.InvalidJsonFormat(w)
		return
	}

	// 3. Validasi struct tersebut
	// Validator akan membaca tag `validate` yang ada di DTO
	if errsMap := utils.ValidateStruct(req); errsMap != nil {
		response.ValidationFailed(w, errsMap)
		return
	}

	res, err := h.profSvc.UpdateProfile(r.Context(), user.ID, req)

	if err != nil {
		response.Error(w, err, res)
		return
	}

	response.OK(w, "UPDATE_PROFILE_SUCCESS", res)
}

func (h *ProfileHandler) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	// 1. Gunakan Struct dari DTO
	var req profile.UpdatePasswordInput
	user, err := middleware.GetUserFromContext(r.Context())

	if err != nil {
		response.Error(w, err, nil)
		return
	}

	if user == nil {
		response.Error(w, errors.New("USER_NOT_FOUND"), nil)
		return
	}

	// 2. Decode langsung ke struct tersebut
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.InvalidJsonFormat(w)
		return
	}

	// 3. Validasi struct tersebut
	// Validator akan membaca tag `validate` yang ada di DTO
	if errsMap := utils.ValidateStruct(req); errsMap != nil {
		response.ValidationFailed(w, errsMap)
		return
	}

	err = h.profSvc.UpdatePassword(r.Context(), user.ID, req)

	if err != nil {
		response.Error(w, err, nil)
		return
	}

	response.OK(w, "UPDATE_PASSWORD_SUCCESS", nil)
}

func (h *ProfileHandler) SetupPassword(w http.ResponseWriter, r *http.Request) {
	// 1. Gunakan Struct yang BENAR (SetupPasswordInput)
	var req profile.SetupPasswordInput

	user, err := middleware.GetUserFromContext(r.Context())
	if err != nil {
		response.Error(w, err, nil)
		return
	}
	if user == nil {
		response.Error(w, errors.New("USER_NOT_FOUND"), nil)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.InvalidJsonFormat(w)
		return
	}

	if errsMap := utils.ValidateStruct(req); errsMap != nil {
		response.ValidationFailed(w, errsMap)
		return
	}

	// 2. Panggil Service yang BENAR (SetupPassword)
	res, err := h.profSvc.SetupPassword(r.Context(), user.ID, req)

	if err != nil {
		// Jika error karena rate limit, kita bisa sertakan data retryAfter
		response.Error(w, err, res)
		return
	}

	// 3. Response setup password biasanya return sukses info
	response.OK(w, "SETUP_PASSWORD_SUCCESS_CHECK_EMAIL", res)
}
