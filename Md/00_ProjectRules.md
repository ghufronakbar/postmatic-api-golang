# Postmatic API - Project Rules & Conventions

> Dokumen ini mengatur standar penulisan kode, struktur folder, dan konvensi penamaan untuk menjaga kualitas dan konsistensi codebase.

---

## 📁 Struktur Folder

```
postmatic-api/
├── cmd/
│   └── api/
│       └── main.go              # Entry point aplikasi
├── config/
│   ├── config.go                # Load environment variables
│   ├── database.go              # Database connection
│   ├── redis.go                 # Redis connection
│   └── asynq.go                 # Asynq (queue) configuration
├── internal/
│   ├── router.go                # HTTP router (Chi)
│   ├── internal_middleware/     # Middleware aplikasi
│   │   ├── auth.go
│   │   ├── logger.go
│   │   ├── owned_business.go
│   │   └── req_filter.go
│   ├── module/                  # Domain modules
│   │   ├── account/
│   │   ├── app/
│   │   ├── business/
│   │   ├── creator/
│   │   ├── affiliator/
│   │   └── headless/           # Headless modules yang tidak ada expose ke HTTP
│   └── repository/
│       ├── entity/              # SQLC generated code
│       └── redis/               # Redis repositories
├── pkg/                         # Shared utilities
│   ├── errs/                    # Custom error types
│   ├── response/                # HTTP response helpers
│   ├── utils/                   # General utilities
│   └── logger/                  # Logging
└── Md/                          # Documentation
```

---

## 📦 Module Structure

Setiap module mengikuti struktur standar:

```
internal/module/{category}/{module_name}/
├── handler/
│   └── handler.go               # HTTP handlers
└── service/
    ├── dto.go                   # Data Transfer Objects (input)
    ├── service.go               # Business logic
    ├── viewmodel.go             # Response models
    └── filter.go                # Query filter structs (optional)
```

### Contoh:

```
internal/module/account/auth/
├── handler/
│   ├── handler.go               # package: auth_handler
│   └── cookie.go                # Helper functions
└── service/
    ├── dto.go                   # package: auth_service
    ├── service.go
    └── viewmodel.go
```

---

## 🏷️ Naming Conventions

### Package Names

| Location               | Package Name          | Contoh         |
| ---------------------- | --------------------- | -------------- |
| `{module}/handler/`    | `{module}_handler`    | `auth_handler` |
| `{module}/service/`    | `{module}_service`    | `auth_service` |
| `internal_middleware/` | `internal_middleware` | -              |

### Constructor Functions

| Type             | Pattern              | Contoh                         |
| ---------------- | -------------------- | ------------------------------ |
| Handler          | `NewHandler()`       | `auth_handler.NewHandler(...)` |
| Service          | `NewService()`       | `auth_service.NewService(...)` |
| Headless Service | `New{Name}Service()` | `NewTimezoneService()`         |

### Struct Names

| Type      | Pattern            | Contoh                          |
| --------- | ------------------ | ------------------------------- |
| Handler   | `Handler`          | `type Handler struct {...}`     |
| Service   | `{Module}Service`  | `type AuthService struct {...}` |
| Input DTO | `{Action}Input`    | `LoginCredentialInput`          |
| Response  | `{Entity}Response` | `LoginResponse`                 |
| Filter    | `{Action}Filter`   | `GetBusinessProductFilter`      |

### File Names

- Gunakan **snake_case** untuk nama file
- Handler: `handler.go`, `cookie.go` (helper)
- Service: `service.go`, `dto.go`, `viewmodel.go`, `filter.go`

---

## 📝 Code Conventions

### Routes Method

```go
// Handler harus memiliki method Routes() yang return chi.Router
func (h *Handler) Routes() chi.Router {
    r := chi.NewRouter()

    r.Get("/", h.GetAll)
    r.Post("/", h.Create)
    r.Put("/{id}", h.Update)
    r.Delete("/{id}", h.Delete)

    return r
}
```

### Handler Method Signature

```go
func (h *Handler) HandlerName(w http.ResponseWriter, r *http.Request) {
    // 1. Parse & validate input
    // 2. Get context data (profile, business, filter)
    // 3. Call service
    // 4. Return response
}
```

### Service Method Signature

```go
func (s *ServiceName) MethodName(ctx context.Context, input InputType) (ResponseType, error) {
    // Business logic
}
```

### Error Handling

Semua error harus menggunakan `pkg/errs` untuk konsistensi. **JANGAN** gunakan `fmt.Errorf` atau error biasa.

#### AppError Struct

```go
type AppError struct {
    Code             int               // HTTP status code
    Message          string            // Error message (constant, e.g. "RESOURCE_NOT_FOUND")
    Err              error             // Original error (for logging, not sent to client)
    ValidationErrors map[string]string // Validation errors (for validation failure)
}
```

#### Factory Functions

