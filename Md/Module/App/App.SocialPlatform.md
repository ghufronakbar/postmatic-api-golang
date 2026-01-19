# Module App.SocialPlatform

Module untuk mengelola social platform yang tersedia di aplikasi.

## Dependency

- None

## Directory

- `internal/module/app/social_platform/handler/*`
- `internal/module/app/social_platform/service/*`
- `internal/repository/queries/app_social_platform.sql`

---

## Endpoints

### GET /api/app/social-platform

**Fungsi**: Menampilkan semua social platforms dengan pagination.

**Auth**: All Allowed

**Note**:

- Admin dapat melihat semua data (termasuk inactive)
- User biasa hanya melihat data dengan `isActive = true`

**Query Params**:
| Param | Type | Required | Description |
|-------|------|----------|-------------|
| search | string | No | Search by name or platform code |
| sortBy | string | No | Sort by field (id, name, platform_code, is_active). Default: id |
| sort | string | No | Sort direction (asc, desc). Default: desc |
| page | int | No | Page number. Default: 1 |
| limit | int | No | Items per page. Default: 10 |

**Response**: List of SocialPlatformResponse with pagination

---

### GET /api/app/social-platform/platform-code

**Fungsi**: Mendapatkan list semua platform code yang valid (dari enum).

**Auth**: All Allowed

**Response**:

```json
[
  "linked_in",
  "facebook_page",
  "instagram_business",
  "whatsapp_business",
  "tiktok",
  "youtube",
  "twitter",
  "pinterest"
]
```

---

### POST /api/app/social-platform

**Fungsi**: Membuat social platform baru.

**Auth**: Admin Only

**Body**:

```json
{
  "platformCode": "linked_in",
  "logo": "https://example.com/logo.png",
  "name": "LinkedIn",
  "hint": "LinkedIn hint",
  "isActive": true
}
```

**Note**:

- `platformCode` harus unik (tidak boleh duplicate dengan yang sudah ada)
- `platformCode` harus valid enum value
- Change akan di-track di `app_social_platform_changes`

**Response**: SocialPlatformResponse

---

### PUT /api/app/social-platform/{id}

**Fungsi**: Update social platform.

**Auth**: Admin Only

**Body**: Sama seperti POST

**Note**:

- Jika `platformCode` diubah, harus tetap unik
- Change akan di-track di `app_social_platform_changes`

**Response**: SocialPlatformResponse

---

### DELETE /api/app/social-platform/{id}

**Fungsi**: Soft delete social platform.

**Auth**: Admin Only

**Note**:

- Soft delete (set `deleted_at`)
- Change akan di-track di `app_social_platform_changes`

**Response**: SocialPlatformResponse

---

## Service Methods

| Method                  | Description                                                 |
| ----------------------- | ----------------------------------------------------------- |
| `GetPlatformCodes()`    | List valid platform code enum values                        |
| `GetAll(filter)`        | List with pagination, admin sees all, user sees active only |
| `GetById(id)`           | Get by ID                                                   |
| `Create(input)`         | Create with duplicate check and change tracking             |
| `Update(id, input)`     | Update with duplicate check and change tracking             |
| `Delete(id, profileID)` | Soft delete with change tracking                            |

---

## SQL Queries

| Query                                | Description                                |
| ------------------------------------ | ------------------------------------------ |
| `CreateAppSocialPlatform`            | Create new record                          |
| `GetAppSocialPlatformById`           | Get by ID                                  |
| `GetAppSocialPlatformByPlatformCode` | Get by platform_code (for duplicate check) |
| `GetAllAppSocialPlatforms`           | List with filter and pagination            |
| `CountAllAppSocialPlatforms`         | Count with filter                          |
| `UpdateAppSocialPlatform`            | Update record                              |
| `DeleteAppSocialPlatform`            | Soft delete                                |
| `CreateAppSocialPlatformChange`      | Track changes                              |

---

## Database Schema

- `app_social_platforms` - Main table with platform info
- `app_social_platform_changes` - Change log table (before/after snapshot)

**Enum**: `social_platform_type`

- linked_in, facebook_page, instagram_business, whatsapp_business, tiktok, youtube, twitter, pinterest
