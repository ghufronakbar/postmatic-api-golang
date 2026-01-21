// internal/module/generative_content/image_post/service/common.go
package image_post_service

import (
	"context"
	"database/sql"
	"encoding/json"

	"postmatic-api/internal/repository/entity"
	"postmatic-api/pkg/errs"
	"postmatic-api/pkg/pagination"
	"postmatic-api/pkg/utils"
)

// GetAllImagePosts returns paginated image posts for a business
func (s *ImagePostService) GetAllImagePosts(ctx context.Context, filter GetImagePostsFilter) ([]ImagePostResponse, *pagination.Pagination, error) {
	// Build params
	params := entity.GetAllGeneratedImagePostsByBusinessIdParams{
		BusinessRootID: filter.BusinessRootID,
		SortBy:         sql.NullString{String: filter.SortBy, Valid: filter.SortBy != ""},
		SortDir:        sql.NullString{String: filter.SortDir, Valid: filter.SortDir != ""},
		DateStart:      utils.NullStringToNullTime(filter.DateStart),
		DateEnd:        utils.NullStringToNullTime(filter.DateEnd),
		PageLimit:      int32(filter.PageLimit),
		PageOffset:     int32(filter.PageOffset),
	}

	// Optional filters
	if filter.Status != nil && *filter.Status != "" {
		params.Status = entity.NullGeneratedImagePostStatus{
			GeneratedImagePostStatus: entity.GeneratedImagePostStatus(*filter.Status),
			Valid:                    true,
		}
	}
	if filter.Mode != nil && *filter.Mode != "" {
		params.Mode = entity.NullGeneratedImagePostMode{
			GeneratedImagePostMode: entity.GeneratedImagePostMode(*filter.Mode),
			Valid:                  true,
		}
	}
	if filter.BusinessProductID != nil {
		params.BusinessProductID = sql.NullInt64{Int64: *filter.BusinessProductID, Valid: true}
	}

	// Query
	rows, err := s.store.GetAllGeneratedImagePostsByBusinessId(ctx, params)
	if err != nil && err != sql.ErrNoRows {
		return nil, nil, errs.NewInternalServerError(err)
	}

	// Count
	countParams := entity.CountGeneratedImagePostsByBusinessIdParams{
		BusinessRootID:    filter.BusinessRootID,
		Status:            params.Status,
		Mode:              params.Mode,
		BusinessProductID: params.BusinessProductID,
		DateStart:         params.DateStart,
		DateEnd:           params.DateEnd,
	}
	total, err := s.store.CountGeneratedImagePostsByBusinessId(ctx, countParams)
	if err != nil && err != sql.ErrNoRows {
		return nil, nil, errs.NewInternalServerError(err)
	}

	pag := pagination.NewPagination(&pagination.PaginationParams{
		Total: int(total),
		Page:  filter.Page,
		Limit: filter.PageLimit,
	})

	// Map to response
	res := make([]ImagePostResponse, 0, len(rows))
	for _, row := range rows {
		resp := mapRowToResponse(row)
		res = append(res, resp)
	}

	return res, &pag, nil
}

