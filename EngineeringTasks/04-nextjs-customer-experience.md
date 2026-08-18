# [ENG-04] Build the Next.js customer integration experience

**Owner:** Engineer 4 — Next.js Customer Experience  
**Suggested labels:** `frontend`, `customer-experience`, `accessibility`  
**Primary AI harness:** [Frontend Harness](../AI-Harnesses/frontend-harness.md)  
**AI initialization:** Start with the [Initialization Evaluator](../AI-Harnesses/initialization-evaluator.md).  
**Authorization:** Frontend implementation begins only after the Integration Lead authorizes it and Product approves the required financial wording and launch scope.

## Goal

Build a clear, responsive customer experience for linked accounts, balance freshness, recent transactions, and feature-gated transfers without moving financial rules or Northwind access into the browser.

## Architecture placement and application dependencies

**Placement:** This feature lives in the Next.js web application. It consumes only Vantaca Go APIs and a Vantaca-approved invalidation channel; it never connects directly to SQL Server, Northwind, the mock, n8n, or the encryption key service. See the [shared runtime dependency map](README.md#architecture-placement-and-runtime-dependencies).

| Relationship | Application/component | Contract |
|---|---|---|
| Receives from | Vantaca identity/session platform | Authenticated customer/tenant context suitable for Go API calls |
| Calls | Go account/transaction API | Masked account data, exact money representation, freshness/version, recent transactions, and typed failures |
| Calls | Go transfer API | Feature-gated validation/submission/status behavior and safe support correlation |
| Receives from | Frontend invalidation transport | Non-sensitive account/version change signal that causes a bounded Go API re-fetch |
| Depends on | Product/QA wording and accessibility standards | Approved stale/unknown/failed/returned language, responsive behavior, and accessible interactions |
| Produces | Customer and support workflow | Clear read state, transfer confirmation/status, and no direct partner or protected-data exposure |

## Architecture workflow logic

This issue renders the customer-facing side of Major Workflows 1–4 in [StartHere.md](../Notes/StartHere.md):

1. Account and balance pages read masked, freshness-aware SQL projections through the Go API.
2. Recent transactions render the current snapshot immediately. After the async worker commits a mismatch, an invalidation signal causes one bounded API re-fetch; matching data sends no signal.
3. Transfer review prevents accidental duplicate clicks and displays definitive, pending, ambiguous, failed, posted, and returned states without changing server truth.
4. Webhook-driven changes appear only after the backend validates and persists them; the browser never processes Northwind webhook payloads directly.

## Sample data

This synthetic client-safe view model illustrates what the UI may receive. It intentionally contains masked rather than full account values:

```json
{
  "account": {
    "id": "acc_1029",
    "display_name": "Checking ••••4321",
    "balance": "4820.55",
    "currency": "USD",
    "status": "open",
    "freshness": {
      "fetched_at": "2026-08-16T18:00:00Z",
      "state": "current"
    }
  },
  "recent_transactions": [
    {
      "id": "txn_88213",
      "amount": "-42.17",
      "currency": "USD",
      "description": "COFFEE HOUSE #42",
      "posted_at": "2026-07-21T14:03:00Z"
    }
  ],
  "invalidation": {
    "type": "recent_transactions.updated",
    "account_id": "acc_1029",
    "version": 7
  },
  "transfer_status": {
    "request_id": "req_demo_001",
    "status": "UNKNOWN",
    "message": "Submission is being verified. Do not submit it again."
  }
}
```

The wording, freshness threshold, and exact money representation remain subject to Product/API approval. The invalidation contains only enough data to re-fetch; it is not a financial-state payload.

## Scope

- Create strongly typed Vantaca API boundaries for accounts, freshness, transactions, transfer submission, and transfer status.
- Display masked account identity, currency, fetched-at/stale/unavailable states, and trustworthy prior data.
- Implement recent-transactions invalidation so the UI re-fetches the Vantaca API after ENG-02 commits a mismatch update.
- Implement transfer review/confirmation and clear `PENDING`, `UNKNOWN`, `FAILED`, `POSTED`, and `RETURNED` experiences when transfer scope is enabled.
- Prevent repeated clicks while preserving navigation and support/recovery information.
- Meet responsive, keyboard, screen-reader, focus, and error-summary expectations using the existing Next.js/TypeScript/Tailwind baseline.

## Acceptance criteria

- [ ] The browser calls only Vantaca APIs and never receives Northwind credentials or direct partner access.
- [ ] Account numbers are masked in rendered HTML, telemetry, errors, and screenshots.
- [ ] Balance and transaction views communicate fetched-at, stale, unavailable, empty, and loading states without calling cached data “live.”
- [ ] A recent-transaction invalidation triggers one bounded re-fetch after the database commit; no invalidation means no refresh loop.
- [ ] Failed refresh preserves the last trustworthy view and displays the approved stale/error state.
- [ ] Transfer confirmation shows source, destination, amount, currency, and non-final status meaning before submission.
- [ ] `UNKNOWN` does not appear as failed or encourage resubmission; `RETURNED` remains possible after `POSTED`.
- [ ] Read and transfer features can be independently gated.

## Required testing

- [ ] Component/behavior tests cover loading, empty, populated, stale, unavailable, and validation states.
- [ ] Test Workflow 2 mismatch invalidation/re-fetch, matching no-op, refresh failure, and repeated invalidation coalescing.
- [ ] Test duplicate-submit prevention and `PENDING`, `UNKNOWN`, `FAILED`, `POSTED`, and `RETURNED` rendering.
- [ ] Test API error categories without exposing sensitive server details.
- [ ] Run keyboard, focus, label, screen-reader-name, contrast, and responsive-layout checks on critical flows.
- [ ] Add a small end-to-end happy path and highest-risk failure paths using synthetic data and the local stack.
- [ ] Verify no secrets, full account/routing numbers, or sensitive transfer payloads reach client logs/telemetry.

## Dependencies and handoffs

- Product/QA wording for freshness, mismatch, ambiguity, returns, and unavailable states.
- Approved Vantaca authentication and tenant context.
- ENG-02 account/transaction/freshness API plus invalidation contract.
- ENG-03 transfer API, status model, feature gate, and support correlation.
- QA-01 acceptance matrix and accessibility/release evidence.

## Out of scope

- Reimplementing backend validation, retry rules, transfer state transitions, or Northwind logic in TypeScript.
- Creating a speculative design system or adding unrelated frontend dependencies.
- Silently deciding Product risk language.
