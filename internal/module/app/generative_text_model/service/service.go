package generative_text_model_service

import (
	"context"
	"database/sql"
	"strings"

	"postmatic-api/internal/repository/entity"
	"postmatic-api/pkg/errs"
	"postmatic-api/pkg/pagination"

	"github.com/google/uuid"
)

type GenerativeTextModelService struct {
	store entity.Store
}

func NewService(store entity.Store) *GenerativeTextModelService {
	return &GenerativeTextModelService{store: store}
}

func (s *GenerativeTextModelService) GetAllGenerativeTextModels(ctx context.Context, filter GetGenerativeTextModelsFilter) ([]GenerativeTextModelResponse, *pagination.Pagination, error) {
	params := entity.GetAllGenerativeTextModelsParams{
		Search:     sql.NullString{String: filter.Search, Valid: filter.Search != ""},
		SortBy:     filter.SortBy,
		SortDir:    filter.SortDir,
		PageOffset: int32(filter.PageOffset),
		PageLimit:  int32(filter.PageLimit),
		IsAdmin:    filter.IsAdmin,
	}

	models, err := s.store.GetAllGenerativeTextModels(ctx, params)
	if err != nil {
		return nil, nil, errs.NewInternalServerError(err)
	}

	countParams := entity.CountAllGenerativeTextModelsParams{
		Search:  sql.NullString{String: filter.Search, Valid: filter.Search != ""},
		IsAdmin: filter.IsAdmin,
	}

	count, err := s.store.CountAllGenerativeTextModels(ctx, countParams)
	if err != nil {
		return nil, nil, errs.NewInternalServerError(err)
	}

	pag := pagination.NewPagination(&pagination.PaginationParams{
		Total: int(count),
		Page:  filter.Page,
		Limit: filter.PageLimit,
	})

	responses := make([]GenerativeTextModelResponse, 0, len(models))
	for _, m := range models {
		responses = append(responses, mapToResponse(m))
	}

	return responses, &pag, nil
}

func (s *GenerativeTextModelService) GetGenerativeTextModelById(ctx context.Context, id int64, isAdmin bool) (GenerativeTextModelResponse, error) {
	var model entity.AppGenerativeTextModel
	var err error

	if isAdmin {
		model, err = s.store.GetGenerativeTextModelByIdAdmin(ctx, id)
	} else {
		model, err = s.store.GetGenerativeTextModelByIdUser(ctx, id)
	}

	if err == sql.ErrNoRows {
		return GenerativeTextModelResponse{}, errs.NewNotFound("GENERATIVE_TEXT_MODEL_NOT_FOUND")
	}
	if err != nil {
		return GenerativeTextModelResponse{}, errs.NewInternalServerError(err)
	}

	return mapToResponse(model), nil
}

func (s *GenerativeTextModelService) GetGenerativeTextModelByModel(ctx context.Context, modelName string, isAdmin bool) (GenerativeTextModelResponse, error) {
	var model entity.AppGenerativeTextModel
	var err error

	normalizedModel := strings.ToLower(modelName)
	if isAdmin {
		model, err = s.store.GetGenerativeTextModelByModelAdmin(ctx, normalizedModel)
	} else {
		model, err = s.store.GetGenerativeTextModelByModelUser(ctx, normalizedModel)
	}

	if err == sql.ErrNoRows {
		return GenerativeTextModelResponse{}, errs.NewNotFound("GENERATIVE_TEXT_MODEL_NOT_FOUND")
	}
	if err != nil {
		return GenerativeTextModelResponse{}, errs.NewInternalServerError(err)
	}

	return mapToResponse(model), nil
}

func (s *GenerativeTextModelService) CreateGenerativeTextModel(ctx context.Context, input CreateGenerativeTextModelInput) (GenerativeTextModelResponse, error) {
	profileID, err := uuid.Parse(input.ProfileID)
	if err != nil {
		return GenerativeTextModelResponse{}, errs.NewBadRequest("INVALID_PROFILE_ID")
	}

	// Check if model already exists
	normalizedModel := strings.ToLower(input.Model)
	existing, err := s.store.GetGenerativeTextModelByModel(ctx, normalizedModel)
	if err != nil && err != sql.ErrNoRows {
		return GenerativeTextModelResponse{}, errs.NewInternalServerError(err)
	}
	if existing.ID != 0 {
		return GenerativeTextModelResponse{}, errs.NewBadRequest("GENERATIVE_TEXT_MODEL_ALREADY_EXISTS")
	}

	var image sql.NullString
	if input.Image != nil && *input.Image != "" {
		image = sql.NullString{String: *input.Image, Valid: true}
	}

	var result entity.AppGenerativeTextModel

	err = s.store.ExecTx(ctx, func(q *entity.Queries) error {
		model, createErr := q.CreateGenerativeTextModel(ctx, entity.CreateGenerativeTextModelParams{
			Model:    normalizedModel,
			Label:    input.Label,
			Image:    image,
			Provider: entity.AppGenerativeTextModelProviderType(input.Provider),
			IsActive: input.IsActive,
		})
		if createErr != nil {
			return createErr
		}

		result = model

		// Log change
		_, logErr := q.CreateGenerativeTextModelChange(ctx, entity.CreateGenerativeTextModelChangeParams{
			Action:                entity.ActionChangeTypeCreate,
			ProfileID:             profileID,
			GenerativeTextModelID: model.ID,
			BeforeModel:           model.Model,
			BeforeLabel:           model.Label,
			BeforeImage:           model.Image,
			BeforeProvider:        model.Provider,
			BeforeIsActive:        model.IsActive,
			AfterModel:            model.Model,
			AfterLabel:            model.Label,
			AfterImage:            model.Image,
			AfterProvider:         model.Provider,
			AfterIsActive:         model.IsActive,
		})
		return logErr
	})

	if err != nil {
		return GenerativeTextModelResponse{}, errs.NewInternalServerError(err)
	}

	return mapToResponse(result), nil
}

