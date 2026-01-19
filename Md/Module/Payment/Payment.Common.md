# Module Payment.Common

Module ini berfungsi untuk mengelola operasi umum payment: list, detail, cancel, dan webhook.

## Dependency

- Headless.Midtrans (untuk check status dan cancel transaction)
- GenerativeToken.ImageToken (untuk credit token saat payment success)
- Queue/Mailer (untuk send email notification)

## Directory

- `internal/module/payment/common/handler/*`
- `internal/module/payment/common/service/*`
- `internal/repository/queries/payment_history.sql`

---

## Endpoints

### 1. GET /api/app/payment/profile

**Fungsi**: Menampilkan list payment histories berdasarkan profile (user logged in).

**Auth**: All Allowed

**Query Params**:
| Param | Type | Required | Description |
|-------|------|----------|-------------|
| search | string | No | Search by product name or payment method |
| status | string | No | Filter by status (pending, success, failed, canceled, etc) |
| sortBy | string | No | Sort by field (created_at, total_amount). Default: created_at |
| sort | string | No | Sort direction (asc, desc). Default: desc |
| page | int | No | Page number. Default: 1 |
| limit | int | No | Items per page. Default: 10 |

**Response**: List of PaymentHistoryResponse with pagination

---

### 2. GET /api/app/payment/{businessId}/

**Fungsi**: Menampilkan list payment histories berdasarkan business.

**Auth**: All Allowed + OwnedBusinessMiddleware

**Middleware**:

- `OwnedBusinessMiddleware` - validasi user adalah member dari business tersebut
- `ReqFilterMiddleware` - parse query params (search, page, limit, sortBy, sort)

**Query Params**: Sama seperti endpoint `/profile`

**Response**: List of PaymentHistoryResponse with pagination

---

### 3. GET /api/app/payment/{businessId}/{id}

**Fungsi**: Menampilkan detail payment history berdasarkan ID dan business.

**Auth**: All Allowed + OwnedBusinessMiddleware

**Note**:

- Dapat diakses oleh semua member dalam business (tidak hanya pembuat payment)
- Jika status masih `pending`, akan cek status ke Midtrans (fallback jika webhook tidak sampai)
- Jika status berubah ke `success`:
  - Update status dalam database transaction (`ExecTx`)
  - Credit token ke `generative_token_image_transactions`
  - Send email notification

**Response**: PaymentHistoryResponse

---

### 4. POST /api/app/payment/{businessId}/{id}/cancel

**Fungsi**: Cancel pending payment.

**Auth**: All Allowed + OwnedBusinessMiddleware

**Note**:

- Dapat diakses oleh semua member dalam business (tidak hanya pembuat payment)
- Hanya payment dengan status `pending` yang dapat di-cancel
- Akan cancel transaction di Midtrans jika ada
- Send email notification

**Response**: PaymentHistoryResponse

---

### 5. POST /api/app/payment/webhook

**Fungsi**: Menerima callback dari Midtrans.

**Auth**: No Auth (public endpoint)

**Request Body**: MidtransNotification

**Logic**:

1. Verify signature dari Midtrans
2. Find payment by midtrans_transaction_id
3. Map status Midtrans ke status internal
4. Jika status berubah, update dalam transaction:
   - Update payment status
   - Update referral record status (if applicable)
   - Credit token (if success)
5. Send email notification

---

## Service Methods

| Method                                                 | Description                               |
| ------------------------------------------------------ | ----------------------------------------- |
| `GetPaymentHistoriesByProfile(filter)`                 | List payments by profile with pagination  |
| `GetPaymentHistoriesByBusiness(filter)`                | List payments by business with pagination |
| `GetPaymentHistoryById(id, profileID)`                 | Get payment detail by ID and profile      |
| `GetPaymentHistoryByIdAndBusiness(id, businessRootID)` | Get payment detail by ID and business     |
| `CancelPaymentByBusiness(id, businessRootID)`          | Cancel payment by ID and business         |
| `HandleWebhook(notification)`                          | Process Midtrans webhook                  |

---

## SQL Queries

| Query                                      | Description                                  |
| ------------------------------------------ | -------------------------------------------- |
| `GetPaymentHistoryById`                    | Get by ID only                               |
| `GetPaymentHistoryByIdAndProfile`          | Get by ID and profile_id                     |
| `GetPaymentHistoryByIdAndBusiness`         | Get by ID and business_root_id               |
| `GetPaymentHistoryByMidtransTransactionId` | Get by midtrans_transaction_id (for webhook) |
| `GetAllPaymentHistories`                   | List by profile with filter                  |
| `GetAllPaymentHistoriesByBusiness`         | List by business with filter                 |
| `CountAllPaymentHistories`                 | Count by profile with filter                 |
| `CountAllPaymentHistoriesByBusiness`       | Count by business with filter                |
| `UpdatePaymentHistoryStatus`               | Update status with timestamp                 |

---

## Background Sync

Pada setiap `GetPaymentHistoriesByProfile` atau `GetPaymentHistoriesByBusiness`, sistem akan menjalankan background sync untuk memastikan semua payment success memiliki token transaction.

```
Query payments → Collect success IDs → (Goroutine) SyncMissingTokenTransactions
```

---

## Flow Diagram

### Status Update Flow

```
┌─────────────────────────────────────────────────────────────────────────┐
│ When Payment Status → "success" (via GetDetail or Webhook)              │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│   ExecTx BEGIN (Database Transaction)                                  │
│   ├─► UpdatePaymentHistoryStatus(status = success)                     │
│   ├─► UpdateReferralRecordStatus (if applicable)                       │
│   └─► CreditTokenFromPayment(payment_id, amount, profile, business)    │
│   ExecTx COMMIT                                                         │
│                                                                         │
│   (async goroutine) sendPaymentSuccessEmail                            │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```
