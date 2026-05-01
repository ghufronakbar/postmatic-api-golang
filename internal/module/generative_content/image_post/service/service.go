// internal/module/generative_content/image_post/service/service.go
package image_post_service

import (
	"context"
	"database/sql"

	generative_image_model_service "postmatic-api/internal/module/app/generative_image_model/service"
	business_knowledge_service "postmatic-api/internal/module/business/business_knowledge/service"
	business_product_service "postmatic-api/internal/module/business/business_product/service"
	business_role_service "postmatic-api/internal/module/business/business_role/service"
	image_token_service "postmatic-api/internal/module/generative_token/image_token/service"
	"postmatic-api/internal/repository/entity"
	"postmatic-api/pkg/errs"
)

// ImagePostService handles generative image post operations
type ImagePostService struct {
	store                entity.Store
	imageModelSvc        *generative_image_model_service.GenerativeImageModelService
	businessKnowledgeSvc *business_knowledge_service.BusinessKnowledgeService
	businessProductSvc   *business_product_service.BusinessProductService
	businessRoleSvc      *business_role_service.BusinessRoleService
	imageTokenSvc        *image_token_service.ImageTokenService
	queueProducer        ImagePostQueueProducer
}

// ImagePostQueueProducer interface for queue producer
type ImagePostQueueProducer interface {
	EnqueueImagePostGenerate(ctx context.Context, payload ImagePostQueuePayload) error
}

// ImagePostQueuePayload payload untuk queue
type ImagePostQueuePayload struct {
	PostID         int64 `json:"postId"`
	BusinessRootID int64 `json:"businessRootId"`
}

// NewService creates a new ImagePostService
func NewService(
	store entity.Store,
	imageModelSvc *generative_image_model_service.GenerativeImageModelService,
	businessKnowledgeSvc *business_knowledge_service.BusinessKnowledgeService,
	businessProductSvc *business_product_service.BusinessProductService,
	businessRoleSvc *business_role_service.BusinessRoleService,
	imageTokenSvc *image_token_service.ImageTokenService,
	queueProducer ImagePostQueueProducer,
) *ImagePostService {
	return &ImagePostService{
		store:                store,
		imageModelSvc:        imageModelSvc,
		businessKnowledgeSvc: businessKnowledgeSvc,
		businessProductSvc:   businessProductSvc,
		businessRoleSvc:      businessRoleSvc,
		imageTokenSvc:        imageTokenSvc,
		queueProducer:        queueProducer,
	}
}

// TODO: fix later
const ESTIMATED_TOKEN_PER_IMAGE = 3000

// CreateImagePost creates a new image post and enqueues it
func (s *ImagePostService) CreateImagePost(ctx context.Context, input CreateImagePostInput) (CreateImagePostResponse, error) {
	// 1. Validate input berdasarkan mode
	if err := s.validateInput(input); err != nil {
		return CreateImagePostResponse{}, err
	}

	// 2. Validate image model exists & active
	imageModel, err := s.imageModelSvc.GetGenerativeImageModelById(ctx, input.AppGenerativeImageModelID, false)
	if err != nil {
		return CreateImagePostResponse{}, err
	}

	// 3. Validate product belongs to business
	product, err := s.store.GetBusinessProductByBusinessProductId(ctx, input.ProductKnowledgeID)
	if err == sql.ErrNoRows {
		return CreateImagePostResponse{}, errs.NewNotFound("PRODUCT_NOT_FOUND")
	}
	if err != nil {
		return CreateImagePostResponse{}, errs.NewInternalServerError(err)
	}
	if product.BusinessRootID != input.BusinessRootID {
		return CreateImagePostResponse{}, errs.NewForbidden("PRODUCT_NOT_BELONGS_TO_BUSINESS")
	}
	// Validate product has at least one image
	if len(product.ImageUrls) == 0 {
		return CreateImagePostResponse{}, errs.NewBadRequest("PRODUCT_HAS_NO_IMAGES")
	}

	// 4. Validate knowledge flags jika ada advanceGenerate
	if input.AdvanceGenerate != nil {
		if err := s.validateKnowledgeFlags(ctx, input.BusinessRootID, input.AdvanceGenerate); err != nil {
			return CreateImagePostResponse{}, err
		}
	}

	// 5. Check token availability
	tokenStatus, err := s.imageTokenSvc.GetTokenStatus(ctx, input.BusinessRootID)
	if err != nil {
		return CreateImagePostResponse{}, err
	}
	if tokenStatus.AvailableToken < int64(input.NumOfImages*ESTIMATED_TOKEN_PER_IMAGE) {
		return CreateImagePostResponse{}, errs.NewBadRequest("INSUFFICIENT_TOKEN")
	}

	// 6. Route berdasarkan mode
	switch input.Mode {
	case "generate":
		return s.handleGenerate(ctx, input, imageModel)
	case "regenerate":
		return s.handleRegenerate(ctx, input)
	case "rss":
		return s.handleRss(ctx, input)
	case "mask":
		return s.handleMask(ctx, input)
	default:
		return CreateImagePostResponse{}, errs.NewBadRequest("INVALID_MODE")
	}
}

