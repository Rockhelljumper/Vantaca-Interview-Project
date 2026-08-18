# Changelog

## 2026-08-16 — Establish discovery governance logs

### What changed

Created the required AI session history, repository changelog, decision record, open-question register, and concern register.

### Why

The Integration Lead's operating prompt requires durable, reviewable records of AI-assisted analysis and explicitly permits AI documentation changes during discovery.

### Files affected

- AI-Log/SESSION-HISTORY.md
- AI-Log/CHANGELOG.md
- AI-Log/DECISIONS.md
- AI-Log/OPEN-QUESTIONS.md
- AI-Log/CONCERNS.md

### Tests performed

- Verified the three source documents were present and readable.
- Verified that no application or infrastructure files were created.
- Validated required question sections and concern fields and checked that no credential-like Northwind example key was copied into the logs.
- No executable tests were applicable.

### Result

Discovery records were created successfully. Two inputs named by the operating prompt remain missing from the authorized workspace: AI-Log/PrePromptHistory.md and Ai-Engineering/README.md (with the entire in-project Ai-Engineering directory absent).

### Architectural impact

None. The records distinguish lead-directed constraints, recommendations, unresolved decisions, and production blockers; they do not authorize implementation.

## 2026-08-16 — Re-rank questions for a two-week production go-live

### What changed

- Reorganized OPEN-QUESTIONS.md into internal Vantaca async decisions, a focused Northwind live architecture agenda, Northwind written follow-ups, and future stability/scalability queues.
- Ranked questions as B0 stop-ship, B1 release blocker, F1 first hardening priority, or F2 planned evolution.
- Recorded the Integration Lead's clarification that two weeks means production-ready go-live.
- Updated the delivery concern to distinguish a clarified deadline from acceptance of unresolved production risk.

### Why

The prior register mixed internal decisions, partner questions, immediate launch blockers, and longer-term topics. The new structure protects Northwind call time, keeps Vantaca-owned risk decisions internal, and focuses the two-week critical path on architectural outcomes.

### Files affected

- AI-Log/OPEN-QUESTIONS.md
- AI-Log/DECISIONS.md
- AI-Log/CONCERNS.md
- AI-Log/CHANGELOG.md
- AI-Log/SESSION-HISTORY.md

### Tests performed

- Validated required routing, ranking, balance, webhook, transfer, future-stability, resolved-question, and concern sections.
- Checked that no credential-like Northwind example key was copied into the logs.

### Result

Questions are now separated by answer owner, communication channel, launch criticality, and deferability. No application artifacts were created or modified.

### Architectural impact

The two-week production target is now the planning baseline. Northwind remains authoritative; Vantaca's tolerance for stale/mismatched cached balances is an internal decision. Read and transfer-write capabilities should remain independently gateable.

## 2026-08-16 — Add reusable post-discovery AI harnesses

### What changed

- Created the Integration Lead-requested Ai_Engineering root with a README documenting the absence of pre-existing local guidance/examples.
- Created a lightweight AI-Harnesses set for architecture, Go backend, database, frontend, integration reliability, QA, security, peer review, and EM delivery.
- Added an initialization evaluator, shared model-routing guide, usage matrix, minimum-context contract, handoff protocol, bounded escalation policy, and loop prevention.
- Updated AI governance records with the routing decision, resolved constraint, remaining source-guidance limitation, and session history.

### Why

Post-discovery work needs reusable specialization without duplicating the master prompt, loading unrelated context, automatically spawning agents, or assigning the strongest model to routine tasks.

### Files affected

- Ai_Engineering/README.md
- AI-Harnesses/README.md
- AI-Harnesses/MODEL-ROUTING.md
- AI-Harnesses/initialization-evaluator.md
- AI-Harnesses/architecture-harness.md
- AI-Harnesses/go-backend-harness.md
- AI-Harnesses/database-harness.md
- AI-Harnesses/frontend-harness.md
- AI-Harnesses/integration-reliability-harness.md
- AI-Harnesses/qa-harness.md
- AI-Harnesses/security-harness.md
- AI-Harnesses/peer-review-harness.md
- AI-Harnesses/em-delivery-harness.md
- AI-Log/DECISIONS.md
- AI-Log/OPEN-QUESTIONS.md
- AI-Log/CONCERNS.md
- AI-Log/CHANGELOG.md
- AI-Log/SESSION-HISTORY.md

