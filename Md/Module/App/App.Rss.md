# Module App.Rss

Module untuk mengambil daftar RSS feed dan kategori RSS.

## Directory

- `internal/module/app/rss/handler/*`
- `internal/module/app/rss/service/*`

---

## Endpoints

### GET /api/app/rss

**Fungsi**: Mendapatkan daftar RSS feed dengan pagination.

**Auth**: All Allowed

**Query Params**:
| Param | Type | Required | Description |
|-------|------|----------|-------------|
| search | string | No | Search feed name |
| sortBy | string | No | Sort by field |
| sort | string | No | Sort direction |
| page | int | No | Page number |
| limit | int | No | Items per page |
| category | int64 | No | Filter by category ID |

**Response**: List of RSS feeds with pagination

**Notes**:

- `category` parameter harus berupa integer64 (ID kategori)
- Jika tidak valid akan return validation error

---

### GET /api/app/rss/category

**Fungsi**: Mendapatkan daftar kategori RSS dengan pagination.

**Auth**: All Allowed

**Query Params**:
| Param | Type | Required | Description |
|-------|------|----------|-------------|
| search | string | No | Search category name |
| sortBy | string | No | Sort by field |
| sort | string | No | Sort direction |
| page | int | No | Page number |
| limit | int | No | Items per page |

**Response**: List of RSS categories with pagination

---

## Service Methods

| Method           | Description                              |
| ---------------- | ---------------------------------------- |
| `GetRSSFeed`     | Get RSS feeds with filter and pagination |
| `GetRSSCategory` | Get RSS categories with pagination       |
