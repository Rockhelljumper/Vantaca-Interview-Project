# Concerns

## C-001 — Ambiguous monetary submission

Concern: POST /transfers has no documented idempotency key/client reference or deterministic lookup after an ambiguous timeout.
Source: Northwind API guide, Transfers and Errors; omission from the documented contract.
Category: Financial integrity / API contract
Severity: Critical
Why it matters: Retrying can create a duplicate transfer; not retrying can leave a customer-facing operation unresolved. GET /transfers alone does not provide a safe correlation mechanism.
Recommendation: Require partner-supported idempotency and exact lookup/correlation. Model an internal UNKNOWN state and never blindly retry a timed-out submission.
Required owner/stakeholder: Northwind, Product, Integration Lead
Blocks demo?: No, if the mock demonstrates the risk and the limitation is explicit.
Blocks production?: Yes
Status: Open

## C-002 — Undefined customer/account scoping

Concern: GET /accounts claims to return accounts linked to the customer, but requests contain only a shared partner key and page number.
Source: Northwind API guide, Authentication and Accounts.
Category: Authorization / tenancy / contract
Severity: Critical
Why it matters: The contract does not explain whose accounts are returned or how cross-customer data exposure is prevented.
Recommendation: Obtain the supported customer identity, consent/linking, tenant-isolation, and revocation contract before detailed design.
Required owner/stakeholder: Northwind, Vantaca Engineering, Security, Product
Blocks demo?: No, with deterministic synthetic identities and a prominently documented assumption.
Blocks production?: Yes
Status: Open

## C-003 — Unauthenticated financial-status webhooks

Concern: Northwind does not sign webhook requests and asks the integration to trust payloads.
Source: Integration thread, Dana item 3; Northwind API guide, Webhooks.
Category: Security / financial integrity
Severity: Critical
Why it matters: A forged or replayed request could falsely change transfer state. IP allowlisting alone is a weaker, operationally fragile compensating control.
Recommendation: Require cryptographic signing, mTLS, or an approved equivalent; use defense in depth, replay protection, durable receipt, and reconciliation.
Required owner/stakeholder: Northwind, Vantaca Security, Operations, Integration Lead
Blocks demo?: No, if synthetic and clearly labeled.
Blocks production?: Yes, unless Security explicitly approves adequate compensating controls.
Status: Open

## C-004 — Contradictory webhook delivery guarantee

Concern: The guide says webhooks are delivered exactly once but also says failed deliveries are retried.
Source: Northwind API guide, Webhooks.
Category: Reliability / contract contradiction
Severity: High
Why it matters: Retries make repeated delivery attempts possible, but the guide does not say whether `transfer_id` or another stable identifier makes those attempts safely deduplicable. Event ordering and replay behavior are also undocumented.
Recommendation: Obtain a precise partner contract for event identity, retries, and ordering. Explicitly ask whether `transfer_id` is globally unique and safe as the consumer idempotency key, whether legitimate status changes require `(transfer_id, status)`, or whether Northwind provides a separate event/delivery ID. Design processing to tolerate retries without dropping valid later transitions until this is confirmed.
Required owner/stakeholder: Northwind, Backend Engineering, QA
Blocks demo?: No
Blocks production?: Yes until a defensible processing contract/control exists.
Status: Open

## C-005 — Unsafe credential transport and environment model

Concern: The API key is sent in query parameters, issued once, and shared across all environments.
Source: Northwind API guide, Authentication.
Category: Security / secrets
Severity: Critical
Why it matters: Query strings are commonly captured in logs and telemetry; shared credentials increase blast radius and weaken environment isolation, rotation, and incident response.
Recommendation: Request scoped, rotatable, environment-specific credentials and header-based auth or stronger controls; redact URLs at every telemetry boundary.
Required owner/stakeholder: Northwind, Vantaca Security, Platform Engineering
Blocks demo?: No; use synthetic local credentials only.
Blocks production?: Yes unless Security approves a controlled integration pattern.
Status: Open

## C-006 — Rate-limit contradiction

Concern: The guide says there are no rate limits while documenting 429 and Retry-After behavior.
Source: Northwind API guide, Rate limits and Errors.
Category: Contract / scalability
Severity: High
Why it matters: The proposed five-second per-customer polling model cannot be capacity-planned and can cause throttling cascades or stale data.
Recommendation: Obtain real quotas; use bounded batching, concurrency, jitter, backoff, caching, and freshness-driven synchronization.
Required owner/stakeholder: Northwind, Integration Lead, Operations
Blocks demo?: No
Blocks production?: Yes for sizing and reliability approval.
Status: Open

