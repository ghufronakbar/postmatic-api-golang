// internal/module/headless/queue/mailer.go
package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"postmatic-api/internal/module/headless/mailer"

	"github.com/hibiken/asynq"
)

// MailerProducer adalah kontrak yang dipakai oleh service lain (mis. AuthService)
// untuk MENAMBAHKAN job mailer ke queue (enqueue), bukan untuk mengirim email langsung.
type MailerProducer interface {
	// AUTH
	EnqueueWelcomeEmail(ctx context.Context, payload mailer.WelcomeInputDTO) error
	EnqueueUserVerification(ctx context.Context, payload mailer.VerificationInputDTO) error
	// MEMBER
	EnqueueInvitation(ctx context.Context, payload mailer.MemberInvitationInputDTO) error
	EnqueueAnnounceRole(ctx context.Context, payload mailer.MemberAnnounceRoleInputDTO) error
	EnqueueAnnounceKick(ctx context.Context, payload mailer.MemberAnnounceKickInputDTO) error
	EnqueueWelcomeBusiness(ctx context.Context, payload mailer.MemberWelcomeBusinessInputDTO) error
	// PAYMENT
	EnqueuePaymentCheckout(ctx context.Context, payload mailer.PaymentCheckoutInputDTO) error
	EnqueuePaymentSuccess(ctx context.Context, payload mailer.PaymentSuccessInputDTO) error
	EnqueuePaymentCanceled(ctx context.Context, payload mailer.PaymentCanceledInputDTO) error
}

// MailerService adalah kontrak yang dipakai oleh worker (consumer) untuk MENGEKSEKUSI job.
// Di sini kita pakai interface dari package mailer agar 1 sumber kebenaran.
type MailerService = mailer.Mailer

// Task type string yang dipakai Asynq untuk routing task ke handler yang sesuai.
// Dibuat private (lowercase) agar tidak menjadi public API package queue.
const (
	// AUTH / WELCOME
	taskMailerWelcome      = "queue:mailer:welcome"
	taskMailerVerification = "queue:mailer:verification"

	// BUSINESS
	taskMailerInvitation      = "queue:mailer:invitation"
	taskMailerAnnounceRole    = "queue:mailer:announce:role"
	taskMailerAnnounceKick    = "queue:mailer:announce:kick"
	taskMailerWelcomeBusiness = "queue:mailer:welcome:business"

	// PAYMENT
	taskMailerPaymentCheckout = "queue:mailer:payment:checkout"
	taskMailerPaymentSuccess  = "queue:mailer:payment:success"
	taskMailerPaymentCanceled = "queue:mailer:payment:canceled"
)

// EnqueueWelcomeEmail adalah API producer untuk mengantrikan email welcome.
// Di sini kamu menetapkan "policy" mailer: queue name, max retry, timeout eksekusi di worker.
func (p *Producer) EnqueueWelcomeEmail(ctx context.Context, payload mailer.WelcomeInputDTO) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	task := asynq.NewTask(taskMailerWelcome, b)

	return p.enqueue(
		ctx,
		task,
		asynq.Queue("default"),
		asynq.MaxRetry(3),
		asynq.Timeout(10*time.Second),
	)
}

// EnqueueUserVerification adalah API producer untuk mengantrikan email verifikasi user.
// Policy-nya bisa sama seperti welcome, atau kamu bedakan nanti (mis. Queue("critical")).
func (p *Producer) EnqueueUserVerification(ctx context.Context, payload mailer.VerificationInputDTO) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	task := asynq.NewTask(taskMailerVerification, b)

	return p.enqueue(
		ctx,
		task,
		asynq.Queue("default"),
		asynq.MaxRetry(3),
		asynq.Timeout(10*time.Second),
	)
}

func (p *Producer) EnqueueInvitation(ctx context.Context, payload mailer.MemberInvitationInputDTO) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	task := asynq.NewTask(taskMailerInvitation, b)

	return p.enqueue(
		ctx,
		task,
		asynq.Queue("default"),
		asynq.MaxRetry(3),
		asynq.Timeout(10*time.Second),
	)
}

func (p *Producer) EnqueueAnnounceRole(ctx context.Context, payload mailer.MemberAnnounceRoleInputDTO) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	task := asynq.NewTask(taskMailerAnnounceRole, b)

	return p.enqueue(
		ctx,
		task,
		asynq.Queue("default"),
		asynq.MaxRetry(3),
		asynq.Timeout(10*time.Second),
	)
}

func (p *Producer) EnqueueAnnounceKick(ctx context.Context, payload mailer.MemberAnnounceKickInputDTO) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	task := asynq.NewTask(taskMailerAnnounceKick, b)

	return p.enqueue(
		ctx,
		task,
		asynq.Queue("default"),
		asynq.MaxRetry(3),
		asynq.Timeout(10*time.Second),
	)
}

