# Module Headless.Mailer

Modul ini bertanggung jawab untuk mengirim email transaksional menggunakan SMTP dengan HTML templates. Modul ini **headless** dan biasanya dipanggil via Queue (async).

## 1. Project Rules & Dependencies

- **Library**: [`gopkg.in/gomail.v2`](https://github.com/go-gomail/gomail)
- **Templates**: Go `html/template` dengan `embed.FS` (compile-time embedding)
- **Headless**: Dipanggil oleh Queue worker, tidak langsung dari HTTP
- **Used By**: Queue workers (consumer)

## 2. Directory Structure

```text
internal/module/headless/mailer/
├── service.go       # MailerService & sendEmail implementation
├── constants.go     # Template name constants
├── dto.go           # Common DTOs
├── dto_auth.go      # Auth-related DTOs
├── dto_member.go    # Member-related DTOs
├── dto_payment.go   # Payment-related DTOs
├── mail.go          # Email method implementations
└── templates/       # HTML email templates
    ├── layout.html
    ├── welcome.html
    ├── verification.html
    ├── invitation.html
    └── ...
```

## 3. Configuration

| Variable            | Type   | Description             |
| ------------------- | ------ | ----------------------- |
| `SMTP_HOST`         | String | SMTP server hostname    |
| `SMTP_PORT`         | Int    | SMTP port (usually 587) |
| `SMTP_USER`         | String | SMTP username           |
| `SMTP_PASS`         | String | SMTP password           |
| `SMTP_SENDER`       | String | From address            |
| `APP_NAME`          | String | App name for templates  |
| `APP_LOGO`          | String | Logo URL for templates  |
| `APP_ADDRESS`       | String | Company address         |
| `APP_CONTACT_EMAIL` | String | Contact email           |

## 4. Service Interface

```go
type Mailer interface {
    // AUTH / WELCOME
    SendWelcomeEmail(ctx context.Context, input WelcomeInputDTO) error
    SendVerificationEmail(ctx context.Context, input VerificationInputDTO) error

    // MEMBER
    SendInvitationEmail(ctx context.Context, input MemberInvitationInputDTO) error
    SendAnnounceRoleEmail(ctx context.Context, input MemberAnnounceRoleInputDTO) error
    SendAnnounceKickEmail(ctx context.Context, input MemberAnnounceKickInputDTO) error
    SendWelcomeBusinessEmail(ctx context.Context, input MemberWelcomeBusinessInputDTO) error

    // PAYMENT
    SendPaymentCheckoutEmail(ctx context.Context, input PaymentCheckoutInputDTO) error
    SendPaymentSuccessEmail(ctx context.Context, input PaymentSuccessInputDTO) error
    SendPaymentCanceledEmail(ctx context.Context, input PaymentCanceledInputDTO) error
}
```

## 5. Email Types & Templates

| Email Type       | Template                | Purpose                   |
| ---------------- | ----------------------- | ------------------------- |
| Welcome          | `welcome.html`          | New user registration     |
| Verification     | `verification.html`     | Email verification link   |
| Invitation       | `invitation.html`       | Invite member to business |
| Announce Role    | `announce_role.html`    | Role change notification  |
| Announce Kick    | `announce_kick.html`    | Removed from business     |
| Welcome Business | `welcome_business.html` | Accepted invitation       |
| Payment Checkout | `payment_checkout.html` | Payment initiated         |
| Payment Success  | `payment_success.html`  | Payment completed         |
| Payment Canceled | `payment_canceled.html` | Payment was canceled      |

## 6. Template System

### Layout Structure

```html
<!-- templates/layout.html -->
{{define "layout"}}
<!DOCTYPE html>
<html>
  <head>
    <title>{{.Subject}}</title>
  </head>
  <body>
    <header>
      <img src="{{.Logo}}" alt="{{.AppName}}" />
    </header>

    {{template "content" .}}

    <footer>
      <p>{{.Address}}</p>
      <p>{{.ContactEmail}}</p>
    </footer>
  </body>
</html>
{{end}}
```

### Content Template

```html
<!-- templates/welcome.html -->
{{define "content"}}
<h1>Welcome, {{.Name}}!</h1>
<p>Thanks for signing up for {{.AppName}}.</p>
{{end}}
```

### Template Data Flow

```
1. Service method receives DTO (e.g., WelcomeInputDTO)

2. sendEmail() builds template data:
   - Global: AppName, Logo, Address, ContactEmail, Subject
   - Payload: DTO fields (Name, Email, etc.)

3. Template execution:
   - Parse layout.html + content.html
   - Execute with merged data
   - Return HTML string

4. Send via SMTP:
   - Set From, To, Subject
   - Set Content-Type: text/html
   - Send via gomail.Dialer
```

## 7. DTO Examples

### Auth DTOs

```go
type WelcomeInputDTO struct {
    Email string `json:"email"`
    Name  string `json:"name"`
}

type VerificationInputDTO struct {
    Email           string `json:"email"`
    Name            string `json:"name"`
    VerificationURL string `json:"verificationUrl"`
}
```

### Member DTOs

```go
type MemberInvitationInputDTO struct {
    Email          string `json:"email"`
    Name           string `json:"name"`
    BusinessName   string `json:"businessName"`
    InvitationURL  string `json:"invitationUrl"`
    ExpiresAt      string `json:"expiresAt"`
}
```

### Payment DTOs

```go
type PaymentSuccessInputDTO struct {
    Email         string `json:"email"`
    Name          string `json:"name"`
    OrderID       string `json:"orderId"`
    Amount        string `json:"amount"`
    TokenAmount   int64  `json:"tokenAmount"`
    PaymentMethod string `json:"paymentMethod"`
    PaidAt        string `json:"paidAt"`
}
```

## 8. sendEmail Business Logic

```
1. Create gomail.Message

2. Set headers:
   - From: cfg.SMTP_SENDER
   - To: input.To
   - Subject: input.Subject

3. Process body:
   ├── If TemplateName empty → plain text body
   └── If TemplateName provided:
       a. Validate template name
       b. Parse layout + content templates
       c. Build template data (global + payload)
       d. Execute template to buffer
       e. Set body as HTML

4. Dial and send via SMTP

5. Return error if any step fails
```

## 9. Usage Example

```go
// Di cmd/worker/main.go
mailerSvc := mailer.NewService(cfg)

// Di queue handler
func handleWelcomeEmail(ctx context.Context, t *asynq.Task) error {
    var payload mailer.WelcomeInputDTO
    json.Unmarshal(t.Payload(), &payload)
    return mailerSvc.SendWelcomeEmail(ctx, payload)
}
```

## 10. Error Handling

| Error                   | Condition                      |
| ----------------------- | ------------------------------ |
| `INVALID_TEMPLATE_NAME` | Template name not in constants |
| Template parse error    | HTML syntax error in template  |
| SMTP dial error         | Cannot connect to SMTP server  |
| Send error              | Email rejected by server       |

## 11. Design Decisions

### Kenapa Embed Templates?

1. **No file dependencies**: Templates compiled into binary
2. **Deploy simple**: Satu binary, no external files
3. **Performance**: No disk I/O saat render

### Kenapa Layout Pattern?

1. **DRY**: Header/footer consistent across emails
2. **Maintainable**: Ubah layout sekali, apply ke semua
3. **Branding**: Logo, colors, footer selalu sama

### Kenapa via Queue?

1. **Non-blocking**: API response tidak tunggu email sent
2. **Retry**: SMTP failure di-retry otomatis
3. **Resilience**: Email queue persist di Redis
