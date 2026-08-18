# QA Acceptance and Evidence Traceability

## Purpose

This is the working requirement → risk → test → evidence matrix requested by the original exercise. It is the concrete output coordinated by [QA-01](../EngineeringTasks/05-qa-pm-acceptance-release.md), not merely a plan to create one.

The matrix separates mock-only evidence, executable end-to-end demo evidence, planned work, blockers, and production-candidate evidence. A demo pass proves the synthetic local topology and checked-in implementation; it does not prove the production integration.

## Evidence status

| Status | Meaning |
|---|---|
| **MOCK PASS** | Automated evidence exists for the local Northwind mock only |
| **DEMO PASS** | Executable evidence exists across the synthetic local Vantaca demo; production controls may still be blocked |
| **PLANNED** | Requirement and test are defined but the required evidence does not exist yet |
| **BLOCKED** | A contract, policy, platform, or accountable-owner decision is required before the test can be finalized or passed |
| **RELEASE PASS** | Reserved for reviewed evidence from the production-candidate build in the approved environment |

No row is currently **RELEASE PASS**.

## Requirement-to-evidence matrix

| ID | Requirement / decision source | Primary risk | Automated test evidence required | Manual / operational evidence required | Owner | Current evidence and release gate |
|---|---|---|---|---|---|---|
| **AUTH-001** | Every account, transaction, and transfer access uses established Vantaca identity, tenant, consent, linked-account, and entitlement rules; [B0-2](../AI-Log/OPEN-QUESTIONS.md#1a-internal-vantaca--async-decision-queue), [C-002](../AI-Log/CONCERNS.md#c-002--undefined-customeraccount-scoping) | Cross-tenant disclosure or unauthorized money movement | API/service/repository positive and negative tenant tests; IDOR tests; revoked-link and insufficient-entitlement cases | Identity/data-flow review; sample audit records; access revocation exercise | EM-01, ENG-02, ENG-03, Security, QA | **BLOCKED:** Vantaca and Northwind scoping contracts are unresolved. Stop-ship for every production capability. |
| **BAL-001** | Balance is a labeled snapshot with meaning, source/fetched timestamps, stale threshold, and outage behavior; [B0-1](../AI-Log/OPEN-QUESTIONS.md#1a-internal-vantaca--async-decision-queue), [C-008](../AI-Log/CONCERNS.md#c-008--five-second-polling-and-local-source-of-truth-request), [C-013](../AI-Log/CONCERNS.md#c-013--undefined-balance-semantics) | Customer acts on stale or misrepresented funds | Sync/service/UI tests for fresh, stale, unavailable, externally changed, and Northwind-recovered states; boundary test at approved stale threshold | Product-approved wording and acceptance examples; stale-age dashboard/alert demonstration | ENG-02, ENG-04, Product, QA | **BLOCKED:** Product tolerance and Northwind balance semantics/freshness are unresolved. |
| **TXN-001** | Recent-transaction reads return tenant-scoped SQL snapshot immediately; mismatches refresh asynchronously, commit before invalidation, and then force a Vantaca API re-fetch; [D-008](../AI-Log/DECISIONS.md#d-008--serve-recent-transactions-from-sql-and-reconcile-asynchronously) | Slow reads, inconsistent UI, or notification before durable data | Repository/service/UI integration test with changed partner data; assert initial snapshot, coalesced job, normalized upsert commit, one post-commit invalidation, then updated API response | Trace/timeline showing request latency, commit, outbox publish, UI re-fetch, and observed/fetched timestamps | ENG-02, ENG-04, QA | **DEMO PASS:** external activity produced one new transaction and exactly one SQL data-version advance; Next.js performs bounded version polling through the API. Production load/transport evidence remains required. |
| **TXN-002** | Matching or failed refreshes do not churn customer state: equality records a successful check without invalidation; failure preserves last-known data and exposes stale/error metadata | Refresh storms, unnecessary rerenders, or loss of usable last-known state | Match/no-op, 429/500/503/timeout, retry-exhaustion, job-coalescing, and outbox-absence tests | Dashboard evidence for sync age, failure, retry, backlog, and recovery | ENG-01, ENG-02, ENG-04, QA | **DEMO PASS:** identical refresh kept the version stable; a 503 preserved all account rows, marked freshness degraded, and recovered on the next successful sync. Production alert/backlog evidence remains planned. |
| **ADP-001** | Adapter honors reviewed authentication, pagination, optional/unknown fields, error taxonomy, timeouts, and mock/production separation; [C-005](../AI-Log/CONCERNS.md#c-005--unsafe-credential-transport-and-environment-model), [C-012](../AI-Log/CONCERNS.md#c-012--incomplete-pagination-and-transaction-schema) | Contract drift, secret leakage, missed/duplicated data | Contract fixtures; multi/empty-page tests; missing optional MCC; unknown field/enum; 401/404/422/429/500/503/latency; URL-redaction assertion; no `/__mock` calls outside tests | Compare captured sandbox contract to reviewed OpenAPI/schema; credential redaction review | ENG-01, QA, Security | **MOCK PASS** for query-key auth, page-size-50 behavior, missing account, and failure shapes. **BLOCKED** for production contract/credential approval. |
| **MONEY-001** | Amounts parse losslessly and use exact decimal/minor-unit semantics; no binary floating point; [C-011](../AI-Log/CONCERNS.md#c-011--imprecise-money-representation) | Silent monetary rounding or invalid transfer amount | Min/max/fraction/currency tests across JSON → domain → SQL → JSON; reject unsupported precision/overflow; sum/comparison tests | Northwind currency/rounding contract review and SQL type review | ENG-01, ENG-02, ENG-03, QA | **DEMO PASS** for exact Go minor units, SQL `BIGINT`, string API/UI money, and parser boundary tests. **BLOCKED** for production currency/precision rules. |
| **TRF-001** | An authorized transfer intent and unique Vantaca request identity are durable before partner submission; duplicate UI/API requests cannot create parallel submissions | Duplicate money movement from double-click or internal retry | Concurrency and repeated-request tests across UI/API/repository; unique constraint/idempotency assertion; crash between intent commit and send | Audit trace from authorization through intent record and feature/kill-switch state | ENG-02, ENG-03, ENG-04, QA | **DEMO PASS:** SQL intent/unique request constraint and repeated API request returned the same record without another partner POST. Crash-window and production authorization evidence remain planned/blocked. |
| **TRF-002** | A timeout after `POST /transfers` becomes `UNKNOWN / RECONCILIATION_REQUIRED` and is never automatically resubmitted without proven partner idempotency/exact lookup; [C-001](../AI-Log/CONCERNS.md#c-001--ambiguous-monetary-submission) | Duplicate transfer after post-commit response loss | Post-commit-timeout test; assert one outbound submission, durable ambiguous state, no automatic resubmit, and exact reconciliation before resolution | Alert/runbook exercise for UNKNOWN transfer; operator permissions and escalation evidence | ENG-03, EM-01, Operations, QA | **DEMO PASS:** post-commit timeout persisted `UNKNOWN`; repeating the Vantaca request left the partner transfer count unchanged. **BLOCKED** for safe production resolution; transfer-write stop-ship. |
| **TRF-003** | Transfer state machine preserves `PENDING`, `POSTED`, `FAILED`, `RETURNED`, and internal `UNKNOWN`; `POSTED → RETURNED` is supported; invalid/regressive events do not corrupt current state; [C-014](../AI-Log/CONCERNS.md#c-014--incomplete-transfer-lifecycle-contract) | Incorrectly claiming success or losing a late return | State-transition table/property tests; duplicate, out-of-order, late-return, unknown-enum, and concurrent webhook/reconciliation cases; immutable history assertion | Product-approved customer wording; Support runbook for failed/returned/unknown states | ENG-03, ENG-04, Product, Operations, QA | **DEMO PASS** for domain transition tests and live `PENDING → POSTED → RETURNED` persistence/UI DTOs. **BLOCKED** for authoritative partner transition/reason contract. |
| **WH-001** | Webhook receipts have a confirmed stable identity rule and tolerate retry without dropping legitimate later status events; `transfer_id` is a candidate, not an assumed sufficient key; [C-004](../AI-Log/CONCERNS.md#c-004--contradictory-webhook-delivery-guarantee) | Duplicate processing or suppression of valid status changes | Inbox uniqueness tests for the confirmed event key; identical retry, same transfer/new status, out-of-order, replay, crash/restart, and concurrent delivery cases | Compare raw receipt metadata and state history; replay/recovery exercise | ENG-03, Northwind, QA | **DEMO PASS:** two identical `POSTED` deliveries produced one inbox row and a later `RETURNED` event remained distinct. **BLOCKED** until Northwind confirms `transfer_id`, `(transfer_id,status)`, or event/delivery ID semantics. |
| **WH-002** | Webhook authenticity and replay controls are verified before an event can advance trusted customer-visible state; [C-003](../AI-Log/CONCERNS.md#c-003--unauthenticated-financial-status-webhooks) | Forged financial status or replay attack | Valid/invalid/missing/expired signature or approved equivalent; body-tamper; replay-window; source-control; fail-closed tests | Security threat-model/sign-off; ingress configuration evidence; key/certificate rotation exercise | ENG-03, Security, Operations, QA | **BLOCKED:** supplied webhooks are unsigned; requires partner control or Security-approved compensation. |
| **REC-001** | Polling/reconciliation finds the exact partner transfer and converges missed/ambiguous states without using approximate matching or resubmission | Wrong transfer matched, unresolved money movement, missed webhook | Exact ID/client-reference lookup; missed-webhook; API outage/recovery; stale pending; mismatch; rerun-idempotency; multiple similar-transfer cases | Reconciliation dashboard/alert and exception-runbook exercise with named owner/SLA | ENG-03, Operations, QA | **BLOCKED:** exact Northwind lookup/correlation and cadence/SLA are unresolved. Transfer-write stop-ship. |
| **REL-001** | Northwind calls use explicit timeouts, bounded retries only for safe operations, `Retry-After` where valid, concurrency limits, jitter/backoff, and stale/fallback behavior; [C-006](../AI-Log/CONCERNS.md#c-006--rate-limit-contradiction), [C-007](../AI-Log/CONCERNS.md#c-007--availability-claims-contradict-documented-failures) | Retry storm, thread exhaustion, cascading outage, unsafe monetary retry | 429/500/503/latency/connection reset tests; retry-budget assertion by method; concurrency/load boundary; cancellation/leak tests; recovery after outage | Metrics/alerts for rate, latency, error, retry, saturation; partner-incident table-top | ENG-01, ENG-02, ENG-03, Operations, QA | **DEMO PASS** for typed 429/500/503/latency behavior, bounded safe-read retry, explicit timeouts, and zero automatic transfer retry. **BLOCKED** for production quotas/SLA and approved policy/load evidence. |
| **DATA-001** | Protected values follow the approved [data-classification controls](Security-Data-Classification.md): minimize, encrypt, mask, tenant-scope, audit, rotate, retain, and delete; [C-009](../AI-Log/CONCERNS.md#c-009--full-account-number-persistence-request) | Financial-data disclosure, excessive retention, unusable encrypted backup | Log/trace/snapshot scanning; SQL permission/tenant-negative tests; ciphertext-at-rest check; decrypt authorization; key version/rotation; backup/restore; retention/deletion tests | Security/Data approval, access review, sample audit records, recovery exercise | ENG-02, Security/Data, Operations, QA | **DEMO PASS** for data minimization: SQL stores opaque account IDs/last-four only, webhook raw payload is reduced to SHA-256, and full account/routing values stay transient. **BLOCKED** for approved production encryption/key/access/retention evidence. |
| **OBS-001** | Every request/job/webhook/reconciliation has non-sensitive correlation and the minimum metrics, logs, traces, alerts, and ownership needed to diagnose freshness and money movement | Silent stale data, unresolved transfer, or unusable incident evidence | Telemetry schema and redaction tests; alert-condition tests; correlation across request/job/outbox/inbox/reconciliation | Dashboard review; alert routing; runbook/table-top with Support/Operations | All engineers, Operations, QA | **DEMO PASS** for structured request/job logs, correlation IDs, freshness/error state, and non-body route logging. Production metrics, alerts, tracing, and ownership remain planned/blocked. |
| **OPS-001** | Feature flags/kill switch isolate reads and transfer writes; deploy/rollback, job pause/resume, replay, backup/restore, and partner escalation are exercised | Incident cannot be contained or recovered safely | Configuration/flag tests; migration forward/backward compatibility; job replay/idempotency; restore validation | Timed rollback/kill-switch exercise; on-call/escalation roster; runbook approval | EM-01, Operations, all engineers, QA | **PLANNED / BLOCKED** on approved production platform and named owners. Required for Production Candidate. |
| **PKG-001** | One root Compose command starts and health-checks every application currently implemented in the repository; [D-007](../AI-Log/DECISIONS.md#d-007--use-root-docker-compose-as-the-local-startup-entry-point) | Reviewer cannot reproduce the local demo or assumes missing services are included | `docker compose config --quiet`; `docker compose up --build -d --wait`; health and handler smoke tests | Follow root README walkthrough; inspect `docker compose ps` | ENG-01, QA | **DEMO PASS:** root Compose built and health-checked SQL Server, Northwind mock, Go API, and Next.js; the automation profile also validates. Not production-hosting evidence. |

## Test layers and ownership

| Layer | Primary purpose | Minimum contents | Evidence owner |
|---|---|---|---|
| Unit/domain | Prove deterministic parsing, comparison, state, and policy functions | Exact money, normalization, allowed transitions, stale calculation, retry classification | Implementing engineer |
| Adapter contract | Detect Northwind request/response drift | Reviewed schemas plus optional/unknown/error/pagination/timeout cases against mock and sandbox | ENG-01 + QA |
| SQL repository/integration | Prove durable, tenant-safe state | Real SQL Server migrations, constraints, transactions, concurrency, outbox/inbox, encrypted fields, rollback | ENG-02 + QA/Security |
| Go API/service | Prove authorization and workflow semantics | Tenant-negative tests, feature flags, response freshness, durable intent, error mapping | ENG-02/03 + QA |
| Next.js component/accessibility | Prove customer states without network ambiguity | Fresh/stale/error, lifecycle, confirmation, duplicate-click prevention, keyboard/screen-reader checks | ENG-04 + QA/Product |
| End-to-end | Prove cross-application sequencing | SQL snapshot → async mismatch → commit → invalidation; transfer intent → ambiguity/webhook/reconciliation | QA with all engineers |
| Security | Prove trust boundaries and protected-data handling | AuthZ/IDOR, ingress authenticity, secret redaction, encryption/key/backup/retention controls | Security + QA |
| Operational | Prove the service can be detected, contained, and recovered | Load boundary, outage, alert, rollback, kill switch, replay, reconciliation, restore, escalation | Operations + EM/QA |

## Human-readable acceptance scenarios

### A-01 — Externally changed recent transaction

**Given** an authorized customer has a last-known SQL transaction snapshot and Northwind contains one changed or new transaction,  
**when** the customer requests recent transactions,  
**then** the API returns the tenant-scoped SQL snapshot immediately with freshness metadata, schedules at most one bounded refresh, commits the normalized difference, publishes invalidation only after commit, and the frontend re-fetches through the Vantaca API to show the new snapshot.

### A-02 — Matching or unavailable partner data

**Given** the customer has a usable last-known snapshot,  
**when** Northwind either matches it or returns 429/500/503/times out,  
**then** an equal result records a successful check without invalidation, while failure preserves the snapshot, updates observable sync error/age, applies bounded safe retry policy, and displays Product-approved stale/unavailable messaging.

### A-03 — Ambiguous transfer submission

**Given** an authorized transfer intent is durable and the partner accepts the transfer but its response is lost,  
**when** the request times out,  
**then** Vantaca records `UNKNOWN / RECONCILIATION_REQUIRED`, performs no automatic resubmission, raises the required operational signal, and changes state only after exact authoritative correlation.

### A-04 — Duplicate and later webhook statuses

**Given** a transfer is `PENDING`,  
**when** the same delivery is retried and a later legitimate `POSTED` or `RETURNED` event arrives,  
**then** Vantaca idempotently acknowledges the retry, preserves event history, does not suppress the new status, applies only a valid authenticated transition, and converges with reconciliation.

### A-05 — Cross-tenant request

**Given** a user belongs to Tenant A,  
**when** the user requests or submits using an account linked to Tenant B,  
**then** the API reveals no account existence or financial data, performs no partner call, records the approved security/audit event, and returns the Vantaca-standard authorization response.

### A-06 — Protected-data recovery

**Given** approved encryption/key policy and a production-like backup,  
**when** the database is restored and keys are rotated according to policy,  
**then** authorized services can read the minimum required protected values, unauthorized principals cannot, old/new key versions behave as designed, and logs/test evidence expose no full values or key material.

## Evidence package requirements

Every Production Candidate evidence link must identify:

- requirement/test ID and exact expected outcome;
- source revision/build identifier and configuration;
- environment, SQL migration version, Northwind schema/mock version, and feature-flag state;
- automated command/result or manual procedure, timestamp, and reviewer;
- sanitized logs/metrics/traces/screenshots with correlation IDs;
- defect or deviation link, accountable risk owner, disposition, and expiration where applicable.

“Tested,” a screenshot without build identity, or a passing mock test does not satisfy a production row.

## Current executable evidence

From the repository root:

```powershell
Push-Location Demo/api
gofmt -l ./cmd ./internal
go test -count=1 ./...
go vet ./...
Pop-Location

Push-Location Demo/mock/northwind
gofmt -l ./cmd ./internal
go test -count=1 ./...
go vet ./...
Pop-Location

Push-Location Demo/web
npm ci
npm test -- --run
npm run typecheck
npm run build
Pop-Location

docker compose config --quiet
docker compose --profile automation config --quiet
docker compose up --build -d --wait
Invoke-RestMethod http://localhost:18080/healthz
Invoke-WebRequest http://localhost:13000 -UseBasicParsing
Invoke-RestMethod http://localhost:8081/healthz
docker compose down
```

Current Go/API and mock tests cover exact money, SQL timestamp comparison, adapter authentication/pagination/error/retry behavior, durable transfer/service behavior, deterministic partner failures, transfer transitions, webhook retry/duplicates, and concurrent unique mock IDs. Frontend tests cover exact money formatting and status presentation; its typecheck and production build also pass. Record these as **DEMO PASS** only after the commands and end-to-end walkthrough pass on the reviewed revision.

## Release exit rules

- **Integration MVP:** AUTH-001 plus the enabled read-path rows have environment-identified automated evidence; transfer rows may remain blocked only when the transfer flag is unavailable/off.
- **Pilot:** enabled rows have production-like evidence; operational alerts/runbooks and explicit tenant/volume caps are exercised.
- **Production Candidate:** every enabled row is **RELEASE PASS**, or a non-stop-ship deviation has a named accountable approver and time-bounded remediation.
- **Production GA:** no stop-ship row is blocked, rollback/kill-switch evidence is current, and Engineering, Product, Security, QA, and Operations complete the human go/no-go.

Transfer submission must remain disabled while TRF-002, WH-002, or REC-001 lacks a safe approved control. Account and transaction access must remain disabled while AUTH-001 is unresolved. The date does not override either rule.
