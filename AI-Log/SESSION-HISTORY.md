# AI Session History

Material, user-visible interactions only. No hidden reasoning, credentials, or sensitive customer data are recorded here.

## 2026-08-16T15:04:56-05:00 — Initial discovery

### User instruction or question

The Integration Lead supplied an operating prompt for the Vantaca Northwind interview exercise. It requires an evidence-led discovery review, explicit separation of observations, recommendations, and decisions, creation of AI governance logs, and no application implementation until separately authorized.

### Visible response summary

- Inventoried the authorized workspace and located the three source documents by descriptive prefix.
- Read the candidate brief, Northwind API guide, and stakeholder thread in full.
- Reported that AI-Log/PrePromptHistory.md and the in-project Ai-Engineering directory/README named by the prompt are absent.
- Did not use a similarly named directory outside the stated Vantaca-Interview-Project boundary.
- Analyzed functional and non-functional requirements, partner and Product requests, contradictions, production blockers, proposed demo and production architecture, mock/database/n8n strategies, delivery ownership, estimates, visuals, repository structure, and prioritized questions.
- Identified transfer submission ambiguity, account/customer scoping, unsigned webhooks, authentication/environment handling, and the two-week production expectation as the leading risks.

### Decisions made

- No application architecture was silently approved; recommendations remain proposals for the Integration Lead.
- The session remained in DISCOVERY mode.
- The explicit technology constraints in the operating prompt are treated as lead-directed baselines unless the lead changes them.
- The undocumented Northwind endpoint is not included in the recommended design without a supported written contract.

### Questions raised

Prioritized questions were recorded in OPEN-QUESTIONS.md. The most urgent concern customer/account scoping, transfer idempotency and ambiguous outcomes, webhook authenticity and delivery semantics, supported environments/authentication, and what two-week go-live actually means.

### Files created

- AI-Log/SESSION-HISTORY.md
- AI-Log/CHANGELOG.md
- AI-Log/DECISIONS.md
- AI-Log/OPEN-QUESTIONS.md
- AI-Log/CONCERNS.md

### Files modified

- None outside the new AI log files.

### Commands/tests executed

- Enumerated repository files and top-level directories.
- Located source documents using descriptive filename prefixes.
- Read all 302 lines across the three supplied source documents.
- Computed source-document hashes for traceability.
- Checked for AGENTS.md, PrePromptHistory.md, Ai-Engineering, and Ai-Engineering/README.md.
- Checked Git status; the workspace is not currently a Git repository.
- Validated that every required log exists, required headings/fields are present, and no credential-like Northwind example key was copied into the logs.
- No application tests were run because implementation is not authorized and no application exists.

### Outcome

Discovery analysis completed from the available evidence, subject to the missing-input limitation. No application, database, mock, frontend, automation workflow, or broad repository structure was created.

## 2026-08-16T15:33:49-05:00 — Production critical-path question triage

### User instruction or question

The Integration Lead clarified that the two-week target means production-ready go-live and requested that open questions be split into internal Vantaca async questions, focused Northwind architecture questions, and future stability/scalability topics. Northwind call time should focus on freshness, webhook retries, account-balance semantics, and other architecture-changing issues.

### Visible response summary

- Separated Vantaca-owned risk/UX/platform decisions from questions Northwind must answer.
- Ranked production questions as B0 stop-ship or B1 release blockers and future questions as F1/F2.
- Reduced the Northwind live agenda to five architectural decision bundles with a 45-minute timebox.
- Moved precise schemas, limits, environment setup, and other auditable details to one written follow-up checklist.
- Explicitly separated Northwind's responsibility to define balance semantics/propagation from Vantaca's responsibility to accept and communicate mismatch/staleness.
- Preserved future extension points without placing speculative scalability work on the two-week critical path.

### Decisions made

- The two-week target is production-ready go-live.
- Deadline clarification does not waive unresolved financial-integrity, security, correctness, testing, or operational gates.
- Read-only account capabilities and transfer-write capabilities should be independently feature-gated if a stop-ship dependency remains unresolved.

### Questions raised

See OPEN-QUESTIONS.md for the ranked internal queue, Northwind call agenda, written follow-ups, and future roadmap questions.

### Files created

- None.

### Files modified

- AI-Log/OPEN-QUESTIONS.md
- AI-Log/DECISIONS.md
- AI-Log/CONCERNS.md
- AI-Log/CHANGELOG.md
- AI-Log/SESSION-HISTORY.md

### Commands/tests executed

- Reviewed the current question, decision, and delivery-concern records.
- Replaced the open-question register with the ranked/routed structure.
- Validated required sections and checked logs for credential-like Northwind example material.

### Outcome

The discovery question set now supports a fast production critical path without spending partner meeting time on Vantaca-owned decisions or silently deferring stop-ship risks. No application implementation was authorized or started.

## 2026-08-16T15:50:27-05:00 — Post-discovery harness specification recorded

### User instruction or question

The Integration Lead supplied a detailed specification for a future set of lightweight, reusable post-discovery AI harnesses and explicitly requested that SESSION-HISTORY.md be updated with new chat entries and kept congruent with the conversation.

The supplied future harness scope includes architecture, Go backend, database, frontend, integration reliability, QA, security, peer review, and EM delivery specialties; an initialization evaluator; shared model routing; context minimization; controlled escalation; and protection against recursive or wasteful sub-agent use. The specification also prohibits application-code changes as part of that future harness task.

### Visible response summary

- Read the attached harness specification in full.
- Reviewed the existing session history for chronological and factual congruency.
- Appended this interaction as a user-visible session record.
- Preserved the distinction between receiving the harness specification and being authorized to create its artifacts.

### Decisions made

- This turn is limited to updating SESSION-HISTORY.md because that is the explicit current request.
- The harness specification is recorded as prospective work; no AI-Harnesses files are created in this turn.
- Any future harness work must augment, not override, the primary project instructions.
- Future model recommendations must be based on models actually available at execution time, using the smallest reliable capability tier and escalating for financial, security, architectural, concurrency, or reliability risk.
- Humans remain responsible for architecture decisions, material risk acceptance, stakeholder decisions, and production approval.

### Questions raised

- Before future harness creation, the required in-project Ai-Engineering/README.md and Ai-Engineering/agent-harnesses examples must be restored to the authorized workspace or the Integration Lead must explicitly authorize the correct source location. They remain absent from Vantaca-Interview-Project.

### Files created

- None.

### Files modified

- AI-Log/SESSION-HISTORY.md

