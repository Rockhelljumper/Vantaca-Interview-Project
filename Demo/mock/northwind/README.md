# Northwind Connect Mock API

A standalone, standard-library Go mock of the documented Northwind Connect v1 API in [the supplied developer guide](<../../../Instructions/02-NORTHWIND-CONNECT-API 1.md>).

This is a development/test dependency. It intentionally does not claim to resolve contradictions or missing production guarantees in the Northwind contract.

## Implemented contract

| Method | Path | Behavior |
|---|---|---|
| GET | `/v1/accounts` | Bare JSON account array, `page` pagination, page size 50 |
| GET | `/v1/accounts/{id}/transactions` | Bare JSON transaction array with pagination |
| POST | `/v1/transfers` | Validates and creates a `PENDING` transfer |
| GET | `/v1/transfers` | Bare JSON transfer array with pagination |
| GET | `/healthz` | Mock-only unauthenticated health endpoint |
| GET | `/openapi.yaml` | Unauthenticated OpenAPI 3.0.3 contract used by the local explorer |
| POST | `/__mock/accounts/{id}/transactions` | Mock-only external-activity control; adds a synthetic transaction and changes the balance |
| POST | `/__mock/transfers/{id}/status` | Mock-only deterministic status/webhook control |

All contract endpoints require `api_key` as a query parameter, matching the guide. The mock logs only the URL path, never the query string.

The seeds include the documented account `acc_1029`, transaction `txn_88213`, and additional synthetic accounts/transactions for useful flows. Generated transfer IDs use `trf_<boot-id>_<sequence>` so rebuilding the in-memory mock cannot collide with transfer IDs retained in the SQL demo volume.

## Run the project stack with Docker Compose

From the repository root, start every currently implemented application with one command:

```powershell
docker compose up --build -d --wait
```

The Northwind mock is then available at `http://localhost:8081`. Compose builds the non-root image, waits for the container-native health check, and uses only synthetic defaults.

Use `http://localhost:18090` to explore both the mock and core APIs in Swagger UI, or download this service's exact contract from `http://localhost:8081/openapi.yaml`. Swagger's **Authorize** dialog accepts the synthetic query key `northwind_mock_local_key`.

```powershell
curl.exe "http://localhost:8081/v1/accounts?api_key=northwind_mock_local_key"
docker compose ps
```

To override local values, copy the root `.env.example` to `.env` and edit the copy. `NORTHWIND_MOCK_HOST_PORT` changes the host-side port without changing the container's internal port.

Stop the stack from the repository root:

```powershell
docker compose down
```

The same root Compose project also starts SQL Server, the Go Vantaca API, and Next.js; the optional `automation` profile adds n8n. All executable assets remain under `Demo/`.

## Run the Go process directly

Requires Go 1.25 or later.

```powershell
cd Demo/mock/northwind
$env:NORTHWIND_MOCK_API_KEY = "northwind_mock_local_key"
go run ./cmd/server
```

The server listens on `http://localhost:8081` by default.

```powershell
curl.exe "http://localhost:8081/v1/accounts?api_key=northwind_mock_local_key"
```

The key above is intentionally synthetic and local-only. Do not reuse real partner credentials.

## Create a transfer

```powershell
curl.exe -X POST "http://localhost:8081/v1/transfers?api_key=northwind_mock_local_key" -H "Content-Type: application/json" -d '{"from_account_number":"000123454321","to_account_number":"000987656789","routing_number":"021000021","amount":250.00,"currency":"USD"}'
```

The first successful transfer has the documented shape and a generated ID:

```json
{
  "id": "trf_<boot-id>_55120",
  "status": "PENDING",
  "amount": 250.00,
  "created_at": "..."
}
```

## Transfer status and webhooks

Configure `NORTHWIND_MOCK_WEBHOOK_URL`, create a transfer, and drive a deterministic status change through the mock-only control endpoint:

```powershell
curl.exe -X POST "http://localhost:8081/__mock/transfers/<created-transfer-id>/status?api_key=northwind_mock_local_key" -H "Content-Type: application/json" -d '{"status":"POSTED","deliveries":1}'
```

Supported transitions:

