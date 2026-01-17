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

```go
// Gunakan custom errors dari pkg/errs
if err != nil {
    return nil, errs.NewInternalServerError(err)
}

if notFound {
    return nil, errs.NewNotFound("RESOURCE_NOT_FOUND")
}

if badRequest {
    return nil, errs.NewBadRequest("INVALID_INPUT")
}

if unauthorized {
    return nil, errs.NewUnauthorized("UNAUTHORIZED_ACCESS")
}
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

## 📚 References

- [Go Project Layout](https://github.com/golang-standards/project-layout)
- [Chi Router](https://github.com/go-chi/chi)
- [SQLC](https://sqlc.dev/)