### Commands/tests executed

- Read the attached post-discovery harness specification.
- Read the current SESSION-HISTORY.md.
- Appended the new chronological entry.
- Validated the new entry's required headings and confirmed that no credential-like Northwind example key was introduced.

### Outcome

The session history is current through this interaction. No harness, application, database, frontend, mock, automation, or infrastructure artifact was created or modified.

## 2026-08-16T15:59:25-05:00 — Post-discovery AI harnesses created

### User instruction or question

The Integration Lead explicitly directed creation of Ai_Engineering in the Vantaca-Interview-Project root and authorized execution of the previously supplied post-discovery AI harness specification.

### Visible response summary

- Used the OpenAI Docs workflow to verify current official model-tier guidance before recommending model IDs.
- Verified that the current Codex runtime exposes gpt-5.6-sol, gpt-5.6-terra, gpt-5.6-luna, gpt-5.5, and gpt-5.4.
- Created Ai_Engineering and confirmed that it contained no pre-existing README or agent-harness examples.
- Read the current candidate/API/stakeholder instructions and AI decision, question, and concern registers.
- Created nine concise specialty harnesses, an initialization evaluator, a model-routing guide, and a harness README.
- Used Luna for low-risk bounded defaults, Terra for balanced work, and Sol for high-risk reasoning defaults, with runtime revalidation and one-step escalation.
- Kept harnesses advisory, context-minimized, non-recursive, and subordinate to primary instructions and human decisions.

### Decisions made

- Adopt evaluator-first, bounded harness routing as recorded in DECISIONS.md.
- Do not spawn a sub-agent when the primary session can complete a bounded task efficiently.
- Use one primary harness and only necessary supporting specialties.
- Treat model availability as runtime-specific; never invent or persist an unavailable model ID.
- Do not create a separate observability harness because reliability, QA, and EM delivery already own its relevant aspects.

### Questions raised

- Historical/upstream Ai-Engineering guidance and PrePromptHistory.md remain unavailable. If restored, their differences from the new harness set require review rather than silent adoption.

### Files created

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

### Files modified

- AI-Log/DECISIONS.md
- AI-Log/OPEN-QUESTIONS.md
- AI-Log/CONCERNS.md
- AI-Log/CHANGELOG.md
- AI-Log/SESSION-HISTORY.md

### Commands/tests executed

- Searched and opened current official OpenAI model documentation using the OpenAI Docs skill.
- Inventoried the requested Ai_Engineering root and the repository.
- Read all governing project/discovery records required by the harness specification.
- Verified expected files, required sections, Model Profile fields, runtime-supported model IDs, and absence of credential-like Northwind examples.
- Corrected an initially missing evaluator principles/checklist section and reran validation successfully.

### Outcome

The post-discovery harness system is complete and ready for review. No application, database, frontend, Northwind mock, n8n workflow, or infrastructure implementation was created or modified.

## 2026-08-16T16:08:52-05:00 — Start Here document created and Mermaid-checked

### User instruction or question

The Integration Lead requested cleanup of Notes/StartHere.md using Notes/Questions-Comments.md and AI-Log/OPEN-QUESTIONS.md, with Mermaid formatting corrected as needed.

### Visible response summary

- Read StartHere.md, Questions-Comments.md, OPEN-QUESTIONS.md, and DECISIONS.md.
- Confirmed that both note files were empty; no informal requirements were inferred from Questions-Comments.md.
- Populated StartHere.md as a concise orientation over the authoritative question, decision, and concern logs.
- Summarized current mode, two-week production target, directed architecture, internal Vantaca decisions, Northwind architecture-call focus, financial-safety rule, launch gates, deferrable work, and next actions.
- Added three Mermaid diagrams for system boundaries, ambiguous transfer submission, and production launch gating.
- Kept detailed questions in OPEN-QUESTIONS.md to reduce duplication and drift.

### Decisions made

- StartHere.md is a navigation/orientation artifact, not a new source of architectural decisions.
- Northwind remains authoritative; Vantaca's cached values are a timestamped read model.
- No new application implementation was authorized.

### Questions raised

- Questions-Comments.md remains empty and can be reconciled into the authoritative logs if working notes are added later.

### Files created

- None.

### Files modified

- Notes/StartHere.md
- AI-Log/CHANGELOG.md
- AI-Log/SESSION-HISTORY.md

### Commands/tests executed

- Inspected source-note sizes and checked for an installed Mermaid CLI; no local mmdc executable was available.
- Validated Mermaid fence pairing, declaration types, sequence control balance/participants, flowchart node definitions, and relative Markdown links.
- Validated key source-aligned statements and absence of credential-like Northwind example material.

### Outcome

StartHere.md is now a reviewer-friendly entry point with structurally valid Mermaid source. No application, database, frontend, mock, automation, or infrastructure code was changed.

## 2026-08-16T16:59:24-05:00 — Northwind Connect Go mock implemented

### User instruction or question

The Integration Lead explicitly authorized creation of a mock Go API based on the supplied Northwind Connect API document.

### Visible response summary

- Applied the initialization evaluator and kept the bounded task in the primary session without a sub-agent.
- Used the Go backend harness with integration-reliability and QA checks.
- Re-read the exact Northwind API contract and current decisions/concerns.
- Created a standalone standard-library Go module under mock/northwind.
- Implemented documented accounts, transactions, transfers, authentication, pagination, error bodies, transfer states, and unsigned webhook delivery using synthetic in-memory data.
- Isolated mock-only health, state-transition, duplicate-webhook, and deterministic failure controls from the public /v1 surface.
- Added integer-cent money handling so internal mock behavior does not use binary floating-point arithmetic.
- Added tests, configuration documentation, .env.example, and a non-root multi-stage Docker image.
- Updated project records so this narrow authorization is not confused with authorization to implement the Vantaca application.

### Decisions made

- The first transfer ID and core seed records emulate the supplied examples.
- The mock represents one synthetic customer because the contract omits customer identity/scoping.
- The mock accepts USD with at most two fractional digits and treats the single routing number as the destination routing number; both remain explicit assumptions.
- Webhooks are unsigned to match the supplied contract and can retry/duplicate to test safe consumers.
- The undocumented internal accounts endpoint is not implemented.
- Mock-only controls use /__mock and X-Northwind-Mock-Scenario and are not part of the emulated public contract.

### Questions raised

- The mock cannot resolve open Northwind questions about customer scoping, partner idempotency, balance semantics, webhook authenticity, real quotas, or production environments.
- Race detection remains unexecuted locally because the Windows Go environment has CGO disabled and no GCC toolchain.

