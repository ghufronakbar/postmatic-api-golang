# Module Business.BusinessInformation

Module untuk mengelola informasi business root dan membership.

## Directory

- `internal/module/business/business_information/handler/*`
- `internal/module/business/business_information/service/*`

---

## Endpoints

### GET /api/business

**Fungsi**: Menampilkan daftar business yang di-join oleh user dengan pagination.

**Auth**: All Allowed

**Query Params**:
| Param | Type | Required | Description |
|-------|------|----------|-------------|
| search | string | No | Search by business name |
| sortBy | string | No | Sort by field |
| sort | string | No | Sort direction (asc, desc) |
| page | int | No | Page number |
| limit | int | No | Items per page |
| dateStart | string | No | Filter by start date |
| dateEnd | string | No | Filter by end date |

**Response**: List of joined businesses with pagination

---

### POST /api/business

**Fungsi**: Setup business root untuk pertama kali (create new business).

**Auth**: All Allowed

**Body**:

```json
{
  "name": "My Business",
  "description": "Business description",
  ...
}
```

**Response**: Business setup response

---

### GET /api/business/{businessId}

**Fungsi**: Mendapatkan detail business by ID.

**Auth**: All Allowed + OwnedBusinessMiddleware

**Response**: Business detail

---

### DELETE /api/business/{businessId}

**Fungsi**: Menghapus business by ID (soft delete).

**Auth**: All Allowed + OwnedBusinessMiddleware

**Response**: Deleted business info

---

## Service Methods

| Method                           | Description                                    |
| -------------------------------- | ---------------------------------------------- |
| `GetJoinedBusinessesByProfileID` | List businesses yang di-join dengan pagination |
| `SetupBusinessRootFirstTime`     | Create new business root                       |
| `GetBusinessById`                | Get business detail by ID                      |
| `DeleteBusinessById`             | Delete business (soft delete)                  |
