# Code Review (Golang)
**Tanggal:** 30 Desember 2025  
**Repo:** postmatic-api

---

## Executive Summary

### Skor Kesiapan Produksi: 5/10

### 3 Kekuatan Utama:
1. **Arsitektur layered yang konsisten** — Pemisahan jelas antara handler → service → repository dengan dependency injection manual yang bersih
2. **Validasi input yang komprehensif** — Package `pkg/utils/validator.go` sangat matang dengan deteksi unknown fields, type mismatch, dan error aggregation
3. **Pattern transaksi database** — Implementasi `ExecTx` di `store.go` dengan proper rollback handling dan SQLC untuk type-safe queries

### 3 Risiko Utama:
1. **Tidak ada unit test sama sekali** — 0% coverage, tidak ada file `*_test.go` di seluruh repository
2. **Email dikirim di dalam database transaction** — Bisa menyebabkan transaction lock berkepanjangan dan data inconsistency jika email gagal
3. **Tidak ada observability stack** — Logging menggunakan `fmt.Println`, tidak ada structured logging, metrics, atau tracing

---

## Arsitektur & Struktur Proyek

### Struktur Saat Ini
```
postmatic-api/
├── cmd/api/main.go          # Entry point
├── config/                   # Configuration loading
├── internal/
│   ├── http/
│   │   ├── handler/          # HTTP handlers per domain
│   │   ├── middleware/       # Auth, owned business, filter
│   │   └── router.go         # Route registration & DI
│   ├── module/               # Business logic services
│   │   ├── account/          # Auth, session, profile
│   │   ├── app/              # RSS, timezone, image upload
│   │   ├── business/         # Business domain modules
│   │   ├── creator/          # Creator features
│   │   └── headless/         # External services (mailer, S3, cloudinary)
│   └── repository/
│       ├── entity/           # SQLC-generated code + Store
│       ├── queries/          # SQL query files
│       └── redis/            # Redis repositories
├── pkg/                      # Shared packages
│   ├── errs/                 # Custom error types
│   ├── filter/               # Query filter structs
│   ├── pagination/           # Pagination logic
│   ├── response/             # HTTP response helpers
│   ├── token/                # JWT handling
│   └── utils/                # Validators, password, etc
├── migrations/               # Goose migration files
└── sqlc.yaml                 # SQLC configuration
```

### Penilaian Objektif

**✅ Kelebihan:**
- Mengikuti konvensi Go standar (`cmd/`, `internal/`, `pkg/`)
- Pemisahan `internal/` untuk kode private dan `pkg/` untuk reusable
- Struktur modular per domain (account, business, app, creator)

**⚠️ Catatan:**
- Tidak ada folder `docs/` untuk dokumentasi API
- Tidak ada `api/` folder untuk OpenAPI specs
- File `test.js` di root tidak relevan dengan Go project

---

## Temuan Utama (Prioritized)

### Critical
- [ ] **T1: Tidak ada test coverage sama sekali**
- [ ] **T2: Email sending di dalam database transaction**
- [ ] **T3: Tidak ada database connection pooling configuration**

### High
- [ ] **T4: Logging menggunakan fmt.Println — tidak production-ready**
- [ ] **T5: Config loaded berulang kali di pkg/token — performance issue**
- [ ] **T6: Tidak ada linter configuration (golangci-lint)**
- [ ] **T7: Tidak ada CI/CD pipeline configuration**

### Medium
- [ ] **T8: Error handling tidak konsisten — beberapa tempat ignore error**
- [ ] **T9: Goroutine leak potential — email sent in background tanpa proper tracking**
- [ ] **T10: Context timeout tidak dikonfigurasi untuk external calls**
- [ ] **T11: Panic digunakan untuk missing config — tidak graceful**

### Low
- [ ] **T12: Makefile tidak ada — build automation manual**
- [ ] **T13: Go version 1.25.4 invalid — tidak ada Go 1.25**
- [ ] **T14: Beberapa debug fmt.Println tersisa (owned_business.go)**

