# Security and Data Classification

## Purpose and approval boundary

This artifact consolidates the integration's data classes and expected handling across transport, storage, masking, access, audit, and retention. It gives Engineering and QA one review surface for [ENG-02](../EngineeringTasks/02-mssql-account-transaction-sync.md), [ENG-03](../EngineeringTasks/03-transfers-webhooks-reliability.md), and [DATA-001](QA-Acceptance-Matrix.md).

The classifications below are **recommendations for Security/Data-owner review**, not declarations of Vantaca policy or regulatory scope. The B0-4/B1-5 decisions in [OPEN-QUESTIONS.md](../AI-Log/OPEN-QUESTIONS.md) remain authoritative and block schema freeze for protected values.

## Proposed classification levels

| Level | Meaning in this integration | Examples |
|---|---|---|
| **Secret** | Credential or key material that grants system access or decrypts protected data | Northwind API credential, webhook verification secret/certificate key, encryption key material |
| **Restricted Financial** | Data that can expose financial activity, enable fraud, or materially harm a customer | Full account number, routing details tied to an account, balances, transactions, transfer instructions/status/history |
| **Confidential** | Non-public identifiers and relationship data that require tenant and role controls | Vantaca tenant/customer/link IDs, Northwind account/transfer IDs, correlation mappings |
| **Internal Operational** | Service metadata that is non-public but should not contain financial payloads or credentials | Job status, sanitized metrics, deployment metadata, runbook evidence |
| **Synthetic** | Deliberately fictional data approved for local tests, demos, AI inputs, and screenshots | `acc_1029` and other documented mock fixtures |

Classification follows the most sensitive combined context. For example, a routing number may be publicly derivable in isolation, but a routing number tied to a customer's full account and transfer instruction is **Restricted Financial**.

## Data-handling matrix

“Pending policy” means Engineering must not invent a production retention period, cipher/key topology, or access rule merely to close the schedule.