// validateInput validates conditional input berdasarkan mode
func (s *ImagePostService) validateInput(input CreateImagePostInput) error {
	switch input.Mode {
	case "generate":
		if input.AdvanceGenerate == nil {
			return errs.NewBadRequest("ADVANCE_GENERATE_REQUIRED_FOR_GENERATE_MODE")
		}
	case "regenerate":
		if input.ReferenceImage == "" {
			return errs.NewBadRequest("REFERENCE_IMAGE_REQUIRED_FOR_REGENERATE_MODE")
		}
		if input.AdvanceGenerate == nil {
			return errs.NewBadRequest("ADVANCE_GENERATE_REQUIRED_FOR_REGENERATE_MODE")
		}
	case "rss":
		if input.Rss == nil {
			return errs.NewBadRequest("RSS_REQUIRED_FOR_RSS_MODE")
		}
		if input.AdvanceGenerate == nil {
			return errs.NewBadRequest("ADVANCE_GENERATE_REQUIRED_FOR_RSS_MODE")
		}
	case "mask":
		if input.MaskImage == "" {
			return errs.NewBadRequest("MASK_IMAGE_REQUIRED_FOR_MASK_MODE")
		}
		if input.ReferenceImage == "" {
			return errs.NewBadRequest("REFERENCE_IMAGE_REQUIRED_FOR_MASK_MODE")
		}
		if input.AdditionalPrompt == "" {
			return errs.NewBadRequest("ADDITIONAL_PROMPT_REQUIRED_FOR_MASK_MODE")
		}
	}
	return nil
}

// validateKnowledgeFlags validates knowledge flags against actual data
func (s *ImagePostService) validateKnowledgeFlags(ctx context.Context, businessRootID int64, adv *AdvanceGenerateInput) error {
	// Get business knowledge
	bk, err := s.businessKnowledgeSvc.GetBusinessKnowledgeByBusinessRootID(ctx, businessRootID)
	if err != nil {
		return err
	}

	// Validate business knowledge flags
	if adv.BusinessKnowledge != nil {
		flags := adv.BusinessKnowledge
		if flags.Website && (bk.WebsiteUrl == nil || *bk.WebsiteUrl == "") {
			return errs.NewBadRequest("BUSINESS_KNOWLEDGE_WEBSITE_NOT_SET")
		}
		if flags.Logo && bk.PrimaryLogoUrl == "" {
			return errs.NewBadRequest("BUSINESS_KNOWLEDGE_LOGO_NOT_SET")
		}
		if flags.Description && bk.Description == "" {
			return errs.NewBadRequest("BUSINESS_KNOWLEDGE_DESCRIPTION_NOT_SET")
		}
		if flags.Location && bk.Location == "" {
			return errs.NewBadRequest("BUSINESS_KNOWLEDGE_LOCATION_NOT_SET")
		}
		if flags.UniqueSellingPoint && bk.UniqueSellingPoint == "" {
			return errs.NewBadRequest("BUSINESS_KNOWLEDGE_USP_NOT_SET")
		}
		if flags.VisionMission && bk.VisionMission == "" {
			return errs.NewBadRequest("BUSINESS_KNOWLEDGE_VISION_MISSION_NOT_SET")
		}
		if flags.ColorTone && bk.ColorTone == "" {
			return errs.NewBadRequest("BUSINESS_KNOWLEDGE_COLOR_TONE_NOT_SET")
		}
	}

	return nil
}
