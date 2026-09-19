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

## 2026-09-19 — multi-agent coordination / self-improvement safety expansion

### Arjia-Labs/clu — MATERIAL CONCEPT / candidate design reference
- MIT; pure-Go single binary; local SQLite through modernc.org/sqlite, no CGO/system database, no daemon/account/network required for the CLI path.
- Reusable delta: atomic task claim using one transactional UPDATE/RETURNING path; dependency graph with cycle detection; TTL locks/heartbeats; explicit human-approval checkpoints; cascading cancellation; append-only work audit.
- Why it matters to NEURA: the current H24 Mission Board is file-based and suitable for one bounded worker, but future subagents can race or duplicate a mission. Atomic claim + lease expiry would make mission ownership crash-safe without adding a cloud coordinator.
- Do not import the whole tracker/UI. Reimplement the narrow claim/lease/dependency contract in NEURA's native coordination store.
- Candidate acceptance tests: 50+ concurrent claim attempts yield one owner per mission; killed worker lease expires and mission becomes reclaimable; dependency cycles rejected; blocked approval cannot be bypassed; cancellation propagates; restart preserves state; stable/main writes remain impossible.
- Benchmark: compare claim latency, binary/RAM delta and recovery time against current file Mission Board. Recurring cost target EUR 0.

### phall1/blackbird — MATERIAL CONCEPT / path-conflict protection
- Go + durable SQLite coordination with repository-scoped agents, tamper-evident event journal, secure resume tokens and shared/exclusive exact/subtree path claims with expiry/renewal/overlap detection.
- Strong reusable delta: path claims before edits and generation numbers to stop two autonomous workers from silently editing overlapping files; refused claim is a normal bounded result rather than a retry-loop error. Also useful: health/status must handshake with the live daemon instead of trusting supervisor metadata, and support bundles redact secrets/home paths.
- Upstream native release path currently targets macOS/Linux, so it is not a direct Windows dependency candidate.
- Decision: extract semantics only. Add a NEURA candidate test where two workers request overlapping path sets; exactly one exclusive claimant may proceed, stale lease can be reclaimed, and an unreachable coordination layer must fail closed for autonomous writes (stricter than Blackbird's optional pre-commit courtesy guard).

### davccavalcante/alkaline — CONCEPT ONLY / durable-loop guardrails
- Apache-2.0; TypeScript/Node 20+, zero required runtime dependencies; deterministic replay, explicit non-deterministic steps, cycle/depth limits, token budgets, durable pause/resume/cancel, heartbeats and zombie reclaim.
- Useful delta: continue-as-new/compaction for never-ending autonomous loops so event history remains bounded, plus explicit cycle/depth budgets independent of model judgement.
- Reject stack import: Node runtime conflicts with NEURA's Go/single-runtime direction; upstream is extremely young and low-adoption, so its guarantees are reference material only.
- Candidate idea: native Go bounded-history checkpoint/continue-as-new test with deterministic replay before and after compaction.

### argszero/emrg — CONCEPT ONLY / self-evolution dirty-tree guard
- MIT self-evolving experiment. Its evolution cycle modifies only its own evolution checkout, runs tests/import checks, rolls back failures, and forces read-only mode when the evolution repository has uncommitted changes.
- Reusable delta: NEURA autonomous evolution should refuse mutation when its candidate workspace is dirty or provenance is ambiguous; it may audit/report but must not overwrite unexplained human/worker changes.
- Do not import Python stack or treat its tests as NEURA evidence. Add a candidate negative test: dirty worktree => READ_ONLY, no source mutation, no promotion; clean isolated candidate => mutation allowed only through normal Warden/test gates.

## Candidate priority from scouting
1. Reliability: atomic mission claim + expiring lease + dependency/cycle gate for H24/subagent coordination.
2. Reliability: path-level edit claims for concurrent workers, with fail-closed autonomous writes when coordination state cannot be verified.
3. Reliability: idempotency-key + intent/result journal for externally mutating actions.
4. Reliability: golden trace behavioral-diff regression gate.
5. Reliability/privacy: NEURA-native read-only Windows UIA context envelope with untrusted-observed-content typing and metadata-only mode.
6. Reliability: deterministic orchestration/non-deterministic activity separation plus bounded-history compaction.
7. Safety: dirty candidate workspace forces autonomous evolution into read-only mode.
8. Safety: default-deny network egress and expanded Windows destructive-command negative tests.
9. Memory: scope isolation + explainable/token-bounded recall benchmarks after reliability gates are green.

No item above is PASS merely because the upstream project claims it. All require NEURA-native tests/benchmarks in an isolated candidate before promotion.
