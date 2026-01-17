package midtrans

import (
	"context"

	"postmatic-api/pkg/errs"
	"postmatic-api/pkg/logger"

	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/coreapi"
)

// ChargeGopay charges a Gopay e-wallet payment
func (s *midtransService) ChargeGopay(ctx context.Context, req ChargeGopayInput) (*ChargeResponse, error) {
	log := logger.From(ctx)
	log.Info("Charging Gopay",
		"orderID", req.OrderID,
		"amount", req.GrossAmount,
	)

	// Build SDK request
	chargeReq := &coreapi.ChargeReq{
		PaymentType: coreapi.PaymentTypeGopay,
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  req.OrderID,
			GrossAmt: req.GrossAmount,
		},
	}

	// Add Gopay specific options
	if req.CallbackURL != "" {
		chargeReq.Gopay = &coreapi.GopayDetails{
			EnableCallback: true,
			CallbackUrl:    req.CallbackURL,
		}
	}

	// Add customer details if provided
	if req.CustomerDetails.Email != "" || req.CustomerDetails.FirstName != "" {
		chargeReq.CustomerDetails = &midtrans.CustomerDetails{
			FName: req.CustomerDetails.FirstName,
			LName: req.CustomerDetails.LastName,
			Email: req.CustomerDetails.Email,
			Phone: req.CustomerDetails.Phone,
		}
	}

	// Add items if provided
	if len(req.Items) > 0 {
		items := make([]midtrans.ItemDetails, len(req.Items))
		for i, item := range req.Items {
			items[i] = midtrans.ItemDetails{
				ID:    item.ID,
				Name:  item.Name,
				Price: item.Price,
				Qty:   item.Quantity,
			}
		}
		chargeReq.Items = &items
	}

	res, err := s.client.ChargeTransaction(chargeReq)
	if err != nil {
		log.Error("Failed to charge Gopay",
			"orderID", req.OrderID,
			"error", err,
		)
		return nil, errs.NewBadRequest("MIDTRANS_CHARGE_GOPAY_FAILED")
	}

	log.Info("Gopay charge successful",
		"orderID", res.OrderID,
		"transactionID", res.TransactionID,
		"status", res.TransactionStatus,
	)

	return mapChargeResponse(res), nil
}