```go
import "postmatic-api/pkg/errs"

// 400 Bad Request - Client error, invalid input
errs.NewBadRequest("INVALID_INPUT")
errs.NewBadRequest("DUPLICATE_ENTRY")

// 401 Unauthorized - Authentication required
errs.NewUnauthorized("UNAUTHORIZED_ACCESS")
errs.NewUnauthorized("TOKEN_EXPIRED")

// 403 Forbidden - Authenticated but not allowed
errs.NewForbidden("FORBIDDEN")
errs.NewForbidden("NOT_OWNER")

// 404 Not Found - Resource not found
errs.NewNotFound("RESOURCE_NOT_FOUND")
errs.NewNotFound("USER_NOT_FOUND")

// 500 Internal Server Error - Server error (wrap original error)
errs.NewInternalServerError(err)

// 400 Validation Failed - With field-level errors
errs.NewValidationFailed(map[string]string{
    "email": "INVALID_EMAIL",
    "name":  "REQUIRED",
})
```

#### Naming Convention untuk Error Messages

- Gunakan **SCREAMING_SNAKE_CASE**
- Format: `{ENTITY}_{ACTION}_{STATUS}` atau `{ENTITY}_{STATUS}`
- Contoh:
  - `USER_NOT_FOUND`
  - `PAYMENT_METHOD_CODE_ALREADY_EXISTS`
  - `MIDTRANS_CHARGE_GOPAY_FAILED`
  - `INVALID_RATIO_FORMAT`

#### Usage in Service

```go
func (s *Service) GetById(ctx context.Context, id int64) (Response, error) {
    data, err := s.store.GetById(ctx, id)
    if err == sql.ErrNoRows {
        return Response{}, errs.NewNotFound("RESOURCE_NOT_FOUND")
    }
    if err != nil {
        return Response{}, errs.NewInternalServerError(err)
    }
    return mapToResponse(data), nil
}
```

#### Usage in Headless Modules (pihak ketiga)

```go
// Log error detail, return generic message
log.Error("Failed to charge", "error", err)
return nil, errs.NewBadRequest("MIDTRANS_CHARGE_FAILED")
```

### Response Format

```go
// Success
response.OK(w, r, "SUCCESS_MESSAGE", data)

// List dengan pagination
response.LIST(w, r, "SUCCESS_MESSAGE", data, &filter, pagination)

// Error
response.Error(w, r, err, nil)

// Validation error
response.ValidationFailed(w, r, validationErrors)
```

---

## 🔧 Dependency Injection

### Di Router (internal/router.go)

```go
func NewRouter(db *sql.DB, cfg *config.Config, asynqClient *asynq.Client, rdb *redis.Client) chi.Router {
    // 1. Initialize repositories
    store := repository.NewStore(db)

    // 2. Initialize services
    authSvc := auth_service.NewService(store, ...)

    // 3. Initialize handlers
    authHandler := auth_handler.NewHandler(authSvc, cfg)

    // 4. Mount routes
    r.Mount("/auth", authHandler.Routes())
}
```

### Service Dependencies

```go
type AuthService struct {
    store        entity.Store
    queueProducer *queue.Producer
    cfg          config.Config
    sessionRepo  *session_repository.SessionRepository
    tokenMaker   token.TokenMaker
}

func NewService(
    store entity.Store,
    queueProducer *queue.Producer,
    cfg config.Config,
    sessionRepo *session_repository.SessionRepository,
    tokenMaker token.TokenMaker,
) *AuthService {
    return &AuthService{
        store:         store,
        queueProducer: queueProducer,
        cfg:           cfg,
        sessionRepo:   sessionRepo,
        tokenMaker:    tokenMaker,
    }
}
```

---

## 🔐 Middleware Usage

### Authentication

```go
// Semua role diizinkan
allAllowed := internal_middleware.AuthMiddleware(*tokenSvc,
    []entity.AppRole{entity.AppRoleAdmin, entity.AppRoleUser})

// Admin only
adminOnly := internal_middleware.AuthMiddleware(*tokenSvc,
    []entity.AppRole{entity.AppRoleAdmin})

// Usage
r.Route("/protected", func(r chi.Router) {
    r.Use(allAllowed)
    r.Mount("/", handler.Routes())
})
```

### Request Filter

```go
r.Use(func(next http.Handler) http.Handler {
    return internal_middleware.ReqFilterMiddleware(next, SORT_BY_FIELDS)
})
```

### Owned Business

```go
// Di handler struct
type Handler struct {
    svc        *service.Service
    middleware *internal_middleware.OwnedBusiness
}

// Di routes
r.Route("/{businessId}", func(r chi.Router) {
    r.Use(h.middleware.ValidateOwnership)
    r.Get("/", h.GetBusiness)
})
```

---

## 📊 Filter & Pagination

### Filter Struct

```go
type GetResourceFilter struct {
    Search     string
    SortBy     string
    SortDir    string
    PageOffset int
    PageLimit  int
    Page       int
    DateStart  *time.Time
    DateEnd    *time.Time
    // Custom fields
    Status     *string
}
```