## C-007 — Availability claims contradict documented failures

Concern: Northwind and its stakeholder describe the API as always available and discourage downtime handling, while the guide documents 500, 503, and maintenance.
Source: Northwind API guide, introduction and Errors; integration thread, Dana.
Category: Reliability / stakeholder contradiction
Severity: High
Why it matters: Omitting timeouts, backoff, circuit protection, stale-state UX, reconciliation, and runbooks would make partner incidents customer incidents.
Recommendation: Treat failures as normal integration behavior and obtain measurable SLA/SLO and maintenance/support details.
Required owner/stakeholder: Northwind, Product, Operations, Integration Lead
Blocks demo?: No
Blocks production?: Yes until resilience criteria are approved and tested.
Status: Open

## C-008 — Five-second polling and local-source-of-truth request

Concern: Dana requests polling every linked customer every five seconds and treating the stored balance as the customer-visible source of truth.
Source: Integration thread, Dana item 1.
Category: Architecture / scalability / product correctness
Severity: High
Why it matters: Cost grows with customers, conflicts with unknown throttling, and still cannot guarantee real time. A cache can be stale and should not become authoritative over the bank.
Recommendation: Define a freshness SLO, use an as-of read model with stale indicators, batch/adapt polling, refresh on demand where safe, and reconcile against Northwind as authority.
Required owner/stakeholder: Product, Northwind, Integration Lead, Operations
Blocks demo?: No
Blocks production?: Yes until scale and freshness behavior are agreed.
Status: Open

## C-009 — Full account-number persistence request

Concern: The partner requests storing full account and routing numbers for later transfer calls.
Source: Integration thread, Dana item 2; Northwind transfer contract.
Category: Sensitive data / data minimization
Severity: High
Why it matters: It expands breach impact, audit scope, access-control needs, and retention obligations; the supported API offers no tokenized alternative.
Recommendation: Ask Northwind to accept opaque account IDs/tokens. If full values are unavoidable, obtain Security approval and use field-level encryption, managed keys, strict access, masking, audit, and retention limits.
Required owner/stakeholder: Security/Compliance, Northwind, Data owner
Blocks demo?: No; use synthetic values only.
Blocks production?: Yes pending data-handling approval.
Status: Open

## C-010 — Undocumented internal endpoint

Concern: Product suggests /internal/accounts/full based on an informal statement, but it is absent from the public partner contract.
Source: Integration thread, Priya.
Category: Supportability / contract drift
Severity: High
Why it matters: It may bypass intended authorization, versioning, rate, security, and support guarantees and can disappear without notice.
Recommendation: Do not depend on it until Northwind supplies a supported written contract and production authorization.
Required owner/stakeholder: Northwind, Product, Integration Lead
Blocks demo?: No; use the public documented contract.
Blocks production?: Yes if the proposed design depends on it.
Status: Open

## C-011 — Imprecise money representation

Concern: Monetary amounts and balances are JSON numbers with no precision/rounding contract.
Source: Northwind API guide, Accounts, Transactions, and Transfers.
Category: Data integrity
Severity: High
Why it matters: Binary floating-point handling can alter monetary values, and currency minor units are not universally two decimals.
Recommendation: Parse losslessly, validate precision by currency, use a decimal/minor-unit domain value and SQL decimal with explicit rules, and reject invalid values.
Required owner/stakeholder: Northwind, Backend Engineering, QA
Blocks demo?: No
Blocks production?: Yes until precision rules are contract-tested.
Status: Open

## C-012 — Incomplete pagination and transaction schema

Concern: Bare-array pagination has no terminal/total/cursor metadata or ordering guarantee; the transaction example and table disagree about merchant_category_code.
Source: Northwind API guide, Pagination and Transactions.
Category: Contract completeness / data synchronization
Severity: Medium
Why it matters: Concurrent changes can cause missed/duplicated rows, exactly-full final pages are ambiguous, and a nominally documented field may be optional or absent.
Recommendation: Clarify termination, stable ordering, incremental sync, optionality, and schema compatibility; make adapter parsing tolerant and contract-tested.
Required owner/stakeholder: Northwind, Adapter Engineering, QA
Blocks demo?: No
Blocks production?: Potentially, for complete/correct synchronization.
Status: Open

## C-013 — Undefined balance semantics

