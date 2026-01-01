# Project Structure

_Generated: 2026-01-02 00:53:16_

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
      - business_image_content/
      - business_information/
      - business_knowledge/
      - business_member/
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
      - queue/
      - s3_uploader/
      - token/
  - repository/
    - entity/
    - queries/
    - redis/
      - email_limiter_repository/
      - invitation_limiter_repository/
      - owned_business_repository/
      - session_repository/
- migrations/
- pkg/
  - errs/
  - filter/
  - hash/
  - logger/
  - pagination/
  - response/
  - token/
  - utils/
- seeders/
```

