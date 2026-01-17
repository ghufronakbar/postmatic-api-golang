package generative_image_model_service

import (
	"context"
	"database/sql"
	"regexp"
	"strings"

	"postmatic-api/internal/repository/entity"
	"postmatic-api/pkg/errs"
	"postmatic-api/pkg/pagination"

	"github.com/google/uuid"
)

// Regex pattern for ratio validation: "N:M" where N and M are positive integers
var ratioPattern = regexp.MustCompile(`^\d+:\d+$`)

// validateRatios validates that all ratios are in the format "N:M"
func validateRatios(ratios []string) error {
	for _, ratio := range ratios {
		if !ratioPattern.MatchString(ratio) {
			return errs.NewBadRequest("INVALID_RATIO_FORMAT")
		}
	}
	return nil
}

type GenerativeImageModelService struct {
	store entity.Store
}

func NewService(store entity.Store) *GenerativeImageModelService {
	return &GenerativeImageModelService{store: store}
}

func (s *GenerativeImageModelService) GetAllGenerativeImageModels(ctx context.Context, filter GetGenerativeImageModelsFilter) ([]GenerativeImageModelResponse, *pagination.Pagination, error) {
	params := entity.GetAllGenerativeImageModelsParams{
		Search:     sql.NullString{String: filter.Search, Valid: filter.Search != ""},
		SortBy:     filter.SortBy,
		SortDir:    filter.SortDir,
		PageOffset: int32(filter.PageOffset),
		PageLimit:  int32(filter.PageLimit),
		IsAdmin:    filter.IsAdmin,
	}

	models, err := s.store.GetAllGenerativeImageModels(ctx, params)
	if err != nil {
		return nil, nil, errs.NewInternalServerError(err)
	}

	countParams := entity.CountAllGenerativeImageModelsParams{
		Search:  sql.NullString{String: filter.Search, Valid: filter.Search != ""},
		IsAdmin: filter.IsAdmin,
	}

	count, err := s.store.CountAllGenerativeImageModels(ctx, countParams)
	if err != nil {
		return nil, nil, errs.NewInternalServerError(err)
	}

	pag := pagination.NewPagination(&pagination.PaginationParams{
		Total: int(count),
		Page:  filter.Page,
		Limit: filter.PageLimit,
	})

	responses := make([]GenerativeImageModelResponse, 0, len(models))
	for _, m := range models {
		responses = append(responses, mapToResponse(m))
	}

	return responses, &pag, nil
}

func (s *GenerativeImageModelService) GetGenerativeImageModelById(ctx context.Context, id int64, isAdmin bool) (GenerativeImageModelResponse, error) {
	var model entity.AppGenerativeImageModel
	var err error

	if isAdmin {
		model, err = s.store.GetGenerativeImageModelByIdAdmin(ctx, id)
	} else {
		model, err = s.store.GetGenerativeImageModelByIdUser(ctx, id)
	}

	if err == sql.ErrNoRows {
		return GenerativeImageModelResponse{}, errs.NewNotFound("GENERATIVE_IMAGE_MODEL_NOT_FOUND")
	}
	if err != nil {
		return GenerativeImageModelResponse{}, errs.NewInternalServerError(err)
	}

	return mapToResponse(model), nil
}

func (s *GenerativeImageModelService) GetGenerativeImageModelByModel(ctx context.Context, modelName string, isAdmin bool) (GenerativeImageModelResponse, error) {
	var model entity.AppGenerativeImageModel
	var err error

	normalizedModel := strings.ToLower(modelName)
	if isAdmin {
		model, err = s.store.GetGenerativeImageModelByModelAdmin(ctx, normalizedModel)
	} else {
		model, err = s.store.GetGenerativeImageModelByModelUser(ctx, normalizedModel)
	}

	if err == sql.ErrNoRows {
		return GenerativeImageModelResponse{}, errs.NewNotFound("GENERATIVE_IMAGE_MODEL_NOT_FOUND")
	}
	if err != nil {
		return GenerativeImageModelResponse{}, errs.NewInternalServerError(err)
	}

	return mapToResponse(model), nil
}

