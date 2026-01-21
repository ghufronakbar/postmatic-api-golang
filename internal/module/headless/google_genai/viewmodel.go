// internal/module/headless/google_genai/viewmodel.go
package google_genai

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
	Base64Data string `json:"base64Data"`
	MimeType   string `json:"mimeType"`
}