### Files created

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

### Files modified

- Notes/StartHere.md
- AI-Log/DECISIONS.md
- AI-Log/OPEN-QUESTIONS.md
- AI-Log/CONCERNS.md
- AI-Log/CHANGELOG.md
- AI-Log/SESSION-HISTORY.md

### Commands/tests executed

- gofmt completed successfully.
- go test -count=1 -cover ./... passed with 53.4% cmd/server and 76.7% internal/mockapi coverage.
- go vet ./... passed.
- go build ./cmd/server passed; its generated executable was removed after validation.
- go test -race was attempted but could not start because CGO/GCC is unavailable.
- Docker build passed and reran the test suite in Linux.
- An isolated container smoke test verified health, authentication failure, accounts, transfer creation, and 429/Retry-After; the temporary container was removed.
- Verified no generated executables/test binaries or smoke containers remained in the workspace/runtime.

### Outcome

The Northwind mock is complete, documented, tested, and containerized. It is development/test infrastructure only and does not constitute Northwind contract certification or authorization to implement the Vantaca application.

## 2026-08-16T17:05:41-05:00 — Root Docker Compose startup added

### User instruction or question

The Integration Lead requested that the mock API be a Docker container and that a root Docker Compose file start all applications at once.

### Visible response summary

- Kept the existing multi-stage, non-root, distroless mock image.
- Added a native `healthcheck` command to the Go binary so Docker can test readiness without adding a shell, curl, or wget.
- Added root `compose.yaml` as the single startup entry point for all currently implemented applications.
- Added synthetic root environment defaults, local `.env` protection, and one-command operating instructions.
- Clarified that the Northwind mock is currently the only authorized runnable application and future authorized services should join the same root Compose file.

### Decisions made

- Root Compose is the local/demo orchestration boundary.
- The container keeps a fixed internal port of 8081; `NORTHWIND_MOCK_HOST_PORT` controls only the host mapping.
- Compose uses synthetic local defaults and waits on a container-native health check.
- This orchestration change does not authorize implementation of the Vantaca API, UI, SQL Server, or n8n.

### Questions raised

- None. Future service dependency and readiness contracts will be defined when those services are authorized.

### Files created

- compose.yaml
- .env.example
- .gitignore

### Files modified

- mock/northwind/Dockerfile
- mock/northwind/cmd/server/main.go
- mock/northwind/cmd/server/main_test.go
- mock/northwind/README.md
- Notes/StartHere.md
- AI-Log/DECISIONS.md
- AI-Log/CHANGELOG.md
- AI-Log/SESSION-HISTORY.md

### Commands/tests executed

- gofmt completed successfully.
- go test -count=1 -cover ./... passed with 50.6% cmd/server and 76.7% internal/mockapi coverage.
- go vet ./... passed.
- docker compose config --quiet passed, and the resolved configuration was inspected.
- An isolated Compose project built and started on host port 18082 with `--wait`.
- Docker health status was `healthy`; `/healthz` returned `ok`; authenticated accounts returned `acc_1029` first.
- The isolated Compose container and network were removed, and no residual project resources remained.

### Outcome

The root command `docker compose up --build -d --wait` now starts the complete currently implemented stack and waits until the Northwind mock is healthy. The default endpoint is `http://localhost:8081`.

## 2026-08-16T17:13:31-05:00 — Recent-transactions and webhook-idempotency guidance refined

### User instruction or question

The Integration Lead asked for StartHere.md to further explain the recent-transactions architectural decision, retain an async mismatch path that updates SQL Server and forces a frontend refresh, and revise Northwind questions to investigate whether webhook `transfer_id` is safe as an idempotency key rather than assuming no webhook idempotency exists.

### Visible response summary

- Added a recent-transactions architecture section and Mermaid sequence to StartHere.md.
- Made SQL Server the initial read model and Northwind the asynchronous authoritative comparison.
- Required a successful SQL commit before a mismatch invalidates the frontend view.
- Defined force refresh as frontend cache/query invalidation followed by a new Vantaca API read, never a direct browser call to Northwind.
- Documented no-refresh behavior for matching values and stale preservation for Northwind failures.
- Updated the focused Northwind question to test `transfer_id`, `(transfer_id, status)`, and separate event-ID alternatives.
- Kept webhook consumer idempotency separate from transfer-creation idempotency after an ambiguous POST outcome.

### Decisions made

- Recent transactions are returned from SQL Server with freshness metadata and asynchronously reconciled with Northwind.
- Frontend invalidation occurs only after changed values are committed to SQL Server.
- A matching Northwind result updates checked-at metadata without forcing a frontend refresh.
- `transfer_id` is an unresolved candidate webhook idempotency key, not assumed sufficient or insufficient.

### Questions raised

- Is `transfer_id` globally unique and immutable for webhook processing?
- Can multiple legitimate status-change events reuse one `transfer_id`?
- Is the supported consumer key `transfer_id`, `(transfer_id, status)`, or a separate event/delivery ID, and is it stable across retries?
- What exact recent window, ordering, mutation/correction behavior, and frontend invalidation transport should be used?

### Files created

- None.

### Files modified

- Notes/StartHere.md
- AI-Log/OPEN-QUESTIONS.md
- AI-Log/CONCERNS.md
- AI-Log/DECISIONS.md
- AI-Log/CHANGELOG.md
- AI-Log/SESSION-HISTORY.md

### Commands/tests executed

- Inspected the supplied Northwind webhook schema and current recent-transaction, idempotency, freshness, and synchronization guidance.
- Validated Markdown/Mermaid structure and cross-document webhook language.

### Outcome

The recent-transactions flow now explicitly provides a fast SQL-backed response, asynchronously converges to Northwind, and refreshes the frontend only after a mismatch is safely persisted. The Northwind call now seeks affirmative webhook-key semantics instead of presuming they do not exist.

## 2026-08-16T17:15:44-05:00 — Docker Compose request reconfirmed

### User instruction or question

The Integration Lead repeated the request that the mock API run as a Docker container and that a root Docker Compose file start all applications at once.

### Visible response summary

- Confirmed the requested container and root orchestration were already implemented.
- Re-read the root Compose file and the Northwind mock Dockerfile.
- Searched for runnable application manifests and confirmed the Northwind Go mock remains the only implemented application.
- Revalidated the resolved Compose configuration.

### Decisions made

- None. Existing decision D-007 remains current.

### Questions raised

- None.

### Files created

- None.

### Files modified