### Tests performed

- Verified all expected files exist.
- Validated every specialty harness and the evaluator contain the required role, objective, inspection, principles, checks, outputs, prohibitions, handoff, and Model Profile sections.
- Validated all Model Profile fields and ensured every named model ID is exposed by the current Codex runtime.
- Checked the harness set for credential-like Northwind example material.
- Reviewed the routing/evaluator behavior for minimum context, no automatic delegation, one-step escalation, and human decision ownership.
- No application tests were run because no application code was created or modified.

### Result

The reusable harness set is complete and structurally validated. One initial evaluator-format gap found during validation was corrected and the full validation passed.

### Architectural impact

None on the application. The harnesses are advisory overlays that augment the primary instructions. Current routing prefers gpt-5.6-luna for bounded low-risk work, gpt-5.6-terra for balanced implementation/reasoning, and gpt-5.6-sol for high-risk architecture, financial, security, reliability, and adversarial review.

## 2026-08-16 — Create the Start Here orientation

### What changed

- Replaced the empty Notes/StartHere.md with a concise orientation derived from the ranked open-question and decision registers.
- Documented that Notes/Questions-Comments.md is currently empty so no informal notes were silently inferred.
- Added focused sections for current mode/target, directed architecture, internal decisions, the Northwind call, transfer safety, production gating, deferrable work, and next actions.
- Added Mermaid system-context, ambiguous-transfer, and production-launch-gate diagrams.

### Why

The project needed a reviewer-friendly starting point that summarizes the critical path without duplicating the full question and concern registers.

### Files affected

- Notes/StartHere.md
- AI-Log/CHANGELOG.md
- AI-Log/SESSION-HISTORY.md

### Tests performed

- Validated three paired Mermaid fences and supported diagram declarations.
- Validated sequence alt/else/end balance, participant declarations, flowchart node definitions, and relative Markdown links.
- Confirmed current mode, two-week target, transfer-safety rule, and empty source-note status are present.
- Checked for credential-like Northwind example material.
- Mermaid CLI is not installed locally, so validation was structural rather than a rendered CLI compilation.

### Result

StartHere.md is populated, internally linked, and structurally validated. It points readers to authoritative logs for detail rather than creating a second source of truth.

### Architectural impact

None. The diagrams visualize existing recommendations and directed constraints; they do not approve implementation or resolve open production blockers.

## 2026-08-16 — Implement the Northwind Connect Go mock

### What changed

- Added a standalone Go 1.25 module under mock/northwind.
- Implemented query-key authentication, accounts, paginated transactions, transfer creation/listing, documented error shapes, precise USD money handling, and structured path-only request logging.
- Added deterministic 429, 500, 503, latency, and post-commit-timeout scenarios.
- Added mock-only transfer-state controls, bounded webhook retries, duplicate webhook delivery, and valid transition enforcement.
- Added synthetic seed data matching the documented account, transaction, and first transfer ID.
- Added configuration documentation, .env.example, multi-stage non-root Docker packaging, and tests.
- Updated authorization, decision, concern, Start Here, changelog, and session records.

### Why

The Integration Lead explicitly authorized a local Go API that emulates the supplied Northwind Connect contract so development and failure behavior can run without proprietary partner infrastructure.

### Files affected

- mock/northwind/go.mod
- mock/northwind/cmd/server/main.go
- mock/northwind/cmd/server/main_test.go
- mock/northwind/internal/mockapi/models.go
- mock/northwind/internal/mockapi/store.go
- mock/northwind/internal/mockapi/server.go
- mock/northwind/internal/mockapi/webhook.go
- mock/northwind/internal/mockapi/server_test.go
- mock/northwind/README.md
- mock/northwind/.env.example
- mock/northwind/Dockerfile
- mock/northwind/.dockerignore
- Notes/StartHere.md
- AI-Log/DECISIONS.md
- AI-Log/OPEN-QUESTIONS.md
- AI-Log/CONCERNS.md
- AI-Log/CHANGELOG.md
- AI-Log/SESSION-HISTORY.md

