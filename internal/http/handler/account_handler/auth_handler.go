// internal/http/handler/account_handler/auth_handler.go
package account_handler

import (
	"encoding/json"
	"net/http"
	"postmatic-api/internal/module/account/auth"
	"postmatic-api/internal/module/account/session"
	"postmatic-api/pkg/response"
	"postmatic-api/pkg/utils"

	"github.com/go-chi/chi/v5"
)

type AuthHandler struct {
	authSvc *auth.AuthService
	sessSvc *session.SessionService
}

func NewAuthHandler(authSvc *auth.AuthService, sessSvc *session.SessionService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc, sessSvc: sessSvc}
}

func (h *AuthHandler) AuthRoutes() chi.Router {
	r := chi.NewRouter()

	r.Post("/login", h.LoginCredentials)
	r.Post("/register", h.Register)
	r.Post("/refresh-token", h.RefreshToken)
	r.Get("/verify/{createAccountToken}", h.CheckVerifyToken)
	r.Post("/verify/{createAccountToken}", h.SubmitVerifyToken)
	r.Post("/resend-email-verification", h.ResendEmailVerification)

	return r
}

func (h *AuthHandler) LoginCredentials(w http.ResponseWriter, r *http.Request) {
	// 1. Gunakan Struct dari DTO
	var req auth.LoginCredentialsInput

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
	clientInfo := utils.ExtractClientInfo(r)
	sessionInput := auth.SessionInput{
		DeviceInfo: clientInfo,
	}
	res, err := h.authSvc.LoginCredentials(r.Context(), req, sessionInput)

	if err != nil {
		response.Error(w, err)
		return
	}

	response.OK(w, "LOGIN_SUCCESS", res)
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	// 1. Gunakan Struct dari DTO
	var req auth.RegisterInput

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
	res, err := h.authSvc.Register(r.Context(), req)

	if err != nil {
		response.Error(w, err)
		return
	}

	response.OK(w, "REGISTER_SUCCESS", res)
}

func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	// 1. Gunakan Struct dari DTO
	var req auth.RefreshTokenInput

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
	res, err := h.authSvc.RefreshToken(r.Context(), req)

	if err != nil {
		response.Error(w, err)
		return
	}

	response.OK(w, "REFRESH_TOKEN_SUCCESS", res)
}

func (h *AuthHandler) CheckVerifyToken(w http.ResponseWriter, r *http.Request) {
	// 1. Gunakan Struct dari DTO
	createAccountToken := chi.URLParam(r, "createAccountToken")

	res, err := h.authSvc.CheckVerifyToken(r.Context(), createAccountToken)

	if err != nil {
		response.Error(w, err)
		return
	}

	response.OK(w, "CHECK_VERIFY_TOKEN_SUCCESS", res)
}

func (h *AuthHandler) SubmitVerifyToken(w http.ResponseWriter, r *http.Request) {
	// 1. Gunakan Struct dari DTO
	createAccountToken := chi.URLParam(r, "createAccountToken")
	clientInfo := utils.ExtractClientInfo(r)

	session := auth.SessionInput{
		DeviceInfo: clientInfo,
	}

	res, err := h.authSvc.SubmitVerifyToken(r.Context(), createAccountToken, session)

	if err != nil {
		response.Error(w, err)
		return
	}

	response.OK(w, "SUBMIT_VERIFY_TOKEN_SUCCESS", res)
}

func (h *AuthHandler) ResendEmailVerification(w http.ResponseWriter, r *http.Request) {
	// 1. Gunakan Struct dari DTO
	var req auth.ResendEmailVerificationInput

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
	res, err := h.authSvc.ResendEmailVerification(r.Context(), req)

	if err != nil {
		response.Error(w, err)
		return
	}

	response.OK(w, "RESEND_EMAIL_VERIFICATION_SUCCESS", res)
}