- AI-Log/SESSION-HISTORY.md

### Commands/tests executed

- Enumerated Dockerfiles and application manifests.
- `docker compose config --quiet` passed.

### Outcome

The root `compose.yaml` remains the single startup entry point and covers every currently implemented application. No functional change was required.

## 2026-08-16T17:23:43-05:00 — Recent-transactions clarification reconfirmed

### User instruction or question

The Integration Lead repeated the request to document asynchronous recent-transaction reconciliation and ask whether webhook `transfer_id` is safe as an idempotency key.

### Visible response summary

- Confirmed StartHere.md already contains the SQL Server-first recent-transactions flow.
- Confirmed its mismatch `alt` branch updates SQL Server before invalidating and force-refreshing the frontend.
- Confirmed the Northwind question explicitly tests `transfer_id`, `(transfer_id, status)`, and separate event-ID alternatives.
- Confirmed transfer-submission idempotency remains a separate question.

### Decisions made

- None. Existing decision D-008 remains current.

### Questions raised

- None beyond the already recorded Northwind webhook-identity questions.

### Files created

- None.

### Files modified

- AI-Log/SESSION-HISTORY.md

### Commands/tests executed

- Verified all required StartHere flow statements and Northwind questions.
- Revalidated Markdown fences and Mermaid `alt`/`else`/`end` balance.

### Outcome

The requested architecture and webhook-idempotency framing were already present and remain structurally valid. No duplicate documentation or functional change was needed.

## 2026-08-16T17:32:41-05:00 — All major workflow charts receive detailed explanations

### User instruction or question

The Integration Lead requested a text description for Major Workflow 2, confirmation that the remaining StartHere.md workflow charts also have detailed explanations, and explicitly directed that the architecture workflows must not be removed unless relocated intact to a detailed Architecture.md file.

### Visible response summary

- Audited all Mermaid charts and the surrounding prose in StartHere.md.
- Retained all six Major Workflow diagrams in place.
- Added detailed purpose, execution, consistency, and failure text to every major workflow.
- Restored Workflow 2 from a synchronous refresh path to the recorded SQL-first asynchronous reconciliation path.
- Required the mismatch branch to commit the MSSQL update before invalidating and force-refreshing Next.js.
- Aligned webhook prose with the authoritative `transfer_id` idempotency question.
- Added an explicit retention/relocation boundary for architecture documentation.

### Decisions made

- Major Workflows 1–6 remain in StartHere.md with their prose explanations.
- If reorganized later, the complete architecture content must move to `Notes/Architecture.md` and remain linked from StartHere.md.
- No Architecture.md file is needed now because no workflow content was removed from StartHere.md.

### Questions raised

- The exact frontend invalidation transport remains open.
- “Recent” transaction window, ordering, corrections, and stable transaction-ID behavior remain open.
- Northwind must still confirm the supported webhook consumer idempotency key.

### Files created

- None.

### Files modified

- Notes/StartHere.md
- AI-Log/DECISIONS.md
- AI-Log/CHANGELOG.md
- AI-Log/SESSION-HISTORY.md

### Commands/tests executed

- Counted and inspected all workflow headings, diagrams, and prose sections.
- Confirmed 12 balanced Mermaid blocks and six sequence diagrams.
- Validated sequence participant references and `alt`/`end` balance.
- Confirmed each Major Workflow section contains a detailed `How it works` explanation.
- Validated relative Markdown links.

### Outcome

StartHere.md retains the complete architecture workflow set, and every workflow can now be understood from text as well as from its Mermaid chart. Workflow 2 specifically preserves the directed asynchronous update-before-frontend-refresh behavior.

## 2026-08-16T17:47:25-05:00 — Role-based GitHub issue drafts created

### User instruction or question

The Integration Lead requested a detailed but straightforward set of local Markdown GitHub issues for each engineer responsibility and QA/PM, with a linked AI harness in every issue and enough goals/testing direction to preserve creative implementation freedom.

### Visible response summary

- Used the six-role Delivery Approach in StartHere.md as the ownership boundary.
- Created an EngineeringTasks index and issue drafts for the Integration Lead, four engineers, and QA/PM.
- Assigned one primary harness per issue: EM Delivery, Go Backend, Database, Integration Reliability, Frontend, and QA.
- Included authorization status, goals, scope, acceptance criteria, required tests/evidence, dependencies/handoffs, and out-of-scope work.
- Kept internal design, naming, component/package layout, and task decomposition with each engineer unless a change crosses an approved boundary.
- Linked the local task index from StartHere.md.

### Decisions made

- Use one outcome-based issue per delivery role rather than a large shared issue or prescriptive implementation recipe.
- Start AI work through the initialization evaluator and use the issue's one primary harness.
- These issue drafts organize future work and do not authorize currently blocked application implementation.

### Questions raised

- Issue details must be updated as Northwind, Product, Security, and production-platform decisions close.
- Actual GitHub assignees, milestones, labels, and publication timing remain with the Integration Lead.

### Files created

- EngineeringTasks/README.md
- EngineeringTasks/00-integration-lead-delivery-readiness.md
- EngineeringTasks/01-northwind-adapter-and-mock.md
- EngineeringTasks/02-mssql-account-transaction-sync.md
- EngineeringTasks/03-transfers-webhooks-reliability.md
- EngineeringTasks/04-nextjs-customer-experience.md
- EngineeringTasks/05-qa-pm-acceptance-release.md

### Files modified

- Notes/StartHere.md
- AI-Log/DECISIONS.md
- AI-Log/CHANGELOG.md
- AI-Log/SESSION-HISTORY.md

### Commands/tests executed

- Read the initialization evaluator and all selected primary harnesses before assigning them.
- Validated file count, required sections, acceptance/testing checklist depth, primary harness links, all relative links, and StartHere discoverability.
- Scanned the task set for credential-like material; none was found.

### Outcome

Six copy-ready GitHub issue drafts now give each owner clear goals and required evidence while preserving implementation freedom and existing authorization boundaries.

## 2026-08-16T18:23:01-05:00 — Workflow logic and synthetic samples added to all role issues

### User instruction or question

The Integration Lead requested that ENG-02 cover secured MSSQL storage under the encryption policy, representative received data, table layouts, and supporting views, then clarified that every role issue should contain sample data and workflow logic from the architecture.

### Visible response summary

