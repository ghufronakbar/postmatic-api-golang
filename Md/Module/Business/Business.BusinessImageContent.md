# Module Business.BusinessImageContent

Module untuk mengelola image content bisnis (konten gambar untuk posting).

## Directory

- `internal/module/business/business_image_content/handler/*`
- `internal/module/business/business_image_content/service/*`

---

## Endpoints

### GET /api/business-image-content/{businessId}

**Fungsi**: Mendapatkan daftar image content dengan pagination.

**Auth**: All Allowed + OwnedBusinessMiddleware

**Query Params**:
| Param | Type | Required | Description |
|-------|------|----------|-------------|
| search | string | No | Search |
| sortBy | string | No | Sort by field |
| sort | string | No | Sort direction |
| page | int | No | Page number |
| limit | int | No | Items per page |
| dateStart | string | No | Filter by start date |
| dateEnd | string | No | Filter by end date |
| category | string | No | `readyToPosts` untuk filter ready to post |

**Response**: List of image contents with pagination

---

### POST /api/business-image-content/{businessId}

**Fungsi**: Create image content baru.

**Auth**: All Allowed + OwnedBusinessMiddleware

**Body**:

```json
{
  "imageUrl": "https://...",
  "caption": "...",
  "hashtags": ["..."],
  ...
}
```

**Response**: Created image content

---

### PUT /api/business-image-content/{businessId}/{businessImageContentId}

**Fungsi**: Update image content.

**Auth**: All Allowed + OwnedBusinessMiddleware

**Body**: Same as POST

**Response**: Updated image content

---

### DELETE /api/business-image-content/{businessId}/{businessImageContentId}

**Fungsi**: Delete image content.

**Auth**: All Allowed + OwnedBusinessMiddleware

**Response**: Deleted image content info

---

## Service Methods

| Method                                     | Description               |
| ------------------------------------------ | ------------------------- |
| `GetBusinessImageContentsByBusinessRootID` | List contents with filter |
| `CreateBusinessImageContent`               | Create new content        |
| `UpdateBusinessImageContent`               | Update content            |
| `DeleteBusinessImageContent`               | Delete content            |
