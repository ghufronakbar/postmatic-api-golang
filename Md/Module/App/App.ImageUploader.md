# Module App.ImageUploader

Module untuk upload gambar ke storage (Cloudinary/S3).

## Directory

- `internal/module/app/image_uploader/handler/*`
- `internal/module/app/image_uploader/service/*`

---

## Endpoints

### POST /api/app/image-uploader/upload-single-image

**Fungsi**: Upload single image file ke storage.

**Auth**: All Allowed

**Content-Type**: `multipart/form-data`

**Body**:
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| image | file | Yes | Image file (max 10MB, must be image/\*) |

**Response**:

```json
{
  "imageUrl": "https://...",
  "publicId": "...",
  "hashkey": "..."
}
```

**Business Logic**:

1. Limit upload size to 10MB
2. Parse multipart form
3. Validate file is an image (content-type: image/\*)
4. Upload to Cloudinary with 30s timeout
5. Return URL, publicId, and hashkey

---

### POST /api/app/image-uploader/presign-upload-image

**Fungsi**: Get presigned URL untuk upload image ke S3.

**Auth**: All Allowed

**Body**:

```json
{
  "fileName": "image.jpg",
  "contentType": "image/jpeg"
}
```

**Response**:

```json
{
  "presignedUrl": "https://s3...",
  "publicUrl": "https://...",
  ...
}
```

---

## Service Methods

| Method               | Description                 |
| -------------------- | --------------------------- |
| `UploadSingleImage`  | Upload file to Cloudinary   |
| `PresignUploadImage` | Get presigned S3 upload URL |
