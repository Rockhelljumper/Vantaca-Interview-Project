# Northwind Integration Interview --- Concerns, Questions, Comments, and Notes

# 1\. Highest-Priority Concerns

## CRITICAL --- Customer identity / account linking is not defined

### Observation

The documentation says:

`GET /accounts` returns all accounts linked to **the customer**.

But the documented request appears to contain only the partner API key.

No customer identifier, customer token, session context, OAuth flow, account-linking flow, or customer-scoped credential is documented.

### Why this matters

We cannot safely design:

- account retrieval;

- tenant isolation;

- authorization;

- caching;

- database relationships;

- transfers;

until we understand how Northwind determines which customer's accounts we're requesting.

### Question for Northwind

> How is the Northwind customer identified and authorized? Is there another customer-scoped token or identifier omitted from this document?

### Production blocker

**Yes.**

---

## CRITICAL --- Transfer idempotency is undefined

### Observation

`POST /transfers` creates an ACH transfer.

No idempotency key or client-generated transaction identifier is documented.

### Failure scenario

1.  We submit a $5,000 transfer.

2.  Northwind receives and creates it.

3.  Our connection times out before receiving the response.

4.  We do not know whether the transfer exists.

5.  Blind retry risks creating another $5,000 transfer.

### Position

Never automatically retry monetary POST operations unless Northwind provides a safe idempotency mechanism or a reliable reconciliation strategy.

### Questions

> Does `/transfers` support an undocumented idempotency key?

> Can we supply our own external transfer/reference ID?

> If a transfer request times out, how should we determine whether Northwind accepted it?

### Production blocker

**Yes.**

---

## CRITICAL --- Unsigned webhook

Northwind currently provides no request signature.

The stakeholder asks us to trust the payload.

The public documentation recommends IP allowlisting.

### Risk

A transfer-state webhook changes financial state in our application.

We should not blindly accept:

```
transfer.updated
POSTED

```

from an unauthenticated internet caller.

### Questions

> Can you provide stable outbound IP ranges before production?

> Is webhook signing on the roadmap?

> Could we use mTLS or another authentication mechanism?

### Recommended interim strategy

Treat webhook state as a signal requiring controlled processing/reconciliation rather than blindly trusting arbitrary state transitions.

---

# 2\. Major Documentation Contradictions

## HIGH --- "No rate limits" vs HTTP 429

Northwind documentation states there are no rate limits.

The documented error table nevertheless defines:

`429 Too Many Requests`

with a `Retry-After` response.

### Position

Design the adapter to handle `429` regardless of the verbal guarantee.

### Question

> Under what circumstances can 429 occur if the API has no rate limit?

---

## HIGH --- "Always available" vs documented maintenance

Northwind repeatedly describes the API as always available.

Their error documentation explicitly says `503` can occur because the API is occasionally taken down for maintenance.

Dana additionally says downtime handling is unnecessary.

### Position

External dependencies fail.

We should implement bounded timeout, appropriate retry/backoff for safe operations, graceful degradation, telemetry, and reconciliation.

Do not argue with the partner.

Just build an integration that fails safely.

---

## HIGH --- "Exactly once" webhooks vs retry delivery

Northwind calls webhook delivery **exactly once**, then states failed deliveries will be retried.

Those statements do not form a useful exactly-once guarantee for the consumer.

### Position

Webhook processing must be idempotent.

### Questions

> Does Northwind provide an event ID?

> Can the same transfer state be delivered more than once?

> Can events arrive out of order?

---

# 3\. Five-Second Balance Polling

## Stakeholder request

Northwind recommends polling `GET /accounts` every five seconds for every linked customer and persisting the response locally.

They also recommend treating our persisted copy as the source of truth displayed to the customer.

## Concerns

### Scaling

Five-second polling becomes significant quickly.

Example:

```
10,000 customers / 5 seconds
= approximately 2,000 customer syncs per second

```

That does not include:

- pagination;

- retries;

- database writes;

- multiple instances;

- transaction synchronization.

### Definition of "real-time"

A locally stored value refreshed every five seconds is not literally Northwind's real-time value.

We should define acceptable staleness.

