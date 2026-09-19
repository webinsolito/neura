# NEURA Zero-Cost Scout Ledger

Purpose: progressive technical scouting ledger. Do not repeatedly re-evaluate the same candidate unless a new release, issue, benchmark or NEURA requirement materially changes the comparison.

## 2026-09-19

### microsoft/BitNet — MATERIAL / MISSION CREATED
- Official, active project; MIT.
- New 2026 material: 270M and 0.6B 1-bit embedding models, plus VibeASR.cpp.
- Reusable concept for NEURA: add a **benchmark-gated ternary embedding/runtime lane** for low-resource local memory retrieval. Do not replace the current runtime by default.
- Why: the 270M embedding model is small enough to be a realistic CPU/local-first candidate and the official project reports optimized x86 kernels.
- Windows: supported, but source setup currently needs VS2022/Clang/CMake/Python/Conda; therefore not yet suitable as NEURA's default single-file runtime.
- Recurring cost: EUR 0.
- Risk: added toolchain/model weight; vendor benchmarks are not NEURA benchmarks.
- Decision: PROPOSE + BENCHMARK. No production integration until Windows-native A/B evidence exists.

### k2-fsa/sherpa-onnx — MATERIAL / MISSION CREATED
- Active (updated 2026-09-17), Apache-2.0.
- Offline/local-first STT, TTS, VAD, keyword spotting, speech enhancement; Windows x64/arm64 and Go bindings exist.
- Reusable concept for NEURA: a unified offline voice pipeline with VAD + denoise + STT behind the existing Warden/capability boundary.
- Recurring cost: EUR 0.
- Risk: native/ONNX dependencies and model footprint; integration must not weaken cancellation, sandbox or capability leases.
- Decision: PROPOSE + BENCHMARK. No automatic voice-stack replacement.

### OpenHands/OpenHands — CONCEPT ONLY
- Active; v1.14.0 published 2026-08-17; core repository MIT (enterprise directory has separate terms).
- Reusable concepts: structured error outcomes, activity-log export, validated model profiles, Git-sync UX.
- NEURA already has stronger stable/candidate/replay/warden constraints, so importing the coding-agent stack would add complexity without proven reliability benefit.
- Decision: do not integrate code. Add structured error/activity-log UX to future mission backlog only.

### ggml-org/llama.cpp — WATCH / BENCHMARK BASELINE
- Very active and broad Windows CPU/CUDA/Vulkan/HIP support; Windows release matrix is real.
- Reusable concept: hardware-adaptive backend selection and portable local serving.
- Risk observed: backend/version-specific Windows issues still occur; never auto-upgrade the runtime without pinned regression tests.
- Decision: keep as a comparison/runtime lane, not an unpinned dependency upgrade.

### openinterpreter/openinterpreter — CONCEPT ONLY
- Active release line (0.0.44 published 2026-09-15).
- Interesting concept: sandboxed/local bridge patterns and broad provider catalog.
- Current releases emphasize provider integrations that can require paid/API services; this conflicts with NEURA's EUR 0 default.
- Decision: no integration this cycle.

### block/goose-plugins — REJECTED AS SOURCE
- Archived 2026-01-26; project explicitly moved from its old plugin system to MCP.
- Reusable lesson: avoid bespoke plugin protocols when a portable tool protocol suffices.
- Decision: do not reuse archived plugin code.

## Next scout rule
Revisit entries only on a material upstream release, a new NEURA bottleneck, a security issue, or when a queued benchmark produces evidence.
