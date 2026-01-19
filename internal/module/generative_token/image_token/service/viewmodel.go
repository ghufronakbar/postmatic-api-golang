// internal/module/generative_token/image_token/service/viewmodel.go
package image_token_service

// TokenStatusResponse is response for GET /status endpoint
type TokenStatusResponse struct {
	AvailableToken int64 `json:"availableToken"`
	UsedToken      int64 `json:"usedToken"`
	TotalToken     int64 `json:"totalToken"`
	IsExhausted    bool  `json:"isExhausted"`
}
