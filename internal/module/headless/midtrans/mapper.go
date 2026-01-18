package midtrans

import "github.com/midtrans/midtrans-go/coreapi"

// mapChargeResponse maps SDK ChargeResponse to our DTO
func mapChargeResponse(res *coreapi.ChargeResponse) *ChargeResponse {
	if res == nil {
		return nil
	}

	resp := &ChargeResponse{
		TransactionID:     res.TransactionID,
		OrderID:           res.OrderID,
		GrossAmount:       res.GrossAmount,
		PaymentType:       res.PaymentType,
		TransactionTime:   res.TransactionTime,
		TransactionStatus: res.TransactionStatus,
		FraudStatus:       res.FraudStatus,
		StatusCode:        res.StatusCode,
		StatusMessage:     res.StatusMessage,
	}

	// Map Actions (for e-wallet)
	if len(res.Actions) > 0 {
		resp.Actions = make([]PaymentAction, len(res.Actions))
		for i, action := range res.Actions {
			resp.Actions[i] = PaymentAction{
				Name:   action.Name,
				Method: action.Method,
				URL:    action.URL,
			}
		}
	}

	// Map VA Numbers (for bank transfer)
	// Handle different response formats from different banks
	if len(res.VaNumbers) > 0 {
		// Standard bank format (BNI, BCA, BRI)
		resp.VANumbers = make([]VANumber, len(res.VaNumbers))
		for i, va := range res.VaNumbers {
			resp.VANumbers[i] = VANumber{
				Bank:     va.Bank,
				VANumber: va.VANumber,
			}
		}
	} else if res.PermataVaNumber != "" {
		// Permata uses separate field: permata_va_number
		resp.VANumbers = []VANumber{
			{
				Bank:     "permata",
				VANumber: res.PermataVaNumber,
			},
		}
		resp.PermataVANumber = res.PermataVaNumber
	} else if res.BillKey != "" && res.BillerCode != "" {
		// Mandiri uses echannel: bill_key + biller_code
		// Combine as "biller_code:bill_key" for display
		resp.VANumbers = []VANumber{
			{
				Bank:     "mandiri",
				VANumber: res.BillerCode + res.BillKey, // Combined format
			},
		}
		resp.BillKey = res.BillKey
		resp.BillerCode = res.BillerCode
	}

	return resp
}

// mapTransactionStatusResponse maps SDK TransactionStatusResponse to our DTO
func mapTransactionStatusResponse(res *coreapi.TransactionStatusResponse) *TransactionStatusResponse {
	if res == nil {
		return nil
	}

	return &TransactionStatusResponse{
		TransactionID:     res.TransactionID,
		OrderID:           res.OrderID,
		GrossAmount:       res.GrossAmount,
		PaymentType:       res.PaymentType,
		TransactionTime:   res.TransactionTime,
		TransactionStatus: res.TransactionStatus,
		FraudStatus:       res.FraudStatus,
		StatusCode:        res.StatusCode,
		StatusMessage:     res.StatusMessage,
	}
}

// mapCancelResponse maps SDK CancelResponse to our DTO
func mapCancelResponse(res *coreapi.CancelResponse) *TransactionStatusResponse {
	if res == nil {
		return nil
	}

	return &TransactionStatusResponse{
		TransactionID:     res.TransactionID,
		OrderID:           res.OrderID,
		GrossAmount:       res.GrossAmount,
		PaymentType:       res.PaymentType,
		TransactionTime:   res.TransactionTime,
		TransactionStatus: res.TransactionStatus,
		FraudStatus:       res.FraudStatus,
		StatusCode:        res.StatusCode,
		StatusMessage:     res.StatusMessage,
	}
}

// mapExpireResponse maps SDK ExpireResponse to our DTO
func mapExpireResponse(res *coreapi.ExpireResponse) *TransactionStatusResponse {
	if res == nil {
		return nil
	}

	return &TransactionStatusResponse{
		TransactionID:     res.TransactionID,
		OrderID:           res.OrderID,
		GrossAmount:       res.GrossAmount,
		PaymentType:       res.PaymentType,
		TransactionTime:   res.TransactionTime,
		TransactionStatus: res.TransactionStatus,
		FraudStatus:       res.FraudStatus,
		StatusCode:        res.StatusCode,
		StatusMessage:     res.StatusMessage,
	}
}
