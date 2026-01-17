# Module App.PaymentMethod

## Endpoint

### GET /api/app/payment-method/type (all allowed)

- Response: `string[]` (enum values: `["bank", "ewallet"]`)
- No pagination
- Used to get available payment method types

### GET /api/app/payment-method (all allowed)

- Response Paginated
- Admin: shows all (active + inactive)
- User: shows only active

### POST /api/app/payment-method (admin only)

- Response Single
- Create new payment method

### GET /api/app/payment-method/:id (all allowed)

- Response Single
- Admin: shows even if inactive
- User: shows only if active

### GET /api/app/payment-method/:code/code (all allowed)

- Response Single
- Admin: shows even if inactive
- User: shows only if active

### PUT /api/app/payment-method/:id (admin only)

- Response Single
- Update payment method

### DELETE /api/app/payment-method/:id (admin only)

- Response Single
- Soft delete payment method

## Note:

- All Allowed: Admin and User (still need use middleware (not public))
- Admin Only: Admin only
- Jangan query is_deleted, pastikan notfound jika is_deleted
- Untuk get all, pastikan ada pagination yang sesuai dengan project rules dan code-code saya lainnya
- Untuk get, pastikan validasi role dari middleware, jika admin maka query yang active maupun tidak, namun jika user maka query yang active saja
- Code selalu uppercase (normalized)
