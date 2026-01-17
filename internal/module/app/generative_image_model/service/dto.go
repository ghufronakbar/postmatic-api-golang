package generative_image_model_service

type CreateGenerativeImageModelInput struct {
	ProfileID   string                           `json:"-"`
	Model       string                           `json:"model" validate:"required,max=255"`
	Label       string                           `json:"label" validate:"required,max=255"`
	Image       *string                          `json:"image" validate:"omitempty,url,max=255"`
	Provider    GenerativeImageModelProviderType `json:"provider" validate:"required,oneof=openai google"`
	IsActive    bool                             `json:"isActive"`
	ValidRatios []string                         `json:"validRatios" validate:"required,min=1"`
	ImageSizes  []string                         `json:"imageSizes"`
}

type UpdateGenerativeImageModelInput struct {
	ID          int64                            `json:"-"`
	ProfileID   string                           `json:"-"`
	Model       string                           `json:"model" validate:"required,max=255"`
	Label       string                           `json:"label" validate:"required,max=255"`
	Image       *string                          `json:"image" validate:"omitempty,url,max=255"`
	Provider    GenerativeImageModelProviderType `json:"provider" validate:"required,oneof=openai google"`
	IsActive    bool                             `json:"isActive"`
	ValidRatios []string                         `json:"validRatios" validate:"required,min=1"`
	ImageSizes  []string                         `json:"imageSizes"`
}

type GenerativeImageModelProviderType string

const (
	GenerativeImageModelProviderTypeOpenAI GenerativeImageModelProviderType = "openai"
	GenerativeImageModelProviderTypeGoogle GenerativeImageModelProviderType = "google"
)
