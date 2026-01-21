// internal/module/headless/openai/service.go
package openai_svc

import (
	"context"

	"postmatic-api/pkg/errs"
	"postmatic-api/pkg/logger"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/packages/param"
	"github.com/openai/openai-go/v3/shared"
)

// Service defines the contract for OpenAI interactions
type Service interface {
	// Chat Completions
	GenerateText(ctx context.Context, input GenerateTextInput) (*GenerateTextResponse, error)

	// Image Generation (DALL-E)
	GenerateImage(ctx context.Context, input GenerateImageInput) (*GenerateImageResponse, error)
}

// openaiService implements the Service interface
type openaiService struct {
	client openai.Client
}

// NewService creates a new OpenAI service instance
func NewService(client openai.Client) Service {
	return &openaiService{
		client: client,
	}
}

// GenerateText generates text using chat completions
func (s *openaiService) GenerateText(ctx context.Context, input GenerateTextInput) (*GenerateTextResponse, error) {
	log := logger.From(ctx)
	log.Info("Generating text", "model", input.Model)

	// Convert messages to OpenAI format
	messages := make([]openai.ChatCompletionMessageParamUnion, len(input.Messages))
	for i, msg := range input.Messages {
		switch msg.Role {
		case "system":
			messages[i] = openai.SystemMessage(msg.Content)
		case "user":
			messages[i] = openai.UserMessage(msg.Content)
		case "assistant":
			messages[i] = openai.AssistantMessage(msg.Content)
		default:
			messages[i] = openai.UserMessage(msg.Content)
		}
	}

	// Build request params
	params := openai.ChatCompletionNewParams{
		Model:    shared.ChatModel(input.Model),
		Messages: messages,
	}

	if input.Temperature != nil {
		params.Temperature = param.NewOpt(*input.Temperature)
	}
	if input.MaxTokens != nil {
		params.MaxTokens = param.NewOpt(int64(*input.MaxTokens))
	}
	if input.TopP != nil {
		params.TopP = param.NewOpt(*input.TopP)
	}
	if input.FrequencyPenalty != nil {
		params.FrequencyPenalty = param.NewOpt(*input.FrequencyPenalty)
	}
	if input.PresencePenalty != nil {
		params.PresencePenalty = param.NewOpt(*input.PresencePenalty)
	}

	// Make request
	result, err := s.client.Chat.Completions.New(ctx, params)
	if err != nil {
		log.Error("Failed to generate text", "model", input.Model, "error", err)
		return nil, errs.NewBadRequest("OPENAI_GENERATE_TEXT_FAILED")
	}

	// Extract response
	var text string
	var finishReason string
	if len(result.Choices) > 0 {
		text = result.Choices[0].Message.Content
		finishReason = string(result.Choices[0].FinishReason)
	}

	log.Info("Text generated successfully", "model", input.Model, "totalTokens", result.Usage.TotalTokens)

	return &GenerateTextResponse{
		Text:             text,
		Model:            result.Model,
		PromptTokenCount: int(result.Usage.PromptTokens),
		OutputTokenCount: int(result.Usage.CompletionTokens),
		TotalTokenCount:  int(result.Usage.TotalTokens),
		FinishReason:     finishReason,
	}, nil
}

// GenerateImage generates images using DALL-E
func (s *openaiService) GenerateImage(ctx context.Context, input GenerateImageInput) (*GenerateImageResponse, error) {
	log := logger.From(ctx)
	log.Info("Generating image", "model", input.Model)

	// Build request params
	params := openai.ImageGenerateParams{
		Model:  openai.ImageModel(input.Model),
		Prompt: input.Prompt,
	}

	if input.N != nil {
		params.N = param.NewOpt(int64(*input.N))
	}
	if input.Size != nil {
		params.Size = openai.ImageGenerateParamsSize(*input.Size)
	}
	if input.Quality != nil {
		params.Quality = openai.ImageGenerateParamsQuality(*input.Quality)
	}
	if input.Style != nil {
		params.Style = openai.ImageGenerateParamsStyle(*input.Style)
	}

	// Make request
	result, err := s.client.Images.Generate(ctx, params)
	if err != nil {
		log.Error("Failed to generate image", "model", input.Model, "error", err)
		return nil, errs.NewBadRequest("OPENAI_GENERATE_IMAGE_FAILED")
	}

	// Map images to response
	images := make([]GeneratedImage, len(result.Data))
	for i, img := range result.Data {
		images[i] = GeneratedImage{
			URL:           img.URL,
			RevisedPrompt: img.RevisedPrompt,
		}
	}

	log.Info("Images generated successfully", "model", input.Model, "count", len(images))

	return &GenerateImageResponse{
		Images: images,
		Model:  input.Model,
	}, nil
}