| Data asset | Source and necessary use | Proposed class | Storage / encryption at rest | Encryption in transit | Masking, telemetry, and AI use | Access and audit | Retention / deletion | Unresolved approval |
|---|---|---|---|---|---|---|---|---|
| **Northwind API credential** | Issued by Northwind; adapter authentication | **Secret** | Approved secrets manager only; never SQL, source, image, fixture, or ordinary environment file; encrypt with platform-managed keys | Vantaca-approved TLS to Northwind; never emit credential in redirect or error | Redact query strings at proxy, HTTP client, APM, logs, traces, alerts, and support tooling; never send to AI | Runtime identity and narrowly approved operators only; audit read/rotation without recording value | Rotate on policy/incident/personnel/environment change; delete on revocation/decommission | Northwind environment scoping, header/stronger auth, issuance/rotation; Security compensation for query parameter |
| **Webhook verification secret, certificate, or equivalent** | Future Northwind webhook trust control | **Secret** | Approved secrets/certificate service; application stores reference/version, not raw key in SQL | Vantaca-approved TLS/mTLS as contracted | Never log header/key/signature input beyond sanitized algorithm/key ID and verification result; no AI use | Webhook verifier runtime and key administrators; audit change/rotation | Policy-driven overlap and revocation window | Northwind authenticity capability and Vantaca-approved compensating control |
| **Encryption key material** | Vantaca key service; protects restricted SQL fields | **Secret** | Never stored in application tables, backups, repository, or container image; managed/HSM-backed service per policy | Protected platform channel and workload identity | No logs, dumps, fixtures, screenshots, support payloads, or AI input | Key administrators and authorized workload only; audit use/admin events; separate duties | Versioned rotation, revocation, and recovery per policy | Approved key platform, rotation/recovery/RPO design, owner |
| **Full account number** | Northwind/authorized customer context; only if partner transfer call cannot use an opaque account ID | **Restricted Financial** | **Prefer not to persist.** If unavoidable and approved: separate secure-values boundary, field/application-layer ciphertext plus key version, approved database/backup encryption, no plaintext index; database encryption alone is not sufficient protection from broad DB access | Vantaca-approved TLS on every hop; never browser-to-Northwind | Render last four only; redact request/response bodies, SQL diagnostics, exceptions, traces, screenshots, exports, fixtures, and AI inputs | Dedicated least-privilege service path; no routine analyst/support access; tenant binding; audit decrypt/read/change and privileged access | Minimum operational window; delete/anonymize on unlink or policy trigger where legally/operationally allowed; prove backup expiration | Whether persistence is allowed at all, data owner, key mechanism, exact retention/deletion, partner token alternative |
| **Routing number associated with customer account/transfer** | Transfer instruction to Northwind | **Restricted Financial** | Do not retain separately unless required. If stored with account/transfer, use the same approved protected-value boundary and backup control | Vantaca-approved TLS | Mask in broad telemetry/AI/support evidence; display only where Product/Security explicitly requires | Transfer service and minimum operational roles; tenant scope; audit protected reads/changes | Align to approved transfer/account retention and deletion | Whether later resubmission/reconciliation requires storage; partner field semantics |
| **Masked account identifier / last four** | Derived for display and support | **Confidential** | SQL Server with approved database/backup encryption; do not make it globally unique authorization proof | Vantaca-approved TLS between browser/API/services | Safe for authorized UI and sanitized test evidence when combined context remains synthetic; avoid broad public logs | Tenant-authorized users/services; support access per role; audit sensitive support lookup as policy requires | Link/account lifecycle plus approved support/audit period | Product display rules and whether last four plus context elevates handling |
| **Balance plus observed/fetched timestamps** | Northwind → SQL read model → Vantaca API/UI | **Restricted Financial** | Tenant-scoped SQL tables/views with approved database/backup encryption; exact value type | Vantaca-approved TLS; browser receives only authorized tenant response | Never general analytics, URL, metric label, log, trace, screenshot, or AI input; telemetry carries age/status, not amount | Authorized customer/account service and limited support/data roles; audit privileged bulk access | Product/Data policy; minimize snapshots/history unless needed for audit/reconciliation | Ledger/current/available meaning, source as-of, stale threshold, history retention |
| **Transaction data** | Northwind → normalized recent-transaction read model | **Restricted Financial** | Tenant-scoped SQL with approved database/backup encryption; minimize raw payload and free text | Vantaca-approved TLS | Mask or synthesize merchant/description/amount in non-production evidence; no payload logs or AI input | Authorized tenant users/services; least-privilege support/data access; audit export/bulk access | Approved recent-window plus correction/audit requirement; deletion on unlink subject to policy | “Recent” window, partner retention/mutation behavior, Vantaca retention/legal hold |
| **Transfer intent, amount, accounts, status, reason, and history** | Vantaca authorization → Northwind → webhook/reconciliation | **Restricted Financial** | Durable tenant-scoped transfer/history tables; exact amounts; protected account references/values separated; approved database/backup encryption | Vantaca-approved TLS; webhook accepted only through approved ingress | Correlation/request IDs may be logged; no full financial payload/account values in logs, metrics, screenshots, or AI | Transfer service, authorized customer/support/operations roles; immutable/append-only history where approved; audit authorization, submission, protected access, transition, manual action | Financial/audit/support policy; retain only needed payload fields; defined deletion/legal-hold behavior | Approval/step-up/limits, full partner lifecycle, data owner, retention, manual-action controls |
| **Webhook receipt / raw payload** | Northwind ingress → durable inbox → transfer processor | **Restricted Financial** when payload contains transfer data | Store minimum fields and body hash/verification metadata. Persist raw body only if Security/Operations approves; encrypt it and isolate from general queries/backups | Vantaca-approved TLS plus signing/mTLS/equivalent; reject or quarantine unverifiable events | Log event/correlation ID, verification outcome, transition result, and payload hash—not body/account/amount; no AI input | Webhook processor and narrowly approved incident/replay role; audit replay/quarantine/manual disposition | Short, explicit replay/forensic window; purge raw payload sooner than normalized history when possible | Authenticity, stable event identity, replay needs/window, exact retention |
| **Tenant, customer, consent/link, partner account/transfer mappings** | Vantaca identity/linking and Northwind correlation | **Confidential**, elevated when joined to financial records | Tenant-scoped relational tables with approved database/backup encryption and uniqueness constraints | Vantaca-approved TLS | Use opaque correlation values; avoid URL/query exposure and public telemetry; synthetic replacements in AI/tests | Identity/link/account/transfer services; tenant-scoped repository views; audit link/revoke/admin changes | Consent/link lifecycle plus approved audit period; remove access immediately on revoke | Vantaca system of record, consent/revocation flow, Northwind customer scoping |
| **Application logs, metrics, traces, alerts, and support evidence** | All runtime boundaries | **Internal Operational** only if redaction succeeds; otherwise inherits highest leaked class | Approved observability platform encryption and access controls; prohibit arbitrary financial payload capture | Vantaca-approved TLS/agent transport | Allow opaque correlation, status, duration, retry count, sync age, error class; prohibit query credentials, full URLs, account/routing numbers, amount/balance, transaction descriptions, raw bodies | Engineering/Operations/Support least privilege; audit search/export/admin based on platform policy | Shortest period meeting operations/security policy; delete incident exports/captures on schedule | Approved platform, redaction standard, access roles, retention, incident-evidence handling |
| **Sync jobs, inbox/outbox, reconciliation metadata** | Go workers and SQL reliability patterns | **Internal Operational** or **Confidential**; payload columns inherit source class | SQL/database encryption; prefer IDs, hashes, versions, outcome metadata over copied financial payloads; protected payload encrypted if unavoidable | Vantaca-approved TLS between worker/platform/database | Log correlation/outcome/age/count, not copied business payload | Worker identity and operations/replay roles; audit manual replay, cancellation, override | Bounded retry/replay window; purge completed operational rows after approved evidence period | Job platform, replay ownership, retention, failure/dead-letter handling |
| **Synthetic mock/test/demo data** | Repository fixtures and local mock | **Synthetic** | May be committed only when obviously fictional and not derived from production; local `.env` remains ignored | Local HTTP only as documented; use TLS in shared test environments per platform rules | Approved for logs, screenshots, and bounded AI use when no secret or production context is mixed in | Development/test access; protect any shared environment credentials separately | Retain with tests; rotate if a value could be confused with real credential/data | Security review of fixture-generation rules and test-environment boundaries |