### Source-of-truth terminology

Northwind is ultimately authoritative for the balance.

Our database is better described as:

- local read model;

- cache;

- synchronized representation;

rather than the ultimate financial source of truth.

### Questions

> What measurable freshness SLA does Product actually need?

> Is five seconds a hard requirement or a suggested implementation?

> Can Northwind provide change notifications for balances?

> Can we refresh on customer activity and use a less aggressive background reconciliation cadence?

> Are there expected customer-volume numbers we should design around?

---

# 4\. Sensitive Account Information

## Stakeholder request

Northwind asks us to persist full:

- account numbers;

- routing numbers;

so they are available for transfers.

## Concern

Convenience alone is not sufficient justification for permanently storing highly sensitive banking identifiers in plaintext.

## Questions

> Can `/transfers` accept Northwind account IDs rather than full account numbers?

> Can Northwind provide tokens representing accounts?

> Are full account numbers required for every transaction because of a contractual limitation, or simply the current API implementation?

> What Vantaca security/compliance standards apply to account-number storage?

## Potential design if storage is unavoidable

- encryption at rest;

- restricted application access;

- masking in UI;

- aggressive log redaction;

- no account numbers in telemetry;

- auditable access;

- documented retention;

- key management through Vantaca's production secret/key infrastructure.

Do not invent exact compliance obligations before confirming them.

---

# 5\. Authentication Concerns

Northwind requires a partner API key in the **query string**.

The same key is described as being used across all environments.

## Concerns

Query parameters frequently appear in:

- reverse-proxy logs;

- request logs;

- observability products;

- browser/history tooling if ever exposed incorrectly;

- debugging output.

Using the same credential across environments also creates unnecessary blast radius.

## Questions

> Can the credential be supplied through an Authorization header instead?

> Are sandbox and production truly using the same credential?

> How is the key rotated?

> Can multiple active keys exist during rotation?

> What is the revocation process?

> Does the credential identify only Vantaca, or does it participate in customer authorization?

---

# 6\. Undocumented Internal Endpoint

Product mentions:

`/internal/accounts/full`

Northwind apparently said informally that Vantaca can use it.

It is not in the public contract.

## Position

Do not make production depend on an undocumented endpoint based solely on an informal conversation.

Possible response:

> I'm happy to evaluate it, but before we make it part of the production design I'd want Northwind to formally document and support the endpoint as part of our integration contract.

## Questions

> Is it versioned?

> Is it supported under the same availability expectations?

> Can Northwind change it without notice?

> Is it intended for third-party integrations?

> Can it be added to the official partner specification?

---

# 7\. Pagination Is Underspecified

The documentation says:

- page number;

- 50 records per page;

- bare JSON array response.

It does not document:

- total records;

- total pages;

- continuation token;

- `has_more`;

- stable sort order.

## Questions

> Is an empty page the official pagination terminator?

> Is result ordering stable while pagination occurs?

> What happens if accounts/transactions are added while pages are being fetched?

This matters for reliable synchronization.

---

# 8\. Transfer Routing-Number Ambiguity

The transfer payload includes:

```
{
  "from_account_number": "...",
  "to_account_number": "...",
  "routing_number": "..."
}

```

There are two accounts but only one routing number.

## Questions

> Does `routing_number` apply to the source or destination account?

> Are both accounts guaranteed to be at Northwind and therefore use the same routing number?

> Is this transfer exclusively between a customer's Northwind accounts, or can another financial institution be involved?

Do not guess this in production code.

---

# 9\. Transfer Lifecycle

Documented states:

```
PENDING
POSTED
FAILED
RETURNED

```

Important distinction:

`POSTED` is not necessarily terminal forever because the transfer can later become `RETURNED`.

## Design implication

Our internal state model must support late state changes.

Do not model:

```
POSTED = immutable success

```

without confirmation.

## Questions

> What is the maximum period during which a POSTED transfer may become RETURNED?

> Are there any additional states not documented?

> Can states arrive out of order through webhook delivery?

> Can a FAILED transfer ever later transition?

---

# 10\. Reconciliation

Northwind provides:

`GET /transfers`

but the documentation does not describe a single-transfer lookup endpoint.

