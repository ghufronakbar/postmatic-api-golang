// internal/module/generative_content/image_post/service/worker.go
package image_post_service

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"io"
	"net/http"

	"postmatic-api/config"
	image_token_service "postmatic-api/internal/module/generative_token/image_token/service"
	"postmatic-api/internal/module/headless/cloudinary_uploader"
	"postmatic-api/internal/module/headless/google_genai"
	openai_svc "postmatic-api/internal/module/headless/openai"
	"postmatic-api/internal/repository/entity"
	"postmatic-api/pkg/logger"
)

// ImagePostWorkerHandler handles queue jobs for image post generation
type ImagePostWorkerHandler struct {
	cfg            *config.Config
	store          entity.Store
	openaiSvc      openai_svc.Service
	googleGenAISvc google_genai.Service
	cloudinarySvc  *cloudinary_uploader.CloudinaryUploaderService
	imageTokenSvc  *image_token_service.ImageTokenService
}

// NewWorkerHandler creates a new ImagePostWorkerHandler
func NewWorkerHandler(
	cfg *config.Config,
	store entity.Store,
	openaiSvc openai_svc.Service,
	googleGenAISvc google_genai.Service,
	cloudinarySvc *cloudinary_uploader.CloudinaryUploaderService,
	imageTokenSvc *image_token_service.ImageTokenService,
) *ImagePostWorkerHandler {
	return &ImagePostWorkerHandler{
		cfg:            cfg,
		store:          store,
		openaiSvc:      openaiSvc,
		googleGenAISvc: googleGenAISvc,
		cloudinarySvc:  cloudinarySvc,
		imageTokenSvc:  imageTokenSvc,
	}
}

