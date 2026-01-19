// internal/module/generative_token/image_token/service/viewmodel.go
package image_token_service

import (
	"time"

	"github.com/google/uuid"
)

// TokenStatusResponse is response for GET /status endpoint
type TokenStatusResponse struct {
	AvailableToken int64 `json:"availableToken"`
	UsedToken      int64 `json:"usedToken"`
	TotalToken     int64 `json:"totalToken"`
	IsExhausted    bool  `json:"isExhausted"`
}

// TokenTransactionResponse is response for token transaction item
type TokenTransactionResponse struct {
	ID               int64      `json:"id"`
	Type             string     `json:"type"`
	Amount           int64      `json:"amount"`
	ProfileID        uuid.UUID  `json:"profileId"`
	BusinessRootID   int64      `json:"businessRootId"`
	PaymentHistoryID *uuid.UUID `json:"paymentHistoryId"` // null if type is 'out'
	CreatedAt        time.Time  `json:"createdAt"`
}
