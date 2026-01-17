package payment_method_service

import (
	"context"
	"database/sql"
	"strings"

	"postmatic-api/internal/repository/entity"
	"postmatic-api/pkg/errs"
	"postmatic-api/pkg/pagination"

	"github.com/google/uuid"
)

type PaymentMethodService struct {
	store entity.Store
}

func NewService(store entity.Store) *PaymentMethodService {
	return &PaymentMethodService{store: store}
}

func (s *PaymentMethodService) GetAllPaymentMethods(ctx context.Context, filter GetPaymentMethodsFilter) ([]PaymentMethodResponse, *pagination.Pagination, error) {
	params := entity.GetAllPaymentMethodsParams{
		Search:     sql.NullString{String: filter.Search, Valid: filter.Search != ""},
		SortBy:     filter.SortBy,
		SortDir:    filter.SortDir,
		PageOffset: int32(filter.PageOffset),
		PageLimit:  int32(filter.PageLimit),
		IsAdmin:    filter.IsAdmin,
	}

	methods, err := s.store.GetAllPaymentMethods(ctx, params)
	if err != nil {
		return nil, nil, errs.NewInternalServerError(err)
	}

	countParams := entity.CountAllPaymentMethodsParams{
		Search:  sql.NullString{String: filter.Search, Valid: filter.Search != ""},
		IsAdmin: filter.IsAdmin,
	}

	count, err := s.store.CountAllPaymentMethods(ctx, countParams)
	if err != nil {
		return nil, nil, errs.NewInternalServerError(err)
	}

	pag := pagination.NewPagination(&pagination.PaginationParams{
		Total: int(count),
		Page:  filter.Page,
		Limit: filter.PageLimit,
	})

	responses := make([]PaymentMethodResponse, 0, len(methods))
	for _, m := range methods {
		responses = append(responses, mapToResponse(m))
	}

	return responses, &pag, nil
}

func (s *PaymentMethodService) GetPaymentMethodById(ctx context.Context, id int64, isAdmin bool) (PaymentMethodResponse, error) {
	var method entity.AppPaymentMethod
	var err error

	if isAdmin {
		method, err = s.store.GetPaymentMethodByIdAdmin(ctx, id)
	} else {
		method, err = s.store.GetPaymentMethodByIdUser(ctx, id)
	}

	if err == sql.ErrNoRows {
		return PaymentMethodResponse{}, errs.NewNotFound("PAYMENT_METHOD_NOT_FOUND")
	}
	if err != nil {
		return PaymentMethodResponse{}, errs.NewInternalServerError(err)
	}

	return mapToResponse(method), nil
}

func (s *PaymentMethodService) GetPaymentMethodByCode(ctx context.Context, code string, isAdmin bool) (PaymentMethodResponse, error) {
	var method entity.AppPaymentMethod
	var err error

	if isAdmin {
		method, err = s.store.GetPaymentMethodByCodeAdmin(ctx, strings.ToUpper(code))
	} else {
		method, err = s.store.GetPaymentMethodByCodeUser(ctx, strings.ToUpper(code))
	}

	if err == sql.ErrNoRows {
		return PaymentMethodResponse{}, errs.NewNotFound("PAYMENT_METHOD_NOT_FOUND")
	}
	if err != nil {
		return PaymentMethodResponse{}, errs.NewInternalServerError(err)
	}

	return mapToResponse(method), nil
}

