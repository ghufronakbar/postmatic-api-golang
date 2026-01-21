// internal/module/headless/google_genai/dto.go
package google_genai

// ================== INPUT DTOs ==================

// GenerateTextInput is the input DTO for generating text
type GenerateTextInput struct {
	Model  string `json:"model" validate:"required"`
	Prompt string `json:"prompt" validate:"required"`
	// Optional parameters
	Temperature     *float64 `json:"temperature"`
	MaxOutputTokens *int     `json:"maxOutputTokens"`
	TopP            *float64 `json:"topP"`
	TopK            *int     `json:"topK"`
}

// GenerateImageInput is the input DTO for generating images (Imagen)
type GenerateImageInput struct {
	Model  string `json:"model" validate:"required"`
	Prompt string `json:"prompt" validate:"required"`
	// Optional parameters
	NumberOfImages *int    `json:"numberOfImages"` // 1-4
	AspectRatio    *string `json:"aspectRatio"`    // e.g., "1:1", "16:9", "9:16"
}
