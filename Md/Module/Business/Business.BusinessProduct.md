# Module Business.BusinessProduct

Module untuk mengelola produk bisnis.

## Directory

- `internal/module/business/business_product/handler/*`
- `internal/module/business/business_product/service/*`

---

## Endpoints

### GET /api/business-product/{businessId}

**Fungsi**: Mendapatkan daftar produk business dengan pagination.

**Auth**: All Allowed + OwnedBusinessMiddleware

**Query Params**:
| Param | Type | Required | Description |
|-------|------|----------|-------------|
| search | string | No | Search by product name |
| sortBy | string | No | Sort by field |
| sort | string | No | Sort direction |
| page | int | No | Page number |
| limit | int | No | Items per page |
| dateStart | string | No | Filter by start date |
| dateEnd | string | No | Filter by end date |
| category | string | No | Filter by category |

**Response**: List of products with pagination

---

### POST /api/business-product/{businessId}

**Fungsi**: Create produk baru.

**Auth**: All Allowed + OwnedBusinessMiddleware

**Body**:

```json
{
  "name": "Product Name",
  "description": "...",
  "price": 100000,
  "category": "...",
  ...
}
```

**Response**: Created product

---

### PUT /api/business-product/{businessId}/{businessProductId}

**Fungsi**: Update produk.

**Auth**: All Allowed + OwnedBusinessMiddleware

**Body**: Same as POST

**Response**: Updated product

---

### DELETE /api/business-product/{businessId}/{businessProductId}

**Fungsi**: Soft delete produk.

**Auth**: All Allowed + OwnedBusinessMiddleware

**Response**: Deleted product info

---

## Service Methods

| Method                                      | Description               |
| ------------------------------------------- | ------------------------- |
| `GetBusinessProductsByBusinessRootID`       | List products with filter |
| `CreateBusinessProduct`                     | Create new product        |
| `UpdateBusinessProduct`                     | Update product            |
| `SoftDeleteBusinessProductByBusinessRootID` | Soft delete product       |
