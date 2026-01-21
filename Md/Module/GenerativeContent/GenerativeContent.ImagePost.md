# Module GenerativeAI.ImagePost

## User story:

Pengguna yang merupakan member dari sebuah bisnis ingin menghasilkan sebuah gambar berdasarkan knowledge yang ada di dalam bisnis tersebut (Termasuk knowledge tentang business, product, role, ).

Selain dari knowledge, pengguna juga dapat memberikan sebuah input.

## Input

```json
{
  // required
  "mode": "generate",
  "rootBusinessId": 1,
  "ratio": "1:1",
  "productKnowledgeId": 1,
  "appGenerativeImageModelId": 1,
  "numOfImages": 1,
  //   additional
  "advanceGenerate": {
    "businessKnowledge": {
      "name": true,
      "category": true,
      "description": true,
      "location": true,
      "logo": true,
      "uniqueSellingPoint": true,
      "website": true,
      "visionMission": true,
      "colorTone": true
    },
    "productKnowledge": {
      "name": true,
      "category": true,
      "description": true,
      "price": true
    },
    "roleKnowledge": {
      "hashtags": true
    }
  },
  "additionalPrompt": "buat desain lebih kekinian",
  "designStyle": "Modern",
  "category": "Promosi & Diskon",
  "referenceImage": "https://example.com/image.jpg",
  "maskImage": "https://example.com/image.jpg",
  "imageSize": "1K",
  "currentCaption": "caption",
  "rss": {
    "title": "Kemenkop-BPS sinkronisasi data desa guna KDMP, pengentasan kemiskinan",
    "url": "https://www.antaranews.com/berita/5020061/kemenkop-bps-sinkronisasi-data-desa-guna-kdmp-pengentasan-kemiskinan",
    "publishedAt": "2025-08-06T15:18:54.000Z",
    "imageUrl": "https://img.antaranews.com/cache/800x533/2025/08/06/928c77f6-944a-4869-aad5-d64ece325f0f.jpeg",
    // rss.imageUrl string|null
    "summary": "Jakarta (ANTARA) - Kementerian Koperasi (Kemenkop) dan ....",
    "publisher": "antara"
  }
}
```

### Keterangan Input:

- mode (required): enum dari "generate", "regenerate", "rss", "mask"
- numOfImages (required): jumlah gambar yang ingin dihasilkan (max: 4)
- rootBusinessId (required): id dari root business yang mengambil data untuk knowledge ini (dari params, dan dari internal_middleware.OwnedBusiness)
- ratio (required): ratio dari gambar yang ingin dihasilkan, seperti "1:1", "4:3", "16:9"
- productKnowledgeId (required): id dari product knowledge yang ingin dijadikan sebagai referensi. selain itu juga untuk relasi melakukan tracking generate untuk product mana
- appGenerativeImageModelId (required): model yang ingin digunakan untuk menghasilkan gambar, seperti "gpt-image-1", "gemini-3-pro-image-preview", dll (diambil dari service lain)

- additionalPrompt (optional): untuk menambahkan instruksi selain dari prompt builder dari aplikasi
- designStyle (optional): seperti "Minimalist", "Vintage", "Modern", ataupun string kosong (default), ini diisi dari sisi client saja tanpa validasi enum
- category (optional): seperti "Promosi & Diskon", "Product Showcase", "Product Catalog", ataupun string kosong (default), ini diisi dari sisi client saja tanpa validasi enum
- referenceImage (optional): berupa url dari gambar yang ingin dijadikan sebagai referensi terhadap gambar yang akan dihasilkan
- maskImage (optional): berupa url dari gambar yang ingin dijadikan sebagai mask terhadap gambar yang akan dihasilkan
- imageSize (optional): ukuran gambar yang dihasilkan, seperti "1K", "2K", "4K" berdasarkan model yang dipilih (dapat dilihat di migrations seed untuk model untuk referensi)
- currentCaption (optional): caption umumnya akan diisi oleh aplikasi, namun jika user melakukan regenerate, yang dimana caption sebelumnya sudah ada dan tidak perlu buat caption lagi, maka caption harus diisi.
- advanceGenerate (optional): untuk memfilter apa saja yang digunakan pada promptbuilder
- model text dipilih berdasarkan provider yang sama dengan model image (ambil 1 yang aktif)

Note:

- Untuk referensi model dapat dilihat "migrations/20260117152211_seed_app_generative_image_models_table.sql" dan "migrations/20260117160526_seed_app_generative_text_models_table.sql"
- jika "mode":"generate" maka "advanceGenerate" wajib diisi
- jika "mode":"regenerate" maka "referenceImage" "advanceGenerate" wajib diisi
- jika "mode":"rss" maka "rss" "advanceGenerate" wajib diisi
- jika "mode":"mask" maka "maskImage" "referenceImage" "prompt" wajib diisi
- Untuk business_roles tetap gunakan semua kecuali hashtags.
- Keseluruhan knowledge dari "business_roles" digunakan untuk generate caption
- Hashtags digunakan untuk generate caption
- Jika sudah ada currentCaption tidak perlu generate caption

## Logic

User melakukan request -> Validasi data dan masukkan ke database -> Masukkan queue -> user mendapatkan response jika berhasil (masuk queue)

Di background -> Di callback untuk queue lakukan generate lalu setelah itu update di database untuk job tersebut

## Before Execute [IMPORTANT]

Karena ini adalah module yang besar. saya ingin anda membagi menjadi beberapa.

Expected file:

- "internal/module/generative_content/image_post/handler/handler.go"
  Sebagai HTTP handler
