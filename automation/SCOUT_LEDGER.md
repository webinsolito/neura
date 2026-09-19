# NEURA Scout Ledger

Progressive technical scouting ledger. Re-check entries only for material new releases, evidence, security findings, or benchmark results.

## 2026-09-19 — MCP process-boundary isolation expansion

### modelcontextprotocol/modelcontextprotocol + mozilla/firefox-devtools-mcp — MATERIAL SECURITY CONTRACT
- Official MCP security guidance makes an important boundary explicit: stdio MCP servers execute as local subprocesses with environment-level privileges equivalent to the client; the protocol/SDK is not an OS sandbox. Therefore Warden tool allow/deny, schema pinning and capability leases alone cannot contain a malicious or compromised local MCP process.
- Mozilla independently recommends dedicated browser profiles and, for untrusted/privileged automation, a sandbox/container/VM/dedicated OS user plus outbound network restrictions. Firefox DevTools MCP is MIT/Apache-2.0 dual licensed, maintained by Mozilla, local, and Windows-capable via Firefox/Node; use it as a security reference, not a runtime dependency.
- MATERIAL NEURA delta: introduce a distinct MCP PROCESS LEASE below the existing tool lease. Every third-party/local MCP server must start under a declared process profile: executable hash/publisher provenance, dedicated working/data directory, environment allow-list, filesystem roots, resource limits, and network policy. Tool authorization must never imply host-process authorization.
- Windows candidate design: separate low-privilege process boundary using restricted token / dedicated OS identity where feasible, Job Object kill-on-close + memory/process limits, workspace ACL isolation, dedicated browser profile for browser automation, and default-deny outbound networking when an enforceable backend is available. If a requested confinement primitive is unavailable, classify it UNKNOWN/UNCONFINED and require a narrower/manual policy rather than silently running with full user privilege.
- Candidate negative tests: malicious MCP tries reading SSH/browser/user-profile secrets outside its roots => denied; child-process escape => contained/terminated; memory/process bomb => Job Object limit; parent crash => children killed; browser automation cannot see normal-user cookies/profile; localhost HTTP MCP requires authentication/binding protection; network exfiltration blocked when network-deny profile is selected; unsupported network confinement cannot be reported PASS.
- Reliability benchmark: startup latency, peak RAM and tool-call overhead vs current stdio launch; no promotion if containment materially destabilizes normal local tools.
- Cost target EUR 0. Status: PROPOSED CANDIDATE, NOT PASS.

### MCP 2026-07-28 stateless request model — MATERIAL INTEROP CONTRACT
- Current MCP specification is stateless: each request carries protocol version/client capabilities and cross-request state uses explicit identifiers. NEURA should not bind security state to a transport connection or assume reconnect == same authorization context.
- NEURA delta: capability/tool/process lease IDs must be explicit, replay-resistant handles bound to tool-definition fingerprint + process identity + task/mission, with expiry checked on every request. Reconnect must not resurrect expired privileges.
- Candidate tests: disconnect/reconnect, server restart, reordered requests and stale-handle replay cannot extend or recover authority; unsupported protocol version fails closed. Status: DESIGN INPUT, NOT PASS.

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
- MIT, Go, young project. Strong reusable delta: fingerprint each MCP tool definition (name + description + input schema) at admission/session start. Definition drift invalidates tool/capability lease and requires re-admission.
- Additional useful contracts: strip denied tools from discovery; deny server-originated sampling/elicitation by default; per-session call budgets; result-size limits; append-only replay.
- Do not import whole gateway; reimplement narrow fingerprint/session invalidation contract natively.
- Cost target EUR 0. Status: PROPOSED, NOT PASS.

### P4ST4S/mcp-audit — MATERIAL REFERENCE / signed MCP evidence boundary
- Apache-2.0, Go. Reusable delta: synchronous allow/deny before tools/call, sensitive-field redaction before persistence, optional HMAC evidence, and no automatic retry of tools/call.
- NEURA extraction: bind evidence to Warden lease ID + pinned tool-definition hash + raw-argument digest; never auto-retry a mutating tool unless idempotency contract explicitly permits it.
- Status: DESIGN INPUT, NOT PASS.

