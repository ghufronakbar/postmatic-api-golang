# Module Business.BusinessRssSubscription

Module untuk mengelola RSS subscription bisnis.

## Directory

- `internal/module/business/business_rss_subscription/handler/*`
- `internal/module/business/business_rss_subscription/service/*`

---

## Endpoints

### GET /api/business-rss-subscription/{businessId}

**Fungsi**: Mendapatkan daftar RSS subscription dengan pagination.

**Auth**: All Allowed + OwnedBusinessMiddleware

**Query Params**:
| Param | Type | Required | Description |
|-------|------|----------|-------------|
| search | string | No | Search |
| sortBy | string | No | Sort by field |
| sort | string | No | Sort direction |
| page | int | No | Page number |
| limit | int | No | Items per page |

**Response**: List of RSS subscriptions with pagination

---

### POST /api/business-rss-subscription/{businessId}

**Fungsi**: Create RSS subscription baru.

**Auth**: All Allowed + OwnedBusinessMiddleware

**Body**:

```json
{
  "feedId": 123,
  ...
}
```

**Response**: Created subscription

---

### PUT /api/business-rss-subscription/{businessId}/{businessRssSubscriptionId}

**Fungsi**: Update RSS subscription.

**Auth**: All Allowed + OwnedBusinessMiddleware

**Body**: Same as POST

**Response**: Updated subscription

---

### DELETE /api/business-rss-subscription/{businessId}/{businessRssSubscriptionId}

**Fungsi**: Hard delete RSS subscription.

**Auth**: All Allowed + OwnedBusinessMiddleware

**Response**: Deleted subscription info

---

## Service Methods

| Method                                       | Description                    |
| -------------------------------------------- | ------------------------------ |
| `GetBusinessRssSubscriptionByBusinessRootID` | List subscriptions with filter |
| `CreateBusinessRssSubscription`              | Create new subscription        |
| `UpdateBusinessRssSubscription`              | Update subscription            |
| `DeleteBusinessRssSubscription`              | Hard delete subscription       |
