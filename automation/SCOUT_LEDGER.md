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

## 2026-09-19 — Windows context / privacy boundary expansion

### YSCJRH/WinChronicle — PROPOSE / extract boundary contracts
- Windows-first, local-first UI Automation context memory; MIT; Python 3.11+ plus .NET UIA helpers; no cloud required.
- Latest inspected release: v0.2.68 (2026-06-29); repository pushed 2026-07-01. Release evidence reports Windows harness + manual UIA smoke. Small adoption footprint means upstream claims must be independently reproduced.
- Strong reusable delta for NEURA: structured UIA metadata before screenshots/OCR; observed UI content explicitly typed as untrusted; fixed read-only context interface; metadata-only mode; confidence/limitations/provenance fields; paths derived from sanitized local session IDs rather than marker-controlled external paths; bounded finite work sessions rather than default background capture.
- Do not import wholesale: Python + .NET helper stack conflicts with NEURA single-binary/Go direction and the project intentionally does not provide desktop control.
- Candidate experiment only: implement a NEURA-native read-only UIA context envelope with fields source/provenance/confidence/limitations/trust=untrusted_observed_content and a metadata-only mode. It must have zero write/control capability and sit outside action execution.
- Required tests before acceptance: malicious UI text cannot become instructions; path/session traversal rejected; no screenshots/OCR/clipboard/audio/network in metadata-only mode; UIA helper failure is fail-closed; Notepad/Edge smoke on real Windows; VS Code/Monaco limitation remains explicit UNKNOWN where UIA coverage is insufficient; RAM/latency benchmark against no-context baseline.
- Expected recurring cost: EUR 0. License permits reuse, but prefer native reimplementation of the narrow contracts rather than dependency import.

### Beledarian/mcp-local-memory — WATCH / memory benchmark
- Local-first SQLite FTS5 + optional sqlite-vec, graph entities/relations, temporal filtering and auditable lifecycle.
- Useful comparison target: time-decayed ranking without rewriting importance on every recall; explicit outdated/incorrect/restored states.
- No integration decision until license, Windows packaging and benchmark are independently verified.

### Sardor-M/lumen — WATCH / scoped-memory reference
- Local SQLite knowledge graph + hybrid retrieval + scope_kind/scope_key isolation + PII scrub + replayable trajectories.
- Useful NEURA delta: hard scope keys on every memory row and secret/PII scrub before journal persistence.
- No wholesale import; deeper license/dependency/Windows validation required.

### AperionAI/shield — WATCH / negative-test corpus reference
- Local MCP guardrail project with explicit Windows/PowerShell destructive-command coverage and install/runtime scanning claims.
- Useful NEURA delta: expand Warden adversarial regression corpus for PowerShell fetch-pipe-IEX, registry deletion, destructive volume commands, force-push and publish/secret mutations.
- Do not import until license, false-positive rate, bypass resistance and runtime overhead are independently tested.

## Candidate priority from scouting
1. Reliability: idempotency-key + intent/result journal for externally mutating actions.
2. Reliability: golden trace behavioral-diff regression gate.
3. Reliability/privacy: NEURA-native read-only Windows UIA context envelope with untrusted-observed-content typing and metadata-only mode.
4. Reliability: deterministic orchestration/non-deterministic activity separation audit.
5. Safety: default-deny network egress and expanded Windows destructive-command negative tests.
6. Memory: scope isolation + explainable/token-bounded recall benchmarks after reliability gates are green.

No item above is PASS merely because the upstream project claims it. All require NEURA-native tests/benchmarks in an isolated candidate before promotion.
