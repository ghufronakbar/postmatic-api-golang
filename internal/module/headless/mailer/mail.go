package mailer

import (
	"context"
	"fmt"
	"net/url"
	"postmatic-api/pkg/errs"
	"postmatic-api/pkg/logger"
)

func (s *MailerService) SendWelcomeEmail(ctx context.Context, input WelcomeInputDTO) error {
	link := s.cfg.AUTH_URL + "&from=" + url.QueryEscape(input.From)

	emailInput := SendEmailInput{
		To:           input.Email,
		Subject:      "Selamat Datang di Postmatic!",
		TemplateName: WelcomeTemplate,
		Data: welcomeInput{
			Name:  input.Name,
			Email: input.Email,
			Link:  link,
		},
	}

	if err := s.sendEmail(ctx, emailInput); err != nil {
		println("Failed to send email:", err.Error())
	}

	return nil
}

func (s *MailerService) SendVerificationEmail(ctx context.Context, input VerificationInputDTO) error {
	u, err := url.Parse(s.cfg.AUTH_URL + s.cfg.VERIFY_EMAIL_ROUTE + "/" + input.Token)
	if err != nil {
		return errs.NewInternalServerError(err)
	}

	q := u.Query()
	q.Set("from", input.From)
	u.RawQuery = q.Encode()

	confirmUrl := u.String()
	fmt.Println(confirmUrl)
	templateData := verificationInput{
		Name:       input.Name,
		ConfirmUrl: confirmUrl,
	}

	err = s.sendEmail(ctx, SendEmailInput{
		To:           input.To,
		Subject:      "Konfirmasi Pendaftaran Akun",
		TemplateName: VerificationTemplate,
		Data:         templateData,
	})
	if err != nil {
		return errs.NewInternalServerError(err)
	}
	return nil
}

func (s *MailerService) SendInvitationEmail(ctx context.Context, input MemberInvitationInputDTO) error {
	logger.From(ctx).Info("SendInvitationEmail", "input", input)
	err := s.sendEmail(ctx, SendEmailInput{
		To:           input.Email,
		Subject:      "Undangan Bergabung",
		TemplateName: MemberInvitationTemplate,
		Data:         input,
	})
	if err != nil {
		return errs.NewInternalServerError(err)
	}
	return nil
}

func (s *MailerService) SendAnnounceRoleEmail(ctx context.Context, input MemberAnnounceRoleInputDTO) error {
	logger.From(ctx).Info("SendAnnounceRoleEmail", "input", input)
	err := s.sendEmail(ctx, SendEmailInput{
		To:           input.Email,
		Subject:      "Perubahan Role",
		TemplateName: MemberAnnounceRoleTemplate,
		Data:         input,
	})
	if err != nil {
		return errs.NewInternalServerError(err)
	}
	return nil
}

func (s *MailerService) SendAnnounceKickEmail(ctx context.Context, input MemberAnnounceKickInputDTO) error {
	logger.From(ctx).Info("SendAnnounceKickEmail", "input", input)
	err := s.sendEmail(ctx, SendEmailInput{
		To:           input.Email,
		Subject:      "Akses Dicabut",
		TemplateName: MemberAnnounceKickTemplate,
		Data:         input,
	})
	if err != nil {
		return errs.NewInternalServerError(err)
	}
	return nil
}

func (s *MailerService) SendWelcomeBusinessEmail(ctx context.Context, input MemberWelcomeBusinessInputDTO) error {
	logger.From(ctx).Info("SendWelcomeBusinessEmail", "input", input)

	err := s.sendEmail(ctx, SendEmailInput{
		To:           input.Email,
		Subject:      "Selamat Datang di " + input.BusinessName,
		TemplateName: MemberWelcomeBusinessTemplate,
		Data:         input,
	})
	if err != nil {
		return errs.NewInternalServerError(err)
	}
	return nil
}

// ==================== PAYMENT ====================

