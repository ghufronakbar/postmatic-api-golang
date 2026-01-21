# Headless.OpenAI

Module headless untuk integrasi dengan OpenAI API.

## Packages

- `github.com/openai/openai-go`

## Config Variables

| Variable         | Description                  |
| ---------------- | ---------------------------- |
| `OPENAI_API_KEY` | API Key dari OpenAI Platform |

## Client Setup

File: `config/openai.go`

```go
client := config.ConnectOpenAI(cfg)
```

---

## Service Interface

```go
type Service interface {
    // Chat Completions
    GenerateText(ctx context.Context, input GenerateTextInput) (*GenerateTextResponse, error)

    // Image Generation (DALL-E)
    GenerateImage(ctx context.Context, input GenerateImageInput) (*GenerateImageResponse, error)
}
```

---

## GenerateText

Generate text menggunakan Chat Completions API (gpt-4o, gpt-4-turbo, gpt-3.5-turbo, dll).

### Input DTO

```go
type GenerateTextInput struct {
    Model            string        `json:"model"`            // Required: gpt-4o, dll
    Messages         []ChatMessage `json:"messages"`         // Required
    Temperature      *float64      `json:"temperature"`      // Optional: 0.0 - 2.0
    MaxTokens        *int          `json:"maxTokens"`        // Optional
    TopP             *float64      `json:"topP"`             // Optional
    FrequencyPenalty *float64      `json:"frequencyPenalty"` // Optional: -2.0 to 2.0
    PresencePenalty  *float64      `json:"presencePenalty"`  // Optional: -2.0 to 2.0
}

type ChatMessage struct {
    Role    string `json:"role"`    // system, user, assistant
    Content string `json:"content"`
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

Generate image menggunakan DALL-E API (dall-e-2, dall-e-3).

### Input DTO

```go
type GenerateImageInput struct {
    Model   string  `json:"model"`   // Required: dall-e-2, dall-e-3
    Prompt  string  `json:"prompt"`  // Required
    N       *int    `json:"n"`       // Optional: 1-10 (dall-e-2), 1 (dall-e-3)
    Size    *string `json:"size"`    // Optional: 256x256, 512x512, 1024x1024, etc.
    Quality *string `json:"quality"` // Optional: standard, hd (dall-e-3 only)
    Style   *string `json:"style"`   // Optional: vivid, natural (dall-e-3 only)
}
```

### Output DTO

```go
type GenerateImageResponse struct {
    Images []GeneratedImage `json:"images"`
    Model  string           `json:"model"`
}

type GeneratedImage struct {
    URL           string `json:"url"`
    RevisedPrompt string `json:"revisedPrompt"`
}
```

---

## Usage

```go
// Di router.go
openaiClient := config.ConnectOpenAI(cfg)
openaiSvc := openai_svc.NewService(openaiClient)

// Di service/handler lain
result, err := openaiSvc.GenerateText(ctx, openai_svc.GenerateTextInput{
    Model: "gpt-4o",
    Messages: []openai_svc.ChatMessage{
        {Role: "system", Content: "You are a helpful assistant."},
        {Role: "user", Content: "Write a poem about AI"},
    },
})
```

---

## Directory

- `config/openai.go` - Client connection
- `internal/module/headless/openai/dto.go` - Input/Output DTOs
- `internal/module/headless/openai/service.go` - Service implementation
