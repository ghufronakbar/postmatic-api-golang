// internal/http/handler/account_handler/auth_handler.go
package account_handler

import (
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

	r.Post("/login", h.LoginCredential)
	r.Post("/register", h.Register)
	r.Post("/refresh-token", h.RefreshToken)
	r.Get("/verify/{createAccountToken}", h.CheckVerifyToken)
	r.Post("/verify/{createAccountToken}", h.SubmitVerifyToken)
	r.Post("/resend-email-verification", h.ResendEmailVerification)

	return r
}

func (h *AuthHandler) LoginCredential(w http.ResponseWriter, r *http.Request) {
	// 1. Gunakan Struct dari DTO
	var req auth.LoginCredentialInput

	if appErr := utils.ValidateStruct(r.Body, &req); appErr != nil {
		response.ValidationFailed(w, r, appErr.ValidationErrors)
		return
	}

	// Tidak perlu mapping manual lagi! (req sudah bertipe DTO)
	clientInfo := utils.ExtractClientInfo(r)
	sessionInput := auth.SessionInput{
		DeviceInfo: clientInfo,
	}
	res, err := h.authSvc.LoginCredential(r.Context(), req, sessionInput)

	if err != nil {
		response.Error(w, r, err, res)
		return
	}

	response.OK(w, r, "LOGIN_SUCCESS", res)
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	// 1. Gunakan Struct dari DTO
	var req auth.RegisterInput

	if appErr := utils.ValidateStruct(r.Body, &req); appErr != nil {
		response.ValidationFailed(w, r, appErr.ValidationErrors)
		return
	}

	res, err := h.authSvc.Register(r.Context(), req)

	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.OK(w, r, "REGISTER_SUCCESS", res)
}

func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	// 1. Gunakan Struct dari DTO
	var req auth.RefreshTokenInput

	if appErr := utils.ValidateStruct(r.Body, &req); appErr != nil {
		response.ValidationFailed(w, r, appErr.ValidationErrors)
		return
	}

	res, err := h.authSvc.RefreshToken(r.Context(), req)

	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.OK(w, r, "REFRESH_TOKEN_SUCCESS", res)
}

func (h *AuthHandler) CheckVerifyToken(w http.ResponseWriter, r *http.Request) {
	// 1. Gunakan Struct dari DTO
	createAccountToken := chi.URLParam(r, "createAccountToken")

	res, err := h.authSvc.CheckVerifyToken(r.Context(), createAccountToken)

	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.OK(w, r, "CHECK_VERIFY_TOKEN_SUCCESS", res)
}

func (h *AuthHandler) SubmitVerifyToken(w http.ResponseWriter, r *http.Request) {
	// 1. Gunakan Struct dari DTO
	createAccountToken := chi.URLParam(r, "createAccountToken")
	from := r.URL.Query().Get("from")
	clientInfo := utils.ExtractClientInfo(r)

	session := auth.SessionInput{
		DeviceInfo: clientInfo,
	}

	input := auth.SubmitVerifyTokenInput{
		Token: createAccountToken,
		From:  from,
	}

	res, err := h.authSvc.SubmitVerifyToken(r.Context(), input, session)

	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.OK(w, r, "SUBMIT_VERIFY_TOKEN_SUCCESS", res)
}

func (h *AuthHandler) ResendEmailVerification(w http.ResponseWriter, r *http.Request) {
	// 1. Gunakan Struct dari DTO
	var req auth.ResendEmailVerificationInput

	if appErr := utils.ValidateStruct(r.Body, &req); appErr != nil {
		response.ValidationFailed(w, r, appErr.ValidationErrors)
		return
	}

	res, err := h.authSvc.ResendEmailVerification(r.Context(), req)

	if err != nil {
		response.Error(w, r, err, res)
		return
	}

	response.OK(w, r, "RESEND_EMAIL_VERIFICATION_SUCCESS", res)
}