### Tests performed

- gofmt completed with no remaining unformatted Go files.
- go test -count=1 -cover ./... passed: cmd/server 53.4%, internal/mockapi 76.7%.
- go vet ./... passed.
- go build ./cmd/server passed; the generated server.exe validation artifact was then removed from the workspace.
- go test -race could not run because CGO is disabled and no GCC toolchain exists in the Windows environment.
- Docker image vantaca-northwind-mock:local built successfully; its Linux build stage reran all normal tests.
- Container smoke test passed for health, 401 auth failure, deterministic accounts, first transfer creation, and 429/Retry-After.
- The temporary smoke container was stopped and removed.

### Result

The mock API is runnable locally or as a non-root container and its contract, reliability, configuration, and concurrency tests pass. No Vantaca application integration was added.

### Architectural impact

The mock is a standalone development/test dependency behind the future Northwind adapter. Mock-only controls are excluded from /v1. Its assumptions remain explicitly non-authoritative and do not close Northwind production blockers.

## 2026-08-16 — Add root Docker Compose orchestration

### What changed

- Added root `compose.yaml` as the single startup entry point for all currently implemented applications.
- Added root `.env.example` with synthetic overrides and `.gitignore` protection for local `.env` values.
- Added an executable health-check mode to the Go server so the distroless non-root image needs no shell or HTTP utility.
- Added image-level and Compose-level health-check configuration.
- Updated the mock README and Start Here guide with the one-command startup, inspection, configuration, and shutdown workflow.
- Recorded the root-orchestration decision.

### Why

The Integration Lead requested that the mock API run as a Docker container and that one root Compose file start all applications together.

### Files affected

- compose.yaml
- .env.example
- .gitignore
- mock/northwind/Dockerfile
- mock/northwind/cmd/server/main.go
- mock/northwind/cmd/server/main_test.go
- mock/northwind/README.md
- Notes/StartHere.md
- AI-Log/DECISIONS.md
- AI-Log/CHANGELOG.md
- AI-Log/SESSION-HISTORY.md

### Tests performed

- gofmt completed successfully.
- go test -count=1 -cover ./... passed: cmd/server 50.6%, internal/mockapi 76.7%.
- go vet ./... passed.
- docker compose config --quiet passed and the fully resolved configuration was inspected.
- An isolated `docker compose up --build -d --wait` smoke run passed on host port 18082.
- Docker reported the container healthy; health and authenticated account requests returned the expected values.
- The isolated Compose container and network were removed, and cleanup was verified.

### Result

`docker compose up --build -d --wait` from the repository root now builds and starts the complete currently implemented local stack. The Northwind mock is available on host port 8081 by default.

### Architectural impact

Root Compose is now the local/demo orchestration boundary. Only the Northwind mock is currently authorized and implemented; this does not authorize or define production deployment for the future Vantaca API, UI, database, or n8n services.

## 2026-08-16 — Clarify recent-transactions reconciliation and webhook identity

### What changed

- Expanded StartHere.md with the recent-transactions architectural decision and an async reconciliation Mermaid sequence.
- Corrected the stale claim that Questions-Comments.md was empty and clarified authoritative-document precedence.
- Defined the initial read as a SQL Server snapshot with freshness metadata.
- Kept Northwind comparison asynchronous and made frontend invalidation conditional on a successfully committed mismatch update.
- Clarified match, Northwind-failure, SQL-failure, and frontend refresh behavior.
- Reframed the Northwind webhook question to determine whether `transfer_id`, `(transfer_id, status)`, or a separate event ID is the safe consumer idempotency key.
- Distinguished webhook-processing idempotency from transfer-submission idempotency.
- Recorded the architectural decision and aligned the concern and open-question registers.

