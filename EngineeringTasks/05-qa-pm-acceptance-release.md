# [QA-01] Define acceptance criteria and assemble production-readiness evidence

**Owner:** QA / PM  
**Suggested labels:** `qa`, `acceptance-criteria`, `release-readiness`  
**Primary AI harness:** [QA Harness](../AI-Harnesses/qa-harness.md)  
**AI initialization:** Start with the [Initialization Evaluator](../AI-Harnesses/initialization-evaluator.md).  
**Authorization:** Acceptance planning can begin immediately. Test implementation/execution follows the authorization and availability of each workstream.

## Goal

Turn approved scope, decisions, open risks, and all six architecture workflows into a traceable, risk-ranked acceptance matrix and an honest production-readiness evidence package.

Maintain the concrete [QA Acceptance and Evidence Traceability Matrix](../Notes/QA-Acceptance-Matrix.md); it is the canonical requirement → risk → test → evidence register for this issue.

## Architecture placement and application dependencies

**Placement:** QA/PM is an evidence and coordination plane, not a production runtime service. It validates the complete application graph and keeps mock/local evidence distinct from partner and production evidence. See the [shared runtime dependency map](README.md#architecture-placement-and-runtime-dependencies).

| Relationship | Application/component | Required evidence |
|---|---|---|
| Validates | Next.js and Go APIs | Customer behavior, authorization, masking, accessibility, API/error contracts, and feature gates |
| Validates | SQL Server/repositories/outbox | Real-engine migrations, constraints, encryption policy, isolation, precision, concurrency, atomicity, and recovery |
| Validates | Northwind adapter and local mock | Contract mappings, deterministic failures, retry safety, pagination, and explicit mock limitations |
| Validates | Webhook ingress and reconciliation | Authenticity, confirmed idempotency identity, ordering, exact correlation, alerts, and runbooks |
| Depends on | CI/CD, synthetic fixtures, test environments, observability, Operations | Repeatable execution, retained evidence, exercised signals, rollback/recovery, and named owners |
| Produces | Integration Lead and accountable reviewers | Traceability matrix, defect/blocker status, evidence summary, and residual-risk visibility—not unilateral launch approval |

## Architecture workflow logic

QA/PM owns cross-workflow traceability for all six Major Workflows in [StartHere.md](../Notes/StartHere.md):

1. Verify account/balance sync preserves previous data on failure and labels freshness accurately.
2. Verify recent transactions return SQL first, reconcile asynchronously, commit a mismatch before invalidation, and remain quiet on a match.
3. Verify transfer submission distinguishes definitive results from ambiguity and never blindly retries an unknown outcome.
4. Verify webhook identity, authenticity, idempotency, and state ordering using Northwind's confirmed contract.
5. Verify reconciliation repairs only exactly correlated transfers and exposes unresolved discrepancies.
6. Verify each error class produces the correct retry/no-retry behavior, telemetry, operator action, and customer state.

## Sample data

Use synthetic scenario records that state both input and observable outcome. Example recent-transactions case:

```json
{
  "case_id": "TXN-REFRESH-001",
  "database_before": {
    "account_id": "acc_1029",
    "transaction_ids": ["txn_88213"],
    "version": 6
  },
  "northwind_response": [
    {
      "id": "txn_88213",
      "amount": -42.17,
      "currency": "USD",
      "description": "COFFEE HOUSE #42",
      "posted_at": "2026-07-21T14:03:00Z"
    },
    {
      "id": "txn_88174",
      "amount": 2500.00,
      "currency": "USD",
      "description": "PAYROLL DEPOSIT",
      "posted_at": "2026-07-18T12:00:00Z"
    }
  ],
  "expected": {
    "database_version": 7,
    "invalidation_count": 1,
    "frontend_refetch_count": 1,
    "protected_values_exposed": false
  }
}
```

Example ambiguous-transfer expectation:

```json
{
  "case_id": "TRANSFER-TIMEOUT-001",
  "mock_scenario": "post-commit-timeout",
  "expected_internal_status": "RECONCILIATION_REQUIRED",
  "automatic_resubmission_count": 0,
  "customer_message": "pending verification",
  "required_operator_signal": "unknown transfer outcome"
}
```

Fixtures must remain synthetic, versioned, deterministic, and traceable to a requirement or risk. They do not represent Northwind certification.

## Scope

- Define observable acceptance examples before implementation for each included customer workflow and failure state.
- Map B0/B1 questions, concerns, decisions, and issue criteria to automated tests, manual checks, stakeholder sign-offs, or explicit stop-ship outcomes.
- Coordinate deterministic mock fixtures and cross-stream integration scenarios.
- Track defects, blocked evidence, scope changes, feature gates, and readiness status daily.
- Verify operational signals, runbooks, rollback, read-only fallback, transfer kill switch, and partner escalation.
- Prepare the go/no-go evidence summary without accepting residual risk on behalf of accountable owners.

## Acceptance criteria

- [ ] The matrix covers Major Workflows 1–6 and links every case to a requirement, decision, concern, or defect.
- [ ] Product has approved examples for freshness, stale/unavailable data, “recent,” transfer ambiguity, failures, and returns.
- [ ] Security and tenant/account isolation have positive and negative acceptance cases.
- [ ] Every B0/B1 launch item has evidence, a named approver, or a visible stop-ship status.
- [ ] Mock evidence, SQL integration evidence, UI evidence, partner confirmation, and production-operability evidence are reported separately.
- [ ] The release checklist names entry/exit criteria, feature-gate state, rollback evidence, known defects, and residual risks.
- [ ] Final go/no-go ownership remains with accountable Product, Security, Engineering, QA, and Operations humans.

## Required testing and evidence

- [ ] Contract tests cover documented payloads, missing/unknown fields, pagination, errors, and drift.
- [ ] SQL Server integration tests cover migrations, constraints, precision, isolation, rollback, concurrency, and idempotent sync.
- [ ] Account/transaction tests cover stale snapshots, mismatch update-before-invalidate, match/no-refresh, outage, and recovery.
- [ ] Transfer tests cover definitive outcomes, duplicate click, pre/post-commit timeout, no blind retry, and exact reconciliation.
- [ ] Webhook tests cover the confirmed idempotency key, retries, multiple valid status changes, reordering, authenticity failure, and late `RETURNED`.
- [ ] Reliability tests cover `429/Retry-After`, `500`, `503`, latency, cancellation, bounded retries, and backlog/alert signals.
- [ ] UI tests cover accessibility, responsive behavior, masking, loading/empty/error states, ambiguity, and feature gates.
- [ ] End-to-end Compose tests use synthetic data, the real SQL Server engine when added, and deterministic mock scenarios.
- [ ] A rollback/recovery exercise and operational tabletop are completed for the approved production topology.

## Dependencies and handoffs

- Launch scope and acceptance wording from Product/Integration Lead.
- Northwind contract answers and Security/Operations decisions.
- Testable increments and evidence from ENG-01 through ENG-04.
- EM-01 owns final sequencing, escalation, and go/no-go coordination.

## Out of scope

- Approving production launch alone or accepting unresolved Product/Security/financial risk.
- Treating happy-path or mock-only success as partner certification.
- Inflating test counts with brittle snapshots or implementation-detail assertions.
