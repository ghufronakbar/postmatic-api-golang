# Module Headless.Token

Modul ini bertanggung jawab untuk generate dan validate JWT tokens. Modul ini **headless** (tidak dipanggil via HTTP Handler langsung).

## 1. Project Rules & Dependencies

- **Library**: [`github.com/golang-jwt/jwt/v5`](https://github.com/golang-jwt/jwt)
- **Algorithm**: HS256 (HMAC-SHA256)
- **Headless**: Modul ini hanya dipanggil oleh module lain (internal)
- **Used By**: Auth, Session, Invitation, Profile services

## 2. Directory Structure

```text
internal/module/headless/token/
├── service.go              # TokenMaker struct & constructor
├── access_token.go         # Access token operations
├── refresh_token.go        # Refresh token operations
├── create_account_token.go # Account creation token
└── invitation_token.go     # Member invitation token
```

## 3. Configuration

| Variable                           | Type          | Description                 |
| ---------------------------------- | ------------- | --------------------------- |
| `JWT_ACCESS_TOKEN_SECRET`          | String        | Secret for access tokens    |
| `JWT_ACCESS_TOKEN_EXPIRED`         | time.Duration | TTL (e.g., 15m)             |
| `JWT_REFRESH_TOKEN_SECRET`         | String        | Secret for refresh tokens   |
| `JWT_REFRESH_TOKEN_EXPIRED`        | time.Duration | TTL (e.g., 7d)              |
| `JWT_CREATE_ACCOUNT_TOKEN_SECRET`  | String        | Secret for account creation |
| `JWT_CREATE_ACCOUNT_TOKEN_EXPIRED` | time.Duration | TTL (e.g., 24h)             |
| `JWT_INVITATION_TOKEN_SECRET`      | String        | Secret for invitations      |
| `JWT_INVITATION_TOKEN_EXPIRED`     | time.Duration | TTL (e.g., 7d)              |

## 4. Token Types & Use Cases

| Token Type           | Purpose                              | TTL      |
| -------------------- | ------------------------------------ | -------- |
| **Access Token**     | API authentication                   | 15 min   |
| **Refresh Token**    | Renew access token                   | 7 days   |
| **Create Account**   | Email verification / complete signup | 24 hours |
| **Invitation Token** | Member invitation to business        | 7 days   |

## 5. Service Interface

```go
type TokenMaker interface {
    // Access Token
    GenerateAccessToken(input GenerateAccessTokenInput) (string, error)
    ValidateAccessToken(tokenString string) (*AccessTokenClaims, error)
    AccessDecodeTokenWithoutVerify(tokenString string) (*AccessTokenClaims, error)

    // Refresh Token
    GenerateRefreshToken(input GenerateRefreshTokenInput) (string, error)
    ValidateRefreshToken(tokenString string) (*RefreshTokenClaims, error)

    // Create Account Token
    GenerateCreateAccountToken(input GenerateCreateAccountTokenInput) (string, error)
    ValidateCreateAccountToken(tokenString string) (*CreateAccountTokenClaims, error)

    // Invitation Token
    GenerateInvitationToken(input GenerateInvitationTokenInput) (string, error)
    ValidateInvitationToken(tokenString string) (*InvitationTokenClaims, error)
}
```

## 6. Access Token

Digunakan untuk autentikasi API request.

### Claims Structure

```go
type AccessTokenClaims struct {
    ID       uuid.UUID      `json:"id"`       // Profile ID
    Email    string         `json:"email"`
    Name     string         `json:"name"`
    ImageUrl *string        `json:"imageUrl"`
    Role     entity.AppRole `json:"role"`     // user, admin
    jwt.RegisteredClaims
}
```

### Methods

| Method                           | Description                               |
| -------------------------------- | ----------------------------------------- |
| `GenerateAccessToken`            | Create new access token                   |
| `ValidateAccessToken`            | Validate and parse claims                 |
| `AccessDecodeTokenWithoutVerify` | Decode without validation (untuk logging) |

## 7. Refresh Token

Digunakan untuk mendapatkan access token baru tanpa re-login.

### Claims Structure

```go
type RefreshTokenClaims struct {
    SessionID string `json:"sessionId"` // Session reference
    jwt.RegisteredClaims
}
```

### Methods

| Method                 | Description               |
| ---------------------- | ------------------------- |
| `GenerateRefreshToken` | Create new refresh token  |
| `ValidateRefreshToken` | Validate and parse claims |

## 8. Create Account Token

Token dikirim via email untuk verifikasi akun / complete registration.

### Claims Structure

```go
type CreateAccountTokenClaims struct {
    Email string `json:"email"`
    jwt.RegisteredClaims
}
```

## 9. Invitation Token

Token untuk invite member ke business.

### Claims Structure

```go
type InvitationTokenClaims struct {
    InvitationID int64  `json:"invitationId"`
    Email        string `json:"email"`
    BusinessID   int64  `json:"businessId"`
    jwt.RegisteredClaims
}
```

## 10. Usage Example

```go
// Di router.go
tokenSvc := token.NewTokenMaker(cfg)

// Generate access token (di Auth service)
accessToken, err := tokenSvc.GenerateAccessToken(token.GenerateAccessTokenInput{
    ID:    profile.ID,
    Email: profile.Email,
    Name:  profile.Name,
    Role:  profile.Role,
})

// Validate access token (di Middleware)
claims, err := tokenSvc.ValidateAccessToken(accessToken)
if err != nil {
    // Token invalid atau expired
}
fmt.Println(claims.Email) // user@example.com
```

## 11. Security Best Practices

1. **Separate secrets**: Setiap token type punya secret berbeda
2. **Short-lived access**: Access token hanya 15 menit
3. **Refresh rotation**: Refresh token bisa di-rotate saat renew
4. **HMAC-SHA256**: Algoritma secure untuk signing
5. **No sensitive data**: Jangan simpan password/secrets di claims

## 12. Error Handling

| Error                   | Condition                   |
| ----------------------- | --------------------------- |
| `jwt.ErrTokenExpired`   | Token sudah expired         |
| `jwt.ErrTokenMalformed` | Token format tidak valid    |
| `INVALID_ACCESS_TOKEN`  | Token signature tidak valid |