## Questions

> Is `GET /transfers/{id}` available?

> Can transfer lists be filtered by updated timestamp or status?

> How should partners efficiently reconcile large transfer histories?

> What retention window applies?

## Architecture position

Webhook delivery should not be the only source of transfer-state recovery.

A periodic reconciliation mechanism is appropriate.

---

# 11\. n8n Position

Use n8n as an **operational orchestration layer**, not the owner of financial business rules.

## Good n8n use cases

- periodic reconciliation;

- stale account refresh;

- failed integration retry orchestration;

- health verification;

- alerting/escalation;

- operational workflow history.

## Architecture

```
n8n
  |
  | scheduled trigger
  v
Internal Go Sync/Reconciliation Endpoint
  |
  | batching / concurrency / business rules
  v
Northwind Adapter
  |
  v
Northwind

```

## Why use it

Do not say:

> Cron is unreliable.

Say:

> Cron and Task Scheduler are perfectly reasonable for isolated jobs. As recurring integration processes grow, however, scheduling, retries, credentials, execution history, alerting and ownership become fragmented. A workflow orchestrator gives the team a centralized operational control plane.

## Demo balance

n8n should be:

- included;

- documented;

- importable;

- optional.

Use an optional Docker Compose profile.

The basic demo should not depend on it.

This shows production thinking without making the architecture unnecessarily complex.

---

# 12\. Northwind Mock

A local Northwind mock is worth building.

It turns the repository from a code sample into a working integration demonstration.

## Happy-path behaviors

- accounts;

- balances;

- transactions;

- transfer submission;

- transfer state changes;

- webhooks.

## Valuable failure simulation

- 401;

- 429 + Retry-After;

- 500;

- 503;

- artificial latency;

- timeout;

- FAILED transfer;

- RETURNED transfer;

- repeated webhook;

- delayed webhook.

## Important principle

The mock should test our assumptions.

It should not become a second giant application.

---

# 13\. Additional Schema/Documentation Inconsistencies

## Transactions

The documented transaction field table contains:

`merchant_category_code`

but the example response does not show that field.

### Question

> Is `merchant_category_code` nullable/optional, or is the example incomplete?

---

## Accounts

The account response example includes:

`currency`

but the accompanying field table does not list it.

### Question

> Is currency guaranteed on every account response?

These are not necessarily major blockers, but noting them demonstrates careful contract review.

---

# 14\. Money Representation

Northwind's JSON represents monetary amounts as JSON numbers.

Internally, avoid binary floating-point for financial calculations.

Possible approaches:

- integer minor units/cents;

- fixed decimal library/type;

- SQL `DECIMAL`.

The exact representation should be an explicit engineering decision.

---

# 15\. Error Model

Northwind provides both:

- HTTP status;

- JSON error code.

Our adapter should normalize partner-specific failures into internal typed categories such as:

```
Validation
Authentication
NotFound
Throttled
PartnerUnavailable
Timeout
UnknownTransferOutcome
UnexpectedPartnerResponse

```

The rest of the application should not need to understand Northwind-specific HTTP semantics.

---

# 16\. Observability

Useful identifiers:

```
Internal correlation ID
Vantaca customer ID
Internal transfer ID
Northwind transfer ID
Northwind request ID if provided

```

Useful metrics:

- Northwind request duration;

- requests by endpoint/status;

- 429 frequency;

- 5xx frequency;

- reconciliation lag;

- stale account count;

- webhook processing failures;

- transfers by current state;

- unknown transfer outcomes.

Never log:

- API keys;

- complete account numbers;

- sensitive transfer payloads unnecessarily.

---

# 17\. Proposed Architectural Boundary

```
                   Next.js
                      |
                      v
                 Vantaca API
                    (Go)
                      |
          +-----------+-----------+
          |                       |
          v                       v
    Application Layer       Domain Models
          |
    +-----+------+
    |            |
    v            v
Repository   Northwind Adapter
    |            |
    v            v
 MSSQL      Northwind / Mock

```

Benefits:

- external API changes remain localized;

- Northwind DTOs do not become our application model;

- adapter is mockable;

