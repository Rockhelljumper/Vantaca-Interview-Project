# Open Questions

## Purpose and routing

The committed target is a production-ready go-live in two weeks. This register is organized to protect Northwind call time and expose the shortest responsible critical path.

Questions are routed as follows:

- **Internal Vantaca — async:** Decisions Vantaca can and should make without Northwind in the room.
- **Northwind — live architecture call:** Only cross-company questions whose answers materially change the integration architecture or determine whether a safe launch is possible.
- **Northwind — async written follow-up:** Precise contract artifacts and operational details that should be answered in writing rather than consuming live design time.
- **Future stability/scalability:** Important work that may follow launch only when the two-week design has a safe extension point and an explicit interim control.

## Ranking

- **B0 — Stop-ship:** No production go-live without an answer and an approved implementation/control. Needed in the first 1–2 business days because it can change architecture.
- **B1 — Release blocker:** Work may proceed behind an explicit assumption, but the answer and evidence are required before the production-readiness review.
- **F1 — First hardening priority:** Important to near-term stability or scale; may be deferred only with a measured interim limit and named owner.
- **F2 — Planned evolution:** Useful future improvement that should not complicate the two-week critical path.

Rank reflects production impact, not preferred meeting order.

# 1. Two-Week Production Go-Live Blockers

## 1A. Internal Vantaca — Async Decision Queue

These do not need Northwind call time.

| Rank | Vantaca area | Question or decision | Why it matters in two weeks | Required outcome / owner |
|---|---|---|---|---|
| **B0-1** | Product / Risk | What balance mismatch is Vantaca willing to expose when customers or third parties can change the account outside Vantaca? What stale threshold, as-of label, refresh behavior, and unavailable-state UX are acceptable? | A cached balance cannot be guaranteed to match Northwind continuously. Calling it live or authoritative would be misleading. | Written freshness/staleness policy and acceptance examples. Product owns UX/risk acceptance; QA makes it testable. Needed Day 1. |
| **B0-2** | Engineering / Product | Which existing Vantaca identity, tenant, linked-account, consent, and entitlement mechanisms authorize account views and transfers? Is step-up authentication or approval required? | Northwind account scoping does not replace Vantaca authorization. Incorrect binding can expose data or move funds for the wrong customer. | Named system of record, authorization rules, and integration owner. Engineering/Product/Security. Needed Day 1–2. |
| **B0-3** | Product / Integration Lead | If Northwind cannot provide safe transfer idempotency and exact lookup immediately, is the approved fallback to launch account/transaction reads while keeping transfer submission disabled? | Blindly retrying an ambiguous monetary request is unsafe. The launch needs a pre-agreed scope reduction rather than a last-day argument. | Explicit feature-gate decision and customer/stakeholder communication. Product and Integration Lead. Needed Day 1. |
| **B0-4** | Security / Compliance | What compensating controls, if any, are acceptable for query-string API credentials, unsigned webhooks, and storage of full account/routing numbers? | These are known partner-contract weaknesses with direct financial-data and state-integrity impact. | Security sign-off or stop-ship criteria; approved secret, encryption, ingress, masking, logging, and retention controls. Needed Day 1–2. |
| **B0-5** | Engineering / Security | What is the production path for webhook ingress, secrets, SQL Server, durable jobs/inbox processing, identity, and network controls? | A local Compose topology is not a production deployment plan. | Named approved platform components and owners. n8n is not required for base launch. Needed Day 2. |
| **B1-1** | Product | Precisely what is in the two-week launch: linked accounts, balances, transaction window, transfers, status history, returns, notifications, and support UX? | Acceptance and safe deferral require a fixed scope. | Signed launch-scope list and feature-flag plan. Needed Day 1. |
| **B1-2** | Product / QA | What does recent transactions mean, and what are the acceptance examples for freshness, stale data, transfer confirmation, FAILED, RETURNED, webhook delay, and Northwind outage? | Engineers otherwise choose customer-visible financial semantics implicitly. | Executable acceptance matrix. Needed Day 2–3. |
| **B1-3** | Operations / Product | Who owns launch-day support, reconciliation exceptions, UNKNOWN transfers, returned transfers, partner escalation, rollback, and the transfer kill switch? | Production-ready includes an operable failure path, not only a working request. | Named on-call/escalation owners, minimum runbooks, dashboards, and kill-switch authority. Needed before readiness review. |
| **B1-4** | QA / Engineering | What evidence is required for release: contract tests, MSSQL integration tests, failure cases, load boundary, security review, and rollback/recovery exercise? | A two-week timeline needs risk-based exit criteria to prevent testing from becoming open-ended or superficial. | Agreed release checklist and daily evidence tracking. QA owns coordination; engineering owns technical evidence. Needed Day 2. |
| **B1-5** | Data / Security | Can Vantaca avoid persisting full account numbers? If not, what is the approved data owner, access policy, encryption/key design, retention, deletion, and audit approach? | The partner request expands sensitive-data scope and affects schema and implementation immediately. | Documented data-handling decision before schema freeze. Needed Day 2–3. |

