// internal/module/payment/image_token/service/service.go
package image_token_service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	referral_basic_service "postmatic-api/internal/module/affiliator/referral_basic/service"
	payment_method_service "postmatic-api/internal/module/app/payment_method/service"
	token_product_service "postmatic-api/internal/module/app/token_product/service"
	"postmatic-api/internal/module/headless/midtrans"
	"postmatic-api/internal/repository/entity"
	"postmatic-api/pkg/errs"
	"postmatic-api/pkg/utils"

	"github.com/google/uuid"
)

// ImageTokenPaymentService handles image token payment operations
type ImageTokenPaymentService struct {
	store         entity.Store
	tokenProduct  *token_product_service.TokenProductService
	paymentMethod *payment_method_service.PaymentMethodService
	referral      *referral_basic_service.ReferralBasicService
	midtrans      midtrans.Service
}

// NewService creates a new ImageTokenPaymentService
func NewService(
	store entity.Store,
	tokenProduct *token_product_service.TokenProductService,
	paymentMethod *payment_method_service.PaymentMethodService,
	referral *referral_basic_service.ReferralBasicService,
	midtrans midtrans.Service,
) *ImageTokenPaymentService {
	return &ImageTokenPaymentService{
		store:         store,
		tokenProduct:  tokenProduct,
		paymentMethod: paymentMethod,
		referral:      referral,
		midtrans:      midtrans,
	}
}

// CheckPrice calculates the total price for checkout preview
func (s *ImageTokenPaymentService) CheckPrice(ctx context.Context, input CheckPriceInput) (CheckPriceResponse, error) {
	var response CheckPriceResponse

	// 1. Get token product price
	tokenCalc, err := s.tokenProduct.CalculateTokenProduct(ctx, token_product_service.TokenCalculateProductFilter{
		Type:         string(entity.TokenTypeImageToken),
		CurrencyCode: strings.ToUpper(input.CurrencyCode),
		From:         "token",
		Amount:       input.TokenAmount,
	})
	if err != nil {
		return response, err
	}

	// 2. Get payment method
	pm, err := s.paymentMethod.GetPaymentMethodByCode(ctx, input.PaymentMethod, false)
	if err != nil {
		return response, err
	}
	if !pm.IsActive {
		return response, errs.NewBadRequest("PAYMENT_METHOD_INACTIVE")
	}

	// 3. Build price calculation input
	calcInput := PriceCalculationInput{
		BasePrice:     tokenCalc.PriceAmount,
		AdminFeeType:  string(pm.AdminType),
		AdminFeeValue: pm.AdminFee,
		TaxPercentage: pm.TaxFee,
	}

	// 4. Validate referral if provided
	if input.ReferralCode != nil && *input.ReferralCode != "" {
		referralValidation, err := s.referral.ValidateReferralForPayment(ctx, referral_basic_service.ValidateReferralInput{
			Code:           *input.ReferralCode,
			ProfileID:      input.ProfileID,
			BusinessRootID: input.BusinessRootID,
		})
		if err != nil {
			return response, err
		}

		response.Referral = &ReferralInfo{
			Valid:   referralValidation.Valid,
			Message: referralValidation.Message,
		}

		if referralValidation.Valid {
			calcInput.DiscountType = referralValidation.DiscountType
			calcInput.DiscountValue = referralValidation.TotalDiscount
			calcInput.MaxDiscount = referralValidation.MaxDiscount
		}
	}

	// 5. Calculate price
	calcResult := CalculatePrice(calcInput)

	// 6. Build response
	response.TokenAmount = input.TokenAmount
	response.Calculation = PriceCalculation{
		ItemPrice:         calcResult.ItemPrice,
		DiscountAmount:    calcResult.DiscountAmount,
		AfterDiscount:     calcResult.AfterDiscount,
		AdminFeeAmount:    calcResult.AdminFeeAmount,
		SubtotalBeforeTax: calcResult.SubtotalBeforeTax,
		TaxAmount:         calcResult.TaxAmount,
		TotalAmount:       calcResult.TotalAmount,
	}
	response.PaymentMethod = PaymentMethodInfo{
		Code: pm.Code,
		Name: pm.Name,
		Type: string(pm.Type),
	}

	return response, nil
}