- Added architecture-workflow logic and role-specific synthetic JSON to all six issue drafts.
- Expanded ENG-02 with inbound Northwind account/transaction examples and a masked read-model projection.
- Defined conceptual link, account, secure-value, transaction, sync-run, and outbox tables.
- Defined customer account, recent transaction, sync health, and unpublished-event views/queries.
- Added required uniqueness/recent-read/work-selection/outbox indexes and concurrency constraints.
- Made protected storage contingent on the approved encryption policy and documented data minimization, managed keys, ciphertext separation, masking, rotation, restore, audit, and failure expectations.
- Added targeted database/security tests without selecting an unapproved encryption product.

### Decisions made

- Every role issue must show how its work implements the documented architecture and provide representative synthetic data.
- ENG-02 may choose physical names/types within repository conventions but must preserve the documented logical boundaries and demonstrable controls.
- Exact encryption technology, key ownership, retention, and physical secured schema remain blocked on Security/data-policy approval.

### Questions raised

- Can opaque Northwind account tokens eliminate full-value storage?
- Which encryption/key platform and rotation/restore process does Vantaca approve?
- What money precision/API representation and transaction correction semantics will be confirmed?

### Files created

- None.

### Files modified

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

### Commands/tests executed

- Re-read the Northwind contract, mock seed models, encryption/data concerns, open questions, and ENG-02 issue.
- Validated six workflow sections, six sample sections, JSON code fences, synthetic-data labels, and relative links.
- Confirmed ENG-02 includes six logical tables, four supporting views/queries, encryption requirements, and plaintext/key-rotation/view tests.

### Outcome

All issue drafts now provide architecture context and concrete synthetic examples. The MSSQL issue additionally describes a secure, policy-driven logical data model and the evidence required to prove it.

## 2026-08-16T18:31:16-05:00 — Application placement and dependency design added

### User instruction or question

The Integration Lead requested that the architecture show where each application/feature is used and its dependencies on other applications and platform capabilities.

### Visible response summary

- Added a shared logical runtime/application dependency Mermaid map to the EngineeringTasks index.
- Showed Next.js, the Go modular application, account/transaction and transfer services, workers, webhook ingress, adapter, repositories, SQL Server/outbox, Northwind/mock, optional n8n/scheduler, identity, secrets, encryption keys, frontend invalidation, and observability.
- Added a feature-placement matrix for account/balance reads, recent transactions, ACH submission, webhooks, reconciliation, and failure/recovery.
- Added role-specific architecture placement/dependency tables to all six issue drafts.
- Distinguished directed application components, external Northwind dependencies, unresolved Vantaca platform dependencies, optional orchestration, and the currently implemented local stack.
- Linked StartHere to the enhanced dependency map.

### Decisions made

- The adapter, domain services, repositories, and workers remain modules in one Go deployable unless a future approved decision changes that boundary.
- Next.js uses only Vantaca APIs/invalidation and never calls Northwind, SQL Server, the mock, n8n, or key services directly.
- The Northwind mock remains a local substitute, not a production application dependency.
- Platform dependencies are shown explicitly without selecting unapproved technologies.

### Questions raised

- Which Vantaca services provide production identity, secrets, encryption keys, durable jobs/outbox execution, invalidation, ingress, observability, CI/CD, and operations ownership?
- Which frontend invalidation transport will be approved?
- Is n8n supported in production or only in the local demonstration?

### Files created

- None.

### Files modified

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

### Commands/tests executed

- Validated the shared Mermaid block structure, component coverage, fence balance, and feature-placement matrix.
- Confirmed all six issues include linked placement/dependency sections with explicit relationship tables.
- Revalidated task document links and StartHere discoverability.

### Outcome

The task set now explains not only what each owner builds, but where it runs, who calls it, which applications/platform services it depends on, and what consumes its output.

## 2026-08-16T18:42:50-05:00 — Production-readiness gap assessment

### User instruction or question

The Integration Lead asked whether anything remains missing beyond the demo applications.

### Visible response summary

- Compared repository artifacts with the B0/B1 question register, concerns, pending decisions, dependency map, Compose services, and production-readiness expectations.
- Confirmed the repository is strong on architecture/workflow documentation, risks, role issues, AI harnesses, sample data, and the Northwind mock.
- Confirmed the Northwind mock is still the only runnable application/service.
- Identified unresolved Product/Northwind/Security/platform decisions and missing production artifacts such as Vantaca API contracts, migrations, CI/CD, infrastructure definitions, threat model, runbooks, dashboards/SLOs, and partner certification evidence.
- Identified potential ownership gaps for Security, Platform/DevOps, and Operations even though their dependencies appear in the existing role issues.

### Decisions made

- None. This was a read-only gap assessment.

### Questions raised

- Should dedicated Security, Platform/DevOps, and Operations issue drafts be added?
- Which missing artifacts are required for the interview deliverable versus the production-readiness claim?
- Who owns closing each external/platform dependency and by what date?

### Files created

- None.

### Files modified

- AI-Log/SESSION-HISTORY.md

### Commands/tests executed

- Enumerated root files, runnable manifests, Compose services, B0/B1 questions, concerns, pending decisions, and dependency status.
- Searched for SQL migrations, OpenAPI/Swagger, CI/CD YAML, infrastructure-as-code, threat models, ADRs, runbooks, SLOs, and dashboard definitions.
- Confirmed the workspace is not currently a Git repository and no production-artifact candidates exist beyond local Compose.

### Outcome

The remaining work is not primarily more architecture diagrams. It is closure of external/internal decisions plus concrete security, platform, contract, operational, QA, and release evidence needed to support a production-ready claim.

## 2026-08-16T18:44:46-05:00 — Original-prompt deliverable gap audit

### User instruction or question

The Integration Lead clarified that the gap assessment should reference the original operating prompt and exclude demo applications, unanswered go-live questions, and requirements for a true full-service demo.

### Visible response summary

- Re-read the full original operating prompt and the later AI harness prompt.
- Mapped their explicit non-application deliverables against the current repository.
- Confirmed that architecture/workflow diagrams, contradiction/risk analysis, mock strategy, database/n8n strategy, team ownership, issue briefs, AI logs (except PrePromptHistory), and the harness suite are substantially covered.
- Narrowed the remaining gaps to reviewer presentation, complete EM delivery analysis, QA traceability, consolidated security/data classification, two useful missing visuals, expected workspace inputs, and Git/GitHub packaging.
- Corrected the earlier broader framing: unresolved Security/Platform/Operations decisions are already captured and are excluded from this prompt-specific gap list.

### Decisions made

- None. This was a read-only audit against the original prompt.

### Questions raised

