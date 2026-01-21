# Headless.GoogleGenAI

Module headless untuk integrasi dengan Google Generative AI (Gemini API).

## Packages

- `google.golang.org/genai`

## Config Variables

| Variable               | Description                   |
| ---------------------- | ----------------------------- |
| `GOOGLE_GENAI_API_KEY` | API Key dari Google AI Studio |

## Client Setup

File: `config/google_gen_ai.go`

```go
client := config.ConnectGoogleGenAI(cfg)
```

Menggunakan `genai.BackendGeminiAPI` (bukan Vertex AI).

---

## Service Interface

```go
type Service interface {
    // Text Generation
    GenerateText(ctx context.Context, input GenerateTextInput) (*GenerateTextResponse, error)

    // Image Generation (Imagen)
    GenerateImage(ctx context.Context, input GenerateImageInput) (*GenerateImageResponse, error)
}
```

---

## GenerateText

Generate text menggunakan model Gemini (gemini-1.5-pro, gemini-1.5-flash, dll).

### Input DTO

```go
type GenerateTextInput struct {
    Model           string   `json:"model"`           // Required: gemini-1.5-pro, dll
    Prompt          string   `json:"prompt"`          // Required
    Temperature     *float64 `json:"temperature"`     // Optional: 0.0 - 2.0
    MaxOutputTokens *int     `json:"maxOutputTokens"` // Optional
    TopP            *float64 `json:"topP"`            // Optional
    TopK            *int     `json:"topK"`            // Optional
}
```

### Output DTO

```go
type GenerateTextResponse struct {
    Text             string `json:"text"`
    Model            string `json:"model"`
    PromptTokenCount int    `json:"promptTokenCount"`
    OutputTokenCount int    `json:"outputTokenCount"`
    TotalTokenCount  int    `json:"totalTokenCount"`
    FinishReason     string `json:"finishReason"`
}
```

---

## GenerateImage

Generate image menggunakan model Imagen (imagen-3.0-generate-002, dll).

### Input DTO

```go
type GenerateImageInput struct {
    Model          string  `json:"model"`          // Required: imagen-3.0-generate-002
    Prompt         string  `json:"prompt"`         // Required
    NumberOfImages *int    `json:"numberOfImages"` // Optional: 1-4
    AspectRatio    *string `json:"aspectRatio"`    // Optional: "1:1", "16:9", "9:16"
}
```

### Output DTO

```go
type GenerateImageResponse struct {
    Images []GeneratedImage `json:"images"`
    Model  string           `json:"model"`
}

type GeneratedImage struct {
    Base64Data string `json:"base64Data"`
    MimeType   string `json:"mimeType"`
}
```

---

## Usage

```go
// Di router.go
genaiClient := config.ConnectGoogleGenAI(cfg)
genaiSvc := google_genai.NewService(genaiClient)

// Di service/handler lain
result, err := genaiSvc.GenerateText(ctx, google_genai.GenerateTextInput{
    Model:  "gemini-1.5-flash",
    Prompt: "Write a poem about AI",
})
```

---

## Directory

- `config/google_gen_ai.go` - Client connection
- `internal/module/headless/google_genai/dto.go` - Input/Output DTOs
- `internal/module/headless/google_genai/service.go` - Service implementation
