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
