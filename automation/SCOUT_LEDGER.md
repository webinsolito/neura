# NEURA Scout Ledger

Progressive technical scouting ledger. Re-check entries only for material new releases, evidence, security findings, or benchmark results.

## 2026-09-19 — durable execution / replay expansion

### clay-good/agent-replay — PROPOSE / extract concept
- 100% local SQLite trace store; time-travel replay, behavioral diff, golden-run regression gates and live tool-call guardrails.
- Useful NEURA delta: canonical trace export + candidate-vs-golden behavioral diff as a CI gate.
- Do not import wholesale: Node.js 20.12+ would add runtime/dependency weight.
- Candidate test: replay the same recorded run on baseline and candidate; fail on tool-order/effect divergence unless explicitly approved.

### avatar-runtime/avatar-engine — CONCEPT ONLY
- Strong crash-safety pattern: append-only step ledger + idempotency key for side effects; explicitly avoids claiming unconditional exactly-once.
- Useful NEURA delta: require an idempotency key for every externally mutating tool action and persist intent/result around dispatch.
- Do not import: Postgres architecture conflicts with lightweight local/single-machine target.

### temporal-sa/go-agent — CONCEPT ONLY
- Deterministic workflow state separated from non-deterministic activities; bounded context; artifact references for large outputs.
- Useful NEURA delta: enforce deterministic orchestration records while wall-clock/network/filesystem/model calls remain worker activities.
- Do not import Temporal stack: too heavy for NEURA zero-cost/local-first target.

### hurtener/Harbor — WATCH / Go-native design reference
- Durable persisted run state, typed event bus, unified pause/resume/cancel/redirect primitive.
- Useful NEURA delta: one serializable control-event mechanism rather than independent interruption paths.
- Requires deeper license/Windows/runtime benchmark before code reuse.

### ionalpha/flynn — WATCH / security architecture reference
- Single Go binary; capability-scoped tools, default-deny egress, governed autonomy, reversible/dry-run actions and tamper-evident records.
- Useful NEURA delta: default-deny egress policy and dry-run/reversible action contract complement Warden leases.
- No wholesale import until Windows containment claims and license/dependency surface are independently tested.

### oxgeneral/agentmem — WATCH / memory benchmark reference
- Single SQLite, hybrid FTS5+semantic retrieval, operational modes from zero dependency to ~12 MB, multi-tier memory.
- Useful NEURA delta: benchmark tiered working/episodic/procedural memory and zero-dependency fallback against current retrieval.
- No replacement until temporal/provenance guarantees match NEURA requirements.

### CaviraOSS/LongMemory — WATCH / temporal-memory reference
- Local-first SQLite, immutable content/provenance/temporal truth, explainable token-bounded evidence selection.
- Useful NEURA delta: evidence selection should expose why each memory entered context and enforce token budget.
- TypeScript engine is not a direct runtime fit; extract semantics, not stack.

## Candidate priority from this scan
1. Reliability: idempotency-key + intent/result journal for externally mutating actions.
2. Reliability: golden trace behavioral-diff regression gate.
3. Reliability: deterministic orchestration/non-deterministic activity separation audit.
4. Safety: default-deny network egress test matrix and dry-run/reversibility contract.
5. Memory: explainable/token-bounded recall benchmark only after reliability gates are green.

No item above is PASS merely because the upstream project claims it. All require NEURA-native tests/benchmarks in an isolated candidate before promotion.