// CreatePayment creates a new payment and charges via midtrans
func (s *ImageTokenPaymentService) CreatePayment(ctx context.Context, input CreatePaymentInput) (CreatePaymentResponse, error) {
	var response CreatePaymentResponse

	// 1. Re-validate and calculate price (same as CheckPrice)
	checkResult, err := s.CheckPrice(ctx, CheckPriceInput{
		TokenAmount:    input.TokenAmount,
		CurrencyCode:   input.CurrencyCode,
		PaymentMethod:  input.PaymentMethod,
		ReferralCode:   input.ReferralCode,
		BusinessRootID: input.BusinessRootID,
		ProfileID:      input.ProfileID,
	})
	if err != nil {
		return response, err
	}

	// 2. Check if referral is valid (if provided)
	var referralRecordID sql.NullInt64
	var referralValidation *referral_basic_service.ReferralValidationResponse
	if input.ReferralCode != nil && *input.ReferralCode != "" {
		referralValidation, err = s.referral.ValidateReferralForPayment(ctx, referral_basic_service.ValidateReferralInput{
			Code:           *input.ReferralCode,
			ProfileID:      input.ProfileID,
			BusinessRootID: input.BusinessRootID,
		})
		if err != nil {
			return response, err
		}
		if !referralValidation.Valid {
			return response, errs.NewBadRequest(referralValidation.Message)
		}
	}

	// 3. Get payment method for type detection
	pm, err := s.paymentMethod.GetPaymentMethodByCode(ctx, input.PaymentMethod, false)
	if err != nil {
		return response, err
	}

	// 4. Get token product for recording
	tokenCalc, err := s.tokenProduct.CalculateTokenProduct(ctx, token_product_service.TokenCalculateProductFilter{
		Type:         string(entity.TokenTypeImageToken),
		CurrencyCode: strings.ToUpper(input.CurrencyCode),
		From:         "token",
		Amount:       input.TokenAmount,
	})
	if err != nil {
		return response, err
	}

	// 5. Generate order ID
	orderID := fmt.Sprintf("IMG-%d-%s", time.Now().UnixMilli(), uuid.New().String()[:8])

	// 6. Execute in transaction
	var paymentHistory entity.PaymentHistory
	var midtransTransactionID string

	err = s.store.ExecTx(ctx, func(q *entity.Queries) error {
		// 6a. Create referral record if applicable
		if referralValidation != nil && referralValidation.Valid {
			referralRecord, err := q.CreateReferralRecord(ctx, entity.CreateReferralRecordParams{
				ConsumerProfileID:       input.ProfileID,
				BusinessRootID:          input.BusinessRootID,
				ProfileReferralCodeID:   referralValidation.ReferralCodeID,
				RecordType:              entity.ReferralTypeBasic,
				RecordTotalDiscount:     referralValidation.TotalDiscount,
				RecordDiscountType:      entity.DiscountType(referralValidation.DiscountType),
				RecordMaxDiscount:       referralValidation.MaxDiscount,
				RecordRewardPerReferral: referralValidation.RewardPerReferral,
				DiscountAmountGranted:   checkResult.Calculation.DiscountAmount,
				DiscountCurrency:        strings.ToUpper(input.CurrencyCode),
				RewardAmountGranted:     referralValidation.RewardPerReferral,
				RewardCurrency:          strings.ToUpper(input.CurrencyCode),
				Status:                  entity.ReferralRecordStatusPending,
			})
			if err != nil {
				return err
			}
			referralRecordID = sql.NullInt64{Int64: referralRecord.ID, Valid: true}
		}

		// 6b. Create payment history
		now := time.Now()
		expiresAt := now.Add(24 * time.Hour) // 24 hours expiry

		// Determine discount percentage
		var discountPct sql.NullInt32
		if checkResult.Referral != nil && checkResult.Referral.Valid && referralValidation != nil {
			if referralValidation.DiscountType == "percentage" {
				discountPct = sql.NullInt32{Int32: int32(referralValidation.TotalDiscount), Valid: true}
			}
		}

		// Determine admin fee percentage
		var adminFeePct sql.NullInt32
		if string(pm.AdminType) == "percentage" {
			adminFeePct = sql.NullInt32{Int32: int32(pm.AdminFee), Valid: true}
		}

		paymentHistory, err = q.CreatePaymentHistory(ctx, entity.CreatePaymentHistoryParams{
			ProfileID:             input.ProfileID,
			BusinessRootID:        input.BusinessRootID,
			ProductAmount:         input.TokenAmount,
			Status:                entity.PaymentStatusPending,
			Currency:              strings.ToUpper(input.CurrencyCode),
			PaymentMethod:         pm.Code,
			PaymentMethodType:     string(pm.Type),
			RecordProductName:     fmt.Sprintf("Image Token x%d", input.TokenAmount),
			RecordProductType:     entity.PaymentProductTypeImageToken,
			RecordProductPrice:    tokenCalc.PriceAmount,
			RecordProductImageUrl: "",
			ReferenceProductID:    tokenCalc.ID,
			SubtotalItemAmount:    checkResult.Calculation.ItemPrice,
			DiscountAmount:        checkResult.Calculation.DiscountAmount,
			DiscountPercentage:    discountPct,
			DiscountType:          entity.DiscountTypeFixed,
			AdminFeeAmount:        checkResult.Calculation.AdminFeeAmount,
			AdminFeePercentage:    adminFeePct,
			AdminFeeType:          entity.DiscountType(pm.AdminType),
			TaxAmount:             checkResult.Calculation.TaxAmount,
			TaxPercentage:         int32(pm.TaxFee),
			ReferralRecordID:      referralRecordID,
			MidtransExpiredAt:     sql.NullTime{Time: expiresAt, Valid: true},
			PaymentPendingAt:      sql.NullTime{Time: now, Valid: true},
			TotalAmount:           checkResult.Calculation.TotalAmount,
		})
		return err
	})
	if err != nil {
		return response, errs.NewInternalServerError(err)
	}

	// 7. Charge via Midtrans based on payment method type
	customerDetails := midtrans.CustomerDetails{}
	items := []midtrans.ItemDetail{
		{
			ID:       tokenCalc.ID.String(),
			Name:     fmt.Sprintf("Image Token x%d", input.TokenAmount),
			Price:    checkResult.Calculation.TotalAmount,
			Quantity: 1,
		},
	}

	// Actions to be saved to DB
	var actionsToSave []entity.CreatePaymentHistoryActionParams

	if string(pm.Type) == "ewallet" || strings.ToLower(pm.Code) == "gopay" {
		// E-wallet (Gopay)
		res, err := s.midtrans.ChargeGopay(ctx, midtrans.ChargeGopayInput{
			OrderID:         orderID,
			GrossAmount:     checkResult.Calculation.TotalAmount,
			CustomerDetails: customerDetails,
			Items:           items,
		})
		if err != nil {
			return response, err
		}
		midtransTransactionID = res.TransactionID

		// Map Midtrans actions to DB actions
		for _, action := range res.Actions {
			label, valueType, isPublic := mapGopayActionToLabel(action.Name)
			actionsToSave = append(actionsToSave, entity.CreatePaymentHistoryActionParams{
				PaymentHistoryID: paymentHistory.ID,
				Name:             action.Name,
				Label:            label,
				Value:            action.URL,
				ValueType:        entity.PaymentActionValueType(valueType),
				PaymentType:      entity.AppPaymentMethodTypeEwallet,
				ActionMethod:     action.Method,
				IsPublic:         isPublic,
			})
		}
	} else {
		// Bank transfer
		res, err := s.midtrans.ChargeBankTransfer(ctx, midtrans.ChargeBankTransferInput{
			OrderID:         orderID,
			GrossAmount:     checkResult.Calculation.TotalAmount,
			Bank:            strings.ToLower(pm.Code),
			CustomerDetails: customerDetails,
			Items:           items,
		})
		if err != nil {
			return response, err
		}
		midtransTransactionID = res.TransactionID

		// Map VA numbers to DB actions
		for _, va := range res.VANumbers {
			actionsToSave = append(actionsToSave, entity.CreatePaymentHistoryActionParams{
				PaymentHistoryID: paymentHistory.ID,
				Name:             "virtual-account",
				Label:            fmt.Sprintf("Virtual Account %s", strings.ToUpper(va.Bank)),
				Value:            va.VANumber,
				ValueType:        entity.PaymentActionValueTypeText,
				PaymentType:      entity.AppPaymentMethodTypeBank,
				ActionMethod:     "GET",
				IsPublic:         true,
			})
		}
	}

	// 8. Save actions to DB
	for _, actionParams := range actionsToSave {
		_, err := s.store.CreatePaymentHistoryAction(ctx, actionParams)
		if err != nil {
			// Log but don't fail - payment is already created
			continue
		}
	}

	// 9. Update payment history with midtrans transaction ID
	_, _ = s.store.UpdatePaymentHistoryMidtransId(ctx, entity.UpdatePaymentHistoryMidtransIdParams{
		ID:                    paymentHistory.ID,
		MidtransTransactionID: utils.StringToNullString(&midtransTransactionID),
	})

	// 10. Fetch public actions for response
	publicActions, _ := s.store.GetPublicPaymentHistoryActionsByPaymentId(ctx, paymentHistory.ID)

	// 11. Build response
	response.PaymentID = paymentHistory.ID.String()
	response.OrderID = orderID
	response.Status = string(paymentHistory.Status)
	response.PaymentMethod = checkResult.PaymentMethod
	response.Calculation = checkResult.Calculation
	response.TokenAmount = input.TokenAmount

	if paymentHistory.MidtransExpiredAt.Valid {
		response.ExpiresAt = &paymentHistory.MidtransExpiredAt.Time
	}

	// Map DB actions to response
	response.Actions = make([]PaymentActionResponse, len(publicActions))
	for i, action := range publicActions {
		response.Actions[i] = PaymentActionResponse{
			Name:      action.Name,
			Label:     action.Label,
			Value:     action.Value,
			ValueType: string(action.ValueType),
			Method:    action.ActionMethod,
		}
	}

	return response, nil
}

// mapGopayActionToLabel maps Midtrans action name to human readable label and metadata
func mapGopayActionToLabel(name string) (label string, valueType string, isPublic bool) {
	switch name {
	case "generate-qr-code":
		return "QR Code", "image", true
	case "generate-qr-code-v2":
		return "QR Code V2", "image", true
	case "deeplink-redirect":
		return "Deeplink Redirect", "link", true
	case "get-status":
		return "Get Status", "link", false
	case "cancel":
		return "Cancel Transaction", "link", false
	default:
		return name, "link", false
	}
}
