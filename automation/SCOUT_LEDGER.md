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
- Strong reusable delta: structured UIA metadata before screenshots/OCR; observed UI content explicitly typed as untrusted; fixed read-only context interface; metadata-only mode; confidence/limitations/provenance fields.
- Do not import wholesale: Python + .NET helper stack conflicts with NEURA single-binary/Go direction.

### Beledarian/mcp-local-memory — WATCH / memory benchmark
- Local-first SQLite FTS5 + optional sqlite-vec, graph entities/relations, temporal filtering and auditable lifecycle.

### Sardor-M/lumen — WATCH / scoped-memory reference
- Local SQLite knowledge graph + hybrid retrieval + scope isolation + PII scrub + replayable trajectories.

### AperionAI/shield — WATCH / negative-test corpus reference
- Useful NEURA delta: expand Warden adversarial regression corpus for destructive PowerShell and secret/publish mutations.

## 2026-09-19 — multi-agent coordination / self-improvement safety expansion

### Arjia-Labs/clu — MATERIAL CONCEPT / candidate design reference
- MIT; pure-Go single binary; local SQLite; atomic task claim, dependency graph, TTL locks/heartbeats and approval checkpoints.
- Candidate acceptance: concurrent claim attempts yield one owner; killed worker lease expires; cycles rejected; approval cannot be bypassed; restart preserves state.

### phall1/blackbird — MATERIAL CONCEPT / path-conflict protection
- Reusable delta: path claims before edits, generation numbers, stale lease recovery, live-daemon handshake and redacted support bundles.

### davccavalcante/alkaline — CONCEPT ONLY / durable-loop guardrails
- Useful delta: continue-as-new/compaction, explicit cycle/depth limits, heartbeat/zombie reclaim.

### argszero/emrg — CONCEPT ONLY / self-evolution dirty-tree guard
- Useful delta: dirty or provenance-ambiguous candidate workspace forces autonomous evolution into read-only mode.

## 2026-09-19 — MCP trust-boundary + pure-Go inference expansion

### hoophq/mcpproxy — MATERIAL CONCEPT / MCP schema rug-pull guard
- MIT, Go, young project (created 2026-07-28; inspected repository pushed 2026-08-06). Windows amd64/arm64 binary release path is documented upstream; local stdio backends are supported.
- Strong reusable delta for NEURA: fingerprint each MCP tool definition (name + description + input schema) at admission/session start. If the definition changes mid-session or after an unapproved update, invalidate the tool/capability lease and require re-admission instead of silently trusting the new schema.
- Additional useful contracts: strip denied tools from discovery; deny server-originated sampling/elicitation by default; per-session call budgets; result-size limits; append-only per-session replay; encrypted outbound token state.
- Why material: Warden can correctly authorize a tool name while the server behind that name has changed its schema/semantics. Definition pinning closes this TOCTOU/supply-chain gap before execution.
- Do not import the whole gateway yet: project is young and low-adoption; NEURA already has Warden/audit infrastructure. Reimplement the narrow fingerprint + session invalidation contract natively first.
- Candidate tests: admitted schema works; description/schema mutation with same tool name => lease invalidated/fail closed; tool-list add/remove => explicit re-admission; malicious server sampling/elicitation => denied; restart preserves approved fingerprint; no network needed for local stdio; benchmark added latency/RAM.
- Recurring cost target: EUR 0. Status: PROPOSED, NOT PASS.

### P4ST4S/mcp-audit — MATERIAL REFERENCE / signed MCP evidence boundary
- Apache-2.0, Go, active repository with 124 commits and current security/audit documentation.
- Reusable delta: transparent interception of MCP JSON-RPC, synchronous allow/deny before tools/call, sensitive-field redaction before persistence, optional HMAC-SHA256 evidence, and a crucial reliability rule: tools/call is not automatically retried.
- NEURA extraction: bind MCP tool-call evidence to Warden lease ID + pinned tool-definition hash + raw-argument digest; never auto-retry a mutating tool unless its idempotency contract explicitly permits it.
- Do not import wholesale before overlap/latency benchmark versus existing Warden ledger. Candidate negative tests should include secret redaction, denied-call evidence, signature tamper detection and duplicate-side-effect prevention.
- Recurring cost target: EUR 0. Status: DESIGN INPUT, NOT PASS.

### townsendmerino/goinfer — WATCH/BENCHMARK / pure-Go static local inference
- MIT; pure Go/no-CGO; repository inspected as pushed 2026-09-19; supports GGUF/safetensors and advertises model embedding into a static executable. Very young/small adoption (13 stars at inspection), so upstream performance claims are not accepted as NEURA evidence.
- Potential NEURA delta: eliminate Python/CUDA-toolkit/llama.cpp runtime dependency for a small fallback model and optionally ship a model+runtime as one static artifact. This aligns unusually well with NEURA's single-runtime/Windows/offline goal.
- Major caution: model redistribution license still applies separately; model-family parity, Windows performance, quantization quality, memory mapping, GPU paths and long-context stability require independent tests.
- Candidate benchmark only: same small Qwen/Gemma checkpoint vs current local runner on Windows CPU; measure cold start, tok/s, peak RSS, binary+weights size, structured-output validity, tool-call JSON validity, 100-run crash rate and answer regression corpus. Require reliability parity before considering speed/size wins.
- Recurring cost target: EUR 0. Status: WATCH/BENCHMARK, NOT PASS.

### skills-mcp/skills-mcp — WATCH / portable-skill interoperability
- Progressive disclosure and portable skill format are useful, but executable skill content is explicitly a trust boundary. NEURA already tracks progressive loading; new delta is interoperability testing, not wholesale server import.
- Candidate idea: import a skill only through existing admission/signature/provenance gates; metadata may be discoverable before trust, but scripts/resources cannot execute until admitted. Status: WATCH.

## Candidate priority from scouting
1. Reliability/security: pin MCP tool-definition fingerprints and invalidate capability leases on schema/description/list drift.
2. Reliability: atomic mission claim + expiring lease + dependency/cycle gate for H24/subagent coordination.
3. Reliability: path-level edit claims for concurrent workers, fail closed when coordination cannot be verified.
4. Reliability: idempotency-key + intent/result journal; mutating tool calls are never blindly retried.
5. Reliability: golden trace behavioral-diff regression gate.
6. Reliability/privacy: read-only Windows UIA context envelope with untrusted-observed-content typing.
7. Safety: deterministic orchestration separation, bounded-history compaction and dirty-worktree read-only guard.
8. Safety: default-deny egress, MCP server-originated request denial, signed/redacted tool evidence.
9. Performance/packaging: benchmark pure-Go static local inference only after reliability gates are green.
10. Memory/skills: scoped memory and portable-skill interoperability after admission/provenance gates.

No item above is PASS merely because upstream claims it. All require NEURA-native tests/benchmarks in an isolated candidate before promotion.
