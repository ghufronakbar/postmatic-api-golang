// internal/module/headless/google_genai/service.go
package google_genai

import (
	"context"

	"postmatic-api/pkg/errs"
	"postmatic-api/pkg/logger"

	"encoding/base64"
	"fmt"

	"google.golang.org/genai"
)

// Service defines the contract for Google GenAI interactions
type Service interface {
	// Text Generation
	GenerateText(ctx context.Context, input GenerateTextInput) (*GenerateTextResponse, error)

	// Image Generation (Imagen)
	GenerateImage(ctx context.Context, input GenerateImageInput) (*GenerateImageResponse, error)
}

// googleGenAIService implements the Service interface
type googleGenAIService struct {
	client *genai.Client
}

// NewService creates a new Google GenAI service instance
func NewService(client *genai.Client) Service {
	return &googleGenAIService{
		client: client,
	}
}

// GenerateText generates text using specified model
func (s *googleGenAIService) GenerateText(ctx context.Context, input GenerateTextInput) (*GenerateTextResponse, error) {
	log := logger.From(ctx)
	log.Info("Generating text", "model", input.Model)

	// Build generation config
	config := &genai.GenerateContentConfig{}
	if input.Temperature != nil {
		temp := float32(*input.Temperature)
		config.Temperature = &temp
	}
	if input.MaxOutputTokens != nil {
		config.MaxOutputTokens = int32(*input.MaxOutputTokens)
	}
	if input.TopP != nil {
		topP := float32(*input.TopP)
		config.TopP = &topP
	}
	if input.TopK != nil {
		topK := float32(*input.TopK)
		config.TopK = &topK
	}

	// Generate content
	result, err := s.client.Models.GenerateContent(ctx, input.Model, genai.Text(input.Prompt), config)
	if err != nil {
		log.Error("Failed to generate text", "model", input.Model, "error", err)
		return nil, errs.NewBadRequest("GOOGLE_GENAI_GENERATE_TEXT_FAILED")
	}

	// Extract text from response
	var text string
	var finishReason string
	if len(result.Candidates) > 0 && result.Candidates[0].Content != nil {
		for _, part := range result.Candidates[0].Content.Parts {
			if part.Text != "" {
				text += part.Text
			}
		}
		finishReason = string(result.Candidates[0].FinishReason)
	}

	// Token counts
	var promptTokens, outputTokens, totalTokens int
	if result.UsageMetadata != nil {
		promptTokens = int(result.UsageMetadata.PromptTokenCount)
		outputTokens = int(result.UsageMetadata.CandidatesTokenCount)
		totalTokens = int(result.UsageMetadata.TotalTokenCount)
	}

	log.Info("Text generated successfully", "model", input.Model, "totalTokens", totalTokens)

	return &GenerateTextResponse{
		Text:             text,
		Model:            input.Model,
		PromptTokenCount: promptTokens,
		OutputTokenCount: outputTokens,
		TotalTokenCount:  totalTokens,
		FinishReason:     finishReason,
	}, nil
}

// GenerateImage generates images using Gemini Multimodal model (via GenerateContent)
func (s *googleGenAIService) GenerateImage(ctx context.Context, input GenerateImageInput) (*GenerateImageResponse, error) {
	log := logger.From(ctx)
	log.Info("Generating image via GenerateContent", "model", input.Model)

	// 1. Build Prompt & Config
	// Gemini menangani Aspect Ratio lebih baik via instruksi prompt daripada config parameter
	finalPrompt := input.Prompt
	if input.AspectRatio != nil {
		finalPrompt = fmt.Sprintf("%s. Aspect ratio: %s", input.Prompt, *input.AspectRatio)
	}

	config := &genai.GenerateContentConfig{}

	// Mapping NumberOfImages ke CandidateCount
	// Note: Tidak semua model Gemini support > 1 candidate.
	// Jika error, mungkin perlu di-force ke 1 atau looping request.
	if input.NumberOfImages != nil {
		config.CandidateCount = int32(*input.NumberOfImages)
	}

	// 2. Call GenerateContent
	// Kita mengirim Text Prompt, tapi mengharapkan balasan berupa Image (Multimodal generation)
	resp, err := s.client.Models.GenerateContent(ctx, input.Model, genai.Text(finalPrompt), config)
	if err != nil {
		log.Error("Failed to generate content", "model", input.Model, "error", err)
		return nil, errs.NewBadRequest("GOOGLE_GENAI_GENERATE_FAILED")
	}

	// 3. Map Response (Parsing InlineData)
	images := make([]GeneratedImage, 0)

	// Loop semua candidates (jawaban alternatif)
	for _, candidate := range resp.Candidates {
		if candidate.Content == nil {
			continue
		}

		// Loop semua parts dalam satu jawaban
		for _, part := range candidate.Content.Parts {
			// Cek apakah part ini adalah InlineData (Gambar/Binary)
			if part.InlineData != nil {
				// Encode raw bytes ke Base64 String agar aman untuk JSON response
				encodedString := base64.StdEncoding.EncodeToString(part.InlineData.Data)

				images = append(images, GeneratedImage{
					Base64Data: encodedString,
					MimeType:   part.InlineData.MIMEType,
				})
			}
		}
	}

	// Validasi apakah ada gambar yang dihasilkan
	if len(images) == 0 {
		log.Warn("Model returned success but no images found in candidates", "model", input.Model)
		// Opsional: Cek apakah model me-reject request (Safety Ratings)
		return nil, errs.NewBadRequest("IMAGE_GENERATION_EMPTY_RESPONSE")
	}

	log.Info("Images generated successfully", "model", input.Model, "count", len(images))

	return &GenerateImageResponse{
		Images: images,
		Model:  input.Model,
	}, nil
}