- repositories can be tested independently;

- frontend remains independent of partner contract quirks.

---

# 18\. Raw SQL Position

Use:

- MSSQL 2022;

- Docker;

- `database/sql`;

- parameterized raw SQL;

- explicit migrations.

Reasoning:

> We don't yet know Vantaca's persistence conventions. Raw SQL keeps the demo dependency-light and makes the actual database behavior obvious. The repository boundary lets the implementation adapt later to whatever database access approach the team standardizes on.

Avoid framing this as:

> ORMs are bad.

Frame it as reducing unknown conventions and unnecessary dependencies in the exercise.

---

# 19\. Demo vs Production

## Demo

Likely:

```
Docker Compose

Next.js
Go API
MSSQL 2022
Northwind Mock
Optional n8n profile

```

## Production

Do not assume the local topology.

Questions include:

- What does Vantaca deploy on?

- Existing database standards?

- Existing secrets platform?

- Existing event/queue infrastructure?

- Existing observability?

- Existing scheduler/orchestrator?

- Existing CI/CD?

- Existing API gateway?

The production architecture should integrate with those standards instead of creating duplicate infrastructure.

---

# 20\. Team Plan

Available:

- Engineering Manager / Integration Lead

- 4 Developers

- 1 QA/PM

A reasonable initial parallelization model:

## EM / Integration Lead

Own:

- architecture;

- Northwind technical discussions;

- Product alignment;

- risk register;

- design decisions;

- sequencing;

- code/design review;

- production-readiness decision;

- unblocking engineers.

Do not personally become the bottleneck implementing every feature.

---

## Developer 1 --- Partner Integration

Potential ownership:

- Northwind client;

- authentication;

- DTO handling;

- pagination;

- retry/backoff;

- mock contract behavior;

- adapter tests.

---

## Developer 2 --- Persistence / Synchronization

Potential ownership:

- MSSQL schema;

- raw SQL repositories;

- account synchronization;

- transaction synchronization;

- migrations;

- reconciliation data structures.

---

## Developer 3 --- Transfers / Reliability

Potential ownership:

- transfer application service;

- state machine;

- idempotency strategy;

- webhook processing;

- reconciliation;

- failure recovery.

This is probably the highest-risk engineering stream.

---

## Developer 4 --- Customer Experience / Full Stack

Potential ownership:

- Next.js account view;

- transactions;

- transfer UX;

- status/error states;

- integration with internal Go APIs.

Could also assist with test harness/demo tooling depending on skill set.

---

## QA / PM

Should participate from the beginning.

Ownership:

- acceptance criteria;

- partner/product questions;

- test scenarios;

- traceability;

- failure cases;

- test execution;

- coordination of demo readiness;

- Product validation.

Important failure scenarios should be designed before development completes.

---

# 21\. Two-Week Timeline

Product has promised Northwind a two-week go-live.

Do not immediately say:

> impossible.

Also do not say:

> sure.

Say:

> I can build a plan around two weeks, but I need to separate implementation effort from unanswered production blockers.

## Illustrative sequence

### Days 1--2

- resolve critical API questions;

- integration spike;

- architecture;

- mock;

- schema;

- contracts;

- QA scenarios.

### Days 2--5

Parallel:

- accounts;

- transactions;

- persistence;

- frontend foundation;

- transfer client.

### Days 4--7

- transfer workflow;

- webhook;

- reconciliation;

- error behavior;

- UI integration.

### Days 6--9

- contract testing;

- integration testing;

- failure testing;

- security review;

- operational tooling;

- observability.

### Days 9--10

- production-readiness review;

- critical fixes;

- rollout decision;

- pilot preparation.

This timeline assumes Northwind quickly answers blocking questions.

Unresolved authentication/idempotency/security concerns can block GA even if all software development is finished.

---

# 22\. Questions for Northwind

Prioritize these live.

1.  How is the end customer identified and authorized when calling `/accounts`?

2.  What is the account-linking/authentication flow?

3.  Does `POST /transfers` support an idempotency key or client reference?

4.  What should we do after a transfer submission times out with an unknown result?

5.  Why is `429` documented if there are no rate limits?

