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

// ================== OUTPUT DTOs ==================

// GenerateTextResponse is the output DTO for text generation
type GenerateTextResponse struct {
	Text             string `json:"text"`
	Model            string `json:"model"`
	PromptTokenCount int    `json:"promptTokenCount"`
	OutputTokenCount int    `json:"outputTokenCount"`
	TotalTokenCount  int    `json:"totalTokenCount"`
	FinishReason     string `json:"finishReason"`
}

// GenerateImageResponse is the output DTO for image generation
type GenerateImageResponse struct {
	Images []GeneratedImage `json:"images"`
	Model  string           `json:"model"`
}

// GeneratedImage represents a single generated image
type GeneratedImage struct {
	Base64Data string `json:"base64Data"`
	MimeType   string `json:"mimeType"`
}
