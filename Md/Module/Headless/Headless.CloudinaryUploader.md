# Module Headless.CloudinaryUploader

Modul ini bertanggung jawab untuk upload gambar ke Cloudinary. Modul ini **headless** (tidak dipanggil via HTTP Handler langsung).

## 1. Project Rules & Dependencies

- **Library**: [`github.com/cloudinary/cloudinary-go/v2`](https://github.com/cloudinary/cloudinary-go)
- **Headless**: Modul ini hanya dipanggil oleh module lain (internal), tidak ada HTTP Handler/Controller di dalamnya.
- **Used By**: `App.ImageUploader` service

## 2. Directory Structure

```text
internal/module/headless/cloudinary_uploader/
├── service.go     # Service implementation
└── viewmodel.go   # Output DTOs / Response types
```

## 3. Configuration

File: `config/cloudinary.go`

| Variable                | Type   | Description                |
| ----------------------- | ------ | -------------------------- |
| `CLOUDINARY_CLOUD_NAME` | String | Cloud name dari Cloudinary |
| `CLOUDINARY_API_KEY`    | String | API key                    |
| `CLOUDINARY_API_SECRET` | String | API secret                 |

```go
cld := config.ConnectCloudinary(cfg)
```

## 4. Service Interface

```go
type CloudinaryUploaderService interface {
    // Upload single image to Cloudinary
    UploadSingleImage(ctx context.Context, file io.Reader) (*CloudinaryUploadSingleImageResponse, error)
}
```

## 5. Method: UploadSingleImage

Upload single image file ke Cloudinary.

### Input

| Parameter | Type            | Description       |
| --------- | --------------- | ----------------- |
| `ctx`     | context.Context | Request context   |
| `file`    | io.Reader       | Image file stream |

### Business Logic

```
1. Call Cloudinary Upload API dengan params:
   - Folder: cfg.APP_NAME (e.g., "postmatic")
   - Tags: ["source:api", "type:image"]

2. Handle response:
   ├── Error → Return errs.NewInternalServerError
   ├── Result nil → Return "CLOUDINARY_RETURNED_EMPTY_URL"
   └── Success → Return response dengan URL

3. Response contains:
   - PublicId: Cloudinary public ID (untuk delete/update)
   - ImageUrl: HTTPS URL gambar (SecureURL)
   - Format: Format file (jpeg, png, webp, etc.)
```

### Output DTO

```go
type CloudinaryUploadSingleImageResponse struct {
    PublicId string `json:"publicId"` // Cloudinary public ID
    ImageUrl string `json:"imageUrl"` // HTTPS URL
    Format   string `json:"format"`   // jpeg, png, etc.
}
```

## 6. Usage Example

```go
// Di router.go
cld := config.ConnectCloudinary(cfg)
cldSvc := cloudinary_uploader.NewService(cfg, cld)

// Di service lain (misal ImageUploaderService)
file, _ := os.Open("image.jpg")
defer file.Close()

result, err := cldSvc.UploadSingleImage(ctx, file)
if err != nil {
    return err
}
fmt.Println(result.ImageUrl) // https://res.cloudinary.com/.../image.jpg
```

## 7. Error Handling

| Error                           | Condition                          |
| ------------------------------- | ---------------------------------- |
| `InternalServerError`           | Cloudinary API call failed         |
| `CLOUDINARY_RETURNED_EMPTY_URL` | Upload success but no URL returned |

## 8. Design Decisions

### Kenapa Simpan di Cloudinary?

1. **Auto-optimization**: Cloudinary otomatis optimize dan resize gambar
2. **CDN**: Gambar di-serve via CDN global
3. **Transformations**: Support on-the-fly image transformations
4. **No storage management**: Tidak perlu kelola S3 bucket/lifecycle

### Kenapa Headless?

1. Module ini tidak perlu HTTP endpoint sendiri
2. Dipanggil oleh module lain (e.g., ImageUploader)
3. Separation of concerns - upload logic terpisah dari business logic
