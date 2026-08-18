# Decisions

This file records decisions explicitly directed or later approved by the Integration Lead. Discovery recommendations are not decisions.

## D-001 — Remain in discovery mode

**Status:** Superseded by later scoped implementation authorizations; retained as discovery history

**Context:** The project needs requirements, risk, architecture, and delivery analysis before implementation.

**Options considered:** Begin building immediately; perform discovery first.

**Decision:** Perform discovery only. Do not scaffold or implement the application, database, Northwind mock, frontend, or n8n workflows without later authorization.

**Why:** The exercise is intended to demonstrate engineering leadership and judgment rather than raw code generation.

**Tradeoffs:** Delays visible code while reducing the risk of encoding unresolved contract and security assumptions.

**Reversibility:** Fully reversible through an explicit mode change from the Integration Lead.

**Production impact:** Positive; prevents unsafe partner assumptions from becoming accidental production behavior.

## D-002 — Use the lead-directed technology baseline

**Status:** Directed by Integration Lead; detailed design pending

**Context:** A coherent demonstration stack has been prescribed.

**Options considered:** Multiple languages/frameworks and persistence approaches were possible before the direction was supplied.

**Decision:** Target Go for the backend, Next.js/TypeScript/Tailwind for the frontend, SQL Server 2022 in Docker with database/sql and parameterized raw SQL, a local Northwind mock, and optional-profile n8n automation. Do not introduce an ORM without an explicit decision change.

**Why:** This makes the implementation understandable, reviewable, and aligned with the interview demonstration goals.

**Tradeoffs:** Raw SQL increases explicit mapping/migration work; the mixed Go/TypeScript stack requires two toolchains; SQL Server containers can be heavier for local development.

**Reversibility:** Reversible before implementation at moderate cost; increasingly costly after schemas and interfaces stabilize.

**Production impact:** Vantaca platform alignment, deployment, database, observability, and support standards remain unknown and must be integrated rather than invented.

## D-003 — Keep demo implementation and production design explicit

**Status:** Directed by Integration Lead

**Context:** Local simplicity is useful for an interview demonstration but can be misleading when financial operations are involved.

**Options considered:** Present one topology as universal; document separate demo and production postures.

**Decision:** Use a clearly labeled local demo topology while separately documenting production controls and unresolved platform integration points.

**Why:** This supports a runnable demonstration without implying that Docker Compose, local secrets, synchronous processing, or a mock partner are production recommendations.

**Tradeoffs:** Requires extra documentation and testing discipline to prevent demo-only behavior from being mistaken for a supported contract.

**Reversibility:** The labels are easily revised; erasing the distinction would introduce substantial communication risk.

**Production impact:** High. Production enablement remains gated by partner, Product, Security, Engineering, QA, and Operations decisions.

## D-004 — Treat the two-week target as production go-live

**Status:** Directed by Integration Lead

**Context:** The earlier stakeholder thread used the term go-live without distinguishing a demo, MVP, pilot, production candidate, or GA. The Integration Lead has now clarified the expectation.

**Options considered:** Two-week demo; integration MVP; limited pilot; production-ready go-live.

**Decision:** Plan and rank discovery questions against a production-ready go-live in two weeks.

**Why:** This is the explicit delivery expectation supplied by the Integration Lead.

**Tradeoffs:** The critical path becomes contract and risk closure, not simply implementation throughput. Low-value questions and future optimizations must be deferred, while financial-integrity, customer-scoping, security, correctness, testing, and operability gates cannot be silently waived.

**Reversibility:** The date or scope can be renegotiated by accountable stakeholders, but the current planning baseline remains production go-live.

**Production impact:** Critical. This decision clarifies the target; it does not approve unresolved production risks. Transfer-write and account-read functionality should be independently gateable if a stop-ship dependency remains unresolved.

## D-005 — Use evaluator-first, bounded AI harness routing

**Status:** Directed by Integration Lead

**Context:** Post-discovery design, implementation, debugging, review, QA, security, and delivery work benefit from specialty instructions, but loading all project history/harnesses or automatically spawning agents would waste context and obscure responsibility.