- Should the missing original `PrePromptHistory.md` and `Ai-Engineering/` source material be restored from their authoritative source rather than recreated?
- Should the remaining non-application artifacts be created now or only when the Integration Lead prepares the final interview package?

### Files created

- None.

### Files modified

- AI-Log/SESSION-HISTORY.md

### Commands/tests executed

- Read both pasted prompt attachments in full.
- Enumerated every repository file and checked required prompt paths.
- Audited StartHere Mermaid coverage, the original team/timeline analysis, workstream delivery fields, and reviewer-entry-point artifacts.
- Confirmed there is no root README, Git repository, PrePromptHistory.md, or Ai-Engineering directory in the current workspace.

### Outcome

Beyond demo implementation and unresolved questions, the project is close on discovery/design coverage. A focused set of non-code artifacts remains to fully satisfy the original prompt and produce the intended five-minute reviewer experience.

## 2026-08-16T19:01:02-05:00 — Complete original-prompt readiness package

### User instruction or question

The Integration Lead requested completion of six original-prompt gaps: a root reviewer README, full EM delivery analysis, phase comparison, QA traceability, consolidated security/data classification, and transfer-lifecycle plus two-week delivery visuals.

### Visible response summary

- Created a five-minute root reviewer path with truthful startup/demo instructions and explicit implementation limitations.
- Created the full workstream planning/evidence matrices and a five-phase path from Prototype/Demo through Production GA.
- Created a 17-row QA requirement → risk → test → evidence register that labels mock proof separately from planned, blocked, and release-candidate evidence.
- Created a consolidated proposed data classification/control matrix without resolving Security/Data-owner policy questions by assumption.
- Added a transfer lifecycle state diagram to Major Workflow 3 and a two-week Gantt to the delivery plan.
- Preserved all six StartHere architecture workflows and linked the new artifacts from the orientation and role issues.

### Decisions made

- Recorded D-011: maintain a concise root entry point backed by focused delivery, QA, and security artifacts.
- Use person-day ranges and dependency assumptions rather than a false point estimate or calendar promise.
- Reserve `RELEASE PASS` for reviewed production-candidate evidence; mock success remains `MOCK PASS`.
- Treat data classifications/controls as proposed until accountable Vantaca owners approve them.

### Questions raised

- No new architecture question was invented. Existing B0/B1 partner, Product, Security, platform, and Operations questions remain open and linked.
- The 8–12 person-day transfer workstream needs pairing/reusable platform support or read-only scope reduction if it trends above one engineer's ten-day capacity.

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

### Commands/tests executed

- Checked required README sections, the five delivery phases, 17 QA rows, security-matrix coverage, and relative links.
- Confirmed all six Major Workflows remain and structurally checked 16 Mermaid declarations; no Mermaid CLI was installed for rendered validation.
- Ran Go formatting check, all mock tests, Go vet, and root Compose configuration validation successfully.

### Outcome

All six requested non-application gaps are now implemented as linked, reviewable artifacts. The package exposes rather than hides production blockers, schedule capacity risk, and the distinction between mock evidence and production approval.

## 2026-08-16T20:40:00-05:00 — Implement and verify the contained interview demo

### User instruction or question

The Integration Lead authorized principal-engineer/principal-architect implementation of an interview demo, then required every application—including the Northwind mock and database assets—to live under a top-level `Demo/` folder separate from AI data/logs/harnesses, Notes, Instructions, and EngineeringTasks.

### Visible response summary

- Implemented the Next.js customer experience, Go Vantaca application API, SQL Server read/operational model, Northwind adapter, deterministic partner controls, webhook/reconciliation paths, and optional n8n trigger.
- Moved the prior API, mock, and database assets into `Demo/`; removed the now-empty legacy `apps/` directory.
- Kept root Compose as the one-command orchestrator and updated current reviewer/architecture/QA/task documentation to distinguish the executable demo from production readiness.
- Added a repeatable PowerShell end-to-end smoke test covering the architectural decision points rather than only health endpoints.

### Decisions made

- Recorded D-012: contain all executable/runtime assets under `Demo/` while retaining root Compose as the repository entry point.
- Keep full account/routing values transient inside the Northwind adapter; persist only opaque IDs, last four, exact minor units, freshness, transfer state, payload hash, and audit/correlation state.
- Serve recent transactions from SQL, compare asynchronously, commit a changed snapshot atomically, and let Next.js perform bounded version polling after refresh begins.
- Retry only safe reads. Persist a transfer intent before exactly one partner POST; represent timeouts, partner 5xx, and post-accept persistence uncertainty as `UNKNOWN` without automatic resubmission.
- Treat unsigned demo webhooks as durable reconciliation signals and use `(transfer_id,status)` only as the synthetic inbox rule; retain the Northwind production identity question.
- Namespace mock transfer IDs per process boot so a rebuilt in-memory mock cannot collide with partner IDs retained in the SQL volume.

### Questions raised

- No new production assumption was closed. Existing B0/B1 identity, balance semantics/freshness, partner transfer idempotency/exact lookup, webhook authenticity/event identity, encryption/key policy, platform, and Operations questions remain release gates.

### Files created

- `Demo/README.md`
- `Demo/api/**`
- `Demo/web/**`
- `Demo/database/migrations/**`
- `Demo/database/seeds/001_demo_link.sql`
- `Demo/automation/n8n/**`
- `Demo/scripts/e2e-smoke.ps1`

### Files moved

- `apps/api` → `Demo/api`
- `mock` → `Demo/mock`
- `database` → `Demo/database`

### Files modified

- `.env.example`
- `.gitignore`
- `compose.yaml`
- `README.md`
- `Notes/StartHere.md`
- `Notes/QA-Acceptance-Matrix.md`
- `EngineeringTasks/README.md`
- `AI-Log/OPEN-QUESTIONS.md`
- `AI-Log/CONCERNS.md`
- `AI-Log/DECISIONS.md`
- `AI-Log/CHANGELOG.md`
- `AI-Log/SESSION-HISTORY.md`

### Commands/tests executed

- Ran Go formatting, full unit tests, and `go vet` for both `Demo/api` and `Demo/mock/northwind`.
- Ran `npm test -- --run`, TypeScript checking, and a Next.js production build in `Demo/web`.
- Validated base and `automation` Compose configurations.
- Built all three local application images and started SQL Server, mock, API, and web with health-dependent Compose ordering.
- Used host ports 18080/13000 for verification because unrelated local projects already owned 8080/3000; the documented defaults remain configurable.
- Ran `Demo/scripts/e2e-smoke.ps1` successfully with 17 live assertions.
- Removed only the Compose-labeled synthetic SQL demo volume after verification, then restarted all four base services healthy with the clean seed (three accounts, zero transfers) for presentation.

