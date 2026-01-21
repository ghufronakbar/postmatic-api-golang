// internal/module/generative_content/image_post/service/generate.go
package image_post_service

import (
	"context"
	"database/sql"
	"time"

	generative_image_model_service "postmatic-api/internal/module/app/generative_image_model/service"
	"postmatic-api/internal/repository/entity"
	"postmatic-api/pkg/errs"
)

// handleGenerate handles mode=generate
func (s *ImagePostService) handleGenerate(
	ctx context.Context,
	input CreateImagePostInput,
	imageModel generative_image_model_service.GenerativeImageModelResponse,
) (CreateImagePostResponse, error) {
	// Build advance generate flags
	advBk := input.AdvanceGenerate.BusinessKnowledge
	advPd := input.AdvanceGenerate.ProductKnowledge
	advRl := input.AdvanceGenerate.RoleKnowledge

	// Default false jika nil
	var bkName, bkCategory, bkDescription, bkLocation, bkLogo, bkUSP, bkWebsite, bkVM, bkColor bool
	var pdName, pdCategory, pdDescription, pdPrice bool
	var rlHashtags bool

	if advBk != nil {
		bkName = advBk.Name
		bkCategory = advBk.Category
		bkDescription = advBk.Description
		bkLocation = advBk.Location
		bkLogo = advBk.Logo
		bkUSP = advBk.UniqueSellingPoint
		bkWebsite = advBk.Website
		bkVM = advBk.VisionMission
		bkColor = advBk.ColorTone
	}
	if advPd != nil {
		pdName = advPd.Name
		pdCategory = advPd.Category
		pdDescription = advPd.Description
		pdPrice = advPd.Price
	}
	if advRl != nil {
		rlHashtags = advRl.Hashtags
	}

	// Insert to DB
	post, err := s.store.CreateGeneratedImagePost(ctx, entity.CreateGeneratedImagePostParams{
		Status:                    entity.GeneratedImagePostStatusPending,
		BusinessRootID:            input.BusinessRootID,
		BusinessProductID:         input.ProductKnowledgeID,
		AppGenerativeImageModelID: input.AppGenerativeImageModelID,
		RecordedModelName:         imageModel.Model,
		RecordedModelProvider:     string(imageModel.Provider),
		Mode:                      entity.GeneratedImagePostModeGenerate,
		NumOfImages:               int32(input.NumOfImages),
		Ratio:                     input.Ratio,
		AdditionalPrompt:          sql.NullString{String: input.AdditionalPrompt, Valid: input.AdditionalPrompt != ""},
		DesignStyle:               sql.NullString{String: input.DesignStyle, Valid: input.DesignStyle != ""},
		Category:                  sql.NullString{String: input.Category, Valid: input.Category != ""},
		ReferenceImageUrl:         sql.NullString{String: input.ReferenceImage, Valid: input.ReferenceImage != ""},
		MaskImageUrl:              sql.NullString{String: input.MaskImage, Valid: input.MaskImage != ""},
		ImageSize:                 sql.NullString{String: input.ImageSize, Valid: input.ImageSize != ""},
		CurrentCaption:            sql.NullString{String: input.CurrentCaption, Valid: input.CurrentCaption != ""},
		AdvBkName:                 bkName,
		AdvBkCategory:             bkCategory,
		AdvBkDescription:          bkDescription,
		AdvBkLocation:             bkLocation,
		AdvBkLogo:                 bkLogo,
		AdvBkUniqueSellingPoint:   bkUSP,
		AdvBkWebsite:              bkWebsite,
		AdvBkVisionMission:        bkVM,
		AdvBkColorTone:            bkColor,
		AdvPdName:                 pdName,
		AdvPdCategory:             pdCategory,
		AdvPdDescription:          pdDescription,
		AdvPdPrice:                pdPrice,
		AdvRlHashtags:             rlHashtags,
	})
	if err != nil {
		return CreateImagePostResponse{}, errs.NewInternalServerError(err)
	}

	// Enqueue job
	if s.queueProducer != nil {
		err = s.queueProducer.EnqueueImagePostGenerate(ctx, ImagePostQueuePayload{
			PostID:         post.ID,
			BusinessRootID: post.BusinessRootID,
		})
		if err != nil {
			// Update status to failed
			_ = s.store.UpdateGeneratedImagePostError(ctx, entity.UpdateGeneratedImagePostErrorParams{
				ID:       post.ID,
				ErrorLog: sql.NullString{String: "FAILED_TO_ENQUEUE: " + err.Error(), Valid: true},
			})
			return CreateImagePostResponse{}, errs.NewInternalServerError(err)
		}
	}

	return CreateImagePostResponse{
		ID:          post.ID,
		Status:      string(post.Status),
		Mode:        string(post.Mode),
		NumOfImages: int(post.NumOfImages),
		CreatedAt:   time.Now(),
	}, nil
}
