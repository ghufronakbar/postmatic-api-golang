// internal/module/generative_token/image_token/service/service.go
package image_token_service

import (
	"context"
	"database/sql"
	"time"

	"postmatic-api/internal/repository/entity"
	"postmatic-api/pkg/errs"
	"postmatic-api/pkg/logger"
	"postmatic-api/pkg/pagination"

	"github.com/google/uuid"
)

// ImageTokenService handles generative image token operations
type ImageTokenService struct {
	store entity.Store
}

// NewService creates a new ImageTokenService
func NewService(store entity.Store) *ImageTokenService {
	return &ImageTokenService{
		store: store,
	}
}

// CreditTokenFromPayment creates a token transaction for successful payment
// This method accepts *entity.Queries to be used within a transaction
func (s *ImageTokenService) CreditTokenFromPayment(ctx context.Context, q *entity.Queries, input CreateTokenTransactionInput) error {
	log := logger.From(ctx)

	// Check if already credited
	existing, err := q.GetGenerativeTokenImageTransactionByPaymentHistoryId(ctx, uuid.NullUUID{
		UUID:  input.PaymentHistoryID,
		Valid: true,
	})
	if err != nil && err != sql.ErrNoRows {
		log.Error("Failed to check existing token transaction", "paymentHistoryId", input.PaymentHistoryID, "error", err)
		return errs.NewInternalServerError(err)
	}
	if existing.ID != 0 {
		log.Info("Token already credited for payment", "paymentHistoryId", input.PaymentHistoryID)
		return nil // Already credited, skip
	}

	// Create token transaction
	_, err = q.CreateGenerativeTokenImageTransaction(ctx, entity.CreateGenerativeTokenImageTransactionParams{
		Type:             entity.TokenTransactionTypeIn,
		Amount:           input.Amount,
		ProfileID:        input.ProfileID,
		BusinessRootID:   input.BusinessRootID,
		PaymentHistoryID: uuid.NullUUID{UUID: input.PaymentHistoryID, Valid: true},
	})
	if err != nil {
		log.Error("Failed to create token transaction", "paymentHistoryId", input.PaymentHistoryID, "error", err)
		return errs.NewInternalServerError(err)
	}

	log.Info("Token credited successfully", "paymentHistoryId", input.PaymentHistoryID, "amount", input.Amount)
	return nil
}

// SyncMissingTokenTransactions syncs token transactions for successful payments that are missing
// This runs in background goroutine, so it uses store directly (not transaction)
func (s *ImageTokenService) SyncMissingTokenTransactions(ctx context.Context, paymentIDs []uuid.UUID) {
	if len(paymentIDs) == 0 {
		return
	}

	log := logger.From(ctx)
	log.Info("Starting sync missing token transactions", "count", len(paymentIDs))

	// Find payments that are success but don't have token transaction
	missingPayments, err := s.store.GetSuccessPaymentIdsWithoutTokenTransaction(ctx, paymentIDs)
	if err != nil {
		log.Error("Failed to get missing payments", "error", err)
		return
	}

	if len(missingPayments) == 0 {
		log.Info("No missing token transactions to sync")
		return
	}

	log.Info("Found missing token transactions", "count", len(missingPayments))

	// Create token transactions for each missing payment
	for _, payment := range missingPayments {
		_, err := s.store.CreateGenerativeTokenImageTransaction(ctx, entity.CreateGenerativeTokenImageTransactionParams{
			Type:             entity.TokenTransactionTypeIn,
			Amount:           payment.ProductAmount,
			ProfileID:        payment.ProfileID,
			BusinessRootID:   payment.BusinessRootID,
			PaymentHistoryID: uuid.NullUUID{UUID: payment.ID, Valid: true},
		})
		if err != nil {
			log.Error("Failed to sync token transaction", "paymentId", payment.ID, "error", err)
			continue
		}
		log.Info("Synced token transaction", "paymentId", payment.ID, "amount", payment.ProductAmount)
	}

	log.Info("Finished sync missing token transactions")
}