### Defects found and corrected during live verification

- Corrected false transaction-version churn caused by SQL `CHAR(4)` padding and finer-than-`DATETIMEOFFSET(7)` timestamp precision.
- Marked every local account degraded when the top-level Northwind account-list request fails while preserving the last-known SQL snapshot.
- Prevented mock transfer-ID reuse across container rebuilds with retained SQL state.
- Added conservative handling for a Northwind acceptance followed by local partner-ID persistence failure: retain the durable intent as `UNKNOWN`, never a definitive failure that invites resubmission.

### Outcome

The repository now has a clean documentation/runtime boundary and a healthy, one-command, synthetic end-to-end demo. The live smoke test proves SQL snapshot behavior, one-change invalidation, failure preservation/recovery, exact money, durable request dedupe, duplicate webhook handling, `PENDING → POSTED → RETURNED`, and ambiguous-outcome non-retry. These are demo results only; production authorization and release evidence remain gated.

## 2026-08-16T20:46:20-05:00 — Enforce `Demo/` as the Docker build boundary

### User instruction or question

The Integration Lead reiterated that all applications, mock assets, and database assets must remain under `Demo/` and separate from AI data/logs/harnesses, Notes, Instructions, and EngineeringTasks.

### Work completed

- Verified no application source, mock code, migration, seed, frontend component, or runtime package remains outside `Demo/`.
- Changed the Go API Compose build context from the repository root to `./Demo`.
- Added `Demo/.dockerignore` so that API builds receive only `Demo/api` and `Demo/database`; documentation and AI/task material cannot enter that build context.
- Updated the API Dockerfile paths for the narrower context.

### Verification

- `docker compose config --quiet` passed.
- The API image rebuilt successfully from the narrowed 134 KB context.
- SQL Server, Northwind mock, Go API, and Next.js remained healthy; API health and UI HTTP checks passed.

### Outcome

The separation is now both physical and enforced at Docker-build time. The only root runtime files are the repository-level Compose orchestrator and its local configuration template; all application and database implementation assets are under `Demo/`.

## 2026-08-16 — Resolve local Compose host-port collision

### User-reported behavior

Running `docker-compose up` failed while publishing the Go API because another local project already owned host port 8080; port 3000 was also occupied.

### Work completed

- Changed the demo's default host ports to API `18080` and UI `13000` while preserving internal container ports 8080 and 3000.
- Updated `.env.example`, the reviewer walkthrough, QA evidence commands, and end-to-end smoke-script defaults.
- Kept every host port configurable through the existing Compose environment variables.

### Verification

- `docker-compose config --quiet` passed using the same hyphenated command invoked by the user.
- `docker-compose up -d --wait` started all four services healthy.
- `http://localhost:18080/healthz` returned a connected database and `http://localhost:13000` returned HTTP 200.

### Outcome

The one-command startup no longer conflicts with the existing services on ports 8080 and 3000.

## 2026-08-16 — Vantaca-aligned interview UI and test guidance

### User instruction or question

Make the demo UI more professional, follow Vantaca's current design and color patterns, use the Vantaca logo, and explain every test action with a tooltip.

### Work completed

- Restyled the Next.js dashboard around Vantaca's public navy, blue, green, and light-surface palette and adopted Montserrat for the interface typography.
- Added the official public Vantaca SVG logo to a restrained product header and clarified the demo's local/synthetic status.
- Reworked the hero, account snapshots, workflow panels, forms, transfer lifecycle cards, notices, and footer into a consistent responsive presentation.
- Added a reusable accessible test-action control. Every scenario-triggering button now exposes a visible information affordance and a detailed `role="tooltip"` explanation on pointer hover or keyboard focus through `aria-describedby`.
- Documented what each action demonstrates: external-activity reconciliation, bounded read failures, durable one-attempt transfer submission, duplicate-webhook idempotency, definitive failure, and late ACH return.
- Kept navigation and dismiss controls free of test tooltips because they do not execute architectural test scenarios.
- Tightened small-viewport behavior for the header, architecture summary rows, and tooltip positioning.

### Verification

- `npm test -- --run`: 2 tests passed.
- `npm run typecheck`: passed.
- `npm run build`: passed with Next.js 16.3.1.
- Rebuilt the web image and started it through the root Compose stack; SQL Server, Northwind mock, Go API, and web all reached healthy state.
- Verified the UI and API returned HTTP 200 on ports 13000 and 18080.
- Visually inspected desktop and narrow-window browser renders. Chrome on Windows enforces a 500-pixel minimum headless layout width, so the final narrow render was checked at that actual layout width rather than treating a cropped 430-pixel bitmap as application overflow.

### Outcome

The demo now presents as a cohesive Vantaca-branded architecture lab while preserving the explicit synthetic-data and no-real-money boundary. Test controls teach the reviewer which reliability or transfer-safety behavior each interaction exercises, including for keyboard users.

## 2026-08-16 — Comprehensive API and Playwright acceptance suite

### User instruction or question

Test every demo feature through API calls and Playwright browser tests.

### Work completed

- Added Playwright 1.62.1 to `Demo/web` with serial `api` and `ui-chromium` projects, bounded timeouts, trace/video/screenshot-on-failure artifacts, and a local HTML report.
- Added 15 direct API tests against the Go API and Northwind mock. Coverage includes health, security/correlation headers, tenant rejection, demo metadata, mock authentication, masked account reads, exact money strings, asynchronous transaction comparison, missing/invalid requests, all 429/500/503/latency safe-read failures and recovery, external activity with exactly one content-version advance, transfer validation, durable request dedupe, partner POST counts, duplicate webhook acknowledgement, `PENDING → POSTED → RETURNED`, `PENDING → FAILED`, every ambiguous submission scenario, unsafe advance rejection, webhook schema validation, and internal reconciliation authorization.
- Added six real Chromium workflow tests. Coverage includes branded/production-boundary content, keyboard-visible tooltips for every test action, account selection, transaction tables, external activity and post-commit UI invalidation, every read-failure selector, all normal/failed/returned/unknown transfer experiences, backend validation messages, and a true 390-pixel responsive overflow assertion.
- Added stable semantic test hooks without coupling tests to layout implementation: named transaction table, account/transfer test IDs, transfer-status attributes, and status/alert roles.
- Documented installation and API/UI/full-suite commands in the root and Demo reviewer guides; ignored local Playwright artifacts in Git and Docker contexts.