func (s *GenerativeImageModelService) CreateGenerativeImageModel(ctx context.Context, input CreateGenerativeImageModelInput) (GenerativeImageModelResponse, error) {
	profileID, err := uuid.Parse(input.ProfileID)
	if err != nil {
		return GenerativeImageModelResponse{}, errs.NewBadRequest("INVALID_PROFILE_ID")
	}

	// Check if model already exists
	normalizedModel := strings.ToLower(input.Model)
	existing, err := s.store.GetGenerativeImageModelByModel(ctx, normalizedModel)
	if err != nil && err != sql.ErrNoRows {
		return GenerativeImageModelResponse{}, errs.NewInternalServerError(err)
	}
	if existing.ID != 0 {
		return GenerativeImageModelResponse{}, errs.NewBadRequest("GENERATIVE_IMAGE_MODEL_ALREADY_EXISTS")
	}

	// Validate ratio format
	if err := validateRatios(input.ValidRatios); err != nil {
		return GenerativeImageModelResponse{}, err
	}

	var image sql.NullString
	if input.Image != nil && *input.Image != "" {
		image = sql.NullString{String: *input.Image, Valid: true}
	}

	var result entity.AppGenerativeImageModel

	err = s.store.ExecTx(ctx, func(q *entity.Queries) error {
		model, createErr := q.CreateGenerativeImageModel(ctx, entity.CreateGenerativeImageModelParams{
			Model:       normalizedModel,
			Label:       input.Label,
			Image:       image,
			Provider:    entity.AppGenerativeImageModelProviderType(input.Provider),
			IsActive:    input.IsActive,
			ValidRatios: input.ValidRatios,
			ImageSizes:  input.ImageSizes,
		})
		if createErr != nil {
			return createErr
		}

		result = model

		// Log change
		_, logErr := q.CreateGenerativeImageModelChange(ctx, entity.CreateGenerativeImageModelChangeParams{
			Action:                 entity.ActionChangeTypeCreate,
			ProfileID:              profileID,
			GenerativeImageModelID: model.ID,
			BeforeModel:            model.Model,
			BeforeLabel:            model.Label,
			BeforeImage:            model.Image,
			BeforeProvider:         model.Provider,
			BeforeIsActive:         model.IsActive,
			BeforeValidRatios:      model.ValidRatios,
			BeforeImageSizes:       model.ImageSizes,
			AfterModel:             model.Model,
			AfterLabel:             model.Label,
			AfterImage:             model.Image,
			AfterProvider:          model.Provider,
			AfterIsActive:          model.IsActive,
			AfterValidRatios:       model.ValidRatios,
			AfterImageSizes:        model.ImageSizes,
		})
		return logErr
	})

	if err != nil {
		return GenerativeImageModelResponse{}, errs.NewInternalServerError(err)
	}

	return mapToResponse(result), nil
}

func (s *GenerativeImageModelService) UpdateGenerativeImageModel(ctx context.Context, input UpdateGenerativeImageModelInput) (GenerativeImageModelResponse, error) {
	profileID, err := uuid.Parse(input.ProfileID)
	if err != nil {
		return GenerativeImageModelResponse{}, errs.NewBadRequest("INVALID_PROFILE_ID")
	}

	// Get existing
	existing, err := s.store.GetGenerativeImageModelById(ctx, input.ID)
	if err == sql.ErrNoRows {
		return GenerativeImageModelResponse{}, errs.NewNotFound("GENERATIVE_IMAGE_MODEL_NOT_FOUND")
	}
	if err != nil {
		return GenerativeImageModelResponse{}, errs.NewInternalServerError(err)
	}

	// Check if new model already exists (different from current)
	normalizedModel := strings.ToLower(input.Model)
	if normalizedModel != strings.ToLower(existing.Model) {
		existingByModel, err := s.store.GetGenerativeImageModelByModel(ctx, normalizedModel)
		if err != nil && err != sql.ErrNoRows {
			return GenerativeImageModelResponse{}, errs.NewInternalServerError(err)
		}
		if existingByModel.ID != 0 {
			return GenerativeImageModelResponse{}, errs.NewBadRequest("GENERATIVE_IMAGE_MODEL_ALREADY_EXISTS")
		}
	}

	// Validate ratio format
	if err := validateRatios(input.ValidRatios); err != nil {
		return GenerativeImageModelResponse{}, err
	}

	var image sql.NullString
	if input.Image != nil && *input.Image != "" {
		image = sql.NullString{String: *input.Image, Valid: true}
	}

	var result entity.AppGenerativeImageModel

	err = s.store.ExecTx(ctx, func(q *entity.Queries) error {
		model, updateErr := q.UpdateGenerativeImageModel(ctx, entity.UpdateGenerativeImageModelParams{
			ID:          input.ID,
			Model:       normalizedModel,
			Label:       input.Label,
			Image:       image,
			Provider:    entity.AppGenerativeImageModelProviderType(input.Provider),
			IsActive:    input.IsActive,
			ValidRatios: input.ValidRatios,
			ImageSizes:  input.ImageSizes,
		})
		if updateErr != nil {
			return updateErr
		}

		result = model

		// Log change
		_, logErr := q.CreateGenerativeImageModelChange(ctx, entity.CreateGenerativeImageModelChangeParams{
			Action:                 entity.ActionChangeTypeUpdate,
			ProfileID:              profileID,
			GenerativeImageModelID: model.ID,
			BeforeModel:            existing.Model,
			BeforeLabel:            existing.Label,
			BeforeImage:            existing.Image,
			BeforeProvider:         existing.Provider,
			BeforeIsActive:         existing.IsActive,
			BeforeValidRatios:      existing.ValidRatios,
			BeforeImageSizes:       existing.ImageSizes,
			AfterModel:             model.Model,
			AfterLabel:             model.Label,
			AfterImage:             model.Image,
			AfterProvider:          model.Provider,
			AfterIsActive:          model.IsActive,
			AfterValidRatios:       model.ValidRatios,
			AfterImageSizes:        model.ImageSizes,
		})
		return logErr
	})

	if err != nil {
		return GenerativeImageModelResponse{}, errs.NewInternalServerError(err)
	}

	return mapToResponse(result), nil
}

