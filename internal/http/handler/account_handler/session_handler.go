package account_handler

import (
	"encoding/json"
	"net/http"
	"postmatic-api/internal/http/middleware"
	"postmatic-api/internal/module/account/session"
	"postmatic-api/pkg/errs"
	"postmatic-api/pkg/response"
	"postmatic-api/pkg/utils"

	"github.com/go-chi/chi/v5"
)

func (h *AuthHandler) SessionRoutes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.GetSession)
	r.Get("/all", h.GetAllSession)
	r.Post("/logout", h.Logout)
	r.Post("/logout-all", h.LogoutAll)

	return r
}

func (h *AuthHandler) LogoutAll(w http.ResponseWriter, r *http.Request) {
	user, err := middleware.GetUserFromContext(r.Context())
	if err != nil {
		response.Error(w, err)
		return
	}
	profileId := user.ID

	input := session.LogoutAllInput{
		ProfileID: profileId,
	}
	// 4. Panggil Service
	// Tidak perlu mapping manual lagi! (req sudah bertipe DTO)
	err = h.sessSvc.LogoutAll(r.Context(), input)

	if err != nil {
		response.Error(w, err)
		return
	}

	response.OK(w, "LOGOUT_ALL_SUCCESS", nil)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req session.LogoutInput

	profile, err := middleware.GetUserFromContext(r.Context())
	if err != nil {
		response.Error(w, err)
		return
	}

	profileId := profile.ID

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

	// 4. Panggil Service
	// Tidak perlu mapping manual lagi! (req sudah bertipe DTO)
	res, err := h.sessSvc.Logout(r.Context(), req, profileId)

	if err != nil {
		response.Error(w, err)
		return
	}

	response.OK(w, "LOGOUT_SUCCESS", res)
}

func (h *AuthHandler) GetSession(w http.ResponseWriter, r *http.Request) {
	// 1. Gunakan Struct dari DTO
	user, err := middleware.GetUserFromContext(r.Context())
	if err != nil {
		response.Error(w, err)
		return
	}

	refreshToken := r.Header.Get("X-Postmatic-RefreshToken")

	if refreshToken == "" {
		response.Error(w, errs.NewUnauthorized("INVALID_REFRESH_TOKEN"))
		return
	}

	res, err := h.sessSvc.GetSession(r.Context(), user, refreshToken)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.OK(w, "GET_SESSION_SUCCESS", res)
}

func (h *AuthHandler) GetAllSession(w http.ResponseWriter, r *http.Request) {
	user, err := middleware.GetUserFromContext(r.Context())
	if err != nil {
		response.Error(w, err)
		return
	}

	profileId := user.ID

	res, err := h.sessSvc.GetAllSession(r.Context(), profileId)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.OK(w, "GET_SESSION_LIST_SUCCESS", res)
}