### Why

The Integration Lead directed a fast SQL-backed recent-transactions read with async convergence and asked that webhook idempotency be investigated rather than assumed absent.

### Files affected

- Notes/StartHere.md
- AI-Log/OPEN-QUESTIONS.md
- AI-Log/CONCERNS.md
- AI-Log/DECISIONS.md
- AI-Log/CHANGELOG.md
- AI-Log/SESSION-HISTORY.md

### Tests performed

- Validated Markdown fence pairing and Mermaid block count.
- Validated sequence participant references and alt/else/end balance.
- Checked authoritative project documents for categorical claims that webhook idempotency is absent.
- Confirmed the question retains the separate unresolved `POST /transfers` idempotency requirement.

### Result

The local-read/async-reconciliation behavior is explicit, including update-before-invalidate ordering. Northwind is now directly asked to confirm the usable webhook identity key, with `transfer_id` treated as a candidate rather than rejected by assumption.

### Architectural impact

The design adds an asynchronous reconciliation and frontend invalidation boundary for recent transactions. It does not select the job or notification technology and does not authorize application implementation.

## 2026-08-16 — Add detailed explanations for every major architecture workflow

### What changed

- Retained all six Major Workflow Mermaid diagrams in StartHere.md.
- Added purpose, numbered execution steps, consistency rules, and failure behavior for account/balance sync, recent transactions, ACH submission, webhook processing, transfer reconciliation, and Northwind recovery.
- Restored Major Workflow 2 to the directed SQL-first asynchronous reconciliation design.
- Made the Workflow 2 mismatch branch commit SQL changes before invalidating and force-refreshing the frontend.
- Aligned StartHere webhook language with the open question about whether `transfer_id`, `(transfer_id, status)`, or a separate event ID is the supported consumer idempotency key.
- Added a documentation boundary requiring the workflow architecture to remain in StartHere.md or be relocated intact to `Notes/Architecture.md` with a clear link.

### Why

The Integration Lead requested prose that explains Major Workflow 2 and confirmation that every remaining workflow chart is understandable without relying on Mermaid alone. The Integration Lead also explicitly prohibited removal of the architecture workflows unless they are preserved in a detailed Architecture.md document.

### Files affected

- Notes/StartHere.md
- AI-Log/DECISIONS.md
- AI-Log/CHANGELOG.md
- AI-Log/SESSION-HISTORY.md

### Tests performed

- Confirmed all six Major Workflow sections and diagrams remain present.
- Confirmed every workflow contains a detailed `How it works` explanation of at least 75 prose words.
- Validated 12 Mermaid blocks, six sequence diagrams, participant references, and branch balance.
- Validated Markdown fence pairing and relative documentation links.

### Result

Each architecture workflow now explains its purpose, sequence, alternate behavior, and safety posture in text. Workflow 2 remains asynchronous and refreshes the frontend only after a changed Northwind snapshot is committed to SQL Server.

### Architectural impact

No new runtime component was selected. The documentation now preserves the recorded async reconciliation design and makes all workflow decisions independently reviewable from the charts.

## 2026-08-16 — Create role-based engineering GitHub issue drafts

### What changed

- Added `EngineeringTasks/` with an index and six local GitHub issue drafts.
- Created one outcome-based issue for the Integration Lead, each of four engineering workstreams, and QA/PM.
- Linked each issue to one primary AI harness and the initialization evaluator.
- Added scope, authorization boundaries, acceptance criteria, required testing/evidence, dependencies/handoffs, and out-of-scope guidance to every issue.
- Preserved engineer freedom over internal structure, naming, decomposition, and implementation approach within approved architecture.
- Linked the issue index from StartHere's Delivery Approach.
- Recorded the role/harness issue-brief decision.

### Why

The Integration Lead requested straightforward but sufficiently detailed local GitHub issues that direct each owner and name the applicable AI harness without over-prescribing implementation.

### Files affected

