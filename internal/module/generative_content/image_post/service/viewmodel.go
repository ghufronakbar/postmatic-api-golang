// internal/module/generative_content/image_post/service/viewmodel.go
package image_post_service

import (
	"time"

	"postmatic-api/internal/repository/entity"
)

// ImagePostResponse adalah response untuk single image post
type ImagePostResponse struct {
	ID                        int64                     `json:"id"`
	Status                    string                    `json:"status"`
	ErrorLog                  *string                   `json:"errorLog"`
	GeneratedImagePrompt      *string                   `json:"generatedImagePrompt"`
	GeneratedCaptionPrompt    *string                   `json:"generatedCaptionPrompt"`
	BusinessRootID            int64                     `json:"businessRootId"`
	BusinessProductID         int64                     `json:"businessProductId"`
	AppGenerativeImageModelID int64                     `json:"appGenerativeImageModelId"`
	AppGenerativeTextModelID  *int64                    `json:"appGenerativeTextModelId"`
	RecordedModelName         string                    `json:"recordedModelName"`
	RecordedModelProvider     string                    `json:"recordedModelProvider"`
	Mode                      string                    `json:"mode"`
	NumOfImages               int                       `json:"numOfImages"`
	Ratio                     string                    `json:"ratio"`
	AdditionalPrompt          *string                   `json:"additionalPrompt"`
	DesignStyle               *string                   `json:"designStyle"`
	Category                  *string                   `json:"category"`
	ReferenceImageURL         *string                   `json:"referenceImageUrl"`
	MaskImageURL              *string                   `json:"maskImageUrl"`
	ImageSize                 *string                   `json:"imageSize"`
	CurrentCaption            *string                   `json:"currentCaption"`
	Items                     []ImagePostItemResponse   `json:"items"`
	Caption                   *ImagePostCaptionResponse `json:"caption"`
	CreatedAt                 time.Time                 `json:"createdAt"`
	UpdatedAt                 time.Time                 `json:"updatedAt"`
}

// ImagePostItemResponse untuk setiap gambar yang dihasilkan
type ImagePostItemResponse struct {
	ID        int64     `json:"id"`
	ImageURL  string    `json:"imageUrl"`
	TokenUsed int       `json:"tokenUsed"`
	CreatedAt time.Time `json:"createdAt"`
}

// ImagePostCaptionResponse untuk caption yang dihasilkan
type ImagePostCaptionResponse struct {
	ID          int64     `json:"id"`
	CaptionText string    `json:"captionText"`
	TokenUsed   int       `json:"tokenUsed"`
	CreatedAt   time.Time `json:"createdAt"`
}

// CreateImagePostResponse respons setelah membuat image post
type CreateImagePostResponse struct {
	ID          int64     `json:"id"`
	Status      string    `json:"status"`
	Mode        string    `json:"mode"`
	NumOfImages int       `json:"numOfImages"`
	CreatedAt   time.Time `json:"createdAt"`
}

// mapToResponse maps entity ke ImagePostResponse
func mapToResponse(post entity.GeneratedImagePost) ImagePostResponse {
	res := ImagePostResponse{
		ID:                        post.ID,
		Status:                    string(post.Status),
		BusinessRootID:            post.BusinessRootID,
		BusinessProductID:         post.BusinessProductID,
		AppGenerativeImageModelID: post.AppGenerativeImageModelID,
		RecordedModelName:         post.RecordedModelName,
		RecordedModelProvider:     post.RecordedModelProvider,
		Mode:                      string(post.Mode),
		NumOfImages:               int(post.NumOfImages),
		Ratio:                     post.Ratio,
		Items:                     []ImagePostItemResponse{},
		CreatedAt:                 post.CreatedAt,
		UpdatedAt:                 post.UpdatedAt,
	}

	// Nullable fields
	if post.ErrorLog.Valid {
		res.ErrorLog = &post.ErrorLog.String
	}
	if post.GeneratedImagePrompt.Valid {
		res.GeneratedImagePrompt = &post.GeneratedImagePrompt.String
	}
	if post.GeneratedCaptionPrompt.Valid {
		res.GeneratedCaptionPrompt = &post.GeneratedCaptionPrompt.String
	}
	if post.AppGenerativeTextModelID.Valid {
		res.AppGenerativeTextModelID = &post.AppGenerativeTextModelID.Int64
	}
	if post.AdditionalPrompt.Valid {
		res.AdditionalPrompt = &post.AdditionalPrompt.String
	}
	if post.DesignStyle.Valid {
		res.DesignStyle = &post.DesignStyle.String
	}
	if post.Category.Valid {
		res.Category = &post.Category.String
	}
	if post.ReferenceImageUrl.Valid {
		res.ReferenceImageURL = &post.ReferenceImageUrl.String
	}
	if post.MaskImageUrl.Valid {
		res.MaskImageURL = &post.MaskImageUrl.String
	}
	if post.ImageSize.Valid {
		res.ImageSize = &post.ImageSize.String
	}
	if post.CurrentCaption.Valid {
		res.CurrentCaption = &post.CurrentCaption.String
	}

	return res
}
