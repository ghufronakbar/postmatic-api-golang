# Module App.TokenProduct

Module untuk menghitung konversi token berdasarkan harga atau jumlah token.

## Directory

- `internal/module/app/token_product/handler/*`
- `internal/module/app/token_product/service/*`

---

## Endpoints

### GET /api/app/token-product

**Fungsi**: Kalkulasi harga token atau jumlah token berdasarkan input.

**Auth**: All Allowed

**Query Params**:
| Param | Type | Required | Allowed Values | Description |
|-------|------|----------|----------------|-------------|
| amount | int64 | Yes | - | Jumlah (price atau token tergantung `from`) |
| currencyCode | string | Yes | `IDR` | Kode mata uang |
| from | string | Yes | `price`, `token` | Konversi dari apa |
| type | string | Yes | `image_token`, `video_token`, `livestream_token` | Jenis token |

**Response**:

```json
{
  "id": 1,
  "type": "image_token",
  "currencyCode": "IDR",
  "tokenAmount": 100,
  "priceAmount": 50000,
  "createdAt": "...",
  "updatedAt": "..."
}
```

---

## Business Logic

### Konversi Token

**1. Dari Price ke Token (`from=price`)**

- Input: jumlah uang (IDR)
- Output: jumlah token yang didapat
- Formula: `tokenAmount = (inputAmount * baseTokenAmount) / basePriceAmount`

**2. Dari Token ke Price (`from=token`)**

- Input: jumlah token
- Output: harga yang harus dibayar (IDR)
- Formula: `priceAmount = (inputToken * basePriceAmount) / baseTokenAmount`

### Validations

1. **Type**: Harus salah satu dari `image_token`, `video_token`, `livestream_token`
2. **CurrencyCode**: Currently hanya support `IDR`
3. **From**: Harus `price` atau `token`
4. **Token Product**: Harus ada data token product dengan type dan currency yang sesuai di database

### Error Codes

| Code                      | Description                                           |
| ------------------------- | ----------------------------------------------------- |
| `INVALID_FROM`            | Parameter `from` bukan `price` atau `token`           |
| `TOKEN_PRODUCT_NOT_FOUND` | Tidak ada token product dengan type/currency tersebut |

---

## Service Methods

| Method                  | Description                              |
| ----------------------- | ---------------------------------------- |
| `CalculateTokenProduct` | Calculate token/price conversion         |
| `convertIDRToTokens`    | Convert price to token amount (internal) |
| `convertTokensToIDR`    | Convert token to price amount (internal) |

---

## Notes

- Module ini digunakan untuk preview harga sebelum checkout
- Tidak termasuk tax/admin fee (hanya base price)
- Base rate diambil dari tabel `app_token_products`