### Defect found and corrected

The real 390-pixel Playwright viewport exposed horizontal page overflow that the Windows headless screenshot check had obscured. The recent-transactions table's 620-pixel minimum content width was expanding its CSS Grid track. The workflow grid and both child panels now explicitly use zero-minimum tracks/items, keeping the table's horizontal scrolling inside its card.

### Verification

- `npm test`: 2 Vitest unit tests passed; Playwright files are excluded from Vitest discovery.
- `npm run typecheck`: passed.
- Web production build and Docker rebuild passed.
- `npm run test:e2e:api`: 15/15 passed.
- `npm run test:e2e:ui`: 6/6 passed.
- `npm run test:e2e`: 21/21 passed in 36.5 seconds in one serial run.
- Independent `Demo/scripts/e2e-smoke.ps1`: 17/17 live assertions passed.
- SQL Server, Northwind mock, Go API, and Next.js remained healthy after testing.

### Outcome

The interview demo now has a repeatable executable acceptance layer spanning raw contracts and actual customer interactions. These synthetic passes improve demo confidence but remain distinct from production identity, partner-contract, security, load, and operational release evidence. The tests intentionally use unique request identities and tolerate retained local demo state; running them adds synthetic transactions and transfers to the current demo volume.

## 2026-08-16 — Prevent external-activity simulation on closed accounts

### User-reported behavior

The dashboard showed “The synthetic Northwind activity could not be created” after the external-deposit test was invoked while the selected account was the closed checking account ending in 1111.

### Root cause

The UI enabled the external-activity control for every selected account. The Go demo endpoint also forwarded the request without first validating the SQL account status, so the Northwind mock correctly rejected activity on the closed account and the API reduced that expected domain condition to a generic 502 control failure.

### Work completed

- The dashboard now derives the selected account and enables external activity only when its status is `open`.
- A closed selection changes the action label to “External activity unavailable” and supplies a specific tooltip explaining the open-account requirement.
- Disabled test controls expose their tooltip through a keyboard-focusable wrapper, preserving accessibility.
- The Go API now validates that the requested account exists and is open before calling Northwind. Missing accounts return `404 account_not_found`; closed accounts return `422 external_activity_unavailable` with safe, precise wording.
- Added direct API regression coverage for closed and missing accounts and browser coverage for the disabled closed-account experience.

### Verification

- Go formatting, all Go API tests, and `go vet` passed.
- Vitest, TypeScript checking, and the Next.js production build passed.
- Rebuilt both API and web images; all four Compose services reached healthy state.
- The expanded combined Playwright suite passed 22/22 tests: 16 direct API cases and six Chromium workflows.

### Outcome

The invalid action shown in the screenshot is no longer possible through the dashboard, and direct callers receive an actionable validation response instead of a misleading upstream failure. Open-account external reconciliation behavior remains unchanged.

## 2026-08-16 — Add Swagger/OpenAPI visibility for both demo APIs

### User instruction or question

Add Swagger OpenAPI so interview reviewers can see and exercise both the Northwind mock endpoints and the core demo API.

### Work completed

- Authored complete OpenAPI 3.0.3 contracts beside the core Go HTTP handlers and Northwind mock handlers. The documents cover all 11 core paths and all seven mock paths, with data examples, exact-money representations, transfer lifecycle enums, error responses, authentication inputs, and clear separation of demo controls from production-shaped endpoints.
- Embedded each contract into its owning Go binary and exposed an unauthenticated `/openapi.yaml` route, so the service and its contract remain versioned together.
- Added a pinned Swagger UI container to the root Compose stack at `http://localhost:18090`. Its definition selector loads the exact checked-in core and mock documents rather than separately generated copies.
- Enabled `Try it out` through a narrowly configured local CORS allowlist. Only the configured Swagger origin is accepted; the code explicitly identifies this as a demo convenience rather than a production browser/security policy.
- Added an **API explorer** link to the interview dashboard and documented the raw specification URLs plus synthetic Swagger authorization inputs.
- Extended Playwright with standards-based parser validation, route/content checks, restricted preflight assertions, explorer-mounted document checks, and a real Chromium test that renders and switches between both contracts.

### Verification

- `gofmt`, `go test -count=1 ./...`, and `go vet ./...` passed for the core and mock Go services.
- `npm test`, `npm run typecheck`, and `npm run build` passed for the web application.
- Both default and automation-profile Compose configurations validated.
- Root Compose rebuilt/started with SQL Server, Northwind mock, core API, Swagger UI, and Next.js all healthy.
- The full serial Playwright run passed 24/24 tests: 17 API/contract cases and seven Chromium workflows.
- The independent PowerShell smoke suite passed all 17 live assertions.

### Outcome

An interviewer can now inspect and execute either API from one local Swagger page while the raw contracts remain available directly from their owning services. This improves demo discoverability and executable-contract evidence without changing the unresolved production identity, partner, webhook, security, or operational gates.

## 2026-08-18 — Document a contingent `/internal/accounts/full` path

### User instruction or question

Update StartHere with a path for handling the undocumented `/internal/accounts/full` endpoint from the integration thread, call out the need for manual testing and Northwind documentation, explain the architectural challenges, and keep it contingent rather than a primary focus.

### Work completed

- Added a dedicated subsection beneath Major Workflow 1 so the option is visible in the account architecture without becoming a new primary workflow.
- Kept the documented `GET /accounts` contract as the default and placed any future `/internal/accounts/full` call behind a tenant/environment feature flag inside the Northwind adapter.
- Added a Mermaid decision flow for written-contract, partner-authorization, Security, and manual-evidence gates; common normalization/persistence; nonconforming-response containment; last-known-good preservation; alerting; and bounded return to the public path.
- Defined manual tests for caller/customer scope, credentials, schema, pagination, balance semantics, timestamps, precision, protected-data exposure, failure behavior, rate limits, parity, performance, SLA, versioning, deprecation, and support.
- Required Northwind-authored versioned documentation and written production authorization before enablement, followed by automated contract tests, safe telemetry, and a kill switch.
- Explicitly identified the endpoint as an asynchronous written follow-up and optional optimization that must not block the public-path implementation or the two-week launch plan.

### Outcome

The architecture now acknowledges the integration-thread proposal without treating an undocumented internal endpoint as a dependable contract. Vantaca can evaluate it later under controlled evidence gates while preserving the supported public path and existing customer-facing freshness semantics.