Concern: A single balance field is labeled current, with no available/ledger distinction or as-of timestamp.
Source: Candidate brief, Northwind API guide, and integration thread.
Category: Product semantics / data correctness
Severity: High
Why it matters: Customers may interpret displayed funds as available to transfer, and the system cannot accurately communicate staleness.
Recommendation: Obtain exact semantics and source timestamp; display type and as-of freshness and do not infer transfer eligibility.
Required owner/stakeholder: Northwind, Product, QA
Blocks demo?: No, with an explicit assumption.
Blocks production?: Yes for customer-facing financial correctness.
Status: Open

## C-014 — Incomplete transfer lifecycle contract

Concern: The status list lacks allowed transitions, reason codes, event timestamps, cancellation behavior, and ordering semantics; RETURNED can occur after POSTED.
Source: Northwind API guide, Transfers and Webhooks.
Category: Domain model / operations
Severity: High
Why it matters: Treating POSTED as permanently final would be wrong, and support/reconciliation cannot explain or safely handle later reversals.
Recommendation: Define an explicit internal state machine that preserves partner history, permits valid late returns, rejects/regards regressions, and reconciles authoritative state.
Required owner/stakeholder: Northwind, Product, Backend Engineering, Operations, QA
Blocks demo?: No
Blocks production?: Yes until lifecycle behavior and acceptance criteria are agreed.
Status: Open

## C-015 — Two-week production target compresses unresolved gates

Concern: The Integration Lead clarified that the executive-promised two-week date means a production-ready go-live while critical partner, security, correctness, and operational contracts remain unresolved.
Source: Integration thread, Priya; Integration Lead clarification on 2026-08-16.
Category: Delivery / governance
Severity: Critical
Why it matters: The target is now clear, but schedule clarity is not risk acceptance. Customer/account scoping, ambiguous transfer submission, webhook trust, balance semantics, security controls, release evidence, and operational ownership can change architecture or block launch.
Recommendation: Run a two-week critical-path plan with B0 stop-ship decisions due in the first 1–2 business days, B1 evidence due before readiness review, independent read/write feature gates, daily risk review, and explicit accountable owners. Defer only work with a safe interim control and extension point.
Required owner/stakeholder: Product, Integration Lead, Executive sponsor, Security, Operations, Northwind
Blocks demo?: No
Blocks production?: Yes until the ranked B0/B1 blockers are resolved; the clarified date does not waive them.
Status: Open — timeline clarified; feasibility and production gates remain at risk

## C-016 — Required discovery inputs are absent

Concern: PrePromptHistory.md and the original Ai-Engineering guidance/examples are still absent. The original prompt named Ai-Engineering with a hyphen; the Integration Lead later explicitly directed creation of Ai_Engineering with an underscore. The latest explicit path was created, but no pre-existing README or agent-harness examples existed to inherit.
Source: Original operating prompt, Integration Lead instruction on 2026-08-16, and repository inventory.
Category: Process / evidence completeness
Severity: Medium
Why it matters: Prior decisions or upstream engineering conventions may still exist elsewhere and could not be incorporated. The project-specific harnesses therefore follow the Integration Lead's supplied format rather than claiming compatibility with unavailable examples.
Recommendation: Treat Ai_Engineering as the current project-local path. If upstream guidance or PrePromptHistory.md is later restored, compare it with AI-Harnesses/ and surface material differences as DECISION REQUIRED before changing the harnesses.
Required owner/stakeholder: Integration Lead
Blocks demo?: No
Blocks production?: No by itself, but blocks claiming the mandated discovery inputs were fully reviewed.
Status: Partially mitigated — requested local directory and harness set created; historical/upstream guidance remains unavailable

## C-017 — Mock fidelity can be mistaken for partner certainty

Concern: The local Northwind mock must fill limited implementation details that the public guide leaves undefined, including customer scoping, money precision, routing-number meaning, and deterministic status control.
Source: Northwind API guide omissions; `Demo/mock/northwind` implementation.
Category: Contract testing / environment fidelity
Severity: Medium
Why it matters: Passing mock tests can create false confidence if locally chosen assumptions are treated as confirmed Northwind production behavior.
Recommendation: Keep mock-only controls outside /v1, document every assumption, preserve contract-focused tests, and revise the mock when Northwind provides supported written schemas and semantics. Do not treat mock success as partner certification.
Required owner/stakeholder: Integration Lead, Northwind, Backend Engineering, QA
Blocks demo?: No
Blocks production?: Yes if the mock is the only contract evidence
Status: Open — controlled through explicit documentation and isolated mock-only behavior
