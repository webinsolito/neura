# NEURA Zero-Cost Scout Ledger

Purpose: progressive technical scouting ledger. Do not repeatedly re-evaluate the same candidate unless a new release, issue, benchmark or NEURA requirement materially changes the comparison.

## 2026-09-19

### microsoft/BitNet — MATERIAL / MISSION CREATED
- Official, active project; MIT.
- New 2026 material: 270M and 0.6B 1-bit embedding models, plus VibeASR.cpp.
- Reusable concept for NEURA: benchmark-gated ternary embedding/runtime lane for low-resource local memory retrieval.
- Decision: PROPOSE + BENCHMARK; no default replacement before Windows-native A/B evidence.

### k2-fsa/sherpa-onnx — MATERIAL / MISSION CREATED
- Active; Apache-2.0; offline STT/TTS/VAD/KWS/speech enhancement with Windows and Go support.
- Reusable concept: unified offline voice pipeline behind Warden/capability boundaries.
- Decision: PROPOSE + BENCHMARK.

### augentic/omnia — MATERIAL / CANDIDATE ARCHITECTURE MISSION
- Apache-2.0; Rust/Wasmtime/WASI.
- Reusable concept: capability-scoped WASI sandbox for generated/portable skills.
- Decision: candidate spike only; reject if reliability gain does not justify packaging/runtime weight.

### viant/sqlite-vec — WATCH
- Apache-2.0, pure Go/CGO-free local vector retrieval with dataset isolation.
- Decision: benchmark only when memory retrieval becomes active focus.

### agentkitai/agentlens — CONCEPT ONLY
- Reusable concept: deterministic mutate/delete/reorder tamper tests against NEURA's own hash-chained recorder.
- Decision: copy verification idea, not Node stack.

### OpenHands/OpenHands — CONCEPT ONLY
- Reusable concepts: structured error outcomes, activity-log export, validated model profiles.
- Decision: no stack import.

### ggml-org/llama.cpp — WATCH / BASELINE
- Keep as hardware-adaptive runtime comparison lane; never unpinned auto-upgrade.

### openinterpreter/openinterpreter — CONCEPT ONLY
- Useful sandbox/local bridge patterns, but provider-heavy defaults conflict with EUR 0 default.

### block/goose-plugins — REJECTED AS SOURCE
- Archived; lesson is to prefer portable protocols over bespoke plugin systems.

### shane-farkas/memento-memory — MATERIAL / MEMORY ARCHITECTURE MISSION
- Local SQLite bitemporal knowledge graph: explicitly separates when a fact was true from when NEURA learned it, resolves entities and contradictions, and exposes MCP compatibility.
- Reusable concept: add `valid_time` + `transaction_time` semantics and contradiction/supersedence tests to NEURA's native memory layer; do NOT import the whole Python/MCP stack by default.
- Reliability upside: prevents today's knowledge from silently rewriting what the agent believed at an earlier time; directly useful for replay, audit and self-improvement attribution.
- Local/offline: database is local; external LLM use for extraction is optional/configurable. Recurring cost can remain EUR 0 with local extraction.
- Risk: upstream benchmark claims are not NEURA evidence; Python/package footprint is undesirable for the default single-file path.
- Decision: MATERIAL CONCEPT. Implement only as a native candidate schema/test spike after current source CI is healthy. Required tests: retroactive fact, correction, revocation, as-of query, replay determinism, project isolation, migration/rollback.

### rawcontext/engram — CONCEPT ONLY / REJECT STACK
- Strong bitemporal graph/event ideas, but operational stack requires Bun, Docker, Python, FalkorDB, Qdrant, NATS and PostgreSQL.
- Reusable concept overlaps Memento's bitemporal semantics; far too heavy for NEURA's zero-cost single-machine/single-file direction.
- Decision: study semantics only; do not integrate stack.

### SecAI-Hub/ai-incident-recorder — MATERIAL / FORENSIC HARDENING MISSION
- Go implementation of a local tamper-evident recorder with fail-closed loading, fsync-before-acknowledgement, corruption/truncation/duplicate detection, atomic retention rewrite and independently retained chain-head checkpoints.
- Reusable concept: strengthen NEURA's existing Warden ledger with durability semantics, not another service: fsync-before-success for consequential events; poison/fail-closed after append/sync failure; externally retained checkpoint before retention/rewrite; deterministic corruption/truncation tests.
- Reliability upside: NEURA already hash-chains audit events, but a hash chain alone does not prove durable acknowledgement or protect against whole-tail truncation without an external head/checkpoint.
- Decision: MATERIAL CONCEPT. Candidate-only native Go anti-regression mission; no stable modification. Benchmark write latency and crash/recovery behavior before acceptance.

### Data-System-School/agent-skills — CONCEPT ONLY / SKILL LOADING
- MIT; portable Markdown skills.
- Reusable concept: progressive disclosure — keep trigger metadata resident, load the selected SKILL body only when needed, and load references/resources on demand.
- Benefit: reduces resident prompt/context without adding executable dependencies.
- Decision: low-risk architecture proposal; integrate only after measuring context/token reduction and ensuring skill hashes/admission remain enforced.

### ellmos-ai/skills — WATCH / PORTABLE SKILL VALIDATION
- MIT portable SKILL.md library with static validation gates and documented runtime/governance invariants.
- Reusable concept: static skill admission tests before a skill is eligible for execution.
- Decision: do not ingest third-party skills wholesale; use its validation pattern as a reference for NEURA's existing skill-admission layer.

### New Windows desktop-agent sweep — WATCH, NO CODE IMPORT
- ClawBridge, Airi, Monaw, SmartiAI, Agentic and WinBridgeAgent were reviewed as current Windows/local-agent references.
- Useful recurring patterns: Plan→Execute→Verify, UI Automation before raw coordinate clicks, visible permission profiles, bounded action counts/emergency stop, localhost-only capability services, portable/self-contained packaging.
- Most stacks add Python/Node/Electron/Playwright or cloud/BYOK assumptions; none currently justifies replacing NEURA's Go + Warden architecture.
- Decision: harvest patterns only; no rewrite and no new dependency this cycle.

## Next scout rule
Revisit entries only on a material upstream release, a new NEURA bottleneck, a security issue, or when a queued benchmark produces evidence.
