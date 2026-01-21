package midtrans

import (
	"context"

	"postmatic-api/pkg/errs"
	"postmatic-api/pkg/logger"

	"github.com/midtrans/midtrans-go/coreapi"
)

// Service defines the contract for Midtrans interactions
type Service interface {
	// Charge e-Wallet (Gopay, ShopeePay, etc) -> implementation in ewallet.go
	ChargeGopay(ctx context.Context, req ChargeGopayInput) (*ChargeResponse, error)

	// Charge Bank Transfer (BCA, BNI, Permata, etc) -> implementation in bank.go
	ChargeBankTransfer(ctx context.Context, req ChargeBankTransferInput) (*ChargeResponse, error)

	// Transaction Management -> implementation in service.go
	CheckStatus(ctx context.Context, orderID string) (*TransactionStatusResponse, error)
	CancelTransaction(ctx context.Context, orderID string) (*TransactionStatusResponse, error)
	ExpireTransaction(ctx context.Context, orderID string) (*TransactionStatusResponse, error)

	// Security -> implementation in helper.go
	VerifySignature(orderID, statusCode, grossAmount, signatureKey string) bool
}

// midtransService implements the Service interface
type midtransService struct {
	client coreapi.Client
}

// NewService creates a new Midtrans service instance
func NewService(client *coreapi.Client) Service {
	return &midtransService{
		client: *client,
	}
}

// CheckStatus checks the transaction status by orderID
func (s *midtransService) CheckStatus(ctx context.Context, orderID string) (*TransactionStatusResponse, error) {
	log := logger.From(ctx)
	log.Info("Checking transaction status", "orderID", orderID)

	res, err := s.client.CheckTransaction(orderID)
	if err != nil {
		log.Error("Failed to check transaction status", "orderID", orderID, "error", err)
		return nil, errs.NewBadRequest("MIDTRANS_CHECK_STATUS_FAILED")
	}

	log.Info("Transaction status retrieved", "orderID", orderID, "status", res.TransactionStatus)
	return mapTransactionStatusResponse(res), nil
}

// CancelTransaction cancels a transaction by orderID
func (s *midtransService) CancelTransaction(ctx context.Context, orderID string) (*TransactionStatusResponse, error) {
	log := logger.From(ctx)
	log.Info("Cancelling transaction", "orderID", orderID)

	res, err := s.client.CancelTransaction(orderID)
	if err != nil {
		log.Error("Failed to cancel transaction", "orderID", orderID, "error", err)
		return nil, errs.NewBadRequest("MIDTRANS_CANCEL_TRANSACTION_FAILED")
	}

	log.Info("Transaction cancelled", "orderID", orderID, "status", res.TransactionStatus)
	return mapCancelResponse(res), nil
}

// ExpireTransaction expires a pending transaction by orderID
func (s *midtransService) ExpireTransaction(ctx context.Context, orderID string) (*TransactionStatusResponse, error) {
	log := logger.From(ctx)
	log.Info("Expiring transaction", "orderID", orderID)

	res, err := s.client.ExpireTransaction(orderID)
	if err != nil {
		log.Error("Failed to expire transaction", "orderID", orderID, "error", err)
		return nil, errs.NewBadRequest("MIDTRANS_EXPIRE_TRANSACTION_FAILED")
	}

	log.Info("Transaction expired", "orderID", orderID, "status", res.TransactionStatus)
	return mapExpireResponse(res), nil
}
