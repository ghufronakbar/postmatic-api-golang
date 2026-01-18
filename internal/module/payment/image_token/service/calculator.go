// internal/module/payment/image_token/service/calculator.go
package image_token_service

import "math"

// PriceCalculationResult contains the detailed breakdown of price calculation
type PriceCalculationResult struct {
	ItemPrice         int64  // harga asli
	DiscountAmount    int64  // nominal diskon yang didapat
	DiscountPercent   *int64 // percentage jika applicable
	AfterDiscount     int64  // harga setelah diskon
	AdminFeeAmount    int64  // nominal admin fee
	AdminFeePercent   *int64 // percentage jika applicable
	SubtotalBeforeTax int64  // subtotal sebelum tax
	TaxAmount         int64  // nominal tax
	TaxPercent        int64  // tax percentage
	TotalAmount       int64  // grand total
}

// CalculatePrice calculates the payment price with discount, admin fee, and tax
// Follows the sequential calculation with rounding up at each step
func CalculatePrice(input PriceCalculationInput) PriceCalculationResult {
	result := PriceCalculationResult{
		ItemPrice:  input.BasePrice,
		TaxPercent: input.TaxPercentage,
	}

	// Step A: Calculate discount
	var discountAmount int64
	if input.DiscountType == "fixed" {
		discountAmount = input.DiscountValue
	} else if input.DiscountType == "percentage" {
		// Calculate percentage of base price, round up
		discountAmount = roundUp(float64(input.BasePrice) * float64(input.DiscountValue) / 100)
		pct := input.DiscountValue
		result.DiscountPercent = &pct
	}
	// Apply max cap
	if input.MaxDiscount > 0 && discountAmount > input.MaxDiscount {
		discountAmount = input.MaxDiscount
	}
	// Discount cannot exceed base price
	if discountAmount > input.BasePrice {
		discountAmount = input.BasePrice
	}
	result.DiscountAmount = discountAmount

	// Step B: Calculate admin fee (from ORIGINAL price, not discounted)
	var adminFeeAmount int64
	if input.AdminFeeType == "fixed" {
		adminFeeAmount = input.AdminFeeValue
	} else if input.AdminFeeType == "percentage" {
		adminFeeAmount = roundUp(float64(input.BasePrice) * float64(input.AdminFeeValue) / 100)
		pct := input.AdminFeeValue
		result.AdminFeePercent = &pct
	}
	result.AdminFeeAmount = adminFeeAmount

	// Step C: Calculate subtotal before tax
	afterDiscount := input.BasePrice - discountAmount
	if afterDiscount < 0 {
		afterDiscount = 0
	}
	result.AfterDiscount = afterDiscount
	result.SubtotalBeforeTax = roundUp(float64(afterDiscount + adminFeeAmount))

	// Step D: Calculate tax (from subtotal before tax)
	taxAmount := roundUp(float64(result.SubtotalBeforeTax) * float64(input.TaxPercentage) / 100)
	result.TaxAmount = taxAmount

	// Step E: Calculate total
	result.TotalAmount = result.SubtotalBeforeTax + taxAmount

	return result
}

// roundUp rounds a float64 up to the nearest integer
func roundUp(f float64) int64 {
	return int64(math.Ceil(f))
}