6.  What maintenance/availability expectations should we actually design for?

7.  Can the API key be sent in a header rather than query string?

8.  Why is the same API key used across environments?

9.  What is the credential rotation process?

10. Can transfers use account IDs instead of full account numbers?

11. Which account does the single `routing_number` field represent?

12. Are transfers exclusively between Northwind-held accounts?

13. Can you provide stable webhook source IPs before production?

14. Is webhook signing planned?

15. Can webhook events be duplicated?

16. Can webhook events arrive out of order?

17. Is there an event ID?

18. What is the official pagination termination mechanism?

19. Is pagination ordering stable?

20. Is `/internal/accounts/full` officially supported for partner use?

21. Can that endpoint be added to the supported public contract?

22. Is `GET /transfers/{id}` available?

23. Can transfers be filtered by updated timestamp?

24. How long after POSTED can a transfer become RETURNED?

25. Are there additional transfer states?

26. Is `merchant_category_code` optional?

27. Is account `currency` always present?

---

# 23\. Questions for Product

1.  What does "real-time balance" mean measurably?

2.  Is five-second polling a requirement or Northwind's suggested implementation?

3.  What customer scale should we design for?

4.  What portion of the two-week date is externally contractual versus internally targeted?

5.  Would Product accept a staged rollout?

6.  What is the minimum launch scope?

7.  Are read-only accounts/transactions separable from transfer enablement?

8.  Is transfer initiation allowed to launch later if banking-security questions remain unresolved?

9.  What user experience is expected when Northwind is unavailable?

10. How stale may a displayed balance be?

11. What support expectations exist after launch?

12. Who owns customer communication if transfers become delayed or returned?

13. What analytics/product telemetry is required?

14. What constitutes successful go-live?

---

# 24\. Questions for Vantaca Engineering

Good questions for the peer-review portion:

- What is the team's preferred persistence/data-access pattern?

- How are integrations typically isolated today?

- Is there an existing background-job/orchestration platform?

- How do you currently manage external API credentials?

- What observability stack does the team use?

- Is there an existing queue/event bus?

- What does deployment look like?

- How are partner contract tests handled?

- Who owns integrations after launch?

- Does the team prefer service ownership or shared platform ownership?

- How are architecture decisions documented?

- How closely does QA work with engineering during implementation?

---

# 25\. Good Interview Phrases

> "I'm separating what Northwind asked us to accomplish from the implementation they suggested."

> "I want to preserve the two-week goal, but first I need to identify which unknowns affect schedule and which ones actually block production."

> "I don't need Northwind to have perfect reliability. I need our side to fail predictably when they don't."

> "A timeout on a read and a timeout while moving money are fundamentally different failure modes."

> "I'd rather make webhook processing idempotent even if the partner believes delivery is exactly once."

> "I'm treating the local database as a synchronized read model, not redefining Northwind's financial data as ours."

> "For this prototype I'm choosing the simplest architecture I can defend. In production I'd integrate into Vantaca's existing platform standards rather than introduce parallel infrastructure."

> "I use AI aggressively for implementation acceleration, test generation and adversarial review, but architectural risk acceptance remains a human decision."

> "My job as the lead isn't to personally write all five workstreams. It's to make sure five workstreams can move safely in parallel."

---

# 26\. What Not to Overdo

Avoid adding complexity solely to impress reviewers.

Do not build:

- Kubernetes;

- Kafka;

- many microservices;

- elaborate event sourcing;

- excessive frontend polish;

- huge workflow engines;

- an enterprise IAM platform;

- dozens of diagrams.

The strongest version of this project is:

**small, runnable, deliberate, well-documented, failure-aware, and easy to explain.**

---

# 27\. Presentation Balance

The repository should communicate:

> This person put serious thought into this.

Not:

> This person spent a week trying to manufacture a production bank.

The ideal "extra effort" artifacts are:

- working end-to-end demo;

- Northwind mock;

- useful failure simulation;

- excellent README;

- 4--6 strong diagrams;

- clear risk/questions document;

- realistic team plan;

- migration scripts;

- tests;

- optional n8n workflow.

That is enough to knock their socks off without looking like unnecessary theater.
