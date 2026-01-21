# Module Business.BusinessRole

Module untuk mengelola role/persona bisnis.

## Directory

- `internal/module/business/business_role/handler/*`
- `internal/module/business/business_role/service/*`

---

## Endpoints

### GET /api/business-role/{businessId}

**Fungsi**: Mendapatkan business role berdasarkan business root ID.

**Auth**: All Allowed + OwnedBusinessMiddleware

**Response**: Business role data

---

### POST /api/business-role/{businessId}

**Fungsi**: Upsert (create or update) business role.

**Auth**: All Allowed + OwnedBusinessMiddleware

**Body**:

```json
{
  "role": "...",
  "persona": "...",
  ...
}
```

**Response**: Updated business role

---

## Service Methods

| Method                               | Description             |
| ------------------------------------ | ----------------------- |
| `GetBusinessRoleByBusinessRootID`    | Get role by business ID |
| `UpsertBusinessRoleByBusinessRootID` | Create or update role   |
