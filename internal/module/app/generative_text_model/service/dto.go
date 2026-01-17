package generative_text_model_service

type CreateGenerativeTextModelInput struct {
	ProfileID string                          `json:"-"`
	Model     string                          `json:"model" validate:"required,max=255"`
	Label     string                          `json:"label" validate:"required,max=255"`
	Image     *string                         `json:"image" validate:"omitempty,url,max=255"`
	Provider  GenerativeTextModelProviderType `json:"provider" validate:"required,oneof=openai google"`
	IsActive  bool                            `json:"isActive"`
}

type UpdateGenerativeTextModelInput struct {
	ID        int64                           `json:"-"`
	ProfileID string                          `json:"-"`
	Model     string                          `json:"model" validate:"required,max=255"`
	Label     string                          `json:"label" validate:"required,max=255"`
	Image     *string                         `json:"image" validate:"omitempty,url,max=255"`
	Provider  GenerativeTextModelProviderType `json:"provider" validate:"required,oneof=openai google"`
	IsActive  bool                            `json:"isActive"`
}

type GenerativeTextModelProviderType string

const (
	GenerativeTextModelProviderTypeOpenAI GenerativeTextModelProviderType = "openai"
	GenerativeTextModelProviderTypeGoogle GenerativeTextModelProviderType = "google"
)
