# Module Affiliator.ReferralBasic

Module untuk mengelola referral code basic profile user.

## Directory

- `internal/module/affiliator/referral_basic/handler/*`
- `internal/module/affiliator/referral_basic/service/*`

---

## Endpoints

### GET /api/affiliator/referral-basic

**Fungsi**: Mendapatkan atau membuat referral code basic untuk profile user.

**Auth**: All Allowed (requires profile context)

**Response**:

```json
{
  "id": 123,
  "code": "ABC12345",
  "totalDiscount": 10000,
  "discountType": "fixed",
  "expiredDays": 30,
  "maxDiscount": 50000,
  "maxUsage": 100,
  "rewardPerReferral": 5000,
  "createdAt": "...",
  "updatedAt": "..."
}
```

---

## Business Logic

### GetReferralBasicByProfileId

Flow untuk mendapatkan/membuat referral code:

```
1. Fast Path: Cek apakah profile sudah punya referral code basic
   ├── YA  → Return referral code yang sudah ada
   └── TIDAK → Lanjut ke step 2

2. Ambil aturan referral dari App.ReferralRule
   - TotalDiscount
   - DiscountType (fixed/percentage)
   - ExpiredDays
   - MaxDiscount
   - MaxUsage
   - RewardPerReferral

3. Generate unique referral code dengan retry loop (max 10x):
   ├── Generate random 8 character code (A-Z, 0-9)
   ├── Coba create di database
   ├── Jika SUCCESS → Return new referral code
   ├── Jika UNIQUE VIOLATION (code sudah ada):
   │   ├── Cek apakah profile keburu punya code → Return existing
   │   └── Jika collision code → Retry dengan code baru
   └── Jika error lain → Return error

4. Jika 10x retry gagal → Return "FAILED_TO_GENERATE_UNIQUE_REFERRAL_CODE"
```

---

### ValidateReferralForPayment (Internal Service)

**Tidak expose ke API**, digunakan oleh payment service untuk validasi referral code.

**7 Tahap Validasi:**

| #   | Check                                   | Error Message                         |
| --- | --------------------------------------- | ------------------------------------- |
| 1   | Code exists                             | `REFERRAL_CODE_NOT_FOUND`             |
| 2   | Code is active                          | `REFERRAL_CODE_INACTIVE`              |
| 3   | Not self-referral (cannot use own code) | `CANNOT_USE_OWN_REFERRAL_CODE`        |
| 4   | Code not expired (based on expiredDays) | `REFERRAL_CODE_EXPIRED`               |
| 5   | Max usage not reached                   | `REFERRAL_CODE_MAX_USAGE_REACHED`     |
| 6   | Profile belum pernah pakai code ini     | `PROFILE_ALREADY_USED_REFERRAL_CODE`  |
| 7   | Business belum pernah pakai code ini    | `BUSINESS_ALREADY_USED_REFERRAL_CODE` |

**Success Response:**

```json
{
  "valid": true,
  "message": "REFERRAL_CODE_VALID",
  "referralCodeId": 123,
  "discountType": "fixed",
  "totalDiscount": 10000,
  "maxDiscount": 50000,
  "ownerProfileId": "uuid...",
  "rewardPerReferral": 5000
}
```

---

### GetReferralCodeByCode (Internal Service)

**Tidak expose ke API**, digunakan internal untuk lookup referral code detail.

---

## Service Methods

| Method                        | Expose      | Description                          |
| ----------------------------- | ----------- | ------------------------------------ |
| `GetReferralBasicByProfileId` | ✅ API      | Get/create referral code for profile |
| `GetReferralCodeByCode`       | ❌ Internal | Get referral detail by code          |
| `ValidateReferralForPayment`  | ❌ Internal | Validate code for payment (7 checks) |

---

## Code Generation

- Format: 8 karakter uppercase alphanumeric (A-Z, 0-9)
- Example: `ABC12345`, `XY9K3MNP`
- Retry up to 10 times jika collision

---

## Dependencies

- `App.ReferralRule`: Untuk mendapatkan aturan default referral (discount, max usage, etc.)