- "internal/module/generative_content/image_post/service/service.go"
  Sebagai contract dan logic service utama untuk validasi di sisi service, seperti validasi input conditional berdasarkan yang sudah dijelaskan.
  validasi apakah business exist dan product dari business itu exist. validasi mengenai advance generate (contoh: businessKnowledge.website = true, tetapi di table business_knowledges tidak ada, maka return error), lalu jika token yang tersedia, jika tidak tersedia return error. jika validasi sudah semua maka lanjut ke logic utama.
- "internal/module/generative_content/image_post/service/common.go" untuk melihat history generate dan paginated juga seperti implementasi pada service lain
- "internal/module/generative_content/image_post/service/generate.go" untuk mode generate
- "internal/module/generative_content/image_post/service/regenerate.go" untuk mode regenerate
- "internal/module/generative_content/image_post/service/rss.go" untuk mode rss
- "internal/module/generative_content/image_post/service/mask.go" untuk mode mask
- "internal/module/generative_content/image_post/service/prompt_builder.go" yang methodnya akan return string untuk promptnya berdasarkan input yang available
- "internal/module/generative_content/image_post/service/caption.go" untuk membuat caption
- "internal/module/generative_content/image_post/service/dto.go" untuk membuat struct input
- "internal/module/generative_content/image_post/service/filter.go" untuk membuat struct filter
- "internal/module/generative_content/image_post/service/viewmodel.go" untuk membuat struct response/return value dari service
- "internal/repository/queries/generated_image_post.sql" tulis sqlc
- "internal/repository/queries/generated_image_post_item.sql" tulis sqlc
- "internal/repository/queries/generated_image_post_caption.sql" tulis sqlc

Untuk query yang tidak berhubungan, ambil dari service lain (atau buat jika belum ada)

Expected Injection:

- App.GenerativeImageModel "internal/module/app/generative_image_model/service"
- Business.BusinessKnowledge "internal/module/business/business_knowledge/service"
- Business.BusinessProduct "internal/module/business/business_product/service"
- Business.BusinessRole "internal/module/business/business_role/service"
- GenerativeToken.ImageToken "internal/module/generative_token/image_token/service"

Karena menggunakan queue buat juga untuk queue yang berada di folder
"internal/module/headless/queue/"

Expected file:

- "internal/module/headless/queue/generative_content_image_post.go"

Note:
Sesuaikan juga untuk producer/consumernya

---

## TODO Implementation Phases

### Phase 1: Foundation & mode=generate (CURRENT)

- [x] Migration schema created
- [ ] SQL Queries (`generated_image_post.sql`, `generated_image_post_item.sql`, `generated_image_post_caption.sql`)
- [ ] Service DTOs (`dto.go`, `viewmodel.go`, `filter.go`)
- [ ] Service Interface (`service.go`) - dengan semua method signatures
- [ ] Common Methods (`common.go`) - GetAllImagePosts dengan pagination
- [ ] Prompt Builder (`prompt_builder.go`) - BuildImagePrompt, BuildCaptionPrompt
- [ ] Caption Generator (`caption.go`) - GenerateCaption via text model
- [ ] Generate Mode (`generate.go`) - FULL IMPLEMENTATION
- [ ] Stub Files (throw error `METHOD_NOT_IMPLEMENTED`):
  - [ ] `regenerate.go`
  - [ ] `rss.go`
  - [ ] `mask.go`
- [ ] Queue Producer & Handler (`generative_content_image_post.go`)
- [ ] HTTP Handler (`handler.go`)
- [ ] Router Wiring

### Phase 2: mode=regenerate (FUTURE)

- [ ] Implement `regenerate.go`
- [ ] Update `prompt_builder.go` for regenerate context

### Phase 3: mode=rss (FUTURE)

- [ ] Implement `rss.go`
- [ ] Update `prompt_builder.go` for RSS context

### Phase 4: mode=mask (FUTURE)

- [ ] Implement `mask.go`
- [ ] Update `prompt_builder.go` for mask context

---

## Detailed Requirements for Phase 1

### Validation Rules for mode=generate

1. **Required Fields**: mode, numOfImages, ratio, productKnowledgeId, appGenerativeImageModelId
2. **advanceGenerate WAJIB** untuk mode=generate
3. **numOfImages**: max 4
4. **Business Validation**:
   - Business harus exist (dari OwnedBusinessMiddleware)
   - Product harus milik business tersebut
5. **Knowledge Validation**:
   - Jika `advBk.website = true` tapi business_knowledge.website_url kosong → error
   - Jika `advBk.logo = true` tapi business_knowledge.primary_logo_url kosong → error
   - dst untuk field lainnya
6. **Token Validation**:
   - Cek availableToken >= numOfImages (atau token cost estimation)
   - Jika tidak cukup → error `INSUFFICIENT_TOKEN`
7. **Model Validation**:
   - appGenerativeImageModelId harus exist dan active
   - Text model dipilih otomatis berdasarkan provider yang sama

### Flow for mode=generate

```
1. Handler menerima request
2. Service.CreateImagePost(input):
   a. Validate conditional fields berdasarkan mode
   b. Validate business ownership (dari middleware)
   c. Validate product belongs to business
   d. Validate image model exists & active
   e. Get text model dengan provider sama (untuk caption)
   f. Validate advanceGenerate fields vs actual knowledge data
   g. Check token availability
   h. Insert ke generated_image_posts dengan status=pending
   i. Enqueue job ke queue
   j. Return response dengan ID
3. Queue Worker:
   a. Update status = processing
   b. Build image prompt (prompt_builder.go)
   c. Call AI service untuk generate images
   d. Upload images ke storage
   e. Insert generated_image_post_items
   f. Jika tidak ada currentCaption:
      - Build caption prompt
      - Call AI text service
      - Insert generated_image_post_captions
   g. Update status = success (atau failed jika error)
   h. Deduct token
```