- `PENDING -> POSTED`
- `PENDING -> FAILED`
- `POSTED -> RETURNED`
- Repeating the current status, which allows duplicate-delivery testing

`deliveries` may be 1–5. Values above 1 intentionally send duplicate webhook payloads. Webhooks are unsigned, matching the supplied guide, and retry up to the configured attempt count after a connection error or non-2xx response.

## Simulate activity outside Vantaca

The demo can change Northwind before Vantaca's SQL snapshot refreshes:

```powershell
curl.exe -X POST "http://localhost:8081/__mock/accounts/acc_1029/transactions?api_key=northwind_mock_local_key" -H "Content-Type: application/json" -d '{"amount":125.50,"description":"EXTERNAL DEMO DEPOSIT","merchant_category_code":"0000"}'
```

This mock-only endpoint creates a synthetic transaction and adjusts the mock account balance atomically. It exists to demonstrate the recent-transaction mismatch → SQL commit → frontend re-fetch workflow and must never be used by production adapter code.

## Deterministic failure scenarios

Set the `X-Northwind-Mock-Scenario` request header:

| Value | Behavior |
|---|---|
| `429` | Returns 429 with `Retry-After: 1` |
| `500` | Returns the documented server-error shape |
| `503` | Returns service unavailable |
| `latency` | Delays before processing for `NORTHWIND_MOCK_SCENARIO_DELAY` |
| `post-commit-timeout` | On POST `/v1/transfers` only, commits the transfer and delays the response |

The post-commit scenario is the important ACH ambiguity test: use a client timeout shorter than the configured scenario delay. The client sees a failure while `GET /v1/transfers` shows that Northwind created the transfer. The mock does not invent a transfer-submission idempotency field that is absent from the supplied public contract; whether Northwind supports one remains an explicit partner question.

## Configuration

The Go process does not automatically load `Demo/mock/northwind/.env.example`; export variables through the shell or IDE when running Go directly. Root Compose reads root `.env` automatically when present.

| Variable | Default | Purpose |
|---|---|---|
| `NORTHWIND_MOCK_PORT` | `8081` | HTTP listen port |
| `NORTHWIND_MOCK_API_KEY` | `northwind_mock_local_key` | Synthetic query-parameter key |
| `NORTHWIND_MOCK_WEBHOOK_URL` | empty | Target for transfer status webhooks |
| `NORTHWIND_MOCK_WEBHOOK_ATTEMPTS` | `3` | Maximum attempts per delivery |
| `NORTHWIND_MOCK_WEBHOOK_BACKOFF` | `100ms` | Linear retry-backoff base |
| `NORTHWIND_MOCK_SCENARIO_DELAY` | `5s` | Latency/post-commit response delay |
| `SWAGGER_UI_ORIGIN` | `http://localhost:18090` | Exact demo explorer origin allowed to call the mock from a browser |

## Contract assumptions kept explicit

- The guide does not define customer identity/account scoping. This mock represents one synthetic customer.
- The guide calls balance current but does not define ledger/available/as-of semantics. The mock emits only the documented field.
- The guide documents JSON numbers without precision rules. This mock accepts USD with at most two fractional digits and stores cents as integers.
- The guide does not identify which account the single transfer routing number describes. The mock treats it as the destination routing number.
- The guide says both no rate limits and that 429 can occur. Normal requests are unlimited; the 429 scenario is explicit.
- The guide says exactly-once webhooks and also promises retries. The mock can produce retries and duplicates so consumers must be idempotent.
- The guide has no webhook signature. The mock sends no signature and must not be treated as production security evidence.
- The undocumented `/internal/accounts/full` endpoint is not implemented.

## Tests

```powershell
cd Demo/mock/northwind
gofmt -l ./cmd ./internal
go test ./...
go test -race ./...
go vet ./...
```

Tests cover:

- Query-key authentication and documented seed payloads
- Page-size-50 pagination and empty pages
- Transactions and missing accounts
- Strict transfer validation and precise money JSON
- 429/500/503 scenarios
- Post-commit timeout ambiguity
- Transfer state transitions
- Webhook retry and duplicate delivery
- Concurrent unique transfer IDs
