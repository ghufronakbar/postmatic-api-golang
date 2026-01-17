# Module App.GenerativeTextModel

## Endpoint

### GET /api/app/generative-text-model/provider (all allowed)

- Response: `string[]` (enum values: `["openai", "google"]`)
- No pagination
- Used to get available generative text model providers

### GET /api/app/generative-text-model (all allowed)

- Response Paginated
- Admin: shows all (active + inactive)
- User: shows only active

### POST /api/app/generative-text-model (admin only)

- Response Single
- Create new generative text model

### GET /api/app/generative-text-model/:id (all allowed)

- Response Single
- Admin: shows even if inactive
- User: shows only if active

### GET /api/app/generative-text-model/:model/model (all allowed)

- Response Single
- Admin: shows even if inactive
- User: shows only if active

### PUT /api/app/generative-text-model/:id (admin only)

- Response Single
- Update generative text model

### DELETE /api/app/generative-text-model/:id (admin only)

- Response Single
- Soft delete generative text model

## Note:

- All Allowed: Admin and User (still need use middleware (not public))
- Admin Only: Admin only
- Jangan query is_deleted, pastikan notfound jika is_deleted
- Untuk get all, pastikan ada pagination yang sesuai dengan project rules dan code-code saya lainnya
- Untuk get, pastikan validasi role dari middleware, jika admin maka query yang active maupun tidak, namun jika user maka query yang active saja
- Provider harus sesuai dengan enum (`openai`, `google`)