func (p *Producer) EnqueueWelcomeBusiness(ctx context.Context, payload mailer.MemberWelcomeBusinessInputDTO) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	task := asynq.NewTask(taskMailerWelcomeBusiness, b)

	return p.enqueue(
		ctx,
		task,
		asynq.Queue("default"),
		asynq.MaxRetry(3),
		asynq.Timeout(10*time.Second),
	)
}

// registerMailerHandlers mendaftarkan consumer handler ke Asynq mux.
// Ini dipanggil dari Worker.RegisterMailer(...).
// Handler akan:
// - decode payload
// - jika payload invalid: SkipRetry (biar tidak loop retry)
// - panggil mailer service yang benar-benar mengirim email
func registerMailerHandlers(mux *asynq.ServeMux, mailerSvc MailerService) {
	mux.HandleFunc(taskMailerWelcome, func(ctx context.Context, t *asynq.Task) error {
		var p mailer.WelcomeInputDTO
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return fmt.Errorf("invalid payload: %v: %w", err, asynq.SkipRetry)
		}
		return mailerSvc.SendWelcomeEmail(ctx, p)
	})

	mux.HandleFunc(taskMailerVerification, func(ctx context.Context, t *asynq.Task) error {
		var p mailer.VerificationInputDTO
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return fmt.Errorf("invalid payload: %v: %w", err, asynq.SkipRetry)
		}
		return mailerSvc.SendVerificationEmail(ctx, p)
	})

	mux.HandleFunc(taskMailerInvitation, func(ctx context.Context, t *asynq.Task) error {
		var p mailer.MemberInvitationInputDTO
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return fmt.Errorf("invalid payload: %v: %w", err, asynq.SkipRetry)
		}
		return mailerSvc.SendInvitationEmail(ctx, p)
	})

	mux.HandleFunc(taskMailerAnnounceRole, func(ctx context.Context, t *asynq.Task) error {
		var p mailer.MemberAnnounceRoleInputDTO
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return fmt.Errorf("invalid payload: %v: %w", err, asynq.SkipRetry)
		}
		return mailerSvc.SendAnnounceRoleEmail(ctx, p)
	})

	mux.HandleFunc(taskMailerAnnounceKick, func(ctx context.Context, t *asynq.Task) error {
		var p mailer.MemberAnnounceKickInputDTO
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return fmt.Errorf("invalid payload: %v: %w", err, asynq.SkipRetry)
		}
		return mailerSvc.SendAnnounceKickEmail(ctx, p)
	})

	mux.HandleFunc(taskMailerWelcomeBusiness, func(ctx context.Context, t *asynq.Task) error {
		var p mailer.MemberWelcomeBusinessInputDTO
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return fmt.Errorf("invalid payload: %v: %w", err, asynq.SkipRetry)
		}
		return mailerSvc.SendWelcomeBusinessEmail(ctx, p)
	})

	// PAYMENT HANDLERS
	mux.HandleFunc(taskMailerPaymentCheckout, func(ctx context.Context, t *asynq.Task) error {
		var p mailer.PaymentCheckoutInputDTO
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return fmt.Errorf("invalid payload: %v: %w", err, asynq.SkipRetry)
		}
		return mailerSvc.SendPaymentCheckoutEmail(ctx, p)
	})

	mux.HandleFunc(taskMailerPaymentSuccess, func(ctx context.Context, t *asynq.Task) error {
		var p mailer.PaymentSuccessInputDTO
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return fmt.Errorf("invalid payload: %v: %w", err, asynq.SkipRetry)
		}
		return mailerSvc.SendPaymentSuccessEmail(ctx, p)
	})

	mux.HandleFunc(taskMailerPaymentCanceled, func(ctx context.Context, t *asynq.Task) error {
		var p mailer.PaymentCanceledInputDTO
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return fmt.Errorf("invalid payload: %v: %w", err, asynq.SkipRetry)
		}
		return mailerSvc.SendPaymentCanceledEmail(ctx, p)
	})
}

// ==================== PAYMENT PRODUCER ====================

func (p *Producer) EnqueuePaymentCheckout(ctx context.Context, payload mailer.PaymentCheckoutInputDTO) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	task := asynq.NewTask(taskMailerPaymentCheckout, b)

	return p.enqueue(
		ctx,
		task,
		asynq.Queue("default"),
		asynq.MaxRetry(3),
		asynq.Timeout(15*time.Second),
	)
}

func (p *Producer) EnqueuePaymentSuccess(ctx context.Context, payload mailer.PaymentSuccessInputDTO) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	task := asynq.NewTask(taskMailerPaymentSuccess, b)

	return p.enqueue(
		ctx,
		task,
		asynq.Queue("default"),
		asynq.MaxRetry(3),
		asynq.Timeout(15*time.Second),
	)
}

func (p *Producer) EnqueuePaymentCanceled(ctx context.Context, payload mailer.PaymentCanceledInputDTO) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	task := asynq.NewTask(taskMailerPaymentCanceled, b)

	return p.enqueue(
		ctx,
		task,
		asynq.Queue("default"),
		asynq.MaxRetry(3),
		asynq.Timeout(15*time.Second),
	)
}
