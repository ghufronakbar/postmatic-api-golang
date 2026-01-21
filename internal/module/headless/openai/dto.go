// internal/module/headless/openai/dto.go
package openai_svc

// ================== INPUT DTOs ==================

// GenerateTextInput is the input DTO for generating text with chat completions
type GenerateTextInput struct {
	Model    string        `json:"model" validate:"required"`
	Messages []ChatMessage `json:"messages" validate:"required,min=1"`
	// Optional parameters
	Temperature      *float64 `json:"temperature"`
	MaxTokens        *int     `json:"maxTokens"`
	TopP             *float64 `json:"topP"`
	FrequencyPenalty *float64 `json:"frequencyPenalty"`
	PresencePenalty  *float64 `json:"presencePenalty"`
}

// ChatMessage represents a message in the chat
type ChatMessage struct {
	Role    string `json:"role" validate:"required,oneof=system user assistant"`
	Content string `json:"content" validate:"required"`
}

// GenerateImageInput is the input DTO for generating images with DALL-E
type GenerateImageInput struct {
	Model  string `json:"model" validate:"required"` // dall-e-2 or dall-e-3
	Prompt string `json:"prompt" validate:"required"`
	// Optional parameters
	N       *int    `json:"n"`       // Number of images (1-10 for dall-e-2, 1 for dall-e-3)
	Size    *string `json:"size"`    // 256x256, 512x512, 1024x1024, 1792x1024, 1024x1792
	Quality *string `json:"quality"` // standard or hd (dall-e-3 only)
	Style   *string `json:"style"`   // vivid or natural (dall-e-3 only)
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
	URL           string `json:"url"`
	RevisedPrompt string `json:"revisedPrompt"`
}
