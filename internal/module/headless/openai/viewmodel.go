// internal/module/headless/openai/viewmodel.go
package openai_svc

// ================== VIEWMODEL ==================

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
