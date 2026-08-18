# [ENG-03] Implement safe transfers, webhook processing, and reconciliation

**Owner:** Engineer 3 — Transfers + Webhooks + Reliability  
**Suggested labels:** `backend`, `financial-safety`, `webhooks`, `reliability`  
**Primary AI harness:** [Integration Reliability Harness](../AI-Harnesses/integration-reliability-harness.md)  
**AI initialization:** Start with the [Initialization Evaluator](../AI-Harnesses/initialization-evaluator.md).  
**Authorization:** This is money-movement implementation and remains blocked until the Integration Lead authorizes it and applicable B0 partner/security questions are resolved.

## Goal

Implement the transfer lifecycle so every definitive, rejected, ambiguous, webhook-driven, and reconciled outcome is explicit, durable, observable, and safe from blind monetary retries.

Use the transfer lifecycle model in [StartHere](../Notes/StartHere.md#transfer-lifecycle-state-model), the protected-data rules in [Security and Data Classification](../Notes/Security-Data-Classification.md), and TRF/WH/REC test IDs in the [QA Acceptance Matrix](../Notes/QA-Acceptance-Matrix.md) as shared boundaries.

## Architecture placement and application dependencies

**Placement:** Transfer submission, webhook processing, state transitions, and reconciliation are modules/workers inside the Go application. Durable state belongs in SQL Server; n8n may trigger work but never owns transfer logic. See the [shared runtime dependency map](README.md#architecture-placement-and-runtime-dependencies).

| Relationship | Application/component | Contract |
|---|---|---|
| Called by | Next.js through the Go HTTP API | Authenticated transfer intent with Vantaca request ID; feature-gated result/status endpoints |
| Calls | ENG-01 Northwind adapter | One safe submission attempt, typed ambiguous outcome, and exact lookup/correlation when confirmed |
| Persists through | ENG-02 SQL conventions / transfer repositories | Intent, partner ID, guarded state/history, webhook inbox, reconciliation state, and operational correlation |
| Receives from | Northwind through approved webhook ingress | Authenticated, schema-valid event with a confirmed consumer-idempotency identity |
| Triggered by | Approved scheduler or operations | Bounded reconciliation; optional n8n invokes Go only and receives a summary |
| Depends on | Identity/authorization, encrypted account-value access, secrets, ingress, observability | Least-privilege customer/action checks, policy-compliant decryption, kill switch, alerts, and support references |
| Produces | Next.js/status APIs and Operations | Explicit pending/unknown/failed/posted/returned state and actionable reconciliation/incident signals |

## Architecture workflow logic

This issue owns Major Workflows 3, 4, and 5 and the transfer branch of Workflow 6 in [StartHere.md](../Notes/StartHere.md):

1. The API authenticates/authorizes, validates the request, and persists one Vantaca transfer intent before one Northwind submission attempt.
2. A definitive acceptance/rejection records its known result. A timeout/lost response records `UNKNOWN/RECONCILIATION_REQUIRED` and does not resubmit.
3. A webhook is authenticated, schema-checked, durably received, deduplicated using the confirmed partner identity, and applied only through valid state transitions.
4. Reconciliation selects unresolved/stale transfers and changes state only after exact partner correlation.
5. Retry policy distinguishes safe reads from ambiguous money movement and emits actionable, redacted signals.

## Sample data

These values are synthetic and represent the core submission and webhook shapes:

```json
{
  "transfer_request": {
    "from_account_number": "000123454321",
    "to_account_number": "000987656789",
    "routing_number": "021000021",
    "amount": 250.00,
    "currency": "USD"
  },
  "northwind_acceptance": {
    "id": "trf_55120",
    "status": "PENDING",
    "amount": 250.00,
    "created_at": "2026-07-28T16:22:00Z"
  },
  "webhook": {
    "event": "transfer.updated",
    "transfer_id": "trf_55120",
    "status": "POSTED"
  },
  "ambiguous_internal_result": {
    "vantaca_request_id": "req_demo_001",
    "partner_transfer_id": null,
    "status": "RECONCILIATION_REQUIRED",
    "retry_allowed": false
  }
}
```

Full account/routing values define the inbound protected-data boundary and must follow the approved encryption/access policy. `transfer_id` is a candidate webhook consumer-idempotency value; tests must not assume it is sufficient until Northwind confirms the identity rule.

## Scope

- Define guarded internal transfer states and allowed transitions, including `PENDING`, `POSTED`, `FAILED`, `RETURNED`, and internal `UNKNOWN/RECONCILIATION_REQUIRED`.
- Authenticate/authorize the customer and persist a Vantaca transfer intent before calling Northwind.
- Apply the confirmed partner idempotency/client-reference mechanism; if none exists, keep automatic transfer retry disabled.
- Implement durable webhook receipt, authenticity controls, schema validation, idempotent processing, and transition validation.
- Confirm whether webhook `transfer_id`, `(transfer_id, status)`, or a separate event ID is the supported consumer idempotency key; do not invent the answer.
- Implement bounded reconciliation only when a deterministic partner correlation/lookup is available.
- Add transfer feature gating, kill-switch behavior, sanitized observability, and operator-facing failure states.

## Acceptance criteria

- [ ] Duplicate customer submissions resolve to one Vantaca transfer intent.
- [ ] A definitive Northwind response records the partner ID and correct state exactly once.
- [ ] A timeout/lost response records `UNKNOWN/RECONCILIATION_REQUIRED` and never triggers blind resubmission.
- [ ] Webhook acknowledgement occurs only after the approved durable-receipt/processing boundary.
- [ ] Webhook retries are idempotent without dropping a legitimate later status change for the same transfer.
- [ ] Invalid, regressive, unauthenticated, or suspicious webhook transitions cannot silently change customer-visible state.
- [ ] Reconciliation updates only an exactly correlated transfer and alerts on impossible discrepancies.
- [ ] Read-only launch, transfer kill switch, and support correlation work without redeployment.

## Required testing

- [ ] Test authorization failures, amount/currency boundaries, account validation, and repeated customer clicks.
- [ ] Reproduce definitive acceptance/rejection, pre-commit failure, and post-commit timeout with the mock.
- [ ] Verify no automatic transfer retry after timeout, cancellation, `500`, or `503` unless proven safe.
- [ ] Test duplicate webhook delivery, same transfer with multiple valid statuses, out-of-order/regressive events, malformed payload, and failed authenticity control.
- [ ] Test late `RETURNED` after `POSTED` and repeated current-state notifications.
- [ ] Test concurrent submissions/webhooks and repository transition guards.
- [ ] Test reconciliation success, no match, multiple/unsafe matches, partner outage, and alert emission.
- [ ] Exercise metrics/runbook signals for unknown outcomes, webhook failures, stale transfers, and reconciliation drift.

## Dependencies and handoffs

- Northwind transfer idempotency, exact lookup, webhook identity, ordering, retry, and authenticity answers.
- Approved customer/account authorization and Security controls.
- ENG-01 transport operations and mock scenarios.
- ENG-02 persistence conventions and transactional repository boundary.
- ENG-04 status/ambiguity UX; QA-01 owns adversarial acceptance evidence.

## Out of scope

- Accepting financial/security risk, guessing partner correlation, or using n8n for banking decisions.
- Blind transfer retries or claiming exactly-once processing without an enforceable identity.
- Frontend presentation details.
