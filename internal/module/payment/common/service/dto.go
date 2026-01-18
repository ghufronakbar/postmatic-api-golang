// internal/module/payment/common/service/dto.go
package payment_common_service

// GetPaymentHistoriesFilter is the input for filtering payment histories
type GetPaymentHistoriesFilter struct {
	ProfileID string
	Search    *string
	Status    *string
	SortBy    string
	SortDir   string
	Page      int
	Limit     int
}