func (s *GenerativeTextModelService) UpdateGenerativeTextModel(ctx context.Context, input UpdateGenerativeTextModelInput) (GenerativeTextModelResponse, error) {
	profileID, err := uuid.Parse(input.ProfileID)
	if err != nil {
		return GenerativeTextModelResponse{}, errs.NewBadRequest("INVALID_PROFILE_ID")
	}

	// Get existing
	existing, err := s.store.GetGenerativeTextModelById(ctx, input.ID)
	if err == sql.ErrNoRows {
		return GenerativeTextModelResponse{}, errs.NewNotFound("GENERATIVE_TEXT_MODEL_NOT_FOUND")
	}
	if err != nil {
		return GenerativeTextModelResponse{}, errs.NewInternalServerError(err)
	}

	// Check if new model already exists (different from current)
	normalizedModel := strings.ToLower(input.Model)
	if normalizedModel != strings.ToLower(existing.Model) {
		existingByModel, err := s.store.GetGenerativeTextModelByModel(ctx, normalizedModel)
		if err != nil && err != sql.ErrNoRows {
			return GenerativeTextModelResponse{}, errs.NewInternalServerError(err)
		}
		if existingByModel.ID != 0 {
			return GenerativeTextModelResponse{}, errs.NewBadRequest("GENERATIVE_TEXT_MODEL_ALREADY_EXISTS")
		}
	}

	var image sql.NullString
	if input.Image != nil && *input.Image != "" {
		image = sql.NullString{String: *input.Image, Valid: true}
	}

	var result entity.AppGenerativeTextModel

	err = s.store.ExecTx(ctx, func(q *entity.Queries) error {
		model, updateErr := q.UpdateGenerativeTextModel(ctx, entity.UpdateGenerativeTextModelParams{
			ID:       input.ID,
			Model:    normalizedModel,
			Label:    input.Label,
			Image:    image,
			Provider: entity.AppGenerativeTextModelProviderType(input.Provider),
			IsActive: input.IsActive,
		})
		if updateErr != nil {
			return updateErr
		}

		result = model

		// Log change
		_, logErr := q.CreateGenerativeTextModelChange(ctx, entity.CreateGenerativeTextModelChangeParams{
			Action:                entity.ActionChangeTypeUpdate,
			ProfileID:             profileID,
			GenerativeTextModelID: model.ID,
			BeforeModel:           existing.Model,
			BeforeLabel:           existing.Label,
			BeforeImage:           existing.Image,
			BeforeProvider:        existing.Provider,
			BeforeIsActive:        existing.IsActive,
			AfterModel:            model.Model,
			AfterLabel:            model.Label,
			AfterImage:            model.Image,
			AfterProvider:         model.Provider,
			AfterIsActive:         model.IsActive,
		})
		return logErr
	})

	if err != nil {
		return GenerativeTextModelResponse{}, errs.NewInternalServerError(err)
	}

	return mapToResponse(result), nil
}

func (s *GenerativeTextModelService) DeleteGenerativeTextModel(ctx context.Context, id int64, profileID string) (GenerativeTextModelResponse, error) {
	pid, err := uuid.Parse(profileID)
	if err != nil {
		return GenerativeTextModelResponse{}, errs.NewBadRequest("INVALID_PROFILE_ID")
	}

	// Get existing
	existing, err := s.store.GetGenerativeTextModelById(ctx, id)
	if err == sql.ErrNoRows {
		return GenerativeTextModelResponse{}, errs.NewNotFound("GENERATIVE_TEXT_MODEL_NOT_FOUND")
	}
	if err != nil {
		return GenerativeTextModelResponse{}, errs.NewInternalServerError(err)
	}

	var result entity.AppGenerativeTextModel

	err = s.store.ExecTx(ctx, func(q *entity.Queries) error {
		model, deleteErr := q.SoftDeleteGenerativeTextModel(ctx, id)
		if deleteErr != nil {
			return deleteErr
		}

		result = model

		// Log change
		_, logErr := q.CreateGenerativeTextModelChange(ctx, entity.CreateGenerativeTextModelChangeParams{
			Action:                entity.ActionChangeTypeDelete,
			ProfileID:             pid,
			GenerativeTextModelID: model.ID,
			BeforeModel:           existing.Model,
			BeforeLabel:           existing.Label,
			BeforeImage:           existing.Image,
			BeforeProvider:        existing.Provider,
			BeforeIsActive:        existing.IsActive,
			AfterModel:            model.Model,
			AfterLabel:            model.Label,
			AfterImage:            model.Image,
			AfterProvider:         model.Provider,
			AfterIsActive:         model.IsActive,
		})
		return logErr
	})

	if err != nil {
		return GenerativeTextModelResponse{}, errs.NewInternalServerError(err)
	}

	return mapToResponse(result), nil
}

func mapToResponse(m entity.AppGenerativeTextModel) GenerativeTextModelResponse {
	var image *string
	if m.Image.Valid {
		image = &m.Image.String
	}

	return GenerativeTextModelResponse{
		ID:        m.ID,
		Model:     m.Model,
		Label:     m.Label,
		Image:     image,
		Provider:  GenerativeTextModelProviderType(m.Provider),
		IsActive:  m.IsActive,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}
