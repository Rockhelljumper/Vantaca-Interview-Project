# [EM-01] Close architecture blockers and coordinate production readiness

**Owner:** Engineering Manager / Integration Lead  
**Suggested labels:** `planning`, `architecture`, `release-blocker`  
**Primary AI harness:** [EM Delivery Harness](../AI-Harnesses/em-delivery-harness.md)  
**AI initialization:** Start with the [Initialization Evaluator](../AI-Harnesses/initialization-evaluator.md).  
**Authorization:** Planning and decision coordination are in scope. This issue does not independently authorize application implementation or production launch.

## Goal

Turn the two-week production target into an owned, evidence-based delivery plan. Close or explicitly escalate architecture-changing B0/B1 questions early enough that the four engineering workstreams and QA/PM can proceed without inventing requirements.

Use the cross-workstream estimates, phase gates, and two-week sequence in the [Engineering Delivery Analysis](../Notes/Delivery-Plan.md) as the current planning baseline.

## Architecture placement and application dependencies

**Placement:** Delivery/control plane only; this issue does not add a runtime service. See the [shared runtime dependency map](README.md#architecture-placement-and-runtime-dependencies).

| Relationship | Application, team, or capability | Required contract |
|---|---|---|
| Inputs | Product, Northwind, Security, Operations, QA | Scope, partner guarantees, policy decisions, platform standards, and release evidence |
| Coordinates | Next.js, Go application, SQL Server, Northwind/mock, optional scheduler | Stable ownership boundaries, interface deadlines, feature gates, and integration sequence |
| Produces | ENG-01 through ENG-04 and QA-01 | Recorded decisions, resolved/escalated blockers, milestones, and go/no-go evidence requirements |
| Production dependencies | Vantaca identity, secrets/keys, hosting/ingress, observability, CI/CD, on-call | Named approved services and accountable owners before readiness approval |

## Architecture workflow logic

This issue coordinates all six Major Workflows in [StartHere.md](../Notes/StartHere.md):

1. Account/balance synchronization and recent transactions provide SQL-backed reads with explicit freshness; recent-transaction reconciliation updates SQL before frontend invalidation.
2. ACH submission distinguishes definitive acceptance/rejection from an ambiguous timeout and never treats a lost response as proof of failure.
3. Webhooks propose state changes through authenticated, durable, idempotent processing; reconciliation repairs missed/late notifications using exact partner correlation.
4. Northwind failures are classified by operation safety so bounded read retries never become blind money-movement retries.

The Integration Lead owns cross-workflow decisions, sequencing, feature gates, and evidence—not each workstream's internal implementation.

## Sample data

Use this synthetic scenario to align planning, tabletop exercises, and evidence:

```json
{
  "customer_ref": "customer_demo_001",
  "account_id": "acc_1029",
  "transaction_id": "txn_88213",
  "transfer_id": "trf_55120",
  "scenario": "post_commit_timeout",
  "expected_control": "mark UNKNOWN, do not resubmit, reconcile by exact partner correlation",
  "release_scope": {
    "account_reads": true,
    "recent_transactions": true,
    "transfer_submission": false
  }
}
```

The feature flags above are an example for planning discussion, not a launch decision. The recorded scope must reflect accountable stakeholder approval.

## Scope

- Freeze the exact read and transfer launch scope, including feature gates and the read-only fallback.
- Assign owners and deadlines to the internal and Northwind questions in [OPEN-QUESTIONS.md](../AI-Log/OPEN-QUESTIONS.md).
- Confirm interfaces and handoffs between the Northwind adapter, SQL repositories/sync, transfer workflows, frontend, and QA.
- Coordinate Product, Security, Operations, and Northwind decisions that engineers cannot make.
- Maintain the daily critical path, integration milestones, readiness evidence, and go/no-go agenda.
- Ensure architecture workflows and approved decisions remain current in [StartHere.md](../Notes/StartHere.md) and [DECISIONS.md](../AI-Log/DECISIONS.md).

## Acceptance criteria

- [ ] Every applicable B0 question has an accountable owner, due date, and recorded answer or explicit stop-ship status.
- [ ] Launch scope names the included account, balance, transaction, transfer, status, and support behaviors.
- [ ] Read and transfer capabilities have independent feature-gate and rollback decisions.
- [ ] Engineer handoffs define the minimum API, persistence, invalidation, and test contracts needed for parallel work.
- [ ] Security, data handling, production platform, operations, and escalation owners are named.
- [ ] The readiness checklist distinguishes demo evidence, mock evidence, partner confirmation, and production evidence.
- [ ] A human go/no-go review is scheduled with Product, Security, Engineering, QA, and Operations.

## Required testing and evidence

- [ ] Walk through all six Major Workflows in StartHere with the assigned owners and capture unresolved decisions.
- [ ] Run tabletop reviews for stale balances/transactions, ambiguous transfer timeout, invalid webhook, missed webhook, and Northwind outage.
- [ ] Confirm each release blocker maps to an observable test, sign-off, or operational control.
- [ ] Confirm rollback, transfer kill switch, read-only fallback, partner escalation, and reconciliation ownership.
- [ ] Publish a final evidence summary listing passed, failed, waived, and still-blocked items; only accountable humans may approve residual risk.

## Dependencies

- Product freshness and UX decisions.
- Northwind architecture-call answers and written artifacts.
- Security/data-handling approval.
- Production platform and operational ownership.
- Evidence and estimates from ENG-01 through ENG-04 and QA-01.

## Out of scope

- Implementing every workstream personally.
- Accepting Product, Security, financial, or production risk on behalf of accountable owners.
- Treating the deadline or a successful demo as production approval.
