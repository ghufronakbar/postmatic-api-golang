# Module App.CategoryCreatorImage

Module untuk mengambil kategori-kategori image creator (type dan product).

## Directory

- `internal/module/app/category_creator_image/handler/*`
- `internal/module/app/category_creator_image/service/*`

---

## Endpoints

### GET /api/app/category-creator-image/type

**Fungsi**: Mendapatkan daftar category creator image type dengan pagination.

**Auth**: All Allowed

**Query Params**:
| Param | Type | Required | Description |
|-------|------|----------|-------------|
| search | string | No | Search category |
| sortBy | string | No | Sort by field |
| sort | string | No | Sort direction (asc, desc) |
| page | int | No | Page number |
| limit | int | No | Items per page |

**Response**: List of category creator image types with pagination

---

### GET /api/app/category-creator-image/product

**Fungsi**: Mendapatkan daftar category creator image product dengan pagination.

**Auth**: All Allowed

**Query Params**:
| Param | Type | Required | Description |
|-------|------|----------|-------------|
| search | string | No | Search product |
| sortBy | string | No | Sort by field |
| sort | string | No | Sort direction |
| page | int | No | Page number |
| limit | int | No | Items per page |
| category | string | No | Filter by locale |

**Response**: List of category creator image products with pagination

---

## Service Methods

| Method                           | Description                        |
| -------------------------------- | ---------------------------------- |
| `GetCategoryCreatorImageType`    | Get image types with pagination    |
| `GetCategoryCreatorImageProduct` | Get image products with pagination |