// ProcessImagePostJob processes the image post job (implements queue.ImagePostHandler)
func (w *ImagePostWorkerHandler) ProcessImagePostJob(ctx context.Context, postID int64, businessRootID int64) error {
	log := logger.L().With("postID", postID, "businessRootID", businessRootID)
	log.Info("[QUEUE] Starting ProcessImagePostJob")

	// 1. Get post from DB
	post, err := w.store.GetGeneratedImagePostById(ctx, postID)
	if err != nil {
		log.Error("[QUEUE] Failed to get post", "error", err)
		return err
	}
	log.Info("[QUEUE] Got post from DB", "status", post.Status, "mode", post.Mode, "provider", post.RecordedModelProvider)

	// 2. Update status to processing
	err = w.store.UpdateGeneratedImagePostStatus(ctx, entity.UpdateGeneratedImagePostStatusParams{
		ID:     postID,
		Status: entity.GeneratedImagePostStatusProcessing,
	})
	if err != nil {
		log.Error("[QUEUE] Failed to update status to processing", "error", err)
		return err
	}
	log.Info("[QUEUE] Updated status to processing")

	// 3. Build image prompt
	imagePrompt := w.buildImagePromptFromPost(ctx, post)
	log.Info("[QUEUE] Built image prompt", "promptLength", len(imagePrompt))

	// 4. Update prompts in DB
	err = w.store.UpdateGeneratedImagePostPrompts(ctx, entity.UpdateGeneratedImagePostPromptsParams{
		ID:                     postID,
		GeneratedImagePrompt:   sql.NullString{String: imagePrompt, Valid: true},
		GeneratedCaptionPrompt: sql.NullString{String: "", Valid: false},
	})
	if err != nil {
		log.Error("[QUEUE] Failed to update prompts", "error", err)
		return err
	}
	log.Info("[QUEUE] Updated prompts in DB")

	// 5. Generate images based on provider
	var imageURLs []string
	var tokenUsed int64

	switch post.RecordedModelProvider {
	case "openai":
		imageURLs, tokenUsed, err = w.generateWithOpenAI(ctx, post, imagePrompt)
	case "google":
		imageURLs, tokenUsed, err = w.generateWithGoogle(ctx, post, imagePrompt)
	default:
		log.Error("[QUEUE] Unknown provider", "provider", post.RecordedModelProvider)
		return w.markAsFailed(ctx, postID, "UNKNOWN_PROVIDER: "+post.RecordedModelProvider)
	}

	if err != nil {
		log.Error("[QUEUE] Failed to generate images", "error", err)
		return w.markAsFailed(ctx, postID, "GENERATE_FAILED: "+err.Error())
	}
	log.Info("[QUEUE] Images generated", "count", len(imageURLs), "tokenUsed", tokenUsed)

	// 6. Insert generated_image_post_items
	for _, url := range imageURLs {
		perImageToken := tokenUsed / int64(len(imageURLs))
		if perImageToken == 0 {
			perImageToken = w.cfg.FALLBACK_TOKEN_PER_IMAGE
		}
		_, err = w.store.CreateGeneratedImagePostItem(ctx, entity.CreateGeneratedImagePostItemParams{
			GeneratedImagePostID: postID,
			ImageUrl:             url,
			TokenUsed:            int32(perImageToken),
		})
		if err != nil {
			log.Error("[QUEUE] Failed to create image post item", "url", url, "error", err)
		}
	}
	log.Info("[QUEUE] Inserted generated_image_post_items", "count", len(imageURLs))

	// 7. Deduct tokens for images
	// Get owner profile from business members
	ownerMember, err := w.store.GetOwnerMemberByBusinessRootId(ctx, businessRootID)
	if err != nil {
		log.Error("[QUEUE] Failed to get owner member", "error", err)
		return w.markAsFailed(ctx, postID, "OWNER_NOT_FOUND: "+err.Error())
	}

	deductAmount := tokenUsed
	if deductAmount == 0 {
		deductAmount = w.cfg.FALLBACK_TOKEN_PER_IMAGE * int64(len(imageURLs))
	}

	err = w.imageTokenSvc.DeductTokenForImagePost(ctx, image_token_service.DeductTokenInput{
		ProfileID:            ownerMember,
		BusinessRootID:       businessRootID,
		GeneratedImagePostID: postID,
		Amount:               deductAmount,
	})
	if err != nil {
		log.Error("[QUEUE] Failed to deduct token", "error", err)
		// Don't fail the job, just log error
	}
	log.Info("[QUEUE] Deducted tokens for images", "amount", deductAmount)

	// 8. Generate caption (TODO: implement later, for now use static)
	// If currentCaption is set, don't generate. Otherwise generate.
	if !post.CurrentCaption.Valid || post.CurrentCaption.String == "" {
		// TODO: Generate caption via AI
		// For now, use static caption
		captionText := "Caption On Progress"
		captionTokenUsed := int32(1500)

		_, err = w.store.CreateGeneratedImagePostCaption(ctx, entity.CreateGeneratedImagePostCaptionParams{
			GeneratedImagePostID: postID,
			CaptionText:          captionText,
			TokenUsed:            captionTokenUsed,
		})
		if err != nil {
			log.Error("[QUEUE] Failed to create caption", "error", err)
		} else {
			log.Info("[QUEUE] Created static caption", "caption", captionText)

			// Deduct caption token
			err = w.imageTokenSvc.DeductTokenForImagePost(ctx, image_token_service.DeductTokenInput{
				ProfileID:            ownerMember,
				BusinessRootID:       businessRootID,
				GeneratedImagePostID: postID,
				Amount:               int64(captionTokenUsed),
			})
			if err != nil {
				log.Error("[QUEUE] Failed to deduct caption token", "error", err)
			} else {
				log.Info("[QUEUE] Deducted tokens for caption", "amount", captionTokenUsed)
			}
		}
	} else {
		log.Info("[QUEUE] currentCaption already set, skipping caption generation")
	}

	// 9. Update status to success
	err = w.store.UpdateGeneratedImagePostStatus(ctx, entity.UpdateGeneratedImagePostStatusParams{
		ID:     postID,
		Status: entity.GeneratedImagePostStatusSuccess,
	})
	if err != nil {
		log.Error("[QUEUE] Failed to update status to success", "error", err)
		return err
	}
	log.Info("[QUEUE] Updated status to success - ProcessImagePostJob completed")

	return nil
}

// generateWithOpenAI generates images using OpenAI DALL-E
func (w *ImagePostWorkerHandler) generateWithOpenAI(ctx context.Context, post entity.GeneratedImagePost, prompt string) ([]string, int64, error) {
	log := logger.L().With("postID", post.ID, "provider", "openai")
	log.Info("[QUEUE] Generating with OpenAI", "model", post.RecordedModelName)

	n := int(post.NumOfImages)
	size := "1024x1024"
	if post.ImageSize.Valid && post.ImageSize.String != "" {
		size = post.ImageSize.String
	}

	result, err := w.openaiSvc.GenerateImage(ctx, openai_svc.GenerateImageInput{
		Model:  post.RecordedModelName,
		Prompt: prompt,
		N:      &n,
		Size:   &size,
	})
	if err != nil {
		return nil, 0, err
	}

	// OpenAI returns temp URLs, need to upload to Cloudinary
	var finalURLs []string
	for _, img := range result.Images {
		if img.URL != "" {
			uploadedURL, err := w.uploadFromURL(ctx, img.URL)
			if err != nil {
				log.Error("[QUEUE] Failed to upload from URL", "url", img.URL, "error", err)
				continue
			}
			finalURLs = append(finalURLs, uploadedURL)
		}
	}

	// OpenAI doesn't return token usage for image generation, use fallback
	tokenUsed := w.cfg.FALLBACK_TOKEN_PER_IMAGE * int64(len(finalURLs))
	return finalURLs, tokenUsed, nil
}

