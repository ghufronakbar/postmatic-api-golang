package mailer

import (
	"context"
	"fmt"
	"net/url"
	"postmatic-api/pkg/errs"
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

func (s *MailerService) SendInvitationEmail(ctx context.Context, input InvitationInputDTO) error {
	fmt.Println("SendInvitationEmail", input)
	err := s.sendEmail(ctx, SendEmailInput{
		To:           input.Email,
		Subject:      "Undangan Bergabung",
		TemplateName: InvitationTemplate,
		Data:         input,
	})
	if err != nil {
		return errs.NewInternalServerError(err)
	}
	return nil
}
