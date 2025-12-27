# Project Structure

_Generated: 2025-12-27 21:16:58_

Root: `.`

```text
.
- .build/
- cmd/
  - api/
  - seed/
    - creator_image_seed/
    - rss_seed/
- config/
- internal/
  - http/
    - handler/
      - account_handler/
      - app_handler/
      - business_handler/
      - creator_handler/
    - middleware/
  - module/
    - account/
      - auth/
      - google_oauth/
      - profile/
      - session/
    - app/
      - category_creator_image/
      - image_uploader/
      - rss/
      - timezone/
    - business/
      - business_information/
      - business_knowledge/
      - business_product/
      - business_role/
      - business_rss_subscription/
      - business_timezone_pref/
    - creator/
      - creator_image/
    - headless/
      - cloudinary_uploader/
      - mailer/
        - templates/
      - s3_uploader/
  - repository/
    - entity/
    - queries/
    - redis/
      - email_limiter_repository/
      - owned_business_repository/
      - session_repository/
- migrations/
- pkg/
  - errs/
  - filter/
  - hash/
  - pagination/
  - response/
  - token/
  - utils/
- seeders/
```

