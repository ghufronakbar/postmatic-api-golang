# Module Payment.ImageToken

Module untuk checkout image token

## Dependency

- Headless.Midtrans (untuk charge)
- Affiliator.Referral (untuk referral)
- App.PaymentMethod (untuk validasi payment method yang aktif dan tidak)
- App.TokenProduct (untuk calculate token)
- **GenerativeToken.ImageToken** (untuk credit token saat payment success)

## Directory

- `internal/module/payment/image_token/handler/*` (untuk handler/http)
- `internal/module/payment/image_token/service/*` (untuk service)
- `internal/module/payment/common/handler/*` (untuk common payment operations)
- `internal/module/payment/common/service/*` (untuk common payment service)

---

## Endpoint: GET /api/app/payment/image-token

### Fungsi:

- Cek harga token beserta admin fee dan tax

### Query Params:

- amount: int64 (required string-> int64) // amount in token (token yang ingin dibeli)
- currencyCode: string (required) // currency code (ex: IDR)
- paymentMethod: string (required) // payment method (ex: bca, bri, gopay)
- referralCode: string (optional) // referral code (ex: POSTMATIC10)
- businessRootId: string (required) // business root id (ex: 1)

### Note:

- amount diambil dari service App.TokenProduct, jangan query table app_token_products
- paymentMethod diambil dari service App.PaymentMethod, jangan query table app_payment_methods.code
- validasi referralCode diambil dari service Affiliator.Referral, jangan query table profile_referral_codes
- jika semisal belum ada fungsi seperti untuk apakah pengguna sudah menggunakan ref atau belum. buat itu di service Affiliator.Referral jangan langsung pada service Payment.ImageToken (begitupula dengan yang lainnya)
- Untuk diskon, itu opsional. namun jika ada, pengecekannya harus "apakah profile pernah menggunakan" dan "apakah business root pernah menggunakan", untuk menghindari abuse. selain itu cek rulesnya sesuai dengan service yang ada

### Response:

```json
{
  "referralCode": {
    "valid": true,
    "message": "REFERRAL_CODE_VALID"
  },
  "token": {
    "itemPrice": 10000,
    "adminFee": 1000,
    "tax": 1000,
    "total": 12000
  },
  "paymentMethod": {
    "code": "bca",
    "name": "BCA",
    "type": "bank"
  }
}
```

---

## Endpoint: POST /api/app/payment/image-token

### Fungsi:

Untuk pertama validasi menggunakan service yang sama yang digunakan di endpoint GET /api/app/payment/image-token, lalu lakukan charge sesuai dengan payment method yang digunakan. jika pada response GET /api/app/payment/image-token menggunakan query params, pada POST pakai body (untuk handler nya)

---

## Token Crediting pada Payment Success

Saat status payment berubah ke `success`, token akan otomatis di-credit ke `generative_token_image_transactions`.

### Flow:

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

### Implementation Details:

1. **Database Transaction**: Status update dan token credit dijalankan dalam satu `ExecTx` transaction untuk menjaga konsistensi data
2. **Idempotent**: Method `CreditTokenFromPayment` akan skip jika payment sudah pernah di-credit
3. **Error Handling**: Jika token credit gagal, transaction tetap commit (log error saja) agar payment status tetap ter-update

---

## Background Sync untuk Token

Pada setiap `GetPaymentHistories` atau `GetPaymentHistoryById`, sistem akan menjalankan background sync untuk memastikan semua payment success memiliki token transaction.

### Flow:

```
┌─────────────────────────────────────────────────────────────────────────┐
│ On GetPaymentHistories / GetPaymentHistoryById                          │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│   1. Query payment histories                                            │
│   2. Collect payment IDs dengan status = 'success'                     │
│   3. (Goroutine) SyncMissingTokenTransactions:                         │
│      ├─► Query GetSuccessPaymentIdsWithoutTokenTransaction             │
│      └─► Bulk insert token transactions for missing records            │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

### Note:

- Background sync berjalan di goroutine dengan `context.Background()` dan timeout 30 detik
- Non-blocking, tidak mempengaruhi response time ke client
- Berfungsi sebagai "self-healing" mechanism

---

## Payment Common Service Dependencies

```go
type PaymentCommonService struct {
    store           entity.Store
    midtrans        midtrans.Service
    queue           queue.MailerProducer
    generativeToken *image_token_service.ImageTokenService  // NEW
}
```

### Constructor:

```go
func NewService(
    store entity.Store,
    midtrans midtrans.Service,
    queue queue.MailerProducer,
    generativeToken *image_token_service.ImageTokenService,
) *PaymentCommonService
```

---

## Perhitungan Admin Fee dan Tax

Perhitungan dilakukan secara berurutan (sekuensial) dengan aturan pembulatan ke atas (rounding up) di setiap langkah.

### Langkah A: Hitung Nominal Diskon

Jika diskon tipe "Fixed" (Tetap): Langsung pakai angka nominal diskonnya (Misal: Diskon Rp 10.000).

Jika diskon tipe "Percentage" (Persen): Hitung persen dari Harga Asli Barang, lalu bulatkan ke atas.

Cek Batas Maksimal: Jika hasil hitungan diskon melebihi maxDiscount (batas maksimal), maka diskon dipotong/mentok sesuai batas maksimal tersebut.

### Langkah B: Hitung Biaya Admin

Penting: Biaya admin dihitung berdasarkan Harga Asli Barang (bukan harga setelah diskon).

Jika admin tipe "Fixed": Pakai nominal tetap.

Jika admin tipe "Percentage": Hitung persen dari harga asli barang.

### Langkah C: Hitung Subtotal (Sebelum Pajak)

Harga Setelah Diskon: (Harga Asli - Diskon) -> Dibulatkan ke atas.

Dasar Pengenaan Pajak (Subtotal Setelah Admin): (Harga Setelah Diskon + Biaya Admin) -> Dibulatkan ke atas.

Catatan: Ini berarti Biaya Admin ikut menjadi dasar perhitungan pajak nanti.

### Langkah D: Hitung Pajak (Tax)

Pajak dihitung dari hasil Langkah C (Harga barang bersih + Biaya Admin).

Rumus: (Harga setelah Diskon + Admin) x Persentase Pajak.

Hasilnya dibulatkan ke atas.

### Langkah E: Hitung Total Akhir

Total yang harus dibayar user adalah penjumlahan dari Langkah C dan Langkah D.

Rumus: (Harga setelah Diskon + Admin) + Pajak.

### Summary Formula:

```
Diskon = Harga Asli × %Diskon (Dibulatkan ke atas, cek max cap).
Admin = Harga Asli × %Admin.
Subtotal 1 = Harga Asli - Diskon.
Subtotal 2 = Subtotal 1 + Admin.
Pajak = Subtotal 2 × %Pajak.
TOTAL BAYAR = Subtotal 2 + Pajak.
```