- EngineeringTasks/README.md
- EngineeringTasks/00-integration-lead-delivery-readiness.md
- EngineeringTasks/01-northwind-adapter-and-mock.md
- EngineeringTasks/02-mssql-account-transaction-sync.md
- EngineeringTasks/03-transfers-webhooks-reliability.md
- EngineeringTasks/04-nextjs-customer-experience.md
- EngineeringTasks/05-qa-pm-acceptance-release.md
- Notes/StartHere.md
- AI-Log/DECISIONS.md
- AI-Log/CHANGELOG.md
- AI-Log/SESSION-HISTORY.md

### Tests performed

- Confirmed seven Markdown files: one index and six issue drafts.
- Confirmed required metadata and Goal, Scope, Acceptance Criteria, Testing, Dependencies, and Out of Scope sections in every issue.
- Confirmed each issue has at least five acceptance checks and five testing/evidence checks.
- Validated every primary harness, evaluator, task-index, decision, question, and StartHere link.
- Scanned task files for credential-like material.

### Result

The repository now contains copy-ready GitHub issue drafts for all roles in the delivery map. The tasks describe results and proof, while leaving implementation choices to the assigned owner.

### Architectural impact

None. The files organize already documented responsibilities and decisions; they do not authorize application implementation or alter runtime architecture.

## 2026-08-16 — Add architecture logic and sample data to every task issue

### What changed

- Added an Architecture Workflow Logic section to all six role-based GitHub issue drafts.
- Added role-appropriate synthetic JSON samples to every issue.
- Expanded ENG-02 with representative Northwind account/transaction input and a masked customer read-model projection.
- Added conceptual layouts for six SQL tables, four supporting views/queries, required indexes/constraints, and the transactional read-model outbox.
- Added encryption-policy requirements covering data minimization, separated ciphertext, managed keys, access restriction, masking, rotation, restore, audit, and fail-closed behavior.
- Expanded ENG-02 acceptance and SQL Server integration tests for plaintext exclusion, authorized/unauthorized access, key rotation/outage, views, sample mapping, and tenant isolation.
- Updated the shared issue rules and D-010 documentation standard.

### Why

The Integration Lead requested secured-data expectations, sample inbound data, tables/views, and workflow logic for ENG-02, then extended the sample-data and architecture-logic requirement to every role issue.

### Files affected

- EngineeringTasks/README.md
- EngineeringTasks/00-integration-lead-delivery-readiness.md
- EngineeringTasks/01-northwind-adapter-and-mock.md
- EngineeringTasks/02-mssql-account-transaction-sync.md
- EngineeringTasks/03-transfers-webhooks-reliability.md
- EngineeringTasks/04-nextjs-customer-experience.md
- EngineeringTasks/05-qa-pm-acceptance-release.md
- AI-Log/DECISIONS.md
- AI-Log/CHANGELOG.md
- AI-Log/SESSION-HISTORY.md

### Tests performed

- Confirmed all six issues contain architecture-workflow and sample-data sections with synthetic JSON.
- Validated balanced code fences and all relative links.
- Confirmed any issue containing full synthetic account/routing values labels them as synthetic/protected boundary data.
- Confirmed ENG-02 documents six logical tables, four supporting views/queries, encryption-policy controls, and expanded security/data tests.

### Result

Every role now has a concrete example of the data it owns or consumes and a textual mapping to the relevant architecture workflow. ENG-02 has enough logical schema/security direction to begin detailed design once the encryption policy is approved, without hardcoding an unapproved encryption product or physical schema.

### Architectural impact

The documented logical data boundaries are more explicit, especially the separation of encrypted account/routing values from customer read views and the update-plus-outbox atomicity requirement. Exact SQL types, object names, encryption technology, and key platform remain implementation/policy decisions.

## 2026-08-16 — Add application placement and dependency architecture to task briefs

### What changed

