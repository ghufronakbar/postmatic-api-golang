// internal/module/headless/google_genai/service.go
package google_genai

import (
	"context"

	"postmatic-api/pkg/errs"
	"postmatic-api/pkg/logger"

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

// GenerateImage generates images using Imagen model
func (s *googleGenAIService) GenerateImage(ctx context.Context, input GenerateImageInput) (*GenerateImageResponse, error) {
	log := logger.From(ctx)
	log.Info("Generating image", "model", input.Model)

	// Build generation config
	config := &genai.GenerateImagesConfig{}
	if input.NumberOfImages != nil {
		config.NumberOfImages = int32(*input.NumberOfImages)
	}
	if input.AspectRatio != nil {
		config.AspectRatio = *input.AspectRatio
	}

	// Generate images
	result, err := s.client.Models.GenerateImages(ctx, input.Model, input.Prompt, config)
	if err != nil {
		log.Error("Failed to generate image", "model", input.Model, "error", err)
		return nil, errs.NewBadRequest("GOOGLE_GENAI_GENERATE_IMAGE_FAILED")
	}

	// Map images to response
	images := make([]GeneratedImage, 0)
	for _, img := range result.GeneratedImages {
		if img.Image != nil {
			images = append(images, GeneratedImage{
				Base64Data: string(img.Image.ImageBytes),
				MimeType:   img.Image.MIMEType,
			})
		}
	}

	log.Info("Images generated successfully", "model", input.Model, "count", len(images))

	return &GenerateImageResponse{
		Images: images,
		Model:  input.Model,
	}, nil
}