**Options considered:** Continue with only the master prompt; create large autonomous agent prompts; create small specialty harnesses selected through an initialization evaluator.

**Decision:** Use nine project-specific markdown harnesses under AI-Harnesses/, selected by initialization-evaluator.md. Prefer no sub-agent when the primary session can complete the task efficiently. When delegation is useful, select one primary harness, provide minimum necessary context, use the lowest reliable current model tier, normally allow one escalation, and return material conflicts as DECISION REQUIRED.

**Why:** This makes AI resource selection explainable, reduces irrelevant context, and preserves human ownership of architecture, risk acceptance, stakeholder decisions, and production approval.

**Tradeoffs:** Routing and initialization records add small process overhead. Model availability must be revalidated at assignment time, and specialty boundaries require deliberate handoffs.

**Reversibility:** Fully reversible; harnesses are advisory markdown overlays and do not change application architecture or runtime behavior.

**Production impact:** Indirect. The approach can improve review and delivery discipline, but harness output is not production evidence without human review and task-appropriate validation.

## D-006 — Implement a standalone contract-oriented Northwind mock

**Status:** Directed by Integration Lead; implemented

**Context:** No Northwind environment is available, and the Integration Lead explicitly authorized a Go mock based on the supplied Northwind Connect API guide.

**Options considered:** Happy-path fixtures only; a standalone mock of documented endpoints; a broader simulator that invents unresolved partner behavior.

**Decision:** Implement a standalone standard-library Go service under mock/northwind. Emulate the documented /v1 accounts, transactions, and transfers contract with synthetic in-memory data, query-key authentication, page-size-50 bare arrays, documented error shapes, transfer states, and unsigned webhook delivery. Keep deterministic failure/status controls isolated under /__mock and X-Northwind-Mock-Scenario.

**Why:** The service provides a runnable end-to-end dependency and meaningful failure testing without coupling the Vantaca application to mock internals or pretending unresolved Northwind semantics are known.

**Tradeoffs:** State resets on restart; the mock represents one synthetic customer; USD uses two-decimal integer cents; the single routing number is treated as the destination routing number; mock controls are intentionally non-production endpoints. These are documented assumptions, not partner decisions.

**Reversibility:** High. The mock is a standalone adapter-side dependency with no application integration yet.

**Production impact:** None by itself. Mock tests are development evidence, not Northwind certification or approval to launch against the real partner.

## D-007 — Use root Docker Compose as the local startup entry point

**Status:** Directed by Integration Lead; implemented

**Context:** The Northwind mock had a container image but no project-level orchestration command. The Integration Lead requested one root Compose file that starts all applications together.

**Options considered:** Per-application Docker commands; multiple nested Compose files; one root Compose file.

**Decision:** Keep `compose.yaml` at the repository root as the single local startup entry point. It builds, starts, and health-checks every currently implemented application. Today that is the authorized Northwind mock; future authorized API, UI, SQL Server, and optional n8n services must be added to the same orchestration boundary.

**Why:** A single command makes the project reproducible and prevents application-specific startup instructions from drifting apart.

**Tradeoffs:** Compose is a local/demo topology, not a production deployment definition. The container port remains fixed at 8081 while the host port is configurable, and adding future services will require explicit dependency and readiness contracts.

**Reversibility:** High. Service definitions can evolve without changing the Northwind HTTP contract.

**Production impact:** None by itself. Docker Compose, synthetic credentials, and local health checks are development conveniences and do not establish production hosting, secret-management, scaling, or availability controls.

## D-008 — Serve recent transactions from SQL and reconcile asynchronously

**Status:** Directed by Integration Lead; design documented

**Context:** Recent transactions need a responsive frontend read path while Northwind remains authoritative and may change independently of Vantaca.

**Options considered:** Block every frontend read on Northwind; serve only the SQL snapshot; serve the SQL snapshot and asynchronously compare it with Northwind.

**Decision:** Return recent transactions from the SQL Server read model with explicit freshness metadata. Start a bounded asynchronous reconciliation against Northwind. When normalized values differ, transactionally update SQL first and then invalidate the frontend view so it forcefully re-fetches through the Vantaca API. When values match, record the successful check without forcing a refresh.

