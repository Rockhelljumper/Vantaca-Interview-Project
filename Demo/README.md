# Demo applications

This directory is the executable interview-demo boundary. It deliberately separates application/runtime assets from discovery notes, AI logs and harnesses, source instructions, and engineering task briefs.

## Contents

| Path | Purpose |
|---|---|
| `api/` | Go Vantaca API, application services, Northwind adapter, and SQL repository |
| `web/` | Next.js/TypeScript/Tailwind interview UI and server-side API proxy |
| `mock/northwind/` | Deterministic Go implementation of the supplied Northwind contract |
| `database/` | SQL Server migrations, tenant-scoped views, inbox/outbox tables, and synthetic seed |
| `automation/n8n/` | Optional reconciliation-trigger workflow and import notes |

The core and mock services each own an OpenAPI 3.0.3 document beside their HTTP handlers. Root Compose mounts those same checked-in files into a pinned Swagger UI image, avoiding a generated/spec drift path.

The root [`compose.yaml`](../compose.yaml) is intentionally outside this folder because it is the repository-wide startup entry point. All service build contexts and runtime data definitions point into `Demo/`.

## Runtime flow

1. SQL Server becomes healthy.
2. The Northwind mock starts with synthetic accounts, transactions, and transfer state.
3. The Go API applies deterministic SQL migrations/seeds and synchronizes the initial read model.
4. Swagger UI starts with a selector for both API contracts after the two HTTP services are healthy.
5. Next.js starts only after the API health check succeeds.
6. The optional n8n profile calls the Go reconciliation boundary; it never writes SQL or calls Northwind directly.

The browser calls only Next.js/Vantaca endpoints. The Go adapter is the only component that sees Northwind account/routing values, and those values are not persisted in the demo database.

## Diagnostic logging

The Go API writes each structured application event to stdout and, after SQL bootstrap, to `dbo.application_logs`. Request logs contain the method, path without query string, response status, duration, and correlation ID. Successful health probes are intentionally excluded. `dbo.vw_recent_application_errors` provides the `WARN`/`ERROR` investigation surface; for example:

```sql
SELECT TOP (100)
    occurred_at,
    severity,
    event_name,
    correlation_id,
    username,
    api_key_last_four,
    attributes_json
FROM dbo.vw_recent_application_errors
ORDER BY occurred_at DESC;
```

Sanitization happens inside both logging handlers rather than at individual call sites. Username is retained when the upstream identity supplies one; passwords, tokens, authorization/cookie values, connection credentials, routing numbers, and raw bodies/payloads are redacted or omitted. Full account/card numbers and API keys are reduced to masked last-four values. Database-log failures fall back to the already-redacted stdout sink without recursively attempting another SQL write. Production still requires an approved retention policy, access controls, and log export/alert ownership.

## Start and test

Run from the repository root:

```powershell
docker compose up --build -d --wait
```

Open `http://localhost:18090` for the combined Swagger explorer. The raw contracts remain available from their owning services at `http://localhost:18080/openapi.yaml` and `http://localhost:8081/openapi.yaml`.

The complete walkthrough, test commands, host-port overrides, and production limitations are in the root [`README.md`](../README.md).

With the stack healthy, the repeatable workflow check is:

```powershell
.\Demo\scripts\e2e-smoke.ps1
```

When host ports are overridden, pass `-ApiBaseUrl` and `-WebBaseUrl` explicitly.

The comprehensive executable acceptance suite lives with the web test harness:

```powershell
Push-Location Demo/web
npm ci
npm run test:e2e:install
npm run test:e2e
Pop-Location
```

The `api` Playwright project calls the Go API and Northwind mock directly. It verifies tenancy, security headers, masking, freshness, every bounded read failure, asynchronous reconciliation, exact money, validation, idempotent submission, webhook deduplication, transfer lifecycle states, ambiguous outcomes, internal reconciliation authorization, both OpenAPI documents, and the restricted Swagger CORS boundary. The `ui-chromium` project drives the actual dashboard and Swagger UI, including contract switching, all scenario controls, keyboard tooltips, account selection, frontend invalidation, safety messages, and a 390-pixel responsive-layout assertion.
