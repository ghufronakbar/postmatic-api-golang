# Module Creator.BusinessCreatorImage

Modul ini untuk mengelola gambar creator yang disimpan oleh business.

## 1. Overview

Tujuan utama module ini untuk menyimpan gambar yang dibuat oleh creator yang disimpan oleh business, mirip dengan fitur "bookmark" atau "favorite".

## 2. Database Schema

Tabel: `business_saved_template_creator_images`

| Column             | Type        | Description             |
| ------------------ | ----------- | ----------------------- |
| `id`               | BIGSERIAL   | Primary key             |
| `business_root_id` | BIGINT      | FK → business_roots(id) |
| `creator_image_id` | BIGINT      | FK → creator_images(id) |
| `created_at`       | TIMESTAMPTZ | Waktu disimpan          |
| `updated_at`       | TIMESTAMPTZ | Waktu update            |
| `deleted_at`       | TIMESTAMPTZ | Soft delete marker      |

**Unique Constraint**: Satu business hanya dapat menyimpan 1 creator_image yang sama (partial unique index dengan `WHERE deleted_at IS NULL`).

## 3. Directory Structure

```text
internal/module/creator/business_creator_image/
├── handler/
│   └── handler.go           # HTTP handlers dengan OwnedBusinessMiddleware
└── service/
    ├── dto.go               # Input DTOs
    ├── viewmodel.go         # Output DTOs
    ├── filter.go            # Sort by constants
    └── service.go           # Business logic
```

## 4. Endpoints

### GET /api/creator/business-saved-creator-image/{businessId}

List all saved creator images untuk business tertentu.

**Authentication**: Required  
**Middleware**: `OwnedBusinessMiddleware`

**Query Parameters**:

| Param            | Type   | Description                      |
| ---------------- | ------ | -------------------------------- |
| `search`         | string | Cari berdasarkan nama            |
| `sortBy`         | string | id, created_at, updated_at, name |
| `sort`           | string | asc / desc (default: desc)       |
| `page`           | int    | Halaman (default: 1)             |
| `limit`          | int    | Limit per page (default: 10)     |
| `dateStart`      | string | Filter date range start          |
| `dateEnd`        | string | Filter date range end            |
| `typeCategoryId` | int64  | Filter by type category          |
| `category`       | int64  | Filter by product category       |

**Response**:

```json
{
  "metaData": { "code": 200, "message": "OK" },
  "responseMessage": "GET_SAVED_CREATOR_IMAGES_SUCCESS",
  "data": [
    {
      "id": 1,
      "name": "Template Name",
      "imageUrl": "https://...", // null jika banned/not published/deleted
      "isPublished": true,
      "price": 10000,
      "publisher": {
        "id": "uuid",
        "name": "Creator Name",
        "image": "https://..."
      },
      "typeCategories": [{"id": 1, "name": "Announcement"}],
      "productCategories": [{"id": 1, "name": "Fashion"}],
      "notShowingReason": null, // "CONTENT_IMAGE_BANNED", "CONTENT_IMAGE_CURRENTLY_NOT_PUBLISHED", "CONTENT_IMAGE_DELETED"
      "savedAt": "2026-01-20T10:00:00Z",
      "createdAt": "2026-01-15T10:00:00Z",
      "updatedAt": "2026-01-15T10:00:00Z"
    }
  ],
  "pagination": {...}
}
```

---

### POST /api/creator/business-saved-creator-image/{businessId}

Simpan creator image ke collection business.

**Authentication**: Required  
**Middleware**: `OwnedBusinessMiddleware`

**Request Body**:

```json
{
  "creatorImageId": 123
}
```

**Validasi**:

1. Creator image harus ada (not found → 404)
2. Tidak boleh deleted (→ 400 `CREATOR_IMAGE_DELETED`)
3. Tidak boleh banned (→ 400 `CREATOR_IMAGE_BANNED`)
4. Harus published (→ 400 `CREATOR_IMAGE_NOT_PUBLISHED`)
5. Belum pernah disave (→ 400 `CREATOR_IMAGE_ALREADY_SAVED`)

**Response (201)**:

```json
{
  "metaData": { "code": 200, "message": "OK" },
  "responseMessage": "CREATE_SAVED_CREATOR_IMAGE_SUCCESS",
  "data": {
    "id": 1,
    "businessRootId": 10,
    "creatorImageId": 123,
    "createdAt": "2026-01-20T10:00:00Z"
  }
}
```

---

### DELETE /api/creator/business-saved-creator-image/{businessId}/{creatorImageId}

Hapus saved creator image dari collection business (soft delete).

**Authentication**: Required  
**Middleware**: `OwnedBusinessMiddleware`

**Validasi**:

- Saved record harus ada (→ 404 `SAVED_CREATOR_IMAGE_NOT_FOUND`)

**Response**:

```json
{
  "metaData": { "code": 200, "message": "OK" },
  "responseMessage": "DELETE_SAVED_CREATOR_IMAGE_SUCCESS",
  "data": {
    "id": 1,
    "businessRootId": 10,
    "creatorImageId": 123,
    "createdAt": "2026-01-20T10:00:00Z"
  }
}
```

## 5. Business Logic Details

### NotShowingReason Priority

Ketika menampilkan saved creator images, cek kondisi dalam urutan:

1. **CONTENT_IMAGE_DELETED**: Creator image sudah di-soft-delete
2. **CONTENT_IMAGE_BANNED**: Creator image dibanned
3. **CONTENT_IMAGE_CURRENTLY_NOT_PUBLISHED**: Creator image tidak dipublish

Jika salah satu kondisi terpenuhi:

- `imageUrl` = `null`
- `notShowingReason` = sesuai kondisi

### Soft Delete

Module menggunakan soft delete (`deleted_at`) karena:

1. Ada partial unique index `WHERE deleted_at IS NULL`
2. User bisa re-save creator image yang sama setelah unsave

### Dependency ke CreatorImageService

Module ini **tidak query langsung** ke tabel creator_images untuk validasi. Sebaliknya, memanggil method dari `CreatorImageService`:

```go
detail, err := s.creatorImageSvc.GetCreatorImageDetailById(ctx, input.CreatorImageID)
```

Ini sesuai requirement separation of concerns.

## 6. Error Codes

| Error Code                      | HTTP Code | Condition                     |
| ------------------------------- | --------- | ----------------------------- |
| `CREATOR_IMAGE_NOT_FOUND`       | 404       | Creator image tidak ditemukan |
| `CREATOR_IMAGE_DELETED`         | 400       | Creator image sudah dihapus   |
| `CREATOR_IMAGE_BANNED`          | 400       | Creator image dibanned        |
| `CREATOR_IMAGE_NOT_PUBLISHED`   | 400       | Creator image tidak publish   |
| `CREATOR_IMAGE_ALREADY_SAVED`   | 400       | Sudah pernah disave           |
| `SAVED_CREATOR_IMAGE_NOT_FOUND` | 404       | Saved record tidak ditemukan  |
| `FORBIDDEN`                     | 403       | Bukan member business         |

## 7. Dependencies

- **CreatorImageService**: Untuk validasi creator image (GetCreatorImageDetailById)
- **OwnedBusinessMiddleware**: Untuk validasi akses business member