**Why:** This preserves a fast and resilient customer read while allowing externally initiated account activity to converge toward Northwind's authoritative state.

**Tradeoffs:** Customers may briefly see a labeled snapshot before reconciliation completes. The design requires coalesced jobs, a comparison contract, freshness/error states, and a frontend invalidation channel. “Recent,” ordering, corrections, and the notification transport remain unresolved details.

**Reversibility:** High. The API/repository boundary permits a different synchronization trigger or notification transport without changing Northwind DTO isolation.

**Production impact:** Material. A database commit must precede invalidation; failed Northwind reads preserve the last known snapshot; the browser never calls Northwind directly. This is a design decision, not authorization to implement the Vantaca application.

## D-009 — Retain the architecture workflows as durable documentation

**Status:** Directed by Integration Lead

**Context:** StartHere.md contains six major workflow diagrams that communicate the integration's operational and failure behavior. Diagram-only or accidentally removed architecture would make the review artifact incomplete.

**Options considered:** Keep diagrams without prose; remove workflows while shortening StartHere; retain diagrams with detailed explanations; relocate the complete architecture to a dedicated document.

**Decision:** Keep Major Workflows 1–6, their diagrams, and their detailed text explanations in StartHere.md. If the document is later reorganized, move the complete workflow and decision content to `Notes/Architecture.md` and leave a clear link from StartHere.md. Do not remove the architecture content.

**Why:** Reviewers must be able to understand the actors, normal path, alternate paths, consistency rules, and failure behavior without inferring architecture from Mermaid notation alone.

**Tradeoffs:** StartHere.md remains longer. A future split is permitted, but only if the dedicated architecture document is complete and discoverable.

**Reversibility:** The content may move between the two named documents, but it may not be discarded without a new Integration Lead decision.

**Production impact:** Indirect but material. Durable workflow documentation improves implementation alignment, QA coverage, operational review, and safe handling of financial failure modes.

## D-010 — Use outcome-based issue briefs with one primary AI harness per role

**Status:** Directed by Integration Lead; issue drafts created

**Context:** The delivery approach assigns an Integration Lead, four engineers, and QA/PM to parallel workstreams. Each owner needs sufficient direction and test expectations without having implementation details prescribed.

**Options considered:** One large implementation issue; highly prescriptive subtask lists; one outcome-based GitHub issue draft per delivery role.

**Decision:** Maintain six local GitHub issue drafts under `EngineeringTasks/`: Integration Lead/readiness, Northwind adapter/mock, SQL Server/sync, transfers/webhooks/reliability, Next.js experience, and QA/PM release evidence. The index contains a shared logical runtime/application dependency map. Each issue identifies one primary AI harness, its runtime or coordination placement, callers and consumers, application/platform dependencies, architecture-derived workflow logic, role-appropriate synthetic sample data, clear outcomes and boundaries, acceptance criteria, testing evidence, and out-of-scope work. Engineers retain freedom over internal implementation within approved architecture and constraints. The database issue documents a logical secured-data/table/view design, but the exact encryption mechanism and physical schema remain gated by the approved Security/data policy and repository conventions.

**Why:** Role-sized ownership reduces handoff overhead, keeps cross-stream goals visible, and supplies enough test direction for parallel delivery without turning planning text into a code recipe.

**Tradeoffs:** Some cross-boundary coordination remains necessary. The issues must be revised as Product, Northwind, Security, and platform questions are answered, and they do not by themselves authorize implementation.

**Reversibility:** High. Drafts can be split or regrouped before publication when ownership or scope changes, while preserving acceptance and risk traceability.

**Production impact:** Indirect. The issue set improves accountability and release evidence but does not replace engineering review, partner confirmation, or human production approval.

## D-011 — Maintain a reviewer-readiness package with explicit evidence status

**Status:** Directed by Integration Lead; documentation implemented

**Context:** The original exercise requires a five-minute repository entry point plus workstream delivery analysis, phase comparison, QA traceability, security/data handling, and useful delivery/domain visuals. The underlying architecture and issue briefs existed, but those cross-cutting artifacts were partial or absent.

