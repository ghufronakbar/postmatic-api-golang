# Module App.GenerativeImageModel

## Endpoint

### GET /api/app/generative-image-model/provider (all allowed)

- Response: `string[]` (enum values: `["openai", "google"]`)
- No pagination
- Used to get available generative image model providers

### GET /api/app/generative-image-model (all allowed)

- Response Paginated
- Admin: shows all (active + inactive)
- User: shows only active

### POST /api/app/generative-image-model (admin only)

- Response Single
- Create new generative image model

### GET /api/app/generative-image-model/:id (all allowed)

- Response Single
- Admin: shows even if inactive
- User: shows only if active

### GET /api/app/generative-image-model/:model/model (all allowed)

- Response Single
- Admin: shows even if inactive
- User: shows only if active

### PUT /api/app/generative-image-model/:id (admin only)

- Response Single
- Update generative image model

### DELETE /api/app/generative-image-model/:id (admin only)

- Response Single
- Soft delete generative image model

## Note:

- All Allowed: Admin and User (still need use middleware (not public))
- Admin Only: Admin only
- Jangan query is_deleted, pastikan notfound jika is_deleted
- Untuk get all, pastikan ada pagination yang sesuai dengan project rules dan code-code saya lainnya
- Untuk get, pastikan validasi role dari middleware, jika admin maka query yang active maupun tidak, namun jika user maka query yang active saja
- Validasi ratio format: "N:M" where N and M are positive integers