// generateWithGoogle generates images using Google Imagen
func (w *ImagePostWorkerHandler) generateWithGoogle(ctx context.Context, post entity.GeneratedImagePost, prompt string) ([]string, int64, error) {
	log := logger.L().With("postID", post.ID, "provider", "google")
	log.Info("[QUEUE] Generating with Google Imagen", "model", post.RecordedModelName)

	n := int(post.NumOfImages)
	aspectRatio := post.Ratio
	if aspectRatio == "" {
		aspectRatio = "1:1"
	}

	result, err := w.googleGenAISvc.GenerateImage(ctx, google_genai.GenerateImageInput{
		Model:          post.RecordedModelName,
		Prompt:         prompt,
		NumberOfImages: &n,
		AspectRatio:    &aspectRatio,
	})
	if err != nil {
		return nil, 0, err
	}

	// Google returns base64, need to decode and upload to Cloudinary
	var finalURLs []string
	for _, img := range result.Images {
		if img.Base64Data != "" {
			uploadedURL, err := w.uploadFromBase64(ctx, img.Base64Data, img.MimeType)
			if err != nil {
				log.Error("[QUEUE] Failed to upload from base64", "error", err)
				continue
			}
			finalURLs = append(finalURLs, uploadedURL)
		}
	}

	// Google doesn't return token usage for image generation, use fallback
	tokenUsed := w.cfg.FALLBACK_TOKEN_PER_IMAGE * int64(len(finalURLs))
	return finalURLs, tokenUsed, nil
}

// uploadFromURL downloads image from URL and uploads to Cloudinary
func (w *ImagePostWorkerHandler) uploadFromURL(ctx context.Context, url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	result, err := w.cloudinarySvc.UploadSingleImage(ctx, resp.Body)
	if err != nil {
		return "", err
	}
	return result.ImageUrl, nil
}

// uploadFromBase64 decodes base64 and uploads to Cloudinary
func (w *ImagePostWorkerHandler) uploadFromBase64(ctx context.Context, base64Data string, mimeType string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return "", err
	}

	reader := bytes.NewReader(decoded)
	result, err := w.cloudinarySvc.UploadSingleImage(ctx, io.NopCloser(reader))
	if err != nil {
		return "", err
	}
	return result.ImageUrl, nil
}

// buildImagePromptFromPost builds image prompt from post + knowledge data
func (w *ImagePostWorkerHandler) buildImagePromptFromPost(ctx context.Context, post entity.GeneratedImagePost) string {
	// TODO: Fetch business knowledge, product, role and build prompt
	// For now, use additionalPrompt + designStyle + category
	var prompt string

	if post.AdditionalPrompt.Valid && post.AdditionalPrompt.String != "" {
		prompt = post.AdditionalPrompt.String
	}

	if post.DesignStyle.Valid && post.DesignStyle.String != "" {
		if prompt != "" {
			prompt += " "
		}
		prompt += "Style: " + post.DesignStyle.String
	}

	if post.Category.Valid && post.Category.String != "" {
		if prompt != "" {
			prompt += ". "
		}
		prompt += "Category: " + post.Category.String
	}

	if prompt == "" {
		prompt = "Generate a high-quality marketing image"
	}

	return prompt
}

// markAsFailed updates post status to failed with error log
func (w *ImagePostWorkerHandler) markAsFailed(ctx context.Context, postID int64, errorMsg string) error {
	log := logger.L().With("postID", postID)
	log.Error("[QUEUE] Marking as failed", "error", errorMsg)

	return w.store.UpdateGeneratedImagePostError(ctx, entity.UpdateGeneratedImagePostErrorParams{
		ID:       postID,
		ErrorLog: sql.NullString{String: errorMsg, Valid: true},
	})
}