- Added a shared Mermaid runtime/application dependency map to EngineeringTasks/README.md.
- Mapped six customer/operational features to their entry points, owning runtime boundaries, application dependencies, and outputs.
- Documented directed, external, unresolved Vantaca platform, and optional dependencies separately.
- Added an Architecture Placement and Application Dependencies section to all six role issues.
- Clarified that the Northwind adapter and domain/workers are modules in one Go deployable, the mock is local-only, Next.js never calls Northwind/SQL directly, and SQL Server owns durable state/read models.
- Named identity, secrets, encryption keys, invalidation, ingress, scheduling, observability, CI/CD, and operations as explicit platform dependencies rather than implied components.
- Updated StartHere and D-010 to point reviewers to the dependency design.

### Why

The Integration Lead requested that the task documentation show where each application/feature is used and what other applications/platform capabilities it depends on.

### Files affected

- EngineeringTasks/README.md
- EngineeringTasks/00-integration-lead-delivery-readiness.md
- EngineeringTasks/01-northwind-adapter-and-mock.md
- EngineeringTasks/02-mssql-account-transaction-sync.md
- EngineeringTasks/03-transfers-webhooks-reliability.md
- EngineeringTasks/04-nextjs-customer-experience.md
- EngineeringTasks/05-qa-pm-acceptance-release.md
- Notes/StartHere.md
- AI-Log/DECISIONS.md
- AI-Log/CHANGELOG.md
- AI-Log/SESSION-HISTORY.md

### Tests performed

- Validated the dependency Mermaid declaration, fence pairing, and presence of eleven required runtime/platform components.
- Confirmed six feature-placement rows and architecture/dependency sections in all six role issues.
- Confirmed every issue links back to the shared map and includes at least four explicit dependency relationships.
- Revalidated all task-local relative links and StartHere discoverability.

### Result

Reviewers and owners can now see both the end-to-end application graph and each role's exact placement, incoming/outgoing contracts, dependencies, and consumers without inferring them from workflow charts.

### Architectural impact

The logical boundaries are clarified, not changed. Production platform technologies remain unresolved where explicitly labeled; the diagram does not imply that unimplemented services already exist or that local Compose defines production deployment.

## 2026-08-16 — Complete original-prompt reviewer and production-readiness artifacts

### What changed

- Added a root `README.md` five-minute entry point covering exercise scope, architecture, technology, startup, walkthrough, tests, decisions, limitations, and demo-versus-production status.
- Added `Notes/Delivery-Plan.md` with two complete EM matrices, 35–52 person-day range, dependency/parallelization analysis, risk ratings, production blockers, QA/observability/DoD, AI suitability, human-review levels, five delivery phases, and a two-week Gantt.
- Added `Notes/QA-Acceptance-Matrix.md` with 17 requirement → risk → test → evidence rows, acceptance scenarios, test layers, evidence metadata, and release exit rules.
- Added `Notes/Security-Data-Classification.md` with proposed classifications and per-data-class transit, at-rest, masking, access, audit, retention/deletion, and unresolved-approval requirements.
- Added the transfer lifecycle state diagram and explanatory state semantics alongside Major Workflow 3 without removing any architecture workflow.
- Cross-linked the new artifacts from StartHere and EngineeringTasks, and added targeted links to the EM, SQL, transfer, and QA issue briefs.
- Recorded D-011 so mock/planned/blocked/release evidence and proposed security policy cannot be mistaken for production approval.

### Why

The Integration Lead asked to close the remaining non-application gaps identified against the original operating prompt while preserving discovery facts, unresolved questions, and the distinction between a demo dependency and production readiness.

### Files created

- README.md
- Notes/Delivery-Plan.md
- Notes/QA-Acceptance-Matrix.md
- Notes/Security-Data-Classification.md

### Files modified

- Notes/StartHere.md
- EngineeringTasks/README.md
- EngineeringTasks/00-integration-lead-delivery-readiness.md
- EngineeringTasks/02-mssql-account-transaction-sync.md
- EngineeringTasks/03-transfers-webhooks-reliability.md
- EngineeringTasks/05-qa-pm-acceptance-release.md
- AI-Log/DECISIONS.md
- AI-Log/CHANGELOG.md
- AI-Log/SESSION-HISTORY.md