## 1B. Northwind — Focused Live Architecture Call

The live call should seek five architectural outcomes. Detailed field lists, IP ranges, schemas, and schedules should be requested as written follow-up.

### Recommended 45-minute agenda

1. **0–5 minutes:** Confirm the launch scope and decisions needed from the call.
2. **5–15 minutes:** Balance meaning, freshness, and multi-channel change propagation.
3. **15–25 minutes:** Transfer idempotency and ambiguous-outcome recovery.
4. **25–35 minutes:** Webhook delivery, retries, ordering, and authenticity.
5. **35–42 minutes:** Customer/account scoping and supported production boundary.
6. **42–45 minutes:** Confirm written follow-ups, owners, and deadlines.

| Rank | Architecture topic and questions | Required outcome | If unanswered |
|---|---|---|---|
| **B0-1** | **Safe transfer submission:** Does POST /transfers accept a durable idempotency key or client reference? If the request commits but the response is lost, how do we find exactly that transfer without resubmitting it? Is there an exact GET-by-ID/client-reference lookup? | A documented, testable method for exactly correlating submission attempts and resolving ambiguous timeouts. | Do not enable production transfer submission; launch read-only if internally approved. |
| **B0-2** | **Customer and account scope:** How does each request identify the customer/link when the documented GET /accounts request contains only a partner key and page number? What are the supported linking, consent, revocation, and account-ownership rules? | Supported identity/scoping flow and stable identifiers for customer, link, and account. | No production account data access or transfers. |
| **B0-3** | **Webhook identity and delivery:** Is `transfer_id` globally unique, immutable, and safe to use as the consumer idempotency/deduplication key? Because one transfer may have legitimate status changes, is the intended event key `transfer_id`, `(transfer_id, status)`, or a separate event/delivery ID? Do retries preserve that key and payload? Can events be duplicated or arrive out of order, and what event/effective timestamp, retry schedule, response timeout, replay ability, and status ordering exist? How can Vantaca authenticate a webhook now—signing, mTLS, fixed source ranges, or another control? | A documented, testable webhook identity rule—explicitly confirming or rejecting `transfer_id` as sufficient—plus retry/order semantics and a verifiable ingress control. | Do not assume webhook idempotency is absent or that `transfer_id` alone is sufficient. Webhooks cannot independently advance trusted customer-visible state until identity, transition, and ingress controls are approved; polling/reconciliation may replace them only with Security approval. |
| **B0-4** | **Balance semantics and freshness:** Is balance ledger, current, or available? Does it include pending activity? What precision and source as-of timestamp exist? When a customer or third party changes the account outside Vantaca, how quickly is the change visible through the API? What consistency applies across paginated reads? Is five-second polling actually the supported integration pattern, and what quota/throttling behavior constrains it? | A documented balance definition and supported freshness/synchronization contract that works for multi-channel account activity. | Vantaca can show only a clearly labeled fetched-at snapshot; Product must approve the mismatch/stale behavior or balance display does not launch. |
| **B0-5** | **Production contract boundary:** What production and sandbox endpoints/credentials exist? Can credentials be environment-specific, scoped, and rotated? Which endpoints are formally supported? What routing field semantics, account combinations, status transitions, and return behavior apply to transfers? | A supportable public production contract and test environment. The undocumented internal endpoint is excluded unless formally documented. | No production credentialing or transfer launch; adapter work proceeds only against documented behavior. |