### Sort By Constants

```go
var SORT_BY = []string{
    "created_at",
    "updated_at",
    "name",
}
```

---

## 📁 Import Order

```go
import (
    // 1. Standard library
    "context"
    "net/http"

    // 2. Internal packages
    "postmatic-api/internal/internal_middleware"
    auth_service "postmatic-api/internal/module/account/auth/service"

    // 3. External packages
    "github.com/go-chi/chi/v5"
)
```

---

## ✅ Checklist Sebelum Commit

- [ ] Jangan buat migrasi, selalu konfirmasi jika perlu table/field yang diperlukan agar dibuatkan oleh saya sendiri
- [ ] Untuk generate sqlc baru tuliskan pada `internal/repository/queries/*.sql` dengan nama file `{nama_tabel_singular}*.sql` (setiap file `.sqlc` harus berdasarkan table tersebut, kecuali dibutuhkan join untuk get )
- [ ] Pastikan logika menggunakan store, bukan query langsung. dan pastikan pakai db transaction jika melakukan lebih dari 1 query mutasi secara paralel
- [ ] Semua handler menggunakan `NewHandler()` constructor
- [ ] Semua service menggunakan `NewService()` constructor
- [ ] Package name sesuai konvensi (`{module}_handler`, `{module}_service`)
- [ ] Import paths menggunakan path lengkap ke subfolder
- [ ] Error handling menggunakan `pkg/errs`
- [ ] Response menggunakan `pkg/response`
- [ ] Tidak ada business logic di handler (pindah ke service)
- [ ] Struct fields di-validate dengan tags jika perlu
- [ ] `go build ./cmd/api/main.go` berhasil tanpa error

---

## 🚫 Anti-Patterns (Hindari)

1. **Jangan** letakkan business logic di handler
2. **Jangan** akses database langsung dari handler
3. **Jangan** hardcode config values
4. **Jangan** return error tanpa wrapping dengan `errs`
5. **Jangan** import package service langsung tanpa alias jika ada nama konflik
6. **Jangan** buat file dengan nama sama di level yang berbeda

---

## 📝 Logger Usage

Logger menggunakan `slog` yang di-wrap dalam `pkg/logger`:

### Inisialisasi (di main.go)

```go
logger.Init(cfg.ENV) // "development" atau "production"
```

### Penggunaan di Code

```go
import "postmatic-api/pkg/logger"

// Global logger (tanpa context)
logger.L().Info("message", "key", value)
logger.L().Error("error message", "error", err)
logger.L().Debug("debug info", "data", someData)

// Logger dari context (di service/handler yang menerima context)
log := logger.From(ctx)
log.Info("Processing request", "userID", userID)
log.Error("Failed operation", "error", err)
```

### Log Levels

| Level   | Penggunaan                 |
| ------- | -------------------------- |
| `Debug` | Development info, verbose  |
| `Info`  | Normal operation, tracking |
| `Warn`  | Potential issues           |
| `Error` | Errors, failures           |

---

## 🔌 Headless Modules

Headless modules adalah service-only modules yang **tidak** terekspos ke HTTP. Digunakan untuk integrasi pihak ketiga atau logic internal.

### Struktur

```
internal/module/headless/{module_name}/
├── service.go     # Interface + Constructor + Common Methods
├── dto.go         # Input/Output DTOs (wrapper untuk SDK types)
├── mapper.go      # SDK response to DTO mappers
├── helper.go      # Helper functions
└── {feature}.go   # Feature-specific implementations
```

### Konvensi

| Aspek       | Pattern                                          |
| ----------- | ------------------------------------------------ |
| Interface   | `type Service interface {...}`                   |
| Struct      | `type {module}Service struct {...}` (unexported) |
| Constructor | `NewService(...) Service`                        |
| DTO Input   | `{Action}Input` (e.g., `ChargeGopayInput`)       |
| DTO Output  | `{Entity}Response` (e.g., `ChargeResponse`)      |

### Contoh (Midtrans)

```go
// Di module caller
midtransSvc := midtrans.NewService(cfg.MIDTRANS_SERVER_KEY, cfg.MIDTRANS_IS_PRODUCTION)

// Usage
res, err := midtransSvc.ChargeGopay(ctx, midtrans.ChargeGopayInput{
    OrderID:     "ORDER-123",
    GrossAmount: 100000,
})
```

### Kenapa DTO Wrapper?

1. **Abstraksi SDK**: Mudah migrasi jika SDK berubah atau ganti provider
2. **Kontrol Types**: Hanya expose fields yang diperlukan
3. **Validation**: Input DTO bisa punya validation tags
4. **Testing**: Mudah mock tanpa depend on SDK types

---

## 📚 References

- [Go Project Layout](https://github.com/golang-standards/project-layout)
- [Chi Router](https://github.com/go-chi/chi)
- [SQLC](https://sqlc.dev/)