## Baseline implementation rules

These controls are the proposed minimum; Security may strengthen them.

1. **Minimize first.** Prefer Northwind opaque IDs/tokens over full account/routing values. Do not persist raw partner payloads merely because they were received.
2. **Separate protected values.** Keep encrypted full values out of customer-facing read views, general repository queries, logs, and analytics. Store ciphertext and key version separately from ordinary account display fields.
3. **Use managed keys and workload identity.** Application code must not contain keys. Rotation, restore, and historical-key availability must be designed and tested together.
4. **Protect backups and replicas.** At-rest control includes backups, restores, replicas, exports, snapshots, local developer copies, and incident evidence—not only the primary database.
5. **Authorize before lookup or partner call.** Tenant/account binding is enforced in the service and repository boundary. A Northwind identifier alone never grants access.
6. **Treat inbound webhooks as untrusted.** Authenticate, bound request size/time, durably record the minimum receipt, protect against replay, validate transition, and reconcile before trusted state changes.
7. **Redact before emission.** Build structured allowlisted telemetry. Never rely only on downstream log scrubbing, especially because the partner credential is documented in a query string.
8. **Use synthetic evidence.** Production payloads and credentials do not enter source control, local fixtures, screenshots, support tickets, or AI prompts.
9. **Fail closed for protected operations.** Key/identity/authenticity failures prevent decrypt, transfer submission, or trusted event processing; a stale read may remain available only under Product-approved semantics.
10. **Audit human power.** Record privileged decrypt/read/export, key/policy change, transfer override, webhook replay, reconciliation disposition, feature-flag/kill-switch use, and retention/deletion action without recording the protected value.

## Trust-boundary review

| Boundary | Minimum control and evidence |
|---|---|
| Browser → Next.js/Go | Vantaca identity/session, tenant authorization, CSRF/request protections as applicable, TLS, no direct Northwind calls, response-cache policy for financial data |
| Go adapter → Northwind | Approved TLS, narrowly scoped environment credential, query redaction, explicit timeout/retry rules, egress restriction where supported, correlation without payload logging |
| Northwind → webhook ingress | Signing/mTLS/approved equivalent, source/size/time limits, replay handling, durable receipt, transition validation, no trust based only on `transfer_id` |
| Go services/workers → SQL Server | Workload identity, least-privilege roles, parameterized SQL, tenant-scoped repository/view controls, field/database/backup encryption per policy, audit |
| SQL/outbox → invalidation transport → frontend | Publish only after commit; opaque tenant/query key; no financial payload; authenticated service path; bounded replay and observability |
| Operations/support/AI → evidence | Role-restricted sanitized evidence, synthetic examples, audited exports, no credential/full financial payload, time-bounded retention |

## Required security evidence

Before Production Candidate, attach evidence for:

- approved classification, data owner, data-flow/threat-model review, and any accepted compensation;
- Vantaca identity/tenant/consent enforcement, including cross-tenant negative tests;
- Northwind credential issuance, storage, use, redaction, rotation, revocation, and environment separation;
- webhook authenticity, replay, event identity, ingress configuration, and failure behavior;
- SQL role grants, tenant views/queries, parameterization, protected-field ciphertext, key versioning, rotation, and unauthorized decrypt denial;
- backup encryption, restore with required keys, recovery access, and expiration/deletion behavior;
- structured telemetry allowlist plus automated scans for credentials, account/routing values, amounts/balances, and raw payloads;
- privileged access/change audit records and alerting for abnormal protected-data access;
- retention/deletion jobs for each persisted class, including inbox/outbox/raw receipts and restored copies;
- dependency/container/CI/CD/platform security review and production secret-injection evidence;
- transfer feature gate and kill-switch access, audit, and table-top exercise.

Use the [QA matrix](QA-Acceptance-Matrix.md#requirement-to-evidence-matrix) to record exact build/environment evidence. A design statement or encrypted primary database alone does not pass DATA-001.

## Decisions still required

The following are deliberately not invented here:

- whether Vantaca may persist full account/routing values at all;
- Vantaca's official classifications and applicable regulatory/control framework;
- approved secrets, encryption-key, SQL, observability, ingress, and workload-identity platforms;
- exact algorithms/key rotation/recovery approach and authorized roles;
- retention, deletion, archive, legal-hold, and production-data-in-lower-environment policy;
- acceptable webhook compensation if Northwind cannot authenticate events;
- whether Northwind can replace protected values with opaque tokens.

Security/Data owners must record these outcomes against B0-4/B1-5 before protected production persistence or transfer submission is enabled.
