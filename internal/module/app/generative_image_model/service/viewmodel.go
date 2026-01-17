package generative_image_model_service

import "time"

type GenerativeImageModelResponse struct {
	ID          int64                            `json:"id"`
	Model       string                           `json:"model"`
	Label       string                           `json:"label"`
	Image       *string                          `json:"image"`
	Provider    GenerativeImageModelProviderType `json:"provider"`
	IsActive    bool                             `json:"isActive"`
	ValidRatios []string                         `json:"validRatios"`
	ImageSizes  *[]string                        `json:"imageSizes"`
	CreatedAt   time.Time                        `json:"createdAt"`
	UpdatedAt   time.Time                        `json:"updatedAt"`
}