func (s *MailerService) SendPaymentCheckoutEmail(ctx context.Context, input PaymentCheckoutInputDTO) error {
	logger.From(ctx).Info("SendPaymentCheckoutEmail", "orderID", input.OrderID, "email", input.Email)

	templateData := paymentCheckoutInput{
		Name:          input.Name,
		OrderID:       input.OrderID,
		ProductName:   input.ProductName,
		PaymentMethod: input.PaymentMethod,
		TotalAmount:   input.TotalAmount,
		ExpiresAt:     input.ExpiresAt,
		Actions:       input.Actions,
	}

	err := s.sendEmail(ctx, SendEmailInput{
		To:           input.Email,
		Subject:      "Konfirmasi Pesanan #" + input.OrderID,
		TemplateName: PaymentCheckoutTemplate,
		Data:         templateData,
	})
	if err != nil {
		logger.From(ctx).Error("Failed to send checkout email", "orderID", input.OrderID, "error", err)
		return errs.NewInternalServerError(err)
	}
	return nil
}

func (s *MailerService) SendPaymentSuccessEmail(ctx context.Context, input PaymentSuccessInputDTO) error {
	logger.From(ctx).Info("SendPaymentSuccessEmail", "orderID", input.OrderID, "email", input.Email)

	// Format amounts with currency
	formatCurrency := func(amount int64, currency string) string {
		if currency == "IDR" {
			return "Rp " + formatNumber(amount)
		}
		return currency + " " + formatNumber(amount)
	}

	templateData := paymentSuccessInput{
		Name:           input.Name,
		OrderID:        input.OrderID,
		ProductName:    input.ProductName,
		ItemPrice:      formatCurrency(input.ItemPrice, input.Currency),
		DiscountAmount: formatCurrency(input.DiscountAmount, input.Currency),
		AdminFeeAmount: formatCurrency(input.AdminFeeAmount, input.Currency),
		TaxAmount:      formatCurrency(input.TaxAmount, input.Currency),
		TotalAmount:    formatCurrency(input.TotalAmount, input.Currency),
		PaymentMethod:  input.PaymentMethod,
		PaidAt:         input.PaidAt.Format("02 Jan 2006, 15:04 WIB"),
	}

	err := s.sendEmail(ctx, SendEmailInput{
		To:           input.Email,
		Subject:      "Invoice Pembayaran #" + input.OrderID,
		TemplateName: PaymentSuccessTemplate,
		Data:         templateData,
	})
	if err != nil {
		logger.From(ctx).Error("Failed to send success email", "orderID", input.OrderID, "error", err)
		return errs.NewInternalServerError(err)
	}
	return nil
}

func (s *MailerService) SendPaymentCanceledEmail(ctx context.Context, input PaymentCanceledInputDTO) error {
	logger.From(ctx).Info("SendPaymentCanceledEmail", "orderID", input.OrderID, "email", input.Email)

	// Format amount with currency
	formatCurrency := func(amount int64, currency string) string {
		if currency == "IDR" {
			return "Rp " + formatNumber(amount)
		}
		return currency + " " + formatNumber(amount)
	}

	templateData := paymentCanceledInput{
		Name:          input.Name,
		OrderID:       input.OrderID,
		ProductName:   input.ProductName,
		TotalAmount:   formatCurrency(input.TotalAmount, input.Currency),
		PaymentMethod: input.PaymentMethod,
		CanceledAt:    input.CanceledAt.Format("02 Jan 2006, 15:04 WIB"),
	}

	err := s.sendEmail(ctx, SendEmailInput{
		To:           input.Email,
		Subject:      "Pembayaran Dibatalkan #" + input.OrderID,
		TemplateName: PaymentCanceledTemplate,
		Data:         templateData,
	})
	if err != nil {
		logger.From(ctx).Error("Failed to send canceled email", "orderID", input.OrderID, "error", err)
		return errs.NewInternalServerError(err)
	}
	return nil
}

// Helper function to format number with thousand separator
func formatNumber(n int64) string {
	if n == 0 {
		return "0"
	}

	negative := n < 0
	if negative {
		n = -n
	}

	// Convert to string first
	numStr := ""
	for n > 0 {
		numStr = string('0'+byte(n%10)) + numStr
		n /= 10
	}

	// Add thousand separators from right to left
	result := ""
	for i, c := range numStr {
		if i > 0 && (len(numStr)-i)%3 == 0 {
			result += "."
		}
		result += string(c)
	}

	if negative {
		return "-" + result
	}
	return result
}
