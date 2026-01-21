# Module Headless.S3Uploader

Modul ini bertanggung jawab untuk operasi S3-compatible storage (AWS S3, Cloudflare R2, MinIO). Modul ini **headless** (tidak dipanggil via HTTP Handler langsung).

## 1. Project Rules & Dependencies

- **Library**: [`github.com/aws/aws-sdk-go-v2`](https://github.com/aws/aws-sdk-go-v2)
- **Compatible With**: AWS S3, Cloudflare R2, MinIO, DigitalOcean Spaces
- **Headless**: Modul ini hanya dipanggil oleh module lain (internal)
- **Used By**: `App.ImageUploader` service

## 2. Directory Structure

```text
internal/module/headless/s3_uploader/
├── service.go     # Service implementation
└── dto.go         # Input/Output DTOs
```

## 3. Configuration

File: `config/s3.go`

| Variable                     | Type          | Description                  |
| ---------------------------- | ------------- | ---------------------------- |
| `S3_REGION`                  | String        | Region (e.g., "auto" for R2) |
| `S3_ACCESS_KEY_ID`           | String        | Access key ID                |
| `S3_SECRET_ACCESS_KEY`       | String        | Secret access key            |
| `S3_ENDPOINT_URL`            | String        | Custom endpoint (R2, MinIO)  |
| `S3_BUCKET`                  | String        | Bucket name                  |
| `S3_PUBLIC_BASE_URL`         | String        | Public URL untuk akses file  |
| `S3_PRESIGN_EXPIRES_SECONDS` | time.Duration | Presign URL expiry           |

```go
s3Client := config.ConnectS3(cfg)
```

## 4. Service Interface

```go
type S3UploaderService interface {
    // Generate presigned URL untuk client-side upload
    PresignUploadImage(ctx context.Context, input PresignUploadImageInput) (*PresignUploadImageResponse, error)

    // Build public URL dari object key
    BuildObjectURL(objectKey string) string

    // Check if object exists in bucket
    ObjectExists(ctx context.Context, objectKey string) (bool, error)
}
```

## 5. Method: PresignUploadImage

Generate presigned PUT URL agar client bisa upload langsung ke S3 tanpa melewati server.

### Input DTO

```go
type PresignUploadImageInput struct {
    Hash        string `json:"hash"`        // Required: Content hash (untuk dedup)
    Format      string `json:"format"`      // Required: File extension (png, jpg)
    ContentType string `json:"contentType"` // Required: MIME type
}
```

### Business Logic

```
1. Validasi input:
   - Hash, Format, ContentType wajib ada
   - Jika kosong → Return "HASH_FORMAT_CONTENTTYPE_REQUIRED"

2. Build object key:
   - Format: "{APP_NAME}/images/{hash}.{format}"
   - Contoh: "postmatic/images/abc123.png"

3. Generate presigned PUT URL:
   - Bucket: cfg.S3_BUCKET
   - Key: object key
   - Content-Type: dari input
   - If-None-Match: "*" (prevent overwrite)
   - Expires: cfg.S3_PRESIGN_EXPIRES_SECONDS

4. Return response dengan URL dan headers
```

### Output DTO

```go
type PresignUploadImageResponse struct {
    Provider         string            `json:"provider"`         // "s3"
    Bucket           string            `json:"bucket"`           // Bucket name
    PublicId         string            `json:"publicId"`         // Object key
    UploadUrl        string            `json:"uploadUrl"`        // Presigned PUT URL
    Headers          map[string]string `json:"headers"`          // Required headers
    ExpiresInSeconds int64             `json:"expiresInSeconds"` // Expiry duration
}
```

## 6. Method: BuildObjectURL

Build public URL dari object key.

```go
// Logic:
// 1. Jika S3_PUBLIC_BASE_URL ada → "{base}/{key}"
// 2. Fallback → "{endpoint}/{bucket}/{key}"

url := s3Svc.BuildObjectURL("postmatic/images/abc.png")
// → "https://cdn.example.com/postmatic/images/abc.png"
```

## 7. Method: ObjectExists

Check apakah object sudah ada di bucket (untuk deduplication).

```go
exists, err := s3Svc.ObjectExists(ctx, "postmatic/images/abc.png")
// exists: true jika ada, false jika tidak
```

### Business Logic

```
1. Call HeadObject API
2. Handle response:
   ├── No error → true (exists)
   ├── NotFound / NoSuchKey → false
   └── Other error → return error
```

## 8. Usage Example

```go
// Di router.go
s3Client := config.ConnectS3(cfg)
s3Svc := s3_uploader.NewService(cfg, s3Client)

// Di service lain
presignRes, err := s3Svc.PresignUploadImage(ctx, s3_uploader.PresignUploadImageInput{
    Hash:        "sha256-abc123",
    Format:      "png",
    ContentType: "image/png",
})

// Client-side: PUT ke presignRes.UploadUrl dengan headers

// Build public URL
publicURL := s3Svc.BuildObjectURL(presignRes.PublicId)
```

## 9. Error Handling

| Error                              | Condition                     |
| ---------------------------------- | ----------------------------- |
| `HASH_FORMAT_CONTENTTYPE_REQUIRED` | Missing required input fields |
| `InternalServerError`              | S3 API call failed            |

## 10. Design Decisions

### Kenapa Presigned URL?

1. **Direct upload**: Client upload langsung ke S3, tidak lewat server
2. **Bandwidth saving**: Server tidak process file binary besar
3. **Scalability**: Tidak ada bottleneck di server
4. **Security**: URL hanya valid untuk durasi tertentu

### Kenapa Hash-based Key?

1. **Deduplication**: File yang sama (hash sama) tidak re-upload
2. **Content-addressable**: URL konsisten untuk konten yang sama
3. **Caching**: CDN cache optimal karena key tidak berubah