### Northwind call discipline

Do not ask Northwind to decide:

- How much balance mismatch Vantaca is willing to show.
- Vantaca UI wording, stale indicators, or refresh controls.
- Whether Vantaca accepts product/reputational risk.
- Vantaca authorization, tenant, data-retention, logging, deployment, n8n, alerting, or on-call design.
- Whether Vantaca should reduce launch scope when a stop-ship item remains open.

Those are internal Vantaca decisions.

## 1C. Northwind — Async Written Follow-Up for the Two-Week Launch

Request these in a single written checklist immediately after the architecture call.

| Rank | Requested artifact or precise answer | Why written |
|---|---|---|
| **B0-6** | Sandbox and production base URLs, environment-specific credential issuance/rotation procedure, webhook source ranges or other ingress setup, and named integration escalation contacts. | Configuration and security details must be exact and auditable. |
| **B0-7** | Updated OpenAPI/JSON schemas for accounts, transactions, transfers, errors, and webhooks, including optional/nullable fields and unknown enum behavior. | Contract tests should derive from a reviewed artifact, not meeting notes. |
| **B0-8** | Pagination termination and stable-ordering rules for every list endpoint, including behavior when records change between pages. | Correct synchronization depends on precise behavior. |
| **B0-9** | Transfer rules: supported currencies/precision, min/max/daily limits, routing-number meaning, allowed account combinations, cutoffs/holidays/fees, failure/return codes, and complete status-transition table. | These are implementation and acceptance inputs better reviewed in tabular form. |
| **B1-6** | Actual 429 quotas/burst rules and Retry-After semantics, request timeout guidance, maintenance behavior, SLA/SLO, and support escalation. | Capacity and timeout settings require exact numbers. |
| **B1-7** | Transaction definition of recent, default ordering, date/incremental filters, mutation/correction behavior, MCC optionality, and retention. | Needed for a bounded, correct launch scope. |
| **B1-8** | Written confirmation that /internal/accounts/full is supported, authorized, versioned, secured, and covered by production support—or confirmation that it is not available to partners. | An informal endpoint must not enter the critical path. Default is not to use it. |

# 2. Future Stability and Scalability

These questions may be deferred only when the two-week launch uses bounded limits, feature flags, observable interim behavior, and a design that does not prevent the future answer.

## 2A. Internal Vantaca — Async Future Queue

| Rank | Vantaca area | Question or issue | Safe interim posture |
|---|---|---|---|
| **F1-1** | Product / Operations | What long-term balance freshness SLO and error budget should the integration meet by customer/account segment? | Launch with explicit fetched-at/stale states, bounded polling, and freshness telemetry. |
| **F1-2** | Engineering / Capacity | What are current and forecast customer, account, transaction, transfer, and webhook volumes and peak patterns? | Configure conservative concurrency/rate limits and pilot/launch caps; measure actual usage. |
| **F1-3** | Engineering / Platform | Which durable queue/job platform should eventually own high-volume sync, webhook processing, and reconciliation? | Use a repository-bounded database inbox/job pattern if approved; keep application interfaces replaceable. |
| **F1-4** | Operations | What are the steady-state SLOs, paging thresholds, reconciliation cadence, support staffing, and partner-incident process? | Start with essential metrics/runbooks and tune from observed production behavior. |
| **F1-5** | Data / Compliance | What are long-term retention, archival, deletion, audit, and legal-hold rules for transactions, transfers, events, and operational payloads? | Minimize stored raw payloads and use conservative documented retention until policy is approved. |
| **F1-6** | Reliability | What RPO/RTO, regional failure, replay, backup/restore, and disaster-recovery requirements apply? | Prove backup/restore and local replay for the initial production topology; document limits. |
| **F2-1** | Operations / Engineering | Is n8n an approved long-term operational dependency, and who owns upgrades, workflow review, credentials, history, and on-call? | Keep n8n optional; base production behavior remains in Go and works without it. |
| **F2-2** | Architecture | Do future bank partners justify a generalized connector framework, or should Northwind remain an explicit adapter behind narrow interfaces? | Use narrow internal ports and Northwind-specific DTOs; do not build a speculative framework. |

