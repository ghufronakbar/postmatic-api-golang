// internal/module/account/auth/handler/handler.go
package auth_handler

import (
	"net/http"
	"postmatic-api/config"
	auth_service "postmatic-api/internal/module/account/auth/service"
	"postmatic-api/pkg/response"
	"postmatic-api/pkg/utils"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	authSvc *auth_service.AuthService
	cfg     *config.Config
}

func NewHandler(authSvc *auth_service.AuthService, cfg *config.Config) *Handler {
	return &Handler{authSvc: authSvc, cfg: cfg}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/login", h.LoginCredential)
	r.Post("/register", h.Register)
	r.Post("/refresh-token", h.RefreshToken)
	r.Get("/verify/{createAccountToken}", h.CheckVerifyToken)
	r.Post("/verify/{createAccountToken}", h.SubmitVerifyToken)
	r.Post("/resend-email-verification", h.ResendEmailVerification)

	return r
}

func (h *Handler) LoginCredential(w http.ResponseWriter, r *http.Request) {
	// 1. Gunakan Struct dari DTO
	var req auth_service.LoginCredentialInput

	if appErr := utils.ValidateStruct(r.Body, &req); appErr != nil {
		response.ValidationFailed(w, r, appErr.ValidationErrors)
		return
	}

	// Tidak perlu mapping manual lagi! (req sudah bertipe DTO)
	clientInfo := utils.ExtractClientInfo(r)
	sessionInput := auth_service.SessionInput{
		DeviceInfo: clientInfo,
	}
	res, err := h.authSvc.LoginCredential(r.Context(), req, sessionInput)

	if err != nil {
		response.Error(w, r, err, res)
		return
	}

	SetAuthCookies(w, r, h.cfg, res.AccessToken, res.RefreshToken)

	response.OK(w, r, "LOGIN_SUCCESS", res)
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	// 1. Gunakan Struct dari DTO
	var req auth_service.RegisterInput

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

func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	// 1. Gunakan Struct dari DTO
	var req auth_service.RefreshTokenInput

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

func (h *Handler) CheckVerifyToken(w http.ResponseWriter, r *http.Request) {
	// 1. Gunakan Struct dari DTO
	createAccountToken := chi.URLParam(r, "createAccountToken")

	res, err := h.authSvc.CheckVerifyToken(r.Context(), createAccountToken)

	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.OK(w, r, "CHECK_VERIFY_TOKEN_SUCCESS", res)
}

func (h *Handler) SubmitVerifyToken(w http.ResponseWriter, r *http.Request) {
	// 1. Gunakan Struct dari DTO
	createAccountToken := chi.URLParam(r, "createAccountToken")
	from := r.URL.Query().Get("from")
	clientInfo := utils.ExtractClientInfo(r)

	session := auth_service.SessionInput{
		DeviceInfo: clientInfo,
	}

	input := auth_service.SubmitVerifyTokenInput{
		Token: createAccountToken,
		From:  from,
	}

	res, err := h.authSvc.SubmitVerifyToken(r.Context(), input, session)

	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	SetAuthCookies(w, r, h.cfg, res.AccessToken, res.RefreshToken)

	response.OK(w, r, "SUBMIT_VERIFY_TOKEN_SUCCESS", res)
}

func (h *Handler) ResendEmailVerification(w http.ResponseWriter, r *http.Request) {
	// 1. Gunakan Struct dari DTO
	var req auth_service.ResendEmailVerificationInput

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