### Tests performed

- Verified the required root README sections and all five delivery phases.
- Checked relative links across the ten created/modified reviewer and task documents.
- Confirmed Major Workflows 1–6 remain and StartHere contains the new lifecycle source among 13 balanced Mermaid blocks.
- Structurally validated 16 Mermaid declarations across the reviewer, architecture, delivery, and engineering-index documents; Mermaid CLI was not installed, so no rendered-diagram validation was claimed.
- Confirmed 17 QA traceability rows and required security-control/data-class coverage.
- Ran `gofmt -l ./cmd ./internal`, `go test -count=1 ./...`, and `go vet ./...` successfully in `mock/northwind`.
- Ran `docker compose config --quiet` successfully and confirmed `northwind-mock` is the single currently implemented service.

### Result

The six identified original-prompt gaps are closed as documentation/evidence artifacts. Reviewers now have a reproducible five-minute entry, a candid two-week feasibility view, phase gates, traceable release tests, one security-control matrix, and the two missing visuals.

### Architectural impact

The transfer state model and delivery/evidence governance are more explicit, but no Vantaca runtime architecture or unresolved partner/security policy was invented. The mock remains the only implemented service, and production transfer submission remains gated by safe correlation, webhook trust, authorization, protected-data, reconciliation, and operational evidence.

## 2026-08-16 — Runnable Vantaca/Northwind interview demo and `Demo/` boundary

### What changed

- Created the top-level `Demo/` runtime boundary and moved the Go API, Northwind mock, and SQL assets into it.
- Added a Next.js/TypeScript/Tailwind interview UI for linked accounts, explicit freshness, recent transactions, external activity, partner failure recovery, and complete transfer-state demonstrations.
- Added the Go Vantaca API with Northwind adapter, exact minor-unit money, bounded safe-read retry, SQL read model, asynchronous transaction comparison, durable transfer intent, webhook inbox, and reconciliation.
- Added SQL Server 2022 migrations, tenant-scoped views, deterministic seed, read-model outbox, transfer history, and webhook inbox. Full account and routing values are intentionally absent from the schema.
- Extended the Go Northwind mock with external-transaction controls and boot-namespaced transfer IDs suitable for repeated container rebuilds with a retained SQL volume.
- Added an optional pinned n8n Compose profile whose workflow invokes one Go-owned reconciliation endpoint.
- Expanded root Compose to build, order, health-check, and start SQL Server, mock, API, and Next.js together.
- Added `Demo/scripts/e2e-smoke.ps1` and updated the root reviewer guide, StartHere current state, QA evidence statuses, task index, authorization statement, concerns path, and decision log.

### Reliability corrections from container-level testing

- Eliminated false read-model invalidations caused by SQL fixed-width MCC padding and sub-100-nanosecond timestamp differences.
- Preserved last-known account rows and marked freshness degraded when the partner account-list call exhausts retry.
- Mapped a partner acceptance followed by local state-update failure to durable `UNKNOWN` when possible, preventing an unsafe implied retry.
- Eliminated mock partner-ID collisions after image recreation while keeping the SQL demo volume persistent.

### Tests performed

- `gofmt`, `go test -count=1 ./...`, and `go vet ./...` passed in `Demo/api` and `Demo/mock/northwind`.
- `npm test -- --run`, `npm run typecheck`, and `npm run build` passed in `Demo/web`.
- `docker compose config --quiet` and `docker compose --profile automation config --quiet` passed.
- Sequential Docker builds passed for mock, API, and web; the four-service base stack reached healthy state.
- The live end-to-end smoke script passed 17 assertions against Next.js, Go, SQL Server, and the Northwind mock.
- After evidence collection, reset only the project-labeled synthetic SQL volume and restarted the healthy stack with clean seed data for the interview walkthrough.

### Result

Reviewers can start and exercise the complete synthetic integration from one root Compose command while executable assets remain cleanly separated under `Demo/`. The demo validates the chosen safety and consistency semantics but does not change any production blocker or authorize real data, credentials, hosting, or money movement.

