# Module Creator.CreatorImage

Module untuk mengelola konten gambar yang dibuat oleh creator (template, desain, dll).

## Directory

- `internal/module/creator/creator_image/handler/*`
- `internal/module/creator/creator_image/service/*`

---

## Endpoints

### GET /api/creator/creator-image

**Fungsi**: Mendapatkan daftar creator image milik profile dengan pagination dan filter.

**Auth**: All Allowed (requires profile context)

**Query Params**:
| Param | Type | Required | Description |
|-------|------|----------|-------------|
| search | string | No | Search by name/description |
| sortBy | string | No | Sort by field |
| sort | string | No | Sort direction (asc, desc) |
| page | int | No | Page number |
| limit | int | No | Items per page |
| dateStart | string | No | Filter by start date |
| dateEnd | string | No | Filter by end date |
| category | int64 | No | Filter by product category ID |
| typeCategoryId | int64 | No | Filter by type category ID |
| published | bool | No | Filter by publish status |

**Response**: List of creator images with pagination

---

### POST /api/creator/creator-image

**Fungsi**: Create creator image baru.

**Auth**: All Allowed (requires profile context)

**Body**:

```json
{
  "name": "Image Name",
  "description": "...",
  "imageUrl": "https://...",
  "typeCategoryId": 1,
  "productCategoryId": 2,
  "published": true,
  ...
}
```

**Response**: Created creator image

---

### PUT /api/creator/creator-image/{creatorImageId}

**Fungsi**: Update creator image.

**Auth**: All Allowed (requires profile context, must be owner)

**Path Params**:
| Param | Type | Required | Description |
|-------|------|----------|-------------|
| creatorImageId | int64 | Yes | ID creator image |

**Body**: Same as POST

**Response**: Updated creator image

---

### DELETE /api/creator/creator-image/{creatorImageId}

**Fungsi**: Soft delete creator image.

**Auth**: All Allowed (requires profile context, must be owner)

**Path Params**:
| Param | Type | Required | Description |
|-------|------|----------|-------------|
| creatorImageId | int64 | Yes | ID creator image |

**Response**: Deleted creator image info

---

## Business Logic

### Ownership

- Creator image hanya bisa diakses/dimodifikasi oleh owner (profile yang membuat)
- `ProfileID` otomatis diambil dari context authentication

### Filtering

Image dapat difilter berdasarkan:

1. **typeCategoryId**: Kategori tipe gambar (dari App.CategoryCreatorImage/type)
2. **productCategoryId** (via `category`): Kategori produk (dari App.CategoryCreatorImage/product)
3. **published**: Status publish (true/false)
4. **dateStart/dateEnd**: Range tanggal pembuatan

### Dependencies

- `App.CategoryCreatorImage`: Untuk validasi category ID

---

## Service Methods

| Method                       | Description                           |
| ---------------------------- | ------------------------------------- |
| `GetCreatorImageByProfileId` | Get images with filter and pagination |
| `CreateCreatorImage`         | Create new image                      |
| `UpdateCreatorImage`         | Update image (owner only)             |
| `SoftDeleteCreatorImage`     | Soft delete image (owner only)        |