### Nit / Style
- [ ] **T15: Komentar mixed Bahasa Indonesia dan English**
- [ ] **T16: Beberapa file tanpa package comment**

---

## Detail Temuan

### T1: Tidak Ada Test Coverage [CRITICAL]

**Bukti:**
```bash
# Pencarian file test
find . -name "*_test.go"
# Hasil: 0 files
```

**Dampak:**
- Tidak ada jaring pengaman untuk refactoring
- Bug bisa lolos ke production tanpa terdeteksi
- Tidak memenuhi standar production-grade

**Rekomendasi:**
1. Mulai dengan table-driven test untuk business logic kritis (auth, transaction)
2. Tambahkan integration test untuk repository layer
3. Target minimal 60% coverage untuk service layer

**Contoh Test (seharusnya ada):**
```go
// internal/module/account/auth/service_test.go
func TestRegister_EmailAlreadyExists(t *testing.T) {
    tests := []struct {
        name    string
        input   RegisterInput
        wantErr string
    }{
        {
            name:    "duplicate email",
            input:   RegisterInput{Email: "existing@test.com"},
            wantErr: "EMAIL_ALREADY_EXISTS",
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // ... test implementation
        })
    }
}
```

---

### T2: Email Sending di Dalam Database Transaction [CRITICAL]

**Bukti:**
`internal/module/account/auth/service.go` (baris 53-117):
```go
err := s.store.ExecTx(ctx, func(q *entity.Queries) error {
    // ... database operations ...
    
    // TODO: add to queue instead synchronous (and place outside db transaction)
    e = s.mailer.SendVerificationEmail(ctx, mailer.VerificationInputDTO{
        // ...
    })
    if e != nil {
        return e // Otomatis Rollback 
    }
    // ...
})
```

**Dampak:**
- Transaction di-hold selama email dikirim (bisa 5-30 detik)
- Jika email server lambat, database connection pool habis
- Rollback transaksi jika email gagal padahal data sudah valid

**Rekomendasi:**
1. Pindahkan email sending ke luar transaction
2. Implementasi message queue (Redis Queue, RabbitMQ)
3. Pattern: commit dulu, kirim email async, handle failure dengan retry

```go
// Seharusnya:
func (s *AuthService) Register(ctx context.Context, input RegisterInput) (...) {
    var profile entity.Profile
    
    // 1. Database transaction ONLY for DB operations
    err := s.store.ExecTx(ctx, func(q *entity.Queries) error {
        // ... create profile & user ...
        profile = newProfile
        return nil
    })
    if err != nil {
        return ..., err
    }
    
    // 2. Email sending OUTSIDE transaction
    token, _ := token.GenerateCreateAccountToken(...)
    
    // Best: use queue
    // s.emailQueue.Enqueue(EmailJob{...})
    
    // Acceptable: goroutine with error handling
    go func() {
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()
        if err := s.mailer.SendVerificationEmail(ctx, ...); err != nil {
            log.Error("failed to send verification email", "error", err)
        }
    }()
    
    return ..., nil
}
```

---

### T3: Tidak Ada Database Connection Pooling [CRITICAL] [NEED_CHECK]

**Bukti:**
`config/database.go`:
```go
func ConnectDB(url string) (*sql.DB, error) {
    db, err := sql.Open("postgres", url)
    if err != nil {
        return nil, err
    }
    if err := db.Ping(); err != nil {
        return nil, err
    }
    return db, nil // No pool configuration!
}
```

**Dampak:**
- Default pool settings bisa menyebabkan connection exhaustion
- Tidak ada idle connection management
- Performance degradation under load

