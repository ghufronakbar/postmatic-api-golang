# Module Business.BusinessMember

Module untuk mengelola member bisnis (invite, edit, remove, invitation flow).

## Directory

- `internal/module/business/business_member/handler/*`
- `internal/module/business/business_member/service/*`

---

## Endpoints

### GET /api/business-member/{businessId}

**Fungsi**: Mendapatkan daftar member business dengan pagination.

**Auth**: All Allowed + OwnedBusinessMiddleware

**Query Params**:
| Param | Type | Required | Description |
|-------|------|----------|-------------|
| category | string | No | Filter: `verified` atau `unverified` |

**Response**: List of business members

---

### POST /api/business-member/{businessId}

**Fungsi**: Invite member baru ke business.

**Auth**: All Allowed + OwnedBusinessMiddleware

**Body**:

```json
{
  "email": "member@example.com",
  "role": "member",
  ...
}
```

**Response**: Invitation result

---

### PUT /api/business-member/{businessId}

**Fungsi**: Edit member (role, etc).

**Auth**: Owner Only + OwnedBusinessMiddleware

**Body**:

```json
{
  "memberId": 123,
  "role": "admin",
  ...
}
```

**Response**: Updated member info

---

### POST /api/business-member/{businessId}/resend-invitation

**Fungsi**: Resend invitation email ke member yang belum verify.

**Auth**: Owner Only + OwnedBusinessMiddleware

**Body**:

```json
{
  "memberId": 123
}
```

---

### DELETE /api/business-member/{businessId}/{memberId}

**Fungsi**: Remove member dari business.

**Auth**: Owner Only + OwnedBusinessMiddleware

---

### GET /api/business-member/{businessId}/{token}/verify

**Fungsi**: Verify invitation token (public endpoint).

**Auth**: None (public)

**Response**: Invitation verification info

---

### POST /api/business-member/{businessId}/{token}/answer

**Fungsi**: Accept/reject invitation.

**Auth**: None (public)

**Body**:

```json
{
  "answer": "accept" // or "reject"
}
```

---

## Service Methods

| Method                               | Description                   |
| ------------------------------------ | ----------------------------- |
| `GetBusinessMembersByBusinessRootID` | List members with filter      |
| `InviteBusinessMember`               | Send invitation email         |
| `EditMember`                         | Edit member role (owner only) |
| `ResendMemberInvitation`             | Resend invitation email       |
| `RemoveBusinessMember`               | Remove member (owner only)    |
| `VerifyMemberInvitation`             | Verify invitation token       |
| `AnswerMemberInvitation`             | Accept/reject invitation      |
