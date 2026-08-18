# Northwind/Vantaca AI Harnesses

## What these are

These are small, reusable instruction overlays for post-discovery work. They narrow an AI assignment to one specialty, reduce irrelevant context, and make model selection and escalation explainable.

They augment the primary project instructions; they never override them. Authority remains:

1. Current Integration Lead instructions.
2. Candidate brief and documented Northwind contract.
3. Approved decisions in ../AI-Log/DECISIONS.md.
4. Relevant open questions and concerns.
5. The selected harness.
6. Explicitly labeled assumptions.

Harnesses may recommend. Humans own architecture, risk acceptance, stakeholder decisions, and production approval.

## Phases

- **DESIGN:** Use architecture, reliability, security, database, backend, frontend, QA, or EM delivery as needed.
- **IMPLEMENTATION:** Load the implementation specialty plus only the risk harnesses relevant to the assigned boundary.
- **DEBUG:** Start with the owning implementation harness; add reliability, database, security, or QA only when evidence points there.
- **REVIEW:** Use peer review as primary, with targeted security/reliability/QA support.
- **DISCOVERY:** Use only for scoped analysis; harnesses do not authorize implementation.

## How to load a harness

1. Run initialization-evaluator.md in the primary session.
2. Decide whether the primary agent can do the work without a sub-agent.
3. Select one primary harness. Add supporting harnesses only for a real cross-boundary risk.
4. Supply the minimum context packet named by the evaluator.
5. Record model and escalation choices for meaningful delegated work.
6. Return findings to the primary session using OBSERVATION, RECOMMENDATION, DECISION REQUIRED, or AUTHORIZED IMPLEMENTATION.

Never load every harness by default.

## Usage matrix

| Task | Primary harness | Supporting harnesses |
|---|---|---|
| Design transfer workflow | architecture-harness | integration-reliability-harness, security-harness |
| Implement Northwind client | go-backend-harness | integration-reliability-harness |
| Build database schema/repositories | database-harness | security-harness |
| Build account/transaction/transfer UI | frontend-harness | qa-harness |
| Diagnose webhook/reconciliation failure | integration-reliability-harness | go-backend-harness, database-harness |
| Define acceptance/release evidence | qa-harness | em-delivery-harness |
| Production design/code review | peer-review-harness | security-harness, integration-reliability-harness, qa-harness |
| Estimate and sequence delivery | em-delivery-harness | architecture-harness, qa-harness |

## Model routing

MODEL-ROUTING.md records the current runtime-supported model IDs and tier defaults. Re-check runtime availability whenever meaningful work is assigned. Prefer:

- Luna for bounded, well-specified, low-risk mechanical work.
- Terra for normal implementation and medium-complexity reasoning.
- Sol for architectural, financial, security, concurrency, reliability, and adversarial review work.

Reasoning effort is part of routing. Escalate one tier or increase effort only when the evaluator's triggers are met.

## Initialization evaluator

initialization-evaluator.md answers:

- Is a harness needed?
- Is a sub-agent useful?
- Which one primary harness applies?
- What is the minimum context?
- What is the lowest reliable model tier?
- What justifies one-step escalation?

Correctness comes first. Token reduction never justifies weak reasoning about money movement, security, distributed state, or production risk.

## Handoff contract

A harness returns only assignment-relevant material:

- Evidence and files inspected.
- OBSERVATIONS.
- RECOMMENDATIONS with tradeoffs.
- DECISIONS REQUIRED with accountable owner.
- Authorized changes and validation performed, when implementation was explicitly approved.
- Remaining risks, confidence, and escalation need.

Material conflicts return to the primary session. Harnesses do not resolve one another's disagreements autonomously.

## Intentional overlap boundaries

- Architecture owns boundaries and tradeoffs; reliability owns failure semantics.
- Go backend owns application code; database owns persistence invariants.
- Security defines controls and questions; it does not invent compliance obligations.
- QA defines evidence and scenarios; peer review critiques the finished result.
- EM delivery owns sequencing and ownership; it does not make domain architecture decisions.
- Observability stays inside reliability, QA, and EM delivery instead of becoming a separate harness.
- Documentation formatting stays with the primary agent or Luna-level mechanical work.

No additional harness is currently justified.

