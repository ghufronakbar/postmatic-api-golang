# Module App.Timezone

Module untuk mengambil daftar timezone yang tersedia.

## Directory

- `internal/module/app/timezone/handler/*`
- `internal/module/app/timezone/service/*`

---

## Endpoints

### GET /api/app/timezone

**Fungsi**: Mendapatkan semua timezone yang tersedia.

**Auth**: All Allowed

**Response**: List of all timezones (tanpa pagination)

```json
[
  {
    "id": 1,
    "name": "Asia/Jakarta",
    "offset": "+07:00",
    ...
  },
  ...
]
```

---

## Service Methods

| Method           | Description                 |
| ---------------- | --------------------------- |
| `GetAllTimezone` | Get all available timezones |

---

## Notes

- Endpoint ini mengembalikan semua timezone tanpa pagination
- Digunakan untuk menampilkan dropdown timezone di UI
