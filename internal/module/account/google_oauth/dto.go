// internal/module/account/google_oauth/dto.go
package google_oauth

import "postmatic-api/pkg/utils"

type SessionInput struct {
	DeviceInfo utils.ClientInfo
}

// // Frontend akan POST ini setelah dapat ?code=...&state=... dari redirect_uri
// type GoogleOAuthCallbackQuery struct {
// 	Code  string
// 	State string
// }

type GoogleOAuthCallbackInput struct {
	From  string `json:"from" validate:"required"`
	Code  string `json:"code" validate:"required"`
	State string `json:"state" validate:"required"`
}