**Rekomendasi:**
```go
func ConnectDB(url string) (*sql.DB, error) {
    db, err := sql.Open("postgres", url)
    if err != nil {
        return nil, err
    }
    
    // Production-ready pool settings
    db.SetMaxOpenConns(25)                 // Sesuaikan dengan kebutuhan
    db.SetMaxIdleConns(10)
    db.SetConnMaxLifetime(5 * time.Minute)
    db.SetConnMaxIdleTime(1 * time.Minute)
    
    if err := db.Ping(); err != nil {
        return nil, err
    }
    return db, nil
}
```

---

### T4: Logging Menggunakan fmt.Println [HIGH] [NEED_CHECK]

**Bukti:**
- `pkg/errs/app_error.go` (baris 57): `fmt.Println(err.Error())`
- `pkg/response/response.go` (baris 90): `fmt.Println(err)`
- `internal/module/account/auth/service.go` (baris 133, 637): `fmt.Printf("Failed to set email limiter: %v\n", err)`
- `internal/http/middleware/owned_business.go` (baris 90, 97, 133, 138): Debug `fmt.Println`

**Dampak:**
- Tidak ada log levels (debug, info, warn, error)
- Tidak ada structured logging untuk aggregation
- Tidak ada request correlation
- Sensitive data bisa ter-log

**Rekomendasi:**
Implementasi structured logger seperti `slog` (stdlib Go 1.21+) atau `zerolog`:

```go
// pkg/logger/logger.go
package logger

import (
    "log/slog"
    "os"
)

var Log *slog.Logger

func Init(mode string) {
    opts := &slog.HandlerOptions{
        Level: slog.LevelInfo,
    }
    if mode == "development" {
        opts.Level = slog.LevelDebug
    }
    
    handler := slog.NewJSONHandler(os.Stdout, opts)
    Log = slog.New(handler)
}

// Usage:
// logger.Log.Error("failed to send email", "error", err, "email", email)
```

---

### T5: Config Loaded Berulang Kali [HIGH]

**Bukti:**
`pkg/token/token.go` — setiap function memanggil `config.Load()`:
```go
func GenerateAccessToken(ID, email, name string, imageUrl *string) (string, error) {
    expirationTime := time.Now().Add(config.Load().JWT_ACCESS_TOKEN_EXPIRED) // Load setiap call!
    // ...
    return token.SignedString([]byte(config.Load().JWT_ACCESS_TOKEN_SECRET)) // Load lagi!
}

func ValidateAccessToken(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        return []byte(config.Load().JWT_ACCESS_TOKEN_SECRET), nil // Load lagi!
    })
    // ...
}
```

**Dampak:**
- Setiap request authentication memanggil `config.Load()` 2-4 kali
- `config.Load()` membaca env vars dan melakukan parsing setiap kali
- Unnecessary CPU cycles dan potential race condition

**Rekomendasi:**
Inject config atau gunakan singleton dengan init:

```go
// Option 1: Inject config ke token package
type TokenMaker struct {
    accessSecret  []byte
    refreshSecret []byte
    accessTTL     time.Duration
}

func NewTokenMaker(cfg *config.Config) *TokenMaker {
    return &TokenMaker{
        accessSecret:  []byte(cfg.JWT_ACCESS_TOKEN_SECRET),
        refreshSecret: []byte(cfg.JWT_REFRESH_TOKEN_SECRET),
        accessTTL:     cfg.JWT_ACCESS_TOKEN_EXPIRED,
    }
}

// Option 2: Package-level init (less ideal but quick fix)
var cfg *config.Config

func init() {
    cfg = config.Load()
}
```

---

### T6: Tidak Ada Linter Configuration [HIGH]

**Bukti:**
Tidak ditemukan file `.golangci.yml` atau `.golangci.yaml` di repository.

**Dampak:**
- Tidak ada enforcement code style
- Potential bugs tidak terdeteksi (staticcheck, gosec)
- Inconsistent code quality

