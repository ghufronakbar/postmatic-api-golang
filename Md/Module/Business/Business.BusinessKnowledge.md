# Module Business.BusinessKnowledge

Module untuk mengelola business knowledge (informasi tambahan tentang bisnis).

## Directory

- `internal/module/business/business_knowledge/handler/*`
- `internal/module/business/business_knowledge/service/*`

---

## Endpoints

### GET /api/business-knowledge/{businessId}

**Fungsi**: Mendapatkan business knowledge berdasarkan business root ID.

**Auth**: All Allowed + OwnedBusinessMiddleware

**Response**: Business knowledge data

---

### POST /api/business-knowledge/{businessId}

**Fungsi**: Upsert (create or update) business knowledge.

**Auth**: All Allowed + OwnedBusinessMiddleware

**Body**:

```json
{
  "knowledge": "...",
  ...
}
```

**Response**: Updated business knowledge

---

## Service Methods

| Method                                    | Description                  |
| ----------------------------------------- | ---------------------------- |
| `GetBusinessKnowledgeByBusinessRootID`    | Get knowledge by business ID |
| `UpsertBusinessKnowledgeByBusinessRootID` | Create or update knowledge   |