## 2B. Northwind — Async Future/Roadmap Questions

| Rank | Northwind roadmap question or issue | Why it matters |
|---|---|---|
| **F1-7** | Will Northwind provide cursor/delta account and transaction synchronization, bulk reads, or balance-change events so polling does not scale per customer? | Enables freshness without continuous full scans. |
| **F1-8** | What is the committed roadmap for signed webhooks, event replay, delivery dashboards, and longer event retention? | Reduces compensating controls and improves recovery. |
| **F1-9** | Will Northwind support OAuth/client credentials, mTLS, scoped/rotatable secrets, and environment isolation? | Reduces credential exposure and operational blast radius. |
| **F1-10** | Can transfers use opaque account tokens/IDs instead of full account/routing numbers? | Reduces sensitive-data persistence and access. |
| **F1-11** | What long-term quotas, fair-use rules, capacity-notification process, and partner-specific scaling reviews exist? | Needed before expanding customer volume or freshness targets. |
| **F1-12** | What formal compatibility, schema versioning, deprecation period, change notification, and certification process applies to /v1 and webhook schemas? | Prevents silent contract drift. |
| **F2-3** | Are transfer-status bulk lookup, exact reconciliation exports, webhook replay, or settlement reports available? | Improves high-volume reconciliation and incident recovery. |
| **F2-4** | Can Northwind provide authoritative source timestamps and both ledger and available balances? | Improves customer communication and transfer eligibility decisions. |

# 3. Working Assumptions Pending Answers

These assumptions permit design work; they are not production approvals.

1. Northwind remains the authoritative financial system. Vantaca stores a read model, not a new source of financial truth.
2. A Vantaca snapshot can differ from Northwind because customers and third parties can change the account through other channels.
3. Vantaca will store and display its own fetched-at time; that is not equivalent to a Northwind source as-of timestamp.
4. Ambiguous transfer submissions are never automatically retried unless Northwind provides proven idempotency/correlation.
5. Webhook idempotency semantics are unverified. `transfer_id` is a candidate consumer key, not assumed insufficient or sufficient; processing tolerates retries without discarding legitimate later status changes until Northwind confirms whether the key is `transfer_id`, `(transfer_id, status)`, or a separate event ID.
6. Only documented public endpoints are used; /internal/accounts/full is excluded by default.
7. Polling is bounded and configurable. One workflow execution per customer every few seconds is not an acceptable production design.
8. Read-only account functionality and transfer-write functionality can be independently feature-gated, subject to Product approval.

# 4. Resolved Questions and Directed Constraints

- **Timeline meaning:** The Integration Lead clarified that the two-week target means production-ready and live, not a demo or pilot.
- **Technology baseline:** Go backend; Next.js, TypeScript, and Tailwind frontend; SQL Server 2022; database/sql with parameterized raw SQL; no ORM unless explicitly reconsidered.
- **Local partner dependency:** A deterministic Northwind mock is required for development and end-to-end validation, but it is not production evidence by itself.
- **Automation posture:** n8n is optional and orchestration-only; Go retains banking/domain behavior.
- **AI harness routing:** Meaningful post-discovery AI work uses the initialization evaluator, one primary specialty harness, minimum necessary context, current runtime-verified model availability, and bounded escalation.
- **Recent-transactions read path:** Serve the SQL Server snapshot with freshness metadata, asynchronously compare it with Northwind, commit differences before invalidating the frontend, and force the frontend to re-fetch through the Vantaca API only when values changed.
- **Current authorization:** The synthetic interview demo is authorized and implemented under `Demo/`, including the Vantaca Go application, SQL Server schema/read model, Next.js frontend, Northwind mock, and optional n8n trigger. Production deployment, real credentials/data, and real-money enablement remain unauthorized and gated by the B0/B1 questions above.