**Options considered:** Leave the information distributed across StartHere and task issues; duplicate all detail in a large root document; add a concise root entry point linked to focused delivery, QA, and security artifacts.

**Decision:** Use the root `README.md` as the five-minute reviewer entry point and keep detailed concerns in linked artifacts: `Notes/Delivery-Plan.md`, `Notes/QA-Acceptance-Matrix.md`, and `Notes/Security-Data-Classification.md`. Estimates are person-day ranges and assumptions rather than commitments. Delivery phases use explicit entry/exit gates. QA evidence distinguishes mock passes, planned tests, blocked tests, and production-candidate release passes. Data classifications/controls remain proposed until Security/Data owners approve them. Keep the transfer lifecycle state model alongside Major Workflow 3 in StartHere.

**Why:** Reviewers can quickly understand and run what exists, then inspect delivery feasibility, test evidence, and data risk without mistaking a mock or proposed control for production readiness.

**Tradeoffs:** The repository contains more linked documentation, and QA/security matrices require active maintenance as decisions and evidence change. The two-week plan exposes that transfer work may exceed one engineer's capacity rather than hiding the risk.

**Reversibility:** High. Estimates, phase scope, matrix rows, and proposed controls can be revised through normal reviewed updates; the evidence-labeling principle and architecture-documentation boundary remain in force unless explicitly superseded.

**Production impact:** Material governance impact but no runtime change. These artifacts make stop-ship conditions, human approvals, evidence gaps, and the read-only fallback explicit; they do not close any unresolved B0/B1 question or authorize application implementation/launch.

## D-012 — Implement a contained synthetic interview demo under `Demo/`

**Status:** Directed by Integration Lead; implemented and verified locally

**Context:** After discovery, architecture, delivery planning, and issue decomposition, the Integration Lead explicitly authorized work on a principal-level interview demo and required every application—including the mock and database assets—to be separated from AI, Notes, Instructions, and EngineeringTasks material.

**Options considered:** Keep the partner mock as the only runtime; scatter applications at the repository root; implement the directed modular demo inside one explicit runtime boundary.

**Decision:** Place every executable/runtime asset under `Demo/`: `api/` for the Go modular application, `web/` for Next.js, `mock/northwind/` for the partner simulator, `database/` for SQL migrations/views/seeds, and `automation/n8n/` for the optional trigger. Retain the root `compose.yaml` as the one-command orchestrator. Implement the architecture's conservative invariants: SQL-backed timestamped reads, async compare/commit/version-poll refresh, transient-only full account/routing values, durable transfer intent, no automatic monetary retry after ambiguity, webhook inbox dedupe, and partner-read reconciliation before state change.

**Why:** The layout makes the interview artifact immediately navigable while the runtime proves the most important architectural decisions and failure semantics end to end. It also keeps discovery evidence and AI working material visibly separate from application code.

**Tradeoffs:** The demo uses a fixed synthetic tenant, local credentials, an unsigned webhook mode, bounded frontend polling, and a single-process Go deployment. It demonstrates boundaries and behavior but does not supply Vantaca production identity, approved key management, partner-certified contracts, scale evidence, or operational controls.

**Reversibility:** High. The modular boundaries allow the local mock, scheduler, invalidation mechanism, identity adapter, or hosting model to change without moving domain rules into the browser or orchestration tool.

**Production impact:** No production authorization. Real credentials/data and transfer enablement remain blocked on the ranked B0/B1 decisions and production-candidate evidence. Read-only remains the required fallback when safe transfer correlation or trust controls are unavailable.

## Pending decisions (not approved)

- Whether ACH transfer initiation may proceed without a Northwind idempotency key/client reference and deterministic lookup mechanism.
- Required balance freshness and whether cached data may be presented as current.
- Customer/account linking and tenant-isolation model.
- Minimum acceptable webhook authenticity control before production.
- Whether full account/routing numbers may be persisted and under which controls.
- Whether n8n is appropriate in Vantaca's production platform after its operational standards are known.
