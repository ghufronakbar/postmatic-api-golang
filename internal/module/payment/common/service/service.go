// internal/module/payment/common/service/service.go
package payment_common_service

import (
	"context"
	"database/sql"

	"postmatic-api/internal/module/headless/midtrans"
	"postmatic-api/internal/repository/entity"
	"postmatic-api/pkg/errs"
	"postmatic-api/pkg/logger"
	"postmatic-api/pkg/pagination"
	"postmatic-api/pkg/utils"

	"github.com/google/uuid"
)

// PaymentCommonService handles common payment operations
type PaymentCommonService struct {
	store    entity.Store
	midtrans midtrans.Service
}

// NewService creates a new PaymentCommonService
func NewService(store entity.Store, midtrans midtrans.Service) *PaymentCommonService {
	return &PaymentCommonService{
		store:    store,
		midtrans: midtrans,
	}
}

// GetPaymentHistories returns paginated payment histories for a profile
func (s *PaymentCommonService) GetPaymentHistories(ctx context.Context, filter GetPaymentHistoriesFilter) ([]PaymentHistoryResponse, *pagination.Pagination, error) {
	profileID, err := uuid.Parse(filter.ProfileID)
	if err != nil {
		return nil, nil, errs.NewBadRequest("INVALID_PROFILE_ID")
	}

	// Convert status string to NullPaymentStatus
	var statusFilter entity.NullPaymentStatus
	if filter.Status != nil {
		statusFilter = entity.NullPaymentStatus{
			PaymentStatus: entity.PaymentStatus(*filter.Status),
			Valid:         true,
		}
	}

	// Count total
	count, err := s.store.CountAllPaymentHistories(ctx, entity.CountAllPaymentHistoriesParams{
		ProfileID: profileID,
		Search:    filter.Search,
		Status:    statusFilter,
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
	data, err := s.store.GetAllPaymentHistories(ctx, entity.GetAllPaymentHistoriesParams{
		ProfileID:  profileID,
		Search:     filter.Search,
		Status:     statusFilter,
		SortBy:     filter.SortBy,
		SortDir:    filter.SortDir,
		PageLimit:  int32(filter.Limit),
		PageOffset: int32(offset),
	})
	if err != nil {
		return nil, nil, errs.NewInternalServerError(err)
	}

	responses := make([]PaymentHistoryResponse, len(data))
	for i, d := range data {
		responses[i] = mapPaymentHistoryToResponse(d)
	}

	return responses, &pag, nil
}

// GetPaymentHistoryById returns a single payment history by ID
// If status is pending, it will check midtrans and update the status
func (s *PaymentCommonService) GetPaymentHistoryById(ctx context.Context, id string, profileID string) (PaymentHistoryResponse, error) {
	log := logger.From(ctx)

	paymentID, err := uuid.Parse(id)
	if err != nil {
		return PaymentHistoryResponse{}, errs.NewBadRequest("INVALID_PAYMENT_ID")
	}

	profID, err := uuid.Parse(profileID)
	if err != nil {
		return PaymentHistoryResponse{}, errs.NewBadRequest("INVALID_PROFILE_ID")
	}

	payment, err := s.store.GetPaymentHistoryByIdAndProfile(ctx, entity.GetPaymentHistoryByIdAndProfileParams{
		ID:        paymentID,
		ProfileID: profID,
	})
	if err == sql.ErrNoRows {
		return PaymentHistoryResponse{}, errs.NewNotFound("PAYMENT_NOT_FOUND")
	}
	if err != nil {
		return PaymentHistoryResponse{}, errs.NewInternalServerError(err)
	}

	// If pending, check midtrans status and update
	if payment.Status == entity.PaymentStatusPending && payment.MidtransTransactionID.Valid {
		log.Info("Payment is pending, checking midtrans status", "paymentID", id, "midtransID", payment.MidtransTransactionID.String)

		midtransStatus, err := s.midtrans.CheckStatus(ctx, payment.MidtransTransactionID.String)
		if err != nil {
			log.Error("Failed to check midtrans status", "error", err)
			// Continue with current data, don't fail
		} else {
			// Map midtrans status to our status
			newStatus := mapMidtransStatusToPaymentStatus(midtransStatus.TransactionStatus)
			if newStatus != string(payment.Status) {
				log.Info("Updating payment status from midtrans", "oldStatus", payment.Status, "newStatus", newStatus)

				updated, err := s.store.UpdatePaymentHistoryStatus(ctx, entity.UpdatePaymentHistoryStatusParams{
					ID:     payment.ID,
					Status: entity.PaymentStatus(newStatus),
				})
				if err != nil {
					log.Error("Failed to update payment status", "error", err)
				} else {
					payment = updated

					// Also update referral record status if applicable
					if payment.ReferralRecordID.Valid {
						_, _ = s.store.UpdateReferralRecordStatus(ctx, entity.UpdateReferralRecordStatusParams{
							ID:     payment.ReferralRecordID.Int64,
							Status: entity.ReferralRecordStatus(newStatus),
						})
					}
				}
			}
		}
	}

	return mapPaymentHistoryToResponse(payment), nil
}

// CancelPayment cancels a pending payment
func (s *PaymentCommonService) CancelPayment(ctx context.Context, id string, profileID string) (PaymentHistoryResponse, error) {
	log := logger.From(ctx)

	paymentID, err := uuid.Parse(id)
	if err != nil {
		return PaymentHistoryResponse{}, errs.NewBadRequest("INVALID_PAYMENT_ID")
	}

	profID, err := uuid.Parse(profileID)
	if err != nil {
		return PaymentHistoryResponse{}, errs.NewBadRequest("INVALID_PROFILE_ID")
	}

	payment, err := s.store.GetPaymentHistoryByIdAndProfile(ctx, entity.GetPaymentHistoryByIdAndProfileParams{
		ID:        paymentID,
		ProfileID: profID,
	})
	if err == sql.ErrNoRows {
		return PaymentHistoryResponse{}, errs.NewNotFound("PAYMENT_NOT_FOUND")
	}
	if err != nil {
		return PaymentHistoryResponse{}, errs.NewInternalServerError(err)
	}

	// Can only cancel pending payments
	if payment.Status != entity.PaymentStatusPending {
		return PaymentHistoryResponse{}, errs.NewBadRequest("PAYMENT_CANNOT_BE_CANCELED")
	}

	// Cancel in midtrans if has transaction ID
	if payment.MidtransTransactionID.Valid {
		log.Info("Canceling payment in midtrans", "paymentID", id, "midtransID", payment.MidtransTransactionID.String)
		_, err := s.midtrans.CancelTransaction(ctx, payment.MidtransTransactionID.String)
		if err != nil {
			log.Error("Failed to cancel in midtrans", "error", err)
			// Continue anyway
		}
	}

	// Update status to canceled
	updated, err := s.store.UpdatePaymentHistoryStatus(ctx, entity.UpdatePaymentHistoryStatusParams{
		ID:     payment.ID,
		Status: entity.PaymentStatusCanceled,
	})
	if err != nil {
		return PaymentHistoryResponse{}, errs.NewInternalServerError(err)
	}

	// Also update referral record status if applicable
	if payment.ReferralRecordID.Valid {
		_, _ = s.store.UpdateReferralRecordStatus(ctx, entity.UpdateReferralRecordStatusParams{
			ID:     payment.ReferralRecordID.Int64,
			Status: entity.ReferralRecordStatusCanceled,
		})
	}

	return mapPaymentHistoryToResponse(updated), nil
}

// HandleWebhook handles midtrans webhook notification
func (s *PaymentCommonService) HandleWebhook(ctx context.Context, notification MidtransNotification) error {
	log := logger.From(ctx)

	// 1. Verify signature
	isValid := s.midtrans.VerifySignature(
		notification.OrderID,
		notification.StatusCode,
		notification.GrossAmount,
		notification.SignatureKey,
	)
	if !isValid {
		log.Error("Invalid webhook signature", "orderID", notification.OrderID)
		return errs.NewUnauthorized("INVALID_SIGNATURE")
	}

	log.Info("Webhook received", "orderID", notification.OrderID, "status", notification.TransactionStatus)

	// 2. Find payment by midtrans transaction ID
	payment, err := s.store.GetPaymentHistoryByMidtransTransactionId(ctx, utils.StringToNullString(&notification.TransactionID))
	if err == sql.ErrNoRows {
		log.Error("Payment not found for webhook", "transactionID", notification.TransactionID)
		return errs.NewNotFound("PAYMENT_NOT_FOUND")
	}
	if err != nil {
		return errs.NewInternalServerError(err)
	}

	// 3. Map status and update
	newStatus := mapMidtransStatusToPaymentStatus(notification.TransactionStatus)
	if newStatus != string(payment.Status) {
		log.Info("Updating payment status from webhook", "oldStatus", payment.Status, "newStatus", newStatus)

		_, err := s.store.UpdatePaymentHistoryStatus(ctx, entity.UpdatePaymentHistoryStatusParams{
			ID:     payment.ID,
			Status: entity.PaymentStatus(newStatus),
		})
		if err != nil {
			return errs.NewInternalServerError(err)
		}

		// Also update referral record status if applicable
		if payment.ReferralRecordID.Valid {
			_, _ = s.store.UpdateReferralRecordStatus(ctx, entity.UpdateReferralRecordStatusParams{
				ID:     payment.ReferralRecordID.Int64,
				Status: entity.ReferralRecordStatus(newStatus),
			})
		}

		// TODO: If success, add tokens to user's balance
	}

	return nil
}

// Helpers

func mapMidtransStatusToPaymentStatus(midtransStatus string) string {
	switch midtransStatus {
	case "capture", "settlement":
		return "success"
	case "pending":
		return "pending"
	case "deny":
		return "denied"
	case "cancel":
		return "canceled"
	case "expire":
		return "expired"
	case "refund", "partial_refund":
		return "refunded"
	case "failure":
		return "failed"
	default:
		return "pending"
	}
}

func mapPaymentHistoryToResponse(p entity.PaymentHistory) PaymentHistoryResponse {
	resp := PaymentHistoryResponse{
		ID:                 p.ID.String(),
		ProductAmount:      p.ProductAmount,
		Status:             string(p.Status),
		Currency:           p.Currency,
		PaymentMethod:      p.PaymentMethod,
		PaymentMethodType:  p.PaymentMethodType,
		ProductName:        p.RecordProductName,
		ProductType:        string(p.RecordProductType),
		ProductPrice:       p.RecordProductPrice,
		ProductImageUrl:    p.RecordProductImageUrl,
		SubtotalItemAmount: p.SubtotalItemAmount,
		DiscountAmount:     p.DiscountAmount,
		AdminFeeAmount:     p.AdminFeeAmount,
		TaxAmount:          p.TaxAmount,
		TotalAmount:        p.TotalAmount,
		CreatedAt:          p.CreatedAt,
		UpdatedAt:          p.UpdatedAt,
	}

	if p.MidtransExpiredAt.Valid {
		resp.MidtransExpiredAt = &p.MidtransExpiredAt.Time
	}
	if p.PaymentPendingAt.Valid {
		resp.PaymentPendingAt = &p.PaymentPendingAt.Time
	}
	if p.PaymentSuccessAt.Valid {
		resp.PaymentSuccessAt = &p.PaymentSuccessAt.Time
	}
	if p.PaymentFailedAt.Valid {
		resp.PaymentFailedAt = &p.PaymentFailedAt.Time
	}
	if p.PaymentCanceledAt.Valid {
		resp.PaymentCanceledAt = &p.PaymentCanceledAt.Time
	}
	if p.PaymentExpiredAt.Valid {
		resp.PaymentExpiredAt = &p.PaymentExpiredAt.Time
	}

	return resp
}
