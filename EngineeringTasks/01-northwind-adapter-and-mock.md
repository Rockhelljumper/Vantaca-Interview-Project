# [ENG-01] Build the Northwind adapter contract and maintain the deterministic mock

**Owner:** Engineer 1 — Northwind Adapter + Mock  
**Suggested labels:** `backend`, `integration`, `northwind`  
**Primary AI harness:** [Go Backend Harness](../AI-Harnesses/go-backend-harness.md)  
**AI initialization:** Start with the [Initialization Evaluator](../AI-Harnesses/initialization-evaluator.md).  
**Authorization:** The existing mock is authorized. Implementing the Vantaca-side adapter begins only after the Integration Lead authorizes the relevant implementation scope.

## Goal

Provide a narrow, well-tested Northwind boundary that translates the confirmed partner contract into internal types without leaking query-key authentication, partner DTOs, pagination quirks, or retry behavior into the rest of the application.

## Architecture placement and application dependencies

**Placement:** The Northwind adapter lives inside the Go modular application, not as a separate production microservice. The mock is a separate local-only container that substitutes for Northwind during development. See the [shared runtime dependency map](README.md#architecture-placement-and-runtime-dependencies).

| Relationship | Application/component | Contract |
|---|---|---|
| Called by | Account/transaction sync and reconciliation workers | Typed account/transaction reads with pagination, cancellation, and freshness/error metadata |
| Called by | Transfer domain/reconciliation service | One safe submission attempt and confirmed exact transfer lookup/correlation operations |
| Calls | Northwind Connect API or local mock | Versioned HTTP DTO/error contract; mock must be selectable only by environment configuration |
| Depends on | Vantaca secrets platform and outbound networking | Query-key retrieval/injection, rotation, TLS/DNS/egress, and redacted telemetry |
| Produces | ENG-02 and ENG-03 application services | Internal normalized values and typed outcomes; never SQL writes, UI models, or partner DTO leakage |

## Architecture workflow logic

This issue supplies the Northwind boundary used by Major Workflows 1, 2, 3, 5, and 6 in [StartHere.md](../Notes/StartHere.md):

1. Account sync calls `GET /accounts` and transaction reconciliation calls `GET /accounts/{id}/transactions` through the adapter.
2. The adapter normalizes partner DTOs and errors before returning them to application services; it never writes SQL or updates the UI.
3. Transfer submission performs one `POST /transfers` attempt under the approved timeout/idempotency policy and returns an explicit ambiguous outcome when the response is lost.
4. Reconciliation uses only a confirmed exact lookup/correlation mechanism.
5. Failure classification exposes typed `429`, transient outage, validation, authorization, and timeout results so callers can apply operation-safe policy.

## Sample data

These synthetic examples come from the supplied guide/mock and define representative adapter shapes:

```json
{
  "account": {
    "id": "acc_1029",
    "account_number": "000123454321",
    "routing_number": "021000021",
    "type": "checking",
    "balance": 4820.55,
    "currency": "USD",
    "status": "open"
  },
  "transaction": {
    "id": "txn_88213",
    "amount": -42.17,
    "currency": "USD",
    "description": "COFFEE HOUSE #42",
    "posted_at": "2026-07-21T14:03:00Z"
  },
  "transfer": {
    "id": "trf_55120",
    "status": "PENDING",
    "amount": 250.00,
    "created_at": "2026-07-28T16:22:00Z"
  },
  "error": {
    "error": "invalid_account",
    "message": "Account not found"
  }
}
```

All values are synthetic. The missing `merchant_category_code` in the transaction example is intentional and must be handled until Northwind resolves the schema inconsistency.

## Scope

- Define small adapter interfaces needed by account sync, recent transactions, transfer submission, and reconciliation.
- Map documented account, transaction, transfer, error, and webhook shapes at the adapter boundary.
- Apply explicit timeouts, cancellation, typed errors, bounded safe-read retries, `Retry-After` handling, and query-string redaction.
- Keep transfer submission non-retrying unless Northwind confirms safe idempotency/correlation.
- Maintain the existing Go mock and add only deterministic scenarios required by confirmed contracts and acceptance tests.
- Keep mock-only controls outside the public `/v1` surface and clearly label assumptions.

## Acceptance criteria

- [ ] Northwind DTOs and authentication details remain inside the adapter.
- [ ] Accounts, account transactions, transfer creation, and transfer listing expose narrow internal operations needed by consumers.
- [ ] Pagination follows confirmed ordering/termination behavior and cannot loop or grow without bounds.
- [ ] Error mapping distinguishes validation, authentication, not found, throttling, transient outage, timeout, and ambiguous transfer outcome.
- [ ] Logs and errors omit API keys, query strings, full account numbers, and sensitive payloads.
- [ ] The mock remains non-root, containerized, synthetic, deterministic, and startable through root Compose.
- [ ] Every unresolved partner behavior is represented as a question or explicit assumption rather than invented code.

## Required testing

- [ ] Unit tests cover DTO mapping, optional/unknown fields, money precision, timestamps, and typed errors.
- [ ] HTTP tests cover cancellation, body closure, malformed JSON, `401`, `404`, `429/Retry-After`, `500`, `503`, and latency.
- [ ] Pagination tests cover first page, multiple pages, empty terminal page, and changing-page assumptions.
- [ ] Transfer tests reproduce pre-commit failure and post-commit timeout without automatic resubmission.
- [ ] Contract tests compare adapter expectations with the reviewed Northwind artifact and mock payloads.
- [ ] Concurrency/race checks run where the toolchain supports them; any limitation is documented.
- [ ] Docker/Compose smoke tests verify health and representative authenticated endpoints.

## Dependencies and handoffs

- Northwind schemas, pagination, quotas, timeouts, environments, and credential answers.
- ENG-02 consumes account/transaction read operations.
- ENG-03 consumes transfer/webhook/reconciliation operations and owns financial state decisions.
- QA-01 owns the contract/failure traceability matrix.

## Out of scope

- Customer-facing UI, SQL schema ownership, or transfer state-machine policy.
- Generalized multi-bank framework design.
- Claiming the mock certifies Northwind production behavior.
