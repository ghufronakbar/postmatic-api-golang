# GenerativeToken.ImageToken

Module untuk mengelola transaksi token generative image (pembelian dan penggunaan).

## Dependency

- Payment.Common (dipanggil oleh payment common service saat payment success)

## Directory

- `internal/module/generative_token/image_token/handler/*`
- `internal/module/generative_token/image_token/service/*`
- `internal/repository/queries/generative_token_image_transaction.sql`

---

## Func CreditTokenFromPayment

Untuk memasukkan data ke `generative_token_image_transactions` dengan type `in`.

### Handler: -Tidak Ada- (Internal Service Method)

Method ini dipanggil oleh `Payment.Common` service saat status payment berubah ke `success`.

### Input DTO:

```go
type CreateTokenTransactionInput struct {
    ProfileID        uuid.UUID
    BusinessRootID   int64
    PaymentHistoryID uuid.UUID
    Amount           int64
}
```

### Logic:

1. Check apakah `payment_history_id` sudah pernah di-credit (cek di `generative_token_image_transactions`)
2. Jika sudah ada, skip (idempotent)
3. Jika belum, insert record baru dengan:
   - `type`: `'in'`
   - `amount`: sesuai input
   - `profile_id`: sesuai input
   - `business_root_id`: sesuai input
   - `payment_history_id`: sesuai input

### Note:

- Method ini menerima `*entity.Queries` sebagai parameter untuk bisa dijalankan dalam database transaction
- Dipanggil dalam `ExecTx` bersama dengan update payment status

---

## Func SyncMissingTokenTransactions

Untuk sinkronisasi token transactions yang belum tercatat (background job).

### Handler: -Tidak Ada- (Internal Service Method)

Method ini dipanggil secara background (goroutine) saat `GetPaymentHistories` atau `GetPaymentHistoryById`.

### Input:

```go
paymentIDs []uuid.UUID
```

### Logic:

1. Query `GetSuccessPaymentIdsWithoutTokenTransaction` untuk menemukan payment yang:
   - Status = `success`
   - Belum ada record di `generative_token_image_transactions`
2. Untuk setiap payment yang ditemukan, insert record token transaction dengan type `in`
3. Logging setiap operasi (success/error)

### Note:

- Dijalankan dengan `context.Background()` dan timeout 30 detik
- Non-blocking, tidak mengganggu response ke client
- Berfungsi sebagai "self-healing" jika token gagal di-credit saat payment success

---

## Func GetTokenStatus

Untuk mendapatkan total token yang ada dan tersedia berdasarkan business yang dipilih.

### Handler: GET /api/app/generative-token/{businessId}/image-token/status

Auth: All Allowed (user harus member dari business tersebut)

### Response:

```json
{
  "availableToken": 100,
  "usedToken": 20,
  "totalToken": 120,
  "isExhausted": false
}
```

### Logic:

1. Validasi `businessId` menggunakan `OwnedBusinessMiddleware`
2. Query `SumTokenByBusinessAndType` dengan type `'in'` → `totalToken`
3. Query `SumTokenByBusinessAndType` dengan type `'out'` → `usedToken`
4. Calculate:
   - `availableToken = totalToken - usedToken`
   - `isExhausted = availableToken <= 0`

---

## SQL Queries

File: `internal/repository/queries/generative_token_image_transaction.sql`

| Query Name                                             | Description                                    |
| ------------------------------------------------------ | ---------------------------------------------- |
| `CreateGenerativeTokenImageTransaction`                | Insert token transaction record                |
| `GetGenerativeTokenImageTransactionByPaymentHistoryId` | Check if payment already credited              |
| `GetSuccessPaymentIdsWithoutTokenTransaction`          | Find success payments missing token records    |
| `SumTokenByBusinessAndType`                            | Sum token amount by business and type (in/out) |

---

## Database Schema

Table: `generative_token_image_transactions`

| Column               | Type                   | Description                             |
| -------------------- | ---------------------- | --------------------------------------- |
| `id`                 | BIGSERIAL              | Primary key                             |
| `type`               | token_transaction_type | `'in'` (purchased) or `'out'` (used)    |
| `amount`             | BIGINT                 | Token amount                            |
| `profile_id`         | UUID                   | FK to profiles                          |
| `business_root_id`   | BIGINT                 | FK to business_roots                    |
| `payment_history_id` | UUID (nullable)        | FK to payment_histories (for type 'in') |
| `created_at`         | TIMESTAMPTZ            | Created timestamp                       |
| `updated_at`         | TIMESTAMPTZ            | Updated timestamp                       |
| `deleted_at`         | TIMESTAMPTZ (nullable) | Soft delete                             |

---

## Files Created

- ✅ `internal/module/generative_token/image_token/handler/handler.go`
- ✅ `internal/module/generative_token/image_token/service/service.go`
- ✅ `internal/module/generative_token/image_token/service/viewmodel.go`
- ✅ `internal/module/generative_token/image_token/service/dto.go`
- ✅ `internal/repository/queries/generative_token_image_transaction.sql`
