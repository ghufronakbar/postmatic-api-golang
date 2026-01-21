# Module Headless.GoogleGenAI

Modul ini bertanggung jawab untuk integrasi dengan Google Generative AI (Gemini API & Imagen). Modul ini **headless** (tidak dipanggil via HTTP Handler langsung).

## 1. Project Rules & Dependencies

- **Library**: [`google.golang.org/genai`](https://pkg.go.dev/google.golang.org/genai)
- **Backend**: Gemini API (`genai.BackendGeminiAPI`), bukan Vertex AI
- **Scope**: Text Generation (Gemini models) dan Image Generation (Imagen models)
- **Headless**: Modul ini hanya dipanggil oleh module lain (internal)
- **DTO Wrapper**: Semua input dan output menggunakan DTO internal

## 2. Directory Structure

```text
internal/module/headless/google_genai/
├── service.go     # Service implementation
├── dto.go         # Input DTOs
└── viewmodel.go   # Output DTOs / Response types
```

## 3. Configuration

File: `config/google_gen_ai.go`

| Variable               | Type   | Description                   |
| ---------------------- | ------ | ----------------------------- |
| `GOOGLE_GENAI_API_KEY` | String | API Key dari Google AI Studio |

```go
client := config.ConnectGoogleGenAI(cfg)
```

**Important**: Menggunakan `genai.BackendGeminiAPI` (bukan Vertex AI).

## 4. Service Interface

```go
type Service interface {
    // Text Generation (Gemini models)
    GenerateText(ctx context.Context, input GenerateTextInput) (*GenerateTextResponse, error)

    // Image Generation (Imagen models)
    GenerateImage(ctx context.Context, input GenerateImageInput) (*GenerateImageResponse, error)
}
```

## 5. Method: GenerateText

Generate text menggunakan Gemini models.

### Supported Models

| Model              | Description          | Context   |
| ------------------ | -------------------- | --------- |
| `gemini-1.5-pro`   | Most capable         | 2M tokens |
| `gemini-1.5-flash` | Fast and efficient   | 1M tokens |
| `gemini-1.0-pro`   | Balanced performance | 32k       |
| `gemini-2.0-flash` | Latest fast model    | 1M tokens |

### Input DTO

```go
type GenerateTextInput struct {
    Model           string   `json:"model"`           // Required: gemini-1.5-pro, dll
    Prompt          string   `json:"prompt"`          // Required: text prompt
    Temperature     *float64 `json:"temperature"`     // Optional: 0.0 - 2.0
    MaxOutputTokens *int     `json:"maxOutputTokens"` // Optional: max tokens
    TopP            *float64 `json:"topP"`            // Optional: nucleus sampling
    TopK            *int     `json:"topK"`            // Optional: top-k sampling
}
```

### Business Logic

```
1. Get model from client:
   - model := client.GenerativeModel(input.Model)

2. Configure generation params (if provided):
   - Temperature
   - MaxOutputTokens
   - TopP
   - TopK (converted to *float32)

3. Generate content:
   - response := model.GenerateContent(ctx, genai.Text(input.Prompt))

4. Handle response:
   ├── Error → Return wrapped error
   └── Success → Extract:
       - Text from candidates[0].content.parts[0]
       - Token usage from metadata
       - Finish reason

5. Return GenerateTextResponse
```

### Output DTO

```go
type GenerateTextResponse struct {
    Text             string `json:"text"`             // Generated text
    Model            string `json:"model"`            // Model used
    PromptTokenCount int    `json:"promptTokenCount"` // Input tokens
    OutputTokenCount int    `json:"outputTokenCount"` // Output tokens
    TotalTokenCount  int    `json:"totalTokenCount"`  // Total tokens
    FinishReason     string `json:"finishReason"`     // STOP, MAX_TOKENS, etc.
}
```

## 6. Method: GenerateImage

Generate images menggunakan Imagen models.

### Supported Models

| Model                     | Description      | Aspect Ratios             |
| ------------------------- | ---------------- | ------------------------- |
| `imagen-3.0-generate-002` | Latest Imagen    | 1:1, 16:9, 9:16, 4:3, 3:4 |
| `imagen-3.0-generate-001` | Previous version | 1:1, 16:9, 9:16           |

### Input DTO

```go
type GenerateImageInput struct {
    Model          string  `json:"model"`          // Required: imagen-3.0-generate-002
    Prompt         string  `json:"prompt"`         // Required: image description
    NumberOfImages *int    `json:"numberOfImages"` // Optional: 1-4
    AspectRatio    *string `json:"aspectRatio"`    // Optional: "1:1", "16:9", "9:16"
}
```

### Business Logic

```
1. Get imagen model:
   - model := client.ImagenModel(input.Model)

2. Configure params:
   - NumberOfImages (converted to int32)
   - AspectRatio

3. Generate images:
   - response := model.GenerateImages(ctx, input.Prompt, config)

4. Handle response:
   ├── Error → Return wrapped error
   └── Success → Extract:
       - Base64 data for each generated image
       - MIME type

5. Return GenerateImageResponse with image list
```

### Output DTO

```go
type GenerateImageResponse struct {
    Images []GeneratedImage `json:"images"`
    Model  string           `json:"model"`
}

type GeneratedImage struct {
    Base64Data string `json:"base64Data"` // Base64 encoded image
    MimeType   string `json:"mimeType"`   // image/png, image/jpeg
}
```

## 7. Usage Example

```go
// Di router.go
genaiClient := config.ConnectGoogleGenAI(cfg)
genaiSvc := google_genai.NewService(genaiClient)

// Generate text
result, err := genaiSvc.GenerateText(ctx, google_genai.GenerateTextInput{
    Model:       "gemini-1.5-flash",
    Prompt:      "Write a poem about AI",
    Temperature: ptr(0.7),
})
fmt.Println(result.Text)

// Generate image
imgResult, err := genaiSvc.GenerateImage(ctx, google_genai.GenerateImageInput{
    Model:          "imagen-3.0-generate-002",
    Prompt:         "A futuristic city with flying cars",
    NumberOfImages: ptr(1),
    AspectRatio:    ptr("16:9"),
})

// Decode base64 dan simpan
imgBytes, _ := base64.StdEncoding.DecodeString(imgResult.Images[0].Base64Data)
os.WriteFile("image.png", imgBytes, 0644)
```

## 8. Error Handling

| Error Type            | Condition                         |
| --------------------- | --------------------------------- |
| `InvalidArgument`     | Invalid model or params           |
| `ResourceExhausted`   | Quota exceeded                    |
| `PermissionDenied`    | Invalid API key                   |
| `SafetyBlocked`       | Content blocked by safety filters |
| `InternalServerError` | Google API server error           |

## 9. Image vs OpenAI

| Feature           | Google GenAI (Imagen) | OpenAI (DALL-E)             |
| ----------------- | --------------------- | --------------------------- |
| **Output Format** | Base64                | URL (temporary)             |
| **Aspect Ratios** | 5 options             | 3 sizes                     |
| **Max Images**    | 4                     | 10 (DALL-E 2), 1 (DALL-E 3) |
| **Persistence**   | Direct embed          | Need download               |

## 10. Design Decisions

### Kenapa Gemini API (bukan Vertex AI)?

1. **Simpler setup**: Hanya butuh API key
2. **No GCP project**: Tidak perlu setup GCP
3. **Direct access**: API sama dengan Google AI Studio
4. **Cost**: Pay-per-use tanpa infrastructure cost

### Kenapa Base64 untuk Image?

Imagen mengembalikan Base64 (bukan URL seperti DALL-E):

1. **No expiry**: Tidak ada URL expiration
2. **Direct embed**: Bisa langsung embed di response
3. **Immediate persist**: Upload langsung ke storage tanpa download step

### Kenapa DTO Wrapper?

1. **Abstraksi SDK**: Mudah migrasi jika SDK berubah
2. **Type safety**: TopK di SDK adalah `*float32`, DTO pakai `*int`
3. **Validation**: Input DTO bisa punya validation tags
4. **Testing**: Mudah mock interface tanpa depend on SDK types

### Kenapa Headless?

1. Module ini tidak perlu HTTP endpoint sendiri
2. Dipanggil oleh module lain (e.g., Content Generator)
3. Separation of concerns - AI logic terpisah dari business logic