func (s *PaymentMethodService) CreatePaymentMethod(ctx context.Context, input CreatePaymentMethodInput) (PaymentMethodResponse, error) {
	profileID, err := uuid.Parse(input.ProfileID)
	if err != nil {
		return PaymentMethodResponse{}, errs.NewBadRequest("INVALID_PROFILE_ID")
	}

	// Check if code already exists
	normalizedCode := strings.ToUpper(input.Code)
	existing, err := s.store.GetPaymentMethodByCode(ctx, normalizedCode)
	if err != nil && err != sql.ErrNoRows {
		return PaymentMethodResponse{}, errs.NewInternalServerError(err)
	}
	if existing.ID != 0 {
		return PaymentMethodResponse{}, errs.NewBadRequest("PAYMENT_METHOD_CODE_ALREADY_EXISTS")
	}

	var image sql.NullString
	if input.Image != nil && *input.Image != "" {
		image = sql.NullString{String: *input.Image, Valid: true}
	}

	var result entity.AppPaymentMethod

	err = s.store.ExecTx(ctx, func(q *entity.Queries) error {
		method, createErr := q.CreatePaymentMethod(ctx, entity.CreatePaymentMethodParams{
			Code:      strings.ToUpper(input.Code),
			Name:      input.Name,
			Type:      entity.AppPaymentMethodType(input.Type),
			Image:     image,
			TaxFee:    input.TaxFee,
			AdminType: entity.AppPaymentAdminType(input.AdminType),
			AdminFee:  input.AdminFee,
			IsActive:  input.IsActive,
		})
		if createErr != nil {
			return createErr
		}

		result = method

		// Log change
		_, logErr := q.CreatePaymentMethodChange(ctx, entity.CreatePaymentMethodChangeParams{
			Action:          entity.ActionChangeTypeCreate,
			ProfileID:       profileID,
			PaymentMethodID: method.ID,
			BeforeCode:      strings.ToUpper(input.Code),
			BeforeName:      method.Name,
			BeforeType:      method.Type,
			BeforeImage:     method.Image,
			BeforeAdminType: method.AdminType,
			BeforeAdminFee:  method.AdminFee,
			BeforeTaxFee:    method.TaxFee,
			BeforeIsActive:  method.IsActive,
			AfterCode:       strings.ToUpper(input.Code),
			AfterName:       method.Name,
			AfterType:       method.Type,
			AfterImage:      method.Image,
			AfterAdminType:  method.AdminType,
			AfterAdminFee:   method.AdminFee,
			AfterTaxFee:     method.TaxFee,
			AfterIsActive:   method.IsActive,
		})
		return logErr
	})

	if err != nil {
		return PaymentMethodResponse{}, errs.NewInternalServerError(err)
	}

	return mapToResponse(result), nil
}

func (s *PaymentMethodService) UpdatePaymentMethod(ctx context.Context, input UpdatePaymentMethodInput) (PaymentMethodResponse, error) {
	profileID, err := uuid.Parse(input.ProfileID)
	if err != nil {
		return PaymentMethodResponse{}, errs.NewBadRequest("INVALID_PROFILE_ID")
	}

	// Get existing
	existing, err := s.store.GetPaymentMethodById(ctx, input.ID)
	if err == sql.ErrNoRows {
		return PaymentMethodResponse{}, errs.NewNotFound("PAYMENT_METHOD_NOT_FOUND")
	}
	if err != nil {
		return PaymentMethodResponse{}, errs.NewInternalServerError(err)
	}

	// Check if new code already exists (different from current)
	normalizedCode := strings.ToUpper(input.Code)
	if normalizedCode != strings.ToUpper(existing.Code) {
		existingByCode, err := s.store.GetPaymentMethodByCode(ctx, normalizedCode)
		if err != nil && err != sql.ErrNoRows {
			return PaymentMethodResponse{}, errs.NewInternalServerError(err)
		}
		if existingByCode.ID != 0 {
			return PaymentMethodResponse{}, errs.NewBadRequest("PAYMENT_METHOD_CODE_ALREADY_EXISTS")
		}
	}

	var image sql.NullString
	if input.Image != nil && *input.Image != "" {
		image = sql.NullString{String: *input.Image, Valid: true}
	}

	var result entity.AppPaymentMethod

	err = s.store.ExecTx(ctx, func(q *entity.Queries) error {
		method, updateErr := q.UpdatePaymentMethod(ctx, entity.UpdatePaymentMethodParams{
			ID:        input.ID,
			Code:      strings.ToUpper(input.Code),
			Name:      input.Name,
			Type:      entity.AppPaymentMethodType(input.Type),
			Image:     image,
			TaxFee:    input.TaxFee,
			AdminType: entity.AppPaymentAdminType(input.AdminType),
			AdminFee:  input.AdminFee,
			IsActive:  input.IsActive,
		})
		if updateErr != nil {
			return updateErr
		}

		result = method

		// Log change
		_, logErr := q.CreatePaymentMethodChange(ctx, entity.CreatePaymentMethodChangeParams{
			Action:          entity.ActionChangeTypeUpdate,
			ProfileID:       profileID,
			PaymentMethodID: method.ID,
			BeforeCode:      strings.ToUpper(existing.Code),
			BeforeName:      existing.Name,
			BeforeType:      existing.Type,
			BeforeImage:     existing.Image,
			BeforeAdminType: existing.AdminType,
			BeforeAdminFee:  existing.AdminFee,
			BeforeTaxFee:    existing.TaxFee,
			BeforeIsActive:  existing.IsActive,
			AfterCode:       strings.ToUpper(input.Code),
			AfterName:       method.Name,
			AfterType:       method.Type,
			AfterImage:      method.Image,
			AfterAdminType:  method.AdminType,
			AfterAdminFee:   method.AdminFee,
			AfterTaxFee:     method.TaxFee,
			AfterIsActive:   method.IsActive,
		})
		return logErr
	})

	if err != nil {
		return PaymentMethodResponse{}, errs.NewInternalServerError(err)
	}

	return mapToResponse(result), nil
}

