# Module Headless.Queue

Modul ini bertanggung jawab untuk async task processing menggunakan Redis-backed queue (Asynq). Modul ini **headless** dan terdiri dari **Producer** (enqueue) dan **Worker** (consumer).

## 1. Project Rules & Dependencies

- **Library**: [`github.com/hibiken/asynq`](https://github.com/hibiken/asynq)
- **Backend**: Redis
- **Pattern**: Producer-Consumer dengan retry mechanism
- **Headless**: Modul ini hanya dipanggil oleh module lain (internal)
- **Used By**: Auth, Member, Payment services

## 2. Directory Structure

```text
internal/module/headless/queue/
├── producer.go   # Producer struct & constructor
├── mailer.go     # Mailer task definitions (producer + handler registration)
├── enqueue.go    # Common enqueue helpers
└── worker.go     # Worker setup & registration
```

## 3. Architecture Overview

```
┌─────────────────┐     ┌───────┐     ┌─────────────────┐
│  Service Layer  │────▶│ Redis │────▶│     Worker      │
│  (Producer)     │     │ Queue │     │   (Consumer)    │
└─────────────────┘     └───────┘     └─────────────────┘
        │                                      │
        │   EnqueueWelcomeEmail(payload)       │   HandleWelcomeEmail(payload)
        │                                      │   → mailerSvc.SendWelcomeEmail()
        ▼                                      ▼
   Return immediately                    Execute async
```

## 4. Task Types

### Mailer Tasks

| Task Name                        | Description                   |
| -------------------------------- | ----------------------------- |
| `queue:mailer:auth:welcome`      | Welcome email setelah signup  |
| `queue:mailer:auth:verification` | Email verification            |
| `queue:mailer:member:invitation` | Member invitation             |
| `queue:mailer:member:role`       | Role change announcement      |
| `queue:mailer:member:kick`       | Removed from business         |
| `queue:mailer:member:welcome`    | Welcome to business           |
| `queue:mailer:payment:checkout`  | Payment checkout notification |
| `queue:mailer:payment:success`   | Payment success notification  |
| `queue:mailer:payment:canceled`  | Payment canceled notification |

## 5. Producer Interface (MailerProducer)

```go
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
```

## 6. Producer Business Logic

Setiap producer method melakukan:

```
1. Serialize payload ke JSON

2. Create asynq.Task dengan:
   - TypeName: task identifier (e.g., "queue:mailer:auth:welcome")
   - Payload: JSON bytes

3. Enqueue dengan options:
   - MaxRetry: 3 (default)
   - Timeout: 30s (default)
   - Queue: "default" (bisa override)

4. Return immediately (non-blocking)
```

### Example Implementation

```go
func (p *Producer) EnqueueWelcomeEmail(ctx context.Context, payload mailer.WelcomeInputDTO) error {
    data, err := json.Marshal(payload)
    if err != nil {
        return err
    }

    task := asynq.NewTask(taskMailerAuthWelcome, data)
    _, err = p.client.Enqueue(task,
        asynq.MaxRetry(3),
        asynq.Timeout(30*time.Second),
    )
    return err
}
```

## 7. Worker (Consumer) Setup

Worker adalah process terpisah yang consume tasks dari queue.

```go
// Di cmd/worker/main.go
worker := queue.NewWorker(redisClient)

// Register mailer handlers
worker.RegisterMailer(mailerService)

// Start processing
worker.Start()
```

### Handler Registration

```go
func registerMailerHandlers(mux *asynq.ServeMux, mailerSvc MailerService) {
    mux.HandleFunc(taskMailerAuthWelcome, func(ctx context.Context, t *asynq.Task) error {
        var payload mailer.WelcomeInputDTO
        if err := json.Unmarshal(t.Payload(), &payload); err != nil {
            return asynq.SkipRetry // Don't retry invalid payload
        }
        return mailerSvc.SendWelcomeEmail(ctx, payload)
    })

    // ... register other handlers
}
```

## 8. Usage Example

```go
// Di router.go - setup Producer
asynqClient := asynq.NewClient(asynq.RedisClientOpt{...})
queueProducer := queue.NewProducer(asynqClient)

// Di AuthService - enqueue task
err := queueProducer.EnqueueWelcomeEmail(ctx, mailer.WelcomeInputDTO{
    Email: user.Email,
    Name:  user.Name,
})
// Returns immediately, email sent async by worker
```

## 9. Retry & Error Handling

| Scenario              | Behavior                  |
| --------------------- | ------------------------- |
| Task success          | Removed from queue        |
| Task failed           | Retry up to MaxRetry      |
| Invalid payload       | SkipRetry (no retry)      |
| All retries exhausted | Move to dead-letter queue |

## 10. Monitoring

Asynq menyediakan web UI untuk monitoring:

```bash
asynqmon --redis-addr=localhost:6379
```

Metrics yang tersedia:

- Active tasks
- Pending tasks
- Retry queue
- Dead-letter queue
- Processing rate

## 11. Design Decisions

### Kenapa Queue?

1. **Non-blocking**: API response cepat, email dikirim async
2. **Reliability**: Retry mechanism untuk transient failures
3. **Scalability**: Worker bisa di-scale horizontal
4. **Resilience**: Task persist di Redis, tidak hilang jika server restart

### Kenapa Asynq?

1. **Simple**: API sederhana dan Go-native
2. **Redis-backed**: Tidak perlu message broker tambahan
3. **Feature-rich**: Retry, timeout, priority queues, cron jobs
4. **Web UI**: Built-in monitoring dashboard