**Rekomendasi:**
Buat `.golangci.yml`:
```yaml
run:
  timeout: 5m

linters:
  enable:
    - errcheck
    - gosimple
    - govet
    - ineffassign
    - staticcheck
    - typecheck
    - unused
    - gosec
    - bodyclose
    - contextcheck
    - nilerr
    
linters-settings:
  errcheck:
    check-blank: true
  gosec:
    excludes:
      - G104 # jika perlu
```

---

### T7: Tidak Ada CI/CD Pipeline [HIGH]

**Bukti:**
Tidak ditemukan file `.github/workflows/`, `Jenkinsfile`, atau CI configuration lainnya.

**Dampak:**
- Tidak ada automated testing
- Manual build process error-prone
- Tidak ada automated code quality checks

**Rekomendasi:**
Buat `.github/workflows/ci.yml`:
```yaml
name: CI

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - run: go mod download
      - run: go vet ./...
      - run: golangci-lint run
      - run: go test -race -coverprofile=coverage.out ./...
      - run: go build -o app ./cmd/api
```

---

### T8: Error Handling Tidak Konsisten [MEDIUM]

**Bukti:**

1. Error diabaikan dengan `_`:
   - `config/config.go` (baris 86): `_ = godotenv.Load()`
   - `auth/service.go` (baris 232): `_ = s.emailLimiterRepo.SaveLimiterEmail(...)`
   - `owned_business.go` (baris 108): `_ = o.repo.UpsertOneBusiness(...)`

2. Error tidak di-wrap dengan context:
   ```go
   // store.go baris 47
   return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr) // %v bukan %w
   ```

**Dampak:**
- Silent failures sulit di-debug
- Error chain terputus, sulit trace root cause

**Rekomendasi:**
```go
// Gunakan %w untuk error wrapping
return fmt.Errorf("transaction failed: %w (rollback: %v)", err, rbErr)

// Untuk error yang sengaja diabaikan, tambahkan comment
_ = godotenv.Load() // ignore: .env optional in production
```

---

### T9: Goroutine Leak Potential [MEDIUM]

**Bukti:**
`auth/service.go` (baris 526-536):
```go
go func() {
    ctxTo, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    s.mailer.SendWelcomeEmail(ctxTo, mailer.WelcomeInputDTO{...})
}() // Error tidak di-handle, tidak ada tracking
```

**Dampak:**
- Jika goroutine stuck, tidak ada visibility
- Error email silently lost
- Sulit debug production issues

**Rekomendasi:**
```go
go func() {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    if err := s.mailer.SendWelcomeEmail(ctx, ...); err != nil {
        // Log error dengan context
        slog.Error("failed to send welcome email",
            "profileId", valid.ID,
            "email", *valid.Email,
            "error", err,
        )
    }
}()
```

Atau lebih baik: gunakan worker pool / message queue.

---

### T10: Context Timeout Tidak Dikonfigurasi [MEDIUM]

**Bukti:**
External calls tanpa timeout:
- `mailer/service.go` baris 179: `s.dialer.DialAndSend(m)` — no timeout
- Database queries tidak ada explicit timeout

**Dampak:**
- Request bisa hang indefinitely
- Cascading failures saat external service lambat

**Rekomendasi:**
```go
// Untuk SMTP
type MailerService struct {
    dialer  *gomail.Dialer
    timeout time.Duration // add timeout config
}

func (s *MailerService) sendEmail(ctx context.Context, input SendEmailInput) error {
    // Check context deadline
    if deadline, ok := ctx.Deadline(); ok {
        if time.Until(deadline) < 5*time.Second {
            return errors.New("insufficient time to send email")
        }
    }
    // ...
}
```

---

### T11: Panic Digunakan untuk Missing Config [MEDIUM]

**Bukti:**
`config/config.go` (baris 189):
```go
func getEnv(key string) string {
    if value, ok := os.LookupEnv(key); ok {
        return value
    }
    panic("ENV " + key + " is required")
}
```

**Dampak:**
- Server crash jika config missing
- Tidak ada graceful error message
- Sulit debug di containerized environment

