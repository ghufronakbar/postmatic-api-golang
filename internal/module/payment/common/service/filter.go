// internal/module/payment/common/service/filter.go
package payment_common_service

// SORT_BY defines allowed sort fields for payment histories
var SORT_BY = []string{"created_at", "total_amount"}

// GetPaymentHistoriesFilter is the input for filtering payment histories by profile
type GetPaymentHistoriesFilter struct {
	ProfileID string
	Search    *string
	Status    *string
	SortBy    string
	SortDir   string
	Page      int
	Limit     int
}

// GetPaymentHistoriesByBusinessFilter is the input for filtering payment histories by business
type GetPaymentHistoriesByBusinessFilter struct {
	BusinessRootID int64
	Search         *string
	Status         *string
	SortBy         string
	SortDir        string
	Page           int
	Limit          int
}