### townsendmerino/goinfer — WATCH/BENCHMARK / pure-Go static local inference
- MIT; pure Go/no-CGO; supports GGUF/safetensors and advertises model embedding into a static executable. Young/small adoption, so upstream performance claims are not NEURA evidence.
- Candidate benchmark: same small checkpoint vs current local runner on Windows CPU; cold start, tok/s, peak RSS, size, structured-output/tool-call validity, 100-run crash rate and regression corpus.
- Cost target EUR 0. Status: WATCH/BENCHMARK, NOT PASS.

### skills-mcp/skills-mcp — WATCH / portable-skill interoperability
- Progressive disclosure and portable skill format useful, but executable skill content is a trust boundary. Import only through admission/signature/provenance gates.

## 2026-09-19 — cross-tool information-flow + voice-loop reliability expansion

### piyushptiwari1/mcpkernel — MATERIAL CONCEPT / taint provenance across tool boundaries
- Apache-2.0, Python. Useful research/reference but not a wholesale dependency.
- Strong reusable delta: taint/provenance labels (secret, PII, untrusted-user, untrusted-tool-result) propagate across tool calls; outbound sinks can deny derived sensitive data. Also useful: scan tool results before agent context and lint policy sets.
- Candidate acceptance: transformed secrets remain tainted; forbidden source->sink blocked; narrow explicit declassification only; malicious tool-result instructions remain untrusted; policy shadowing/unreachable deny fails lint; replay preserves provenance.
- Cost target EUR 0. Status: PROPOSED, NOT PASS.

### JenkinsRobotics/VoiceLLM — MATERIAL VOICE CONCEPT / self-speech rejection + two-pass STT
- Fully offline voice-loop reference with two-pass fast->accurate STT, mic pause during TTS, similarity rejection against last spoken reply, streaming TTS and interruptibility. Apple Silicon upstream, so design reference only.
- Candidate Windows test: play NEURA TTS into microphone at several volumes/noise levels and require zero privileged tool dispatches; verify barge-in/cancel and second-pass correction.
- Status: PROPOSED TEST CONTRACT, NOT PASS.

### SpandanNagale/ARIA and recent Windows Jarvis projects — WATCH / voice UX references
- Reusable concept only: wake word must have deterministic push-to-talk fallback, visible listening/thinking/speaking state, and false-trigger benchmark. Do not copy heavy runtime stack.

### razzant/ouroboros — WATCH / reviewed self-evolution reference
- Useful architectural confirmation: self-evolution emits patch/evidence bundle and requires a distinct reviewer role before integration; author should not self-approve core changes.
- Status: WATCH / INDEPENDENT-REVIEW CONTRACT, NOT PASS.

## Candidate priority from scouting
1. Reliability/security: OS-level MCP process lease/isolation; never treat stdio/tool authorization as a sandbox.
2. Reliability/security: propagate taint/provenance across tool boundaries and block forbidden source->sink flows; lint Warden policy.
3. Reliability/security: pin MCP tool-definition fingerprints and invalidate capability leases on schema/description/list drift.
4. Reliability: atomic mission claim + expiring lease + dependency/cycle gate for H24/subagent coordination.
5. Reliability: path-level edit claims for concurrent workers, fail closed when coordination cannot be verified.
6. Reliability: idempotency-key + intent/result journal; mutating tool calls are never blindly retried.
7. Reliability: golden trace behavioral-diff regression gate and independent reviewer for self-evolution patches.
8. Voice reliability: self-speech rejection + two-pass STT + push-to-talk fallback before autonomous voice tool execution.
9. Reliability/privacy: read-only Windows UIA context envelope with untrusted-observed-content typing.
10. Safety: deterministic orchestration separation, bounded-history compaction and dirty-worktree read-only guard.
11. Safety: default-deny egress, signed/redacted tool evidence, explicit stateless MCP lease handles.
12. Performance/packaging: benchmark pure-Go static local inference only after reliability gates are green.
13. Memory/skills: scoped memory and portable-skill interoperability after admission/provenance gates.

No item above is PASS merely because upstream claims it. All require NEURA-native tests/benchmarks in an isolated candidate before promotion.