// GetTokenStatus returns token status for a business
func (s *ImageTokenService) GetTokenStatus(ctx context.Context, businessRootID int64) (TokenStatusResponse, error) {
	var response TokenStatusResponse

	// Get total token in (purchased)
	totalIn, err := s.store.SumTokenByBusinessAndType(ctx, entity.SumTokenByBusinessAndTypeParams{
		BusinessRootID: businessRootID,
		Type:           entity.TokenTransactionTypeIn,
	})
	if err != nil {
		return response, errs.NewInternalServerError(err)
	}

	// Get total token out (used)
	totalOut, err := s.store.SumTokenByBusinessAndType(ctx, entity.SumTokenByBusinessAndTypeParams{
		BusinessRootID: businessRootID,
		Type:           entity.TokenTransactionTypeOut,
	})
	if err != nil {
		return response, errs.NewInternalServerError(err)
	}

	response.TotalToken = totalIn
	response.UsedToken = totalOut
	response.AvailableToken = totalIn - totalOut
	response.IsExhausted = response.AvailableToken <= 0

	return response, nil
}

// GetTokenTransactions returns paginated token transactions for a business
func (s *ImageTokenService) GetTokenTransactions(ctx context.Context, filter GetTokenTransactionsFilter) ([]TokenTransactionResponse, *pagination.Pagination, error) {
	// Convert type string to NullTokenTransactionType
	var typeFilter entity.NullTokenTransactionType
	if filter.Type != nil && (*filter.Type == "in" || *filter.Type == "out") {
		typeFilter = entity.NullTokenTransactionType{
			TokenTransactionType: entity.TokenTransactionType(*filter.Type),
			Valid:                true,
		}
	}

	// Convert date strings to sql.NullTime
	dateStart := parseDateToNullTime(filter.DateStart)
	dateEnd := parseDateToNullTime(filter.DateEnd)

	// Count total
	count, err := s.store.CountAllTokenTransactionsByBusiness(ctx, entity.CountAllTokenTransactionsByBusinessParams{
		BusinessRootID: filter.BusinessRootID,
		Type:           typeFilter,
		DateStart:      dateStart,
		DateEnd:        dateEnd,
	})
	if err != nil {
		return nil, nil, errs.NewInternalServerError(err)
	}

	pag := pagination.NewPagination(&pagination.PaginationParams{
		Total: int(count),
		Page:  filter.Page,
		Limit: filter.Limit,
	})

	// Calculate offset
	offset := (filter.Page - 1) * filter.Limit
	if offset < 0 {
		offset = 0
	}

	// Get data
	data, err := s.store.GetAllTokenTransactionsByBusiness(ctx, entity.GetAllTokenTransactionsByBusinessParams{
		BusinessRootID: filter.BusinessRootID,
		Type:           typeFilter,
		DateStart:      dateStart,
		DateEnd:        dateEnd,
		SortBy:         filter.SortBy,
		SortDir:        filter.SortDir,
		PageLimit:      int32(filter.Limit),
		PageOffset:     int32(offset),
	})
	if err != nil {
		return nil, nil, errs.NewInternalServerError(err)
	}

	responses := make([]TokenTransactionResponse, len(data))
	for i, d := range data {
		responses[i] = mapTokenTransactionToResponse(d)
	}

	return responses, &pag, nil
}

// mapTokenTransactionToResponse maps entity to response
func mapTokenTransactionToResponse(t entity.GenerativeTokenImageTransaction) TokenTransactionResponse {
	var paymentHistoryID *uuid.UUID
	if t.PaymentHistoryID.Valid {
		paymentHistoryID = &t.PaymentHistoryID.UUID
	}

	return TokenTransactionResponse{
		ID:               t.ID,
		Type:             string(t.Type),
		Amount:           t.Amount,
		ProfileID:        t.ProfileID,
		BusinessRootID:   t.BusinessRootID,
		PaymentHistoryID: paymentHistoryID,
		CreatedAt:        t.CreatedAt,
	}
}

// parseDateToNullTime converts a date string (YYYY-MM-DD) to sql.NullTime
func parseDateToNullTime(dateStr *string) sql.NullTime {
	if dateStr == nil || *dateStr == "" {
		return sql.NullTime{Valid: false}
	}
	t, err := time.Parse("2006-01-02", *dateStr)
	if err != nil {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{Time: t, Valid: true}
}
