// internal/module/headless/openai/service.go
package openai_svc

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

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

// GenerateImage generates images using DALL-E or gpt-image-1
// If ReferenceImageURLs are provided, uses Images.Edit API for image editing
func (s *openaiService) GenerateImage(ctx context.Context, input GenerateImageInput) (*GenerateImageResponse, error) {
	log := logger.From(ctx)

	// If reference images are provided, use Images.Edit for gpt-image-1
	if len(input.ReferenceImageURLs) > 0 {
		return s.generateImageWithEdit(ctx, input)
	}

	// Standard image generation without references
	log.Info("Generating image (no reference)", "model", input.Model)

	// Build request params
	params := openai.ImageGenerateParams{
		Model:  openai.ImageModel(input.Model),
		Prompt: input.Prompt,
	}

	if input.N != nil {
		params.N = param.NewOpt(int64(*input.N))
	}
	// if input.Size != nil {
	// 	params.Size = openai.ImageGenerateParamsSize(*input.Size)
	// } else {
	// 	params.Size = openai.ImageGenerateParamsSize("1024x1024")
	// }
	params.Size = openai.ImageGenerateParamsSize("auto")
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
			Base64Data:    img.B64JSON,
			MimeType:      "image/png",
			RevisedPrompt: img.RevisedPrompt,
		}
	}

	log.Info("Images generated successfully", "model", input.Model, "count", len(images))

	return &GenerateImageResponse{
		Images: images,
		Model:  input.Model,
	}, nil
}

// generateImageWithEdit uses Images.Edit API for gpt-image-1 with reference images
func (s *openaiService) generateImageWithEdit(ctx context.Context, input GenerateImageInput) (*GenerateImageResponse, error) {
	log := logger.From(ctx)
	log.Info("[OPENAI] Generating image with references via Images.Edit",
		"model", input.Model,
		"refCount", len(input.ReferenceImageURLs),
		"refURLs", input.ReferenceImageURLs)

	// Download reference images
	var imageReaders []io.Reader
	for i, url := range input.ReferenceImageURLs {
		log.Info("[OPENAI] Downloading reference image", "index", i, "url", url)

		resp, err := http.Get(url)
		if err != nil {
			log.Error("[OPENAI] Failed to download reference image", "url", url, "error", err)
			continue
		}

		// Read body into buffer
		data, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			log.Error("[OPENAI] Failed to read image body", "url", url, "error", err)
			continue
		}

		// Sniff MIME type from bytes and create named reader with proper extension and content type
		mimeType := sniffMimeType(data)
		ext := mimeToExtension(mimeType)
		namedFile := &namedReader{
			Reader:      bytes.NewReader(data),
			name:        fmt.Sprintf("image_%d%s", i, ext),
			contentType: mimeType,
		}

		imageReaders = append(imageReaders, namedFile)
		log.Info("[OPENAI] Added reference image", "index", i, "size", len(data), "mimeType", mimeType, "filename", namedFile.name)
	}

	if len(imageReaders) == 0 {
		log.Error("[OPENAI] No reference images could be downloaded")
		return nil, errs.NewBadRequest("OPENAI_NO_REFERENCE_IMAGES")
	}

	// Build edit params
	size := "1024x1024"
	// if input.Size != nil {
	// 	size = *input.Size
	// }

	// Create ImageEditParamsImageUnion with OfFileArray for multiple images
	imageUnion := openai.ImageEditParamsImageUnion{
		OfFileArray: imageReaders,
	}

	params := openai.ImageEditParams{
		Model:   openai.ImageModel(input.Model),
		Prompt:  input.Prompt,
		Image:   imageUnion,
		Size:    openai.ImageEditParamsSize(size),
		Quality: openai.ImageEditParamsQualityHigh,
	}

	if input.N != nil {
		params.N = param.NewOpt(int64(*input.N))
	}

	log.Info("[OPENAI] Calling Images.Edit API", "model", input.Model, "imageCount", len(imageReaders), "prompt_length", len(input.Prompt))

	// Make request
	result, err := s.client.Images.Edit(ctx, params)
	if err != nil {
		log.Error("[OPENAI] Failed to edit/generate image", "model", input.Model, "error", err)
		return nil, errs.NewBadRequest("OPENAI_EDIT_IMAGE_FAILED")
	}

	// Map images to response
	images := make([]GeneratedImage, len(result.Data))
	for i, img := range result.Data {
		images[i] = GeneratedImage{
			URL:           img.URL,
			Base64Data:    img.B64JSON,
			MimeType:      "image/png",
			RevisedPrompt: img.RevisedPrompt,
		}
	}

	log.Info("[OPENAI] Images generated successfully via Edit API", "model", input.Model, "count", len(images))

	return &GenerateImageResponse{
		Images: images,
		Model:  input.Model,
	}, nil
}

// namedReader wraps an io.Reader with filename and contentType for OpenAI SDK
// OpenAI SDK checks for Name() and ContentType() interfaces to determine filename and MIME type
type namedReader struct {
	io.Reader
	name        string
	contentType string
}

// Name returns the filename (used by SDK for Content-Disposition header)
func (n *namedReader) Name() string {
	return n.name
}

// ContentType returns the MIME type (used by SDK for Content-Type header)
func (n *namedReader) ContentType() string {
	return n.contentType
}

// sniffMimeType detects MIME type from magic bytes (like TypeScript sniffImageMime)
func sniffMimeType(data []byte) string {
	if len(data) < 12 {
		return "application/octet-stream"
	}

	// PNG: 89 50 4E 47 0D 0A 1A 0A
	if data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 &&
		data[4] == 0x0D && data[5] == 0x0A && data[6] == 0x1A && data[7] == 0x0A {
		return "image/png"
	}

	// JPEG: FF D8 FF
	if data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
		return "image/jpeg"
	}

	// WEBP: RIFF....WEBP
	if data[0] == 0x52 && data[1] == 0x49 && data[2] == 0x46 && data[3] == 0x46 &&
		data[8] == 0x57 && data[9] == 0x45 && data[10] == 0x42 && data[11] == 0x50 {
		return "image/webp"
	}

	return "application/octet-stream"
}

// mimeToExtension converts MIME type to file extension
func mimeToExtension(mimeType string) string {
	switch mimeType {
	case "image/png":
		return ".png"
	case "image/jpeg":
		return ".jpg"
	case "image/webp":
		return ".webp"
	default:
		return ".png" // Default to PNG
	}
}
