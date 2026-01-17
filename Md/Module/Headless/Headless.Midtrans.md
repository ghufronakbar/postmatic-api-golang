# Module Headless.Midtrans

Modul ini bertanggung jawab untuk menangani komunikasi _Server-to-Server_ (Core API) dengan Payment Gateway Midtrans. Modul ini **TIDAK** menggunakan Snap/Popup, melainkan API langsung untuk kebutuhan custom UI.

## 1. Project Rules & Dependencies

- **Library**: Wajib menggunakan **Official Go SDK**: [`github.com/midtrans/midtrans-go`](https://github.com/midtrans/midtrans-go).
- **Scope**: Hanya implementasi **Core API** (`github.com/midtrans/midtrans-go/coreapi`).
- **Logger**: Gunakan `pkg/logger` dengan pattern `logger.From(ctx)` untuk logging.
- **Environment**: Support Sandbox dan Production berdasarkan config.
- **Headless**: Modul ini hanya dipanggil oleh module lain (internal), tidak boleh ada HTTP Handler/Controller di dalamnya.
- **DTO Wrapper**: Semua input dan output menggunakan DTO internal, bukan SDK types langsung.

## 2. Directory Structure

```text
internal/module/headless/midtrans/
├── service.go     # Interface, Struct, Constructor, Common Methods (Check/Cancel/Expire)
├── dto.go         # Input/Output DTOs (wrapper untuk SDK types)
├── mapper.go      # SDK response to DTO mappers
├── helper.go      # Helper functions (Signature Validation)
├── bank.go        # Bank Transfer implementation
└── ewallet.go     # E-Wallet/Gopay implementation
```

## 3. Configuration

Environment Variables yang diperlukan:

| Variable                 | Type    | Description               |
| ------------------------ | ------- | ------------------------- |
| `MIDTRANS_SERVER_KEY`    | String  | Server key dari dashboard |
| `MIDTRANS_CLIENT_KEY`    | String  | Client key dari dashboard |
| `MIDTRANS_MERCHANT_ID`   | String  | Merchant ID               |
| `MIDTRANS_IS_PRODUCTION` | Boolean | `true` untuk Production   |

## 4. Service Interface

```go
type Service interface {
    // E-Wallet
    ChargeGopay(ctx context.Context, req ChargeGopayInput) (*ChargeResponse, error)

    // Bank Transfer
    ChargeBankTransfer(ctx context.Context, req ChargeBankTransferInput) (*ChargeResponse, error)

    // Transaction Management
    CheckStatus(ctx context.Context, orderID string) (*TransactionStatusResponse, error)
    CancelTransaction(ctx context.Context, orderID string) (*TransactionStatusResponse, error)
    ExpireTransaction(ctx context.Context, orderID string) (*TransactionStatusResponse, error)

    // Security
    VerifySignature(orderID, statusCode, grossAmount, signatureKey string) bool
}
```

## 5. DTO Structures

### Input DTOs

```go
type ChargeGopayInput struct {
    OrderID         string          `json:"orderId" validate:"required"`
    GrossAmount     int64           `json:"grossAmount" validate:"required,min=1"`
    CustomerDetails CustomerDetails `json:"customerDetails"`
    Items           []ItemDetail    `json:"items"`
    CallbackURL     string          `json:"callbackUrl"`
}

type ChargeBankTransferInput struct {
    OrderID         string          `json:"orderId" validate:"required"`
    GrossAmount     int64           `json:"grossAmount" validate:"required,min=1"`
    Bank            string          `json:"bank" validate:"required,oneof=bca bni bri permata mandiri"`
    CustomerDetails CustomerDetails `json:"customerDetails"`
    Items           []ItemDetail    `json:"items"`
}
```

### Output DTOs

```go
type ChargeResponse struct {
    TransactionID     string        `json:"transactionId"`
    OrderID           string        `json:"orderId"`
    GrossAmount       string        `json:"grossAmount"`
    PaymentType       string        `json:"paymentType"`
    TransactionStatus string        `json:"transactionStatus"`
    Actions           []PaymentAction `json:"actions,omitempty"`      // E-Wallet
    VANumbers         []VANumber      `json:"vaNumbers,omitempty"`    // Bank Transfer
}

type TransactionStatusResponse struct {
    TransactionID     string `json:"transactionId"`
    OrderID           string `json:"orderId"`
    TransactionStatus string `json:"transactionStatus"`
    // ... other fields
}
```

## 6. Usage Example

```go
// Di router.go atau service lain
midtransSvc := midtrans.NewService(
    cfg.MIDTRANS_SERVER_KEY,
    cfg.MIDTRANS_IS_PRODUCTION,
)

// Charge Gopay
res, err := midtransSvc.ChargeGopay(ctx, midtrans.ChargeGopayInput{
    OrderID:     "ORDER-123",
    GrossAmount: 100000,
    CustomerDetails: midtrans.CustomerDetails{
        Email: "customer@example.com",
    },
})

// Charge Bank Transfer
res, err := midtransSvc.ChargeBankTransfer(ctx, midtrans.ChargeBankTransferInput{
    OrderID:     "ORDER-456",
    GrossAmount: 150000,
    Bank:        "bca",
})

// Check Status
status, err := midtransSvc.CheckStatus(ctx, "ORDER-123")

// Verify Webhook Signature
isValid := midtransSvc.VerifySignature(orderID, statusCode, grossAmount, signatureKey)
```

## 7. Logging Pattern

```go
// Gunakan logger.From(ctx) di dalam method
log := logger.From(ctx)
log.Info("Charging Gopay", "orderID", req.OrderID, "amount", req.GrossAmount)

// Error logging
log.Error("Failed to charge", "orderID", req.OrderID, "error", err)
```

## 8. Signature Verification

Untuk validasi callback notification dari Midtrans:

```go
// Formula: SHA512(orderId + statusCode + grossAmount + serverKey)
isValid := midtransSvc.VerifySignature(
    notification.OrderID,
    notification.StatusCode,
    notification.GrossAmount,
    notification.SignatureKey,
)
```

## 9. Design Decisions

### Kenapa DTO Wrapper?

1. **Abstraksi SDK**: Mudah migrasi jika SDK berubah atau ganti provider
2. **Kontrol Types**: Hanya expose fields yang diperlukan aplikasi
3. **Validation**: Input DTO bisa punya validation tags
4. **Testing**: Mudah mock interface tanpa depend on SDK types
5. **Future-proof**: Jika ingin migrasi ke HTTP client langsung, hanya ubah internal implementation

### Kenapa Headless?

1. Module ini tidak perlu HTTP endpoint sendiri
2. Dipanggil oleh module lain (e.g., Order, Payment)
3. Separation of concerns - payment logic terpisah dari business logic