// GetImagePostById returns single image post by ID
func (s *ImagePostService) GetImagePostById(ctx context.Context, id int64, businessRootID int64) (ImagePostResponse, error) {
	post, err := s.store.GetGeneratedImagePostById(ctx, id)
	if err == sql.ErrNoRows {
		return ImagePostResponse{}, errs.NewNotFound("IMAGE_POST_NOT_FOUND")
	}
	if err != nil {
		return ImagePostResponse{}, errs.NewInternalServerError(err)
	}

	// Check ownership
	if post.BusinessRootID != businessRootID {
		return ImagePostResponse{}, errs.NewForbidden("IMAGE_POST_NOT_BELONGS_TO_BUSINESS")
	}

	// Get items
	items, err := s.store.GetGeneratedImagePostItemsByPostId(ctx, id)
	if err != nil && err != sql.ErrNoRows {
		return ImagePostResponse{}, errs.NewInternalServerError(err)
	}

	// Get caption
	caption, err := s.store.GetGeneratedImagePostCaptionByPostId(ctx, id)
	if err != nil && err != sql.ErrNoRows {
		return ImagePostResponse{}, errs.NewInternalServerError(err)
	}

	resp := mapToResponse(post)

	// Map items
	for _, item := range items {
		resp.Items = append(resp.Items, ImagePostItemResponse{
			ID:        item.ID,
			ImageURL:  item.ImageUrl,
			TokenUsed: int(item.TokenUsed),
			CreatedAt: item.CreatedAt,
		})
	}

	// Map caption
	if caption.ID != 0 {
		resp.Caption = &ImagePostCaptionResponse{
			ID:          caption.ID,
			CaptionText: caption.CaptionText,
			TokenUsed:   int(caption.TokenUsed),
			CreatedAt:   caption.CreatedAt,
		}
	}

	return resp, nil
}

// mapRowToResponse maps GetAllGeneratedImagePostsByBusinessIdRow ke ImagePostResponse
func mapRowToResponse(row entity.GetAllGeneratedImagePostsByBusinessIdRow) ImagePostResponse {
	res := ImagePostResponse{
		ID:                        row.ID,
		Status:                    string(row.Status),
		BusinessRootID:            row.BusinessRootID,
		BusinessProductID:         row.BusinessProductID,
		AppGenerativeImageModelID: row.AppGenerativeImageModelID,
		RecordedModelName:         row.RecordedModelName,
		RecordedModelProvider:     row.RecordedModelProvider,
		Mode:                      string(row.Mode),
		NumOfImages:               int(row.NumOfImages),
		Ratio:                     row.Ratio,
		Items:                     []ImagePostItemResponse{},
		CreatedAt:                 row.CreatedAt,
		UpdatedAt:                 row.UpdatedAt,
	}

	// Nullable fields
	if row.ErrorLog.Valid {
		res.ErrorLog = &row.ErrorLog.String
	}
	if row.GeneratedImagePrompt.Valid {
		res.GeneratedImagePrompt = &row.GeneratedImagePrompt.String
	}
	if row.GeneratedCaptionPrompt.Valid {
		res.GeneratedCaptionPrompt = &row.GeneratedCaptionPrompt.String
	}
	if row.AppGenerativeTextModelID.Valid {
		res.AppGenerativeTextModelID = &row.AppGenerativeTextModelID.Int64
	}
	if row.AdditionalPrompt.Valid {
		res.AdditionalPrompt = &row.AdditionalPrompt.String
	}
	if row.DesignStyle.Valid {
		res.DesignStyle = &row.DesignStyle.String
	}
	if row.Category.Valid {
		res.Category = &row.Category.String
	}
	if row.ReferenceImageUrl.Valid {
		res.ReferenceImageURL = &row.ReferenceImageUrl.String
	}
	if row.MaskImageUrl.Valid {
		res.MaskImageURL = &row.MaskImageUrl.String
	}
	if row.ImageSize.Valid {
		res.ImageSize = &row.ImageSize.String
	}
	if row.CurrentCaption.Valid {
		res.CurrentCaption = &row.CurrentCaption.String
	}

	// Parse items JSON - Items is interface{}, need to marshal first
	if row.Items != nil {
		var items []ImagePostItemResponse
		// row.Items is already deserialized by sqlc as interface{}
		b, err := json.Marshal(row.Items)
		if err == nil {
			_ = json.Unmarshal(b, &items)
		}
		if items != nil {
			res.Items = items
		}
	}

	// Parse caption JSON - Caption is also interface{}
	if row.Caption != nil {
		var caption ImagePostCaptionResponse
		b, err := json.Marshal(row.Caption)
		if err == nil {
			if err := json.Unmarshal(b, &caption); err == nil && caption.ID != 0 {
				res.Caption = &caption
			}
		}
	}

	return res
}
