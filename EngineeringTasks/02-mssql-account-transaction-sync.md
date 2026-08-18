# [ENG-02] Build SQL Server persistence and account/transaction synchronization

**Owner:** Engineer 2 — MSSQL + Account/Transaction Sync  
**Suggested labels:** `database`, `synchronization`, `backend`  
**Primary AI harness:** [Database Harness](../AI-Harnesses/database-harness.md)  
**AI initialization:** Start with the [Initialization Evaluator](../AI-Harnesses/initialization-evaluator.md).  
**Authorization:** Schema, repository, and application implementation begin only after the Integration Lead authorizes the relevant scope and data-handling decisions.

## Goal

Create a reliable SQL Server read model and synchronization path for linked accounts, balances, and recent transactions. Preserve Northwind as authoritative while giving the frontend fast reads with accurate freshness states.

Treat the proposed controls in [Security and Data Classification](../Notes/Security-Data-Classification.md) and test IDs AUTH-001, TXN-001/002, MONEY-001, and DATA-001 in the [QA Acceptance Matrix](../Notes/QA-Acceptance-Matrix.md) as shared review inputs; Security/Data approval still governs protected production storage.

## Architecture placement and application dependencies

**Placement:** SQL Server is the durable persistence/read-model application. Repository code, sync workers, comparison logic, and the outbox publisher remain modules inside the Go application. See the [shared runtime dependency map](README.md#architecture-placement-and-runtime-dependencies).

| Relationship | Application/component | Contract |
|---|---|---|
| Receives from | ENG-01 Northwind adapter | Normalized accounts/transactions and typed fetch outcomes; no partner DTOs in SQL repositories |
| Called by | Go account/transaction APIs | Tenant-scoped, masked account and recent-transaction views with explicit freshness/version metadata |
| Runs from | Go workers triggered by approved scheduler/job mechanism | Bounded sync/reconciliation selection, concurrency, retry eligibility, and cancellation |
| Publishes to | Frontend invalidation transport through a Go outbox publisher | Non-sensitive aggregate/version event only after the data transaction commits |
| Depends on | Vantaca identity/tenant mapping | Stable customer/link/account ownership used in every repository query and constraint |
| Depends on | Approved encryption key service and policy | Encrypt/decrypt authorization, managed key version, rotation, outage, audit, and restore behavior |
| Observed by | Vantaca logs/metrics/alerts and QA | Stale age, sync outcomes, backlog, outbox attempts, constraint failures, and sanitized diagnostics |

## Architecture workflow logic

This issue owns the persistence portion of Major Workflows 1 and 2 in [StartHere.md](../Notes/StartHere.md):

1. Account synchronization receives normalized accounts from ENG-01 and transactionally upserts the account read model plus fetched-at metadata.
2. The frontend-facing API reads only customer-scoped SQL views; it does not wait for or call Northwind directly.
3. A recent-transactions request returns the current SQL snapshot immediately and schedules bounded asynchronous reconciliation.
4. The worker compares a normalized Northwind window with the stored values.
5. If values differ, one database transaction updates the read model and writes the invalidation/outbox event. The frontend is signaled only after commit.
6. If values match, only checked-at metadata changes. If Northwind or SQL fails, prior trustworthy data remains and no false invalidation is published.

## Sample Northwind data received

The following payloads are synthetic examples from the supplied Northwind guide/mock. Full account and routing numbers are shown only to define the inbound trust boundary; they must not appear in customer views, logs, routine fixtures, or unapproved plaintext storage.

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
  }
}
```

Expected safe read-model projection:

```json
{
  "account_id": "acc_1029",
  "display_account": "checking ••••4321",
  "balance": "4820.55",
  "currency": "USD",
  "status": "open",
  "fetched_at": "2026-08-16T18:00:00Z",
  "transactions": [
    {
      "id": "txn_88213",
      "amount": "-42.17",
      "currency": "USD",
      "description": "COFFEE HOUSE #42",
      "merchant_category_code": null,
      "posted_at": "2026-07-21T14:03:00Z"
    }
  ]
}
```

The response uses decimal strings illustratively; the final API representation must follow the approved money contract and must never depend on binary floating-point persistence.

## Logical table layout

Names and exact SQL types may follow repository conventions, but the implementation must preserve these boundaries:

| Logical table | Key columns | Purpose and important fields | Security/integrity notes |
|---|---|---|---|
| `northwind_links` | internal link ID; Vantaca tenant/customer reference | Partner-link status, created/revoked timestamps, last successful sync | Never treat a view or partner account ID as customer authorization. Enforce the approved tenant/link ownership model. |
| `linked_accounts` | internal account ID; link FK; unique Northwind account ID within the approved scope | Type, currency, status, exact-numeric balance, masked last four, fetched/checked timestamps, row version | Contains display-safe/read-model fields only. No plaintext full account/routing values. |
| `linked_account_secure_values` | account FK | Encrypted account number, encrypted routing number when approved/required, key version, encrypted-at timestamp | Optional until policy approval. Restrict access to the narrow transfer boundary; never expose through customer views. |
| `account_transactions` | internal transaction ID; account FK; unique Northwind transaction ID per account | Exact-numeric amount, currency, description, nullable MCC, posted timestamp, first/last seen timestamps | Preserve missing MCC as null. Index the customer-scoped recent read path. |
| `sync_runs` | sync-run ID; link/account reference | Sync type, status, start/end, page/count metadata, safe error category, retry/next-attempt metadata | No raw sensitive payloads or secrets in error text. Supports bounded work and operational evidence. |
| `read_model_outbox` | event ID; aggregate ID/version | Invalidation type, non-sensitive account reference, created/published timestamps, attempts | Written in the same transaction as a mismatch update; publication occurs after commit and is idempotent. |

Transfer intent, webhook inbox, and transfer status-history tables belong to ENG-03, but both engineers must agree on identifiers, transaction conventions, encryption access, and outbox patterns.

## Supporting views and indexes

| Logical view | Consumer | Minimum projection |
|---|---|---|
| `vw_customer_linked_accounts` | Go account API | Tenant/customer scope key, internal/partner account ID, type, masked display value, exact balance, currency, status, fetched/checked/stale metadata |
| `vw_customer_recent_transactions` | Go recent-transactions API | Tenant/customer scope key, account ID, partner transaction ID, exact amount, currency, description, nullable MCC, posted timestamp, freshness/version |
| `vw_account_sync_health` | Operations/reconciliation | Link/account reference, last success/attempt, stale age, last safe error class, pending work state |
| `vw_unpublished_read_model_events` or equivalent repository query | Outbox publisher | Event ID, aggregate/version, type, created time, attempt metadata; no protected account/routing values |

Views simplify stable projections but are not an authorization boundary. Every repository call must still carry and filter by the authenticated tenant/customer context.

Minimum indexes/constraints should support:

- unique partner account identity within the confirmed customer/link scope;
- unique partner transaction identity per linked account;
- customer/account plus descending `posted_at` for recent transactions;
- stale/eligible sync work ordered by next-attempt time;
- unpublished outbox work ordered by creation time;
- guarded row-version/status updates for concurrent sync.

The engineer may rename, combine, or split objects when repository conventions require it, provided the same access, atomicity, security, and query behavior remain demonstrable.

## Encryption-policy requirements

- Treat the approved answer to B0-4/B1-5 in [OPEN-QUESTIONS.md](../AI-Log/OPEN-QUESTIONS.md) as the source of truth for encryption, access, retention, deletion, and audit controls; unresolved policy blocks schema freeze for protected values.
- Treat full account and routing values as protected data under [Concern C-009](../AI-Log/CONCERNS.md) and the eventual Security-approved encryption policy.
- Do not freeze the secure-value schema until the owner, retention, access, key-management, rotation, deletion, audit, and recovery requirements are approved.
- Prefer avoiding full-value persistence if Northwind supports opaque account tokens/IDs. If persistence is required, keep ciphertext separate from the display read model.
- Use the approved field-encryption mechanism and managed key service. Store only ciphertext and non-secret key/version metadata; never store keys in SQL, source, Compose, migrations, or fixtures.
- TDE/volume encryption may protect media but does not automatically satisfy field-level access or masking requirements; implement the policy selected by Security rather than assuming one control is sufficient.
- Permit decryption only in the narrow authorized transfer path. Customer/operations views use masked last-four values and must not expose ciphertext as a substitute for access control.
- Define key rotation/re-encryption, restore, audit, and failure behavior before production. A key-service failure must fail closed for protected reads without corrupting the public read model.

## Scope

- Design deterministic SQL Server migrations, constraints, indexes, and repository interfaces for customer/account links, account snapshots, transactions, and sync metadata.
- Use `database/sql`, parameterized raw SQL, explicit transactions, and lossless money/time representations.
- Implement bounded account synchronization and the async recent-transactions workflow described by [Decision D-008](../AI-Log/DECISIONS.md#d-008--serve-recent-transactions-from-sql-and-reconcile-asynchronously).
- Compare normalized recent-transaction values, commit mismatches atomically, and emit frontend invalidation only after commit.
- Preserve last known data on Northwind or persistence failure and update checked/fetched/stale metadata accurately.
- Minimize sensitive account data and implement the approved encryption, managed-key, masking, audit, retention, deletion, backup/restore, and tenant-isolation decisions.

## Acceptance criteria

- [ ] Migrations create a fresh database deterministically and are safe for the approved lifecycle.
- [ ] Database constraints enforce customer/account scoping, stable external identities, currencies, required timestamps, and valid precision.
- [ ] Repository queries are parameterized and scoped to the authenticated tenant/customer.
- [ ] Repeated account or transaction sync produces one correct current record rather than duplicates.
- [ ] External HTTP calls never occur inside a SQL transaction.
- [ ] Recent-transaction reads return the committed SQL snapshot plus freshness metadata immediately.
- [ ] A Northwind mismatch commits changed values before one invalidation is published; matching values cause no frontend refresh.
- [ ] Failed fetches or commits preserve the prior snapshot and cannot publish a false refresh signal.
- [ ] Table/view/index design documents each stored field, source, classification, retention owner, and customer/operational access path.
- [ ] Full account/routing values are absent unless explicitly approved; if approved, only policy-compliant ciphertext and key-version metadata are stored outside customer read tables/views.
- [ ] Customer and operational views expose masked/display-safe projections and cannot return protected values.
- [ ] Key rotation and key-service failure behavior are implemented and documented according to the approved encryption policy.

## Required testing

- [ ] Run repository integration tests against SQL Server 2022, not an in-memory substitute.
- [ ] Test migrations from empty state, constraints, indexes used by expected queries, and transaction rollback.
- [ ] Test money precision, UTC timestamps, pagination ingestion, missing optional data, and duplicate external IDs.
- [ ] Test negative tenant/customer isolation for account and transaction reads/writes.
- [ ] Test concurrent/repeated sync runs for duplicate prevention and safe update behavior.
- [ ] Test Workflow 2 match, mismatch, Northwind failure, and SQL commit failure paths.
- [ ] Verify mismatch publishes invalidation after commit; verify match/failure publishes none.
- [ ] Verify logs and fixtures contain only synthetic, masked data.
- [ ] Query SQL directly in an integration test to confirm protected sample values are not stored as plaintext in tables, views, indexes, logs, or outbox payloads.
- [ ] Test authorized decrypt, unauthorized access denial, key-version rotation/re-encryption, key-service outage, and restore behavior required by policy.
- [ ] Test all supporting views for expected projections, stable types, tenant filtering, masking, freshness, nullable MCC, and exclusion of secured columns.
- [ ] Load the documented synthetic account/transaction samples and verify their exact mapping, uniqueness, precision, and recent-transaction ordering.

## Dependencies and handoffs

- Approved identity/customer/account-link model and encryption/data-handling policy, including managed key owner and rotation/restore procedure.
- Product definition of “recent,” freshness, stale behavior, and unavailable-state UX.
- ENG-01 adapter interfaces and confirmed pagination/order behavior.
- ENG-04 consumes the read/freshness API and invalidation contract.
- QA-01 supplies traceable acceptance cases and SQL integration evidence requirements.

## Out of scope

- Frontend implementation, transfer submission/state tables owned by ENG-03, or partner contract invention.
- ORM introduction or long-running SQL transactions around Northwind calls.
- Selecting a production notification/queue platform without approval.
