package generative_text_model_service

import "time"

type GenerativeTextModelResponse struct {
	ID        int64                           `json:"id"`
	Model     string                          `json:"model"`
	Label     string                          `json:"label"`
	Image     *string                         `json:"image"`
	Provider  GenerativeTextModelProviderType `json:"provider"`
	IsActive  bool                            `json:"isActive"`
	CreatedAt time.Time                       `json:"createdAt"`
	UpdatedAt time.Time                       `json:"updatedAt"`
}
