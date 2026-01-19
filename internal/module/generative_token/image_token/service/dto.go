// internal/module/generative_token/image_token/service/dto.go
package image_token_service

import "github.com/google/uuid"

// CreateTokenTransactionInput is input for crediting token from payment
type CreateTokenTransactionInput struct {
	ProfileID        uuid.UUID
	BusinessRootID   int64
	PaymentHistoryID uuid.UUID
	Amount           int64
}

// SyncMissingTokenInput is input for bulk sync missing token transactions
type SyncMissingTokenInput struct {
	PaymentIDs []uuid.UUID
}
