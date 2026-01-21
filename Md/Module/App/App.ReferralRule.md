# Module App.ReferralRule

Module untuk mengelola aturan referral (admin only untuk upsert).

## Directory

- `internal/module/app/referral_rule/handler/*`
- `internal/module/app/referral_rule/service/*`

---

## Endpoints

### GET /api/app/referral-rule

**Fungsi**: Mendapatkan aturan referral yang berlaku.

**Auth**: All Allowed

**Response**: Referral rule data (bonus amounts, etc.)

---

### POST /api/app/referral-rule

**Fungsi**: Create atau update aturan referral.

**Auth**: All Allowed (should be Admin Only)

**Body**:

```json
{
  "bonusAmount": 10000,
  "minTransaction": 50000,
  ...
}
```

**Response**: Updated referral rule

---

## Service Methods

| Method               | Description                    |
| -------------------- | ------------------------------ |
| `GetRuleReferral`    | Get active referral rule       |
| `UpsertRuleReferral` | Create or update referral rule |

---

## Notes

- Referral rule digunakan untuk menghitung bonus yang diberikan kepada referrer saat referred user melakukan transaksi.
- Hanya ada satu active referral rule pada satu waktu.
