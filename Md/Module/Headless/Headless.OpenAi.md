# Module Headless.OpenAI

Modul ini bertanggung jawab untuk integrasi dengan OpenAI API (Chat Completions & DALL-E Image Generation). Modul ini **headless** (tidak dipanggil via HTTP Handler langsung).

## 1. Project Rules & Dependencies

- **Library**: [`github.com/openai/openai-go/v3`](https://github.com/openai/openai-go)
- **Scope**: Chat Completions (GPT-4o, GPT-4-turbo, GPT-3.5-turbo) dan Image Generation (DALL-E 2, DALL-E 3)
- **Headless**: Modul ini hanya dipanggil oleh module lain (internal)
- **DTO Wrapper**: Semua input dan output menggunakan DTO internal

## 2. Directory Structure

```text
internal/module/headless/openai/
├── service.go     # Service implementation
├── dto.go         # Input DTOs
└── viewmodel.go   # Output DTOs / Response types
```

## 3. Configuration

File: `config/openai.go`

| Variable         | Type   | Description                  |
| ---------------- | ------ | ---------------------------- |
| `OPENAI_API_KEY` | String | API Key dari OpenAI Platform |

```go
client := config.ConnectOpenAI(cfg)
```

## 4. Service Interface

```go
type Service interface {
    // Chat Completions (GPT models)
    GenerateText(ctx context.Context, input GenerateTextInput) (*GenerateTextResponse, error)

    // Image Generation (DALL-E models)
    GenerateImage(ctx context.Context, input GenerateImageInput) (*GenerateImageResponse, error)
}
```

## 5. Method: GenerateText

Generate text menggunakan Chat Completions API.

### Supported Models

| Model           | Description             | Context |
| --------------- | ----------------------- | ------- |
| `gpt-4o`        | Latest GPT-4 multimodal | 128k    |
| `gpt-4-turbo`   | GPT-4 Turbo with vision | 128k    |
| `gpt-4`         | Original GPT-4          | 8k      |
| `gpt-3.5-turbo` | Fast and cost-effective | 16k     |

### Input DTO

```go
type GenerateTextInput struct {
    Model            string        `json:"model"`            // Required: gpt-4o, dll
    Messages         []ChatMessage `json:"messages"`         // Required: conversation
    Temperature      *float64      `json:"temperature"`      // Optional: 0.0 - 2.0
    MaxTokens        *int          `json:"maxTokens"`        // Optional: max output tokens
    TopP             *float64      `json:"topP"`             // Optional: nucleus sampling
    FrequencyPenalty *float64      `json:"frequencyPenalty"` // Optional: -2.0 to 2.0
    PresencePenalty  *float64      `json:"presencePenalty"`  // Optional: -2.0 to 2.0
}

type ChatMessage struct {
    Role    string `json:"role"`    // system, user, assistant
    Content string `json:"content"`
}
```

### Business Logic

```
1. Build chat completion request:
   - Messages: Convert []ChatMessage to SDK format
   - Model: Use specified model
   - Optional params: Temperature, MaxTokens, TopP, etc.

2. Call OpenAI API:
   - client.Chat.Completions.New(ctx, params)

3. Handle response:
   ├── Error → Return wrapped error
   └── Success → Extract:
       - Text from choices[0].message.content
       - Token usage counts
       - Finish reason

4. Return GenerateTextResponse
```

### Output DTO

```go
type GenerateTextResponse struct {
    Text             string `json:"text"`             // Generated text
    Model            string `json:"model"`            // Model used
    PromptTokenCount int    `json:"promptTokenCount"` // Input tokens
    OutputTokenCount int    `json:"outputTokenCount"` // Output tokens
    TotalTokenCount  int    `json:"totalTokenCount"`  // Total tokens
    FinishReason     string `json:"finishReason"`     // stop, length, etc.
}
```

## 6. Method: GenerateImage

Generate images menggunakan DALL-E API.

### Supported Models

| Model      | Description             | Sizes                           | Max N |
| ---------- | ----------------------- | ------------------------------- | ----- |
| `dall-e-3` | Latest, highest quality | 1024x1024, 1792x1024, 1024x1792 | 1     |
| `dall-e-2` | Faster, more sizes      | 256x256, 512x512, 1024x1024     | 10    |

### Input DTO

```go
type GenerateImageInput struct {
    Model   string  `json:"model"`   // Required: dall-e-2, dall-e-3
    Prompt  string  `json:"prompt"`  // Required: image description
    N       *int    `json:"n"`       // Optional: 1-10 (dall-e-2), 1 (dall-e-3)
    Size    *string `json:"size"`    // Optional: 256x256, 512x512, 1024x1024, etc.
    Quality *string `json:"quality"` // Optional: standard, hd (dall-e-3 only)
    Style   *string `json:"style"`   // Optional: vivid, natural (dall-e-3 only)
}
```

### Business Logic

```
1. Build image generation request:
   - Model: Validate dall-e-2 or dall-e-3
   - Prompt: Required, max 1000 chars (dall-e-2) or 4000 chars (dall-e-3)
   - Optional: N, Size, Quality, Style

2. Call OpenAI API:
   - client.Images.Generate(ctx, params)

3. Handle response:
   ├── Error → Return wrapped error
   └── Success → Extract image URLs

4. Return GenerateImageResponse with image list
```

### Output DTO

```go
type GenerateImageResponse struct {
    Images []GeneratedImage `json:"images"`
    Model  string           `json:"model"`
}

type GeneratedImage struct {
    URL           string `json:"url"`           // Temporary URL (valid ~1 hour)
    RevisedPrompt string `json:"revisedPrompt"` // DALL-E 3 revised prompt
}
```

## 7. Usage Example

```go
// Di router.go
openaiClient := config.ConnectOpenAI(cfg)
openaiSvc := openai_svc.NewService(openaiClient)

// Generate text
result, err := openaiSvc.GenerateText(ctx, openai_svc.GenerateTextInput{
    Model: "gpt-4o",
    Messages: []openai_svc.ChatMessage{
        {Role: "system", Content: "You are a helpful assistant."},
        {Role: "user", Content: "Write a poem about AI"},
    },
    Temperature: ptr(0.7),
})
fmt.Println(result.Text)

// Generate image
imgResult, err := openaiSvc.GenerateImage(ctx, openai_svc.GenerateImageInput{
    Model:  "dall-e-3",
    Prompt: "A futuristic city with flying cars",
    Size:   ptr("1024x1024"),
})
fmt.Println(imgResult.Images[0].URL)
```

## 8. Error Handling

| Error Type               | Condition                      |
| ------------------------ | ------------------------------ |
| `InvalidRequestError`    | Invalid params or prompt       |
| `RateLimitError`         | Too many requests              |
| `InsufficientQuotaError` | No credits remaining           |
| `ContentPolicyViolation` | Prompt violates content policy |
| `InternalServerError`    | OpenAI server error            |

## 9. Design Decisions

### Kenapa DTO Wrapper?

1. **Abstraksi SDK**: Mudah migrasi jika SDK berubah
2. **Kontrol Types**: Hanya expose fields yang diperlukan
3. **Validation**: Input DTO bisa punya validation tags
4. **Testing**: Mudah mock interface tanpa depend on SDK types

### Kenapa Headless?

1. Module ini tidak perlu HTTP endpoint sendiri
2. Dipanggil oleh module lain (e.g., Content Generator)
3. Separation of concerns - AI logic terpisah dari business logic

### Image URL Expiry

DALL-E image URLs expire setelah ~1 jam. Jika perlu persist:

1. Download image setelah generate
2. Upload ke CloudinaryUploader atau S3Uploader
3. Simpan permanent URL