## 2026-08-16 — Enforce the `Demo/` Docker-build boundary

- Changed the Vantaca API build context from the repository root to `Demo/`.
- Added `Demo/.dockerignore` to admit only `api/` and `database/` to that build.
- Updated API Dockerfile copy paths for the narrowed context.
- Confirmed no application/runtime source remains beside AI, Notes, Instructions, or EngineeringTasks directories.
- Rebuilt the API image and revalidated healthy API/UI responses.

## 2026-08-16 — Non-conflicting default demo ports

- Changed the default published API port from 8080 to 18080 and UI port from 3000 to 13000.
- Left internal service ports and container-to-container URLs unchanged.
- Updated environment examples, startup documentation, QA commands, and smoke-test defaults.
- Verified the exact `docker-compose up -d --wait` path with four healthy services.

## 2026-08-16 — Vantaca-aligned demo presentation

- Updated `Demo/web` to use Vantaca's public navy/blue/green visual language, Montserrat typography, official hosted SVG logo, and a more polished responsive product layout.
- Added an accessible reusable tooltip treatment to every architectural test action, with scenario-specific descriptions for reads, reconciliation, submissions, webhooks, failures, and returns.
- Improved narrow-viewport header, architecture-summary, and tooltip layout without changing the demo workflows.
- Added `@fontsource-variable/montserrat` as a deterministic UI dependency.
- Passed frontend tests, TypeScript checking, production build, Docker image rebuild, Compose health checks, HTTP checks, and desktop/narrow visual review.

## 2026-08-16 — API and browser acceptance automation

- Added Playwright configuration and scripts for independent API, Chromium UI, browser-install, and combined test runs.
- Added 15 live API tests covering boundaries, safe-read failures, SQL preservation/recovery, reconciliation, exact money, idempotency, webhooks, transfer lifecycles, ambiguous outcomes, and authorization.
- Added six live browser tests covering every dashboard workflow and scenario, all test-action tooltips, safety validation, production-boundary content, and 390-pixel responsive behavior.
- Added semantic UI testing hooks and status/alert roles.
- Corrected mobile horizontal overflow by allowing the recent-transactions Grid track and panels to shrink while retaining table-local scrolling.
- Updated reviewer test instructions and excluded generated Playwright reports/results from source and Docker build contexts.
- Passed 21/21 combined Playwright tests and the independent 17/17 assertion PowerShell smoke suite with all four containers healthy.

## 2026-08-16 — Closed-account external-activity guard

- Disabled the external-deposit demo action when the selected account is closed and added clear unavailable-state wording.
- Kept the disabled action's explanatory tooltip accessible from the keyboard.
- Added API-side account existence/status validation before invoking the Northwind mock, returning specific 404/422 errors.
- Added API and Chromium regression assertions for closed-account behavior.
- Passed all Go/API/frontend checks, rebuilt the affected containers, and passed the expanded 22/22 Playwright suite.

## 2026-08-16 — OpenAPI contracts and combined Swagger explorer

- Added embedded OpenAPI 3.0.3 documents for every core demo API and Northwind mock endpoint, including schemas, examples, security inputs, failure scenarios, and explicit demo-only control boundaries.
- Exposed the owning specifications at each service's unauthenticated `/openapi.yaml` route.
- Added a pinned Swagger UI Compose service at `http://localhost:18090` with a selector for both exact checked-in contracts and browser `Try it out` support.
- Restricted cross-origin API access to the configured local Swagger origin and documented that the demo CORS policy is not a production security design.
- Linked the explorer from the dashboard and updated the reviewer/runtime guides with raw-spec URLs and synthetic authorization values.
- Added parser-level OpenAPI validation, live spec/CORS/explorer API assertions, and a Chromium contract-switching test.
- Passed Go formatting/tests/vet for both services, Vitest, TypeScript, the Next.js production build, both Compose configuration modes, all 24 Playwright tests, and the independent 17-assertion smoke suite with five healthy base containers.
