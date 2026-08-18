# Model Routing

## Verification basis

Verified on 2026-08-16 from:

- The current Codex runtime's exposed model overrides.
- Official OpenAI model guidance: https://developers.openai.com/api/docs/models

Current runtime-exposed models:

- gpt-5.6-sol
- gpt-5.6-terra
- gpt-5.6-luna
- gpt-5.5
- gpt-5.4

Official guidance describes Sol for complex professional reasoning/coding, Terra as the intelligence/cost balance, and Luna for cost-sensitive workloads. Runtime availability can change; re-check before assigning a meaningful sub-agent. Do not substitute or invent an unavailable ID.

## Capability tiers

| Tier | Current preferred model | Typical effort | Use |
|---|---|---|---|
| Light | gpt-5.6-luna | low or medium | Bounded, well-specified, low-risk mechanical work |
| Balanced | gpt-5.6-terra | medium or high | Normal implementation, debugging, contracts, and cross-file reasoning |
| Strong | gpt-5.6-sol | high or xhigh | Architecture, financial semantics, security, concurrency, reliability, and adversarial review |
| Highest available | gpt-5.6-sol | max only when justified | Unresolved high-risk conflict or unusually complex production decision |

gpt-5.5 and gpt-5.4 are runtime-available fallbacks, not defaults for new routing while the current 5.6 tiered family is available. Preserve an explicitly requested model.

## Harness routing table

| Harness | Default tier | Recommended available model | Escalation tier | Recommended escalation model | Typical reason |
|---|---|---|---|---|---|
| Architecture | Strong | gpt-5.6-sol, high | Highest available | gpt-5.6-sol, max | Cross-system decisions and overengineering tradeoffs |
| Go Backend | Balanced | gpt-5.6-terra, medium | Strong | gpt-5.6-sol, high | Domain/concurrency or ambiguous partner behavior |
| Database | Balanced | gpt-5.6-terra, medium | Strong | gpt-5.6-sol, high | Transactions, integrity, encryption, or concurrency |
| Frontend | Light | gpt-5.6-luna, medium | Balanced | gpt-5.6-terra, high | Complex financial state or accessibility/workflow ambiguity |
| Integration Reliability | Strong | gpt-5.6-sol, high | Highest available | gpt-5.6-sol, max | ACH idempotency, retries, distributed state, outages |
| QA | Balanced | gpt-5.6-terra, medium | Strong | gpt-5.6-sol, high | Adversarial financial/reliability release scenarios |
| Security | Strong | gpt-5.6-sol, high | Highest available | gpt-5.6-sol, max | Sensitive financial data and trust boundaries |
| Peer Review | Strong | gpt-5.6-sol, high | Highest available | gpt-5.6-sol, max | Adversarial production review across boundaries |
| EM Delivery | Balanced | gpt-5.6-terra, high | Strong | gpt-5.6-sol, high | Conflicting dependencies, timeline, or risk |
| Initialization Evaluator | Light | gpt-5.6-luna, low | Balanced | gpt-5.6-terra, medium | Ambiguous classification or cross-harness routing |

## Task-level overrides

Use Luna even when a harness defaults higher for genuinely mechanical subtasks such as:

- Formatting documentation.
- Generating repetitive test cases from approved acceptance criteria.
- Straightforward UI implementation from an approved design.
- Mechanical migrations or repository boilerplate from an approved schema.
- Changelog updates and narrow refactoring.

Use Sol regardless of the harness default when work involves:

- Monetary retry/idempotency or ambiguous transfer outcomes.
- Security or sensitive-data decisions.
- Cross-system architecture and material tradeoffs.
- Concurrency, ordering, reconciliation, or distributed state.
- Conflicting requirements or adversarial production approval.

## Escalation

1. Start at the lowest tier that reliably fits the evaluated risk.
2. Increase reasoning or move one capability tier when an escalation trigger occurs.
3. Normally allow one escalation attempt.
4. If uncertainty remains or harnesses materially disagree, stop and return DECISION REQUIRED.
5. Never use model escalation as authorization to broaden scope or accept risk.

