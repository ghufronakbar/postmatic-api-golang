// internal/http/handler/account_handler/google_oauth_handler.go
package account_handler

import (
	"net/http"

	"postmatic-api/config"
	"postmatic-api/internal/module/account/google_oauth"
	"postmatic-api/pkg/response"
	"postmatic-api/pkg/utils"

	"github.com/go-chi/chi/v5"
)

type GoogleOAuthHandler struct {
	authSvc *google_oauth.GoogleOAuthService
	cfg     *config.Config
}

func NewGoogleOAuthHandler(authSvc *google_oauth.GoogleOAuthService, cfg *config.Config) *GoogleOAuthHandler {
	return &GoogleOAuthHandler{authSvc: authSvc, cfg: cfg}
}

func (h *GoogleOAuthHandler) GoogleOAuthRoutes() chi.Router {
	r := chi.NewRouter()

	r.Get("/login", h.GetGoogleAuthURL)
	r.Get("/callback", h.LoginGoogleCallback)

	return r
}

// GET /url?from=web
func (h *GoogleOAuthHandler) GetGoogleAuthURL(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("from")
	if from == "" {
		response.ValidationFailed(w, r, map[string]string{"from": "from is required"})
		return
	}

	res, err := h.authSvc.GetGoogleAuthURL(r.Context(), from)
	if err != nil {
		response.Error(w, r, err, res)
		return
	}

	response.OK(w, r, "GOOGLE_AUTH_URL_SUCCESS", res)
}

func (h *GoogleOAuthHandler) LoginGoogleCallback(w http.ResponseWriter, r *http.Request) {
	req := google_oauth.GoogleOAuthCallbackInput{
		Code:  r.URL.Query().Get("code"),
		State: r.URL.Query().Get("state"),
		From:  "", // biarkan kosong, nanti diisi dari state (yang sudah signed)
	}

	if req.Code == "" || req.State == "" {
		response.ValidationFailed(w, r, map[string]string{
			"code":  "code is required",
			"state": "state is required",
		})
		return
	}

	clientInfo := utils.ExtractClientInfo(r)
	sessionInput := google_oauth.SessionInput{DeviceInfo: clientInfo}

	res, err := h.authSvc.LoginGoogleCallback(r.Context(), req, sessionInput)
	if err != nil {
		response.Error(w, r, err, res)
		return
	}

	// ✅ mode redirect (callback dipanggil browser)
	SetAuthCookies(w, r, h.cfg, res.AccessToken, res.RefreshToken)
	http.Redirect(w, r, res.From, http.StatusFound)

	// (opsional) kalau kamu kadang pingin lihat JSON via Postman:
	// response.OK(w, r, "LOGIN_GOOGLE_CALLBACK_SUCCESS", res)
}