**Rekomendasi:**
Return error instead of panic:
```go
func Load() (*Config, error) {
    cfg := &Config{}
    
    var missing []string
    
    cfg.PORT = os.Getenv("PORT")
    if cfg.PORT == "" {
        missing = append(missing, "PORT")
    }
    // ... repeat for other required vars
    
    if len(missing) > 0 {
        return nil, fmt.Errorf("missing required env vars: %v", missing)
    }
    
    return cfg, nil
}
```

---

### T13: Go Version Invalid [LOW]

**Bukti:**
`go.mod` baris 3:
```
go 1.25.4
```

**Dampak:**
- Go 1.25 tidak exist (latest stable adalah 1.22 per Dec 2024)
- Mungkin typo dari 1.23 atau 1.22
- Bisa menyebabkan build issues

**Rekomendasi:**
Perbaiki ke versi valid:
```
go 1.22
```

---

## Rekomendasi Roadmap Perbaikan

### Quick Wins (1-2 hari)
1. [ ] Fix go.mod version ke Go 1.22
2. [ ] Tambahkan database connection pool config
3. [ ] Ganti `fmt.Println` dengan `log.Printf` minimal (sementara)
4. [ ] Refactor `config.Load()` di token package — load sekali saja
5. [ ] Tambahkan `.golangci.yml` dan run linter

### Perbaikan Menengah (1-2 minggu)
1. [ ] Pindahkan email sending ke luar database transaction
2. [ ] Implementasi structured logging dengan `slog`
3. [ ] Setup CI pipeline basic (build + lint)
4. [ ] Tulis unit test untuk auth service (minimal 3-5 test cases)
5. [ ] Tambahkan context timeout untuk external calls

### Perbaikan Besar/Arsitektur (2+ minggu)
1. [ ] Implementasi message queue untuk async operations (email, notifications)
2. [ ] Comprehensive test suite dengan table-driven tests
3. [ ] Observability stack: structured logging + metrics + tracing
4. [ ] API documentation (OpenAPI/Swagger)
5. [ ] Graceful shutdown implementation

---

## Checklist Best Practice Go (Ringkas)

### Packaging & Modules
- [x] Struktur folder standar (`cmd/`, `internal/`, `pkg/`)
- [x] `go.mod` ada dan proper
- [ ] Go version valid di `go.mod`
- [x] Tidak ada cyclic dependencies terlihat

### Error Handling
- [x] Custom error types (`AppError`)
- [x] Centralized error response
- [ ] Consistent error wrapping dengan `%w`
- [ ] Semua error di-handle (tidak ada `_` untuk error kritikal)

### Context
- [x] Context propagation di service methods
- [x] Context digunakan untuk auth claims
- [ ] Context timeout untuk external calls
- [ ] Graceful cancellation handling

### Logging
- [ ] Structured logging
- [ ] Log levels (debug, info, warn, error)
- [ ] Request correlation ID di logs
- [ ] Sensitive data tidak di-log

### Testing
- [ ] Unit tests ada
- [ ] Table-driven tests
- [ ] Mocking untuk dependencies
- [ ] Coverage untuk critical paths

### Concurrency
- [x] Tidak ada obvious race conditions
- [ ] Goroutine tracking/error handling
- [x] Redis operations menggunakan context

### Linting/CI
- [ ] golangci-lint configuration
- [ ] CI pipeline
- [x] gofmt (code terformat)
- [ ] Automated testing di CI

### Security Basics
- [x] JWT untuk authentication
- [x] Password hashing (bcrypt assumed dari `utils.HashPassword`)
- [x] Input validation
- [ ] Rate limiting global
- [ ] CORS configuration
- [ ] Secrets tidak di-hardcode (env vars)

---

*Review ini dibuat berdasarkan static code analysis. Untuk assessment lebih lengkap, diperlukan runtime testing dan load testing.*