func (s *GenerativeImageModelService) DeleteGenerativeImageModel(ctx context.Context, id int64, profileID string) (GenerativeImageModelResponse, error) {
	pid, err := uuid.Parse(profileID)
	if err != nil {
		return GenerativeImageModelResponse{}, errs.NewBadRequest("INVALID_PROFILE_ID")
	}

	// Get existing
	existing, err := s.store.GetGenerativeImageModelById(ctx, id)
	if err == sql.ErrNoRows {
		return GenerativeImageModelResponse{}, errs.NewNotFound("GENERATIVE_IMAGE_MODEL_NOT_FOUND")
	}
	if err != nil {
		return GenerativeImageModelResponse{}, errs.NewInternalServerError(err)
	}

	var result entity.AppGenerativeImageModel

	err = s.store.ExecTx(ctx, func(q *entity.Queries) error {
		model, deleteErr := q.SoftDeleteGenerativeImageModel(ctx, id)
		if deleteErr != nil {
			return deleteErr
		}

		result = model

		// Log change
		_, logErr := q.CreateGenerativeImageModelChange(ctx, entity.CreateGenerativeImageModelChangeParams{
			Action:                 entity.ActionChangeTypeDelete,
			ProfileID:              pid,
			GenerativeImageModelID: model.ID,
			BeforeModel:            existing.Model,
			BeforeLabel:            existing.Label,
			BeforeImage:            existing.Image,
			BeforeProvider:         existing.Provider,
			BeforeIsActive:         existing.IsActive,
			BeforeValidRatios:      existing.ValidRatios,
			BeforeImageSizes:       existing.ImageSizes,
			AfterModel:             model.Model,
			AfterLabel:             model.Label,
			AfterImage:             model.Image,
			AfterProvider:          model.Provider,
			AfterIsActive:          model.IsActive,
			AfterValidRatios:       model.ValidRatios,
			AfterImageSizes:        model.ImageSizes,
		})
		return logErr
	})

	if err != nil {
		return GenerativeImageModelResponse{}, errs.NewInternalServerError(err)
	}

	return mapToResponse(result), nil
}

func mapToResponse(m entity.AppGenerativeImageModel) GenerativeImageModelResponse {
	var image *string
	if m.Image.Valid {
		image = &m.Image.String
	}

	var imageSizes *[]string
	if len(m.ImageSizes) > 0 {
		imageSizes = &m.ImageSizes
	}

	return GenerativeImageModelResponse{
		ID:          m.ID,
		Model:       m.Model,
		Label:       m.Label,
		Image:       image,
		Provider:    GenerativeImageModelProviderType(m.Provider),
		IsActive:    m.IsActive,
		ValidRatios: m.ValidRatios,
		ImageSizes:  imageSizes,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}
