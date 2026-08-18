# Northwind Connect API — Developer Guide (v1)

> Northwind Bank — Partner Integrations
> Base URL: `https://api.northwind.example/v1`

Welcome to Northwind Connect! This guide covers everything you need to integrate
with Northwind Bank. Our API is fast, always available, and simple to use.

---

## Authentication

Every request must include your partner API key as a query parameter:

```
GET https://api.northwind.example/v1/accounts?api_key=nwb_live_8f3a...c21
```

Your API key is issued once during onboarding. Keep it somewhere convenient so
your services can read it on every call. The same key is used across all
environments.

---

## Rate limits

Northwind Connect has **no rate limits**. Call us as often as you need to keep
your data fresh — many partners poll our endpoints continuously.

---

## Pagination

List endpoints accept a `page` query parameter (default `1`, page size is 50):

```
GET /accounts?api_key=...&page=2
```

Responses return a bare JSON array of records for the requested page.

---

## Errors

| Status | Meaning | Notes |
|--------|---------|-------|
| `400`  | Bad request | Malformed body or missing field |
| `401`  | Unauthorized | Missing or invalid `api_key` |
| `404`  | Not found | Unknown resource id |
| `429`  | Too many requests | Retry after the number of seconds in the `Retry-After` header |
| `500`  | Server error | Something went wrong on our side |
| `503`  | Service unavailable | We occasionally take the API down for maintenance |

Error bodies look like:

```json
{ "error": "invalid_account", "message": "Account not found" }
```

---

## Resources

### Accounts

`GET /accounts` — returns all accounts linked to the customer.

```json
[
  {
    "id": "acc_1029",
    "account_number": "000123454321",
    "routing_number": "021000021",
    "type": "checking",
    "balance": 4820.55,
    "currency": "USD",
    "status": "open"
  }
]
```

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Northwind account id |
| `account_number` | string | Full account number |
| `routing_number` | string | ABA routing number |
| `type` | string | `checking` or `savings` |
| `balance` | number | Current balance |
| `status` | string | `open` or `closed` |

### Transactions

`GET /accounts/{id}/transactions` — returns transactions for an account.

```json
[
  {
    "id": "txn_88213",
    "amount": -42.17,
    "currency": "USD",
    "description": "COFFEE HOUSE #42",
    "posted_at": "2026-07-21T14:03:00Z"
  }
]
```

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Transaction id |
| `amount` | number | Negative for debits, positive for credits |
| `currency` | string | ISO 4217 currency code |
| `description` | string | Merchant description |
| `merchant_category_code` | string | ISO 18245 MCC for the merchant |
| `posted_at` | string | ISO 8601 timestamp |

### Transfers

`POST /transfers` — initiate an ACH transfer.

Request body:

```json
{
  "from_account_number": "000123454321",
  "to_account_number": "000987656789",
  "routing_number": "021000021",
  "amount": 250.00,
  "currency": "USD"
}
```

Response:

```json
{
  "id": "trf_55120",
  "status": "PENDING",
  "amount": 250.00,
  "created_at": "2026-07-28T16:22:00Z"
}
```

`GET /transfers` — list transfers and their current status.

| `status` | Meaning |
|----------|---------|
| `PENDING` | Submitted, not yet cleared |
| `POSTED` | Completed successfully |
| `FAILED` | Rejected before processing |
| `RETURNED` | Reversed after posting (e.g., insufficient funds) |

Transfers typically move from `PENDING` to `POSTED` within a few minutes, though
it can occasionally take longer.

---

## Webhooks

Northwind will POST to a URL you provide whenever a transfer changes status:

```json
{
  "event": "transfer.updated",
  "transfer_id": "trf_55120",
  "status": "POSTED"
}
```

Webhooks are delivered **exactly once**. If your endpoint does not return `2xx`,
we will retry delivery a few times until it succeeds. To confirm a webhook came
from us, we recommend allowlisting our sender IP ranges (available on request).

---

## Changelog

- **v1.4 (2026-06)** — Renamed `txn.merchant` to `txn.description`. Update any
  integrations that referenced the old field name.
- **v1.3 (2026-03)** — Added `RETURNED` transfer status.
- **v1.0 (2025-11)** — Initial release.