func (s *PaymentMethodService) DeletePaymentMethod(ctx context.Context, id int64, profileID string) (PaymentMethodResponse, error) {
	pid, err := uuid.Parse(profileID)
	if err != nil {
		return PaymentMethodResponse{}, errs.NewBadRequest("INVALID_PROFILE_ID")
	}

	// Get existing
	existing, err := s.store.GetPaymentMethodById(ctx, id)
	if err == sql.ErrNoRows {
		return PaymentMethodResponse{}, errs.NewNotFound("PAYMENT_METHOD_NOT_FOUND")
	}
	if err != nil {
		return PaymentMethodResponse{}, errs.NewInternalServerError(err)
	}

	var result entity.AppPaymentMethod

	err = s.store.ExecTx(ctx, func(q *entity.Queries) error {
		method, deleteErr := q.SoftDeletePaymentMethod(ctx, id)
		if deleteErr != nil {
			return deleteErr
		}

		result = method

		// Log change
		_, logErr := q.CreatePaymentMethodChange(ctx, entity.CreatePaymentMethodChangeParams{
			Action:          entity.ActionChangeTypeDelete,
			ProfileID:       pid,
			PaymentMethodID: method.ID,
			BeforeCode:      strings.ToUpper(existing.Code),
			BeforeName:      existing.Name,
			BeforeType:      existing.Type,
			BeforeImage:     existing.Image,
			BeforeAdminType: existing.AdminType,
			BeforeAdminFee:  existing.AdminFee,
			BeforeTaxFee:    existing.TaxFee,
			BeforeIsActive:  existing.IsActive,
			AfterCode:       strings.ToUpper(method.Code),
			AfterName:       method.Name,
			AfterType:       method.Type,
			AfterImage:      method.Image,
			AfterAdminType:  method.AdminType,
			AfterAdminFee:   method.AdminFee,
			AfterTaxFee:     method.TaxFee,
			AfterIsActive:   method.IsActive,
		})
		return logErr
	})

	if err != nil {
		return PaymentMethodResponse{}, errs.NewInternalServerError(err)
	}

	return mapToResponse(result), nil
}

func mapToResponse(m entity.AppPaymentMethod) PaymentMethodResponse {
	var image *string
	if m.Image.Valid {
		image = &m.Image.String
	}

	return PaymentMethodResponse{
		ID:        m.ID,
		Code:      strings.ToUpper(m.Code),
		Name:      m.Name,
		Type:      PaymentMethodType(m.Type),
		Image:     image,
		TaxFee:    m.TaxFee,
		AdminType: PaymentMethodAdminType(m.AdminType),
		AdminFee:  m.AdminFee,
		IsActive:  m.IsActive,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}
