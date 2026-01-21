# Module Business.BusinessTimezonePref

Module untuk mengelola preferensi timezone bisnis.

## Directory

- `internal/module/business/business_timezone_pref/handler/*`
- `internal/module/business/business_timezone_pref/service/*`

---

## Endpoints

### GET /api/business-timezone-pref/{businessId}

**Fungsi**: Mendapatkan timezone preference berdasarkan business root ID.

**Auth**: All Allowed + OwnedBusinessMiddleware

**Response**: Timezone preference data

---

### POST /api/business-timezone-pref/{businessId}

**Fungsi**: Upsert (create or update) timezone preference.

**Auth**: All Allowed + OwnedBusinessMiddleware

**Body**:

```json
{
  "timezoneId": 123
}
```

**Response**: Updated timezone preference

---

## Service Methods

| Method                                       | Description                      |
| -------------------------------------------- | -------------------------------- |
| `GetBusinessTimezonePrefByBusinessRootID`    | Get timezone pref by business ID |
| `UpsertBusinessTimezonePrefByBusinessRootID` | Create or update timezone pref   |
