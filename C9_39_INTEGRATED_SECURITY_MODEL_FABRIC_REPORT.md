# NEURA C9.39 — Integrated Security + Local Model Fabric

## Goal
Merge the two real C9.38 candidates without weakening either boundary:
- External Warden + bounded single-use capability leases.
- Local Model Fabric 2026 + task-aware loopback-only multi-runtime routing.

Stable was not modified.

## Implemented
- External Warden process, pinned fail-closed policy and hash-chained ledger integrated into the Model Fabric baseline.
- Shipped config requires the External Warden.
- Model Fabric remains task-aware and loopback-only for llama.cpp, Bonsai 2, Qwen3.6, gpt-oss-20b and BitNet compatible endpoints.
- Local-AI authorization is now bound to the exact frozen endpoint selected by the router before a real call.
- Replay preflight prevents model probing on cassette hits; STRICT_REPLAY remains offline.
- Status requires both Model Fabric readiness and Warden readiness when Warden is required.
- Recovery bundle rebuilt with 212 critical files from both C9.38 lines.
- Windows amd64 core and Warden binaries cross-built.

## Real tests in this workspace
- Focused Model Fabric + Warden tests: PASS.
- C9.39 integration tests with a real external Warden process: PASS.
- Full `go test -count=1 ./...`: PASS (core ~18 s).
- `go vet ./...`: PASS.
- Core `go test -race -count=1 .`: PASS (~22.7 s).
- Warden `go test -race -count=1 ./warden`: PASS (~1.0 s).
- Critical integration/recovery tests repeated 3x: PASS.
- Linux production smoke: PASS for startup + External Warden readiness. Backend/model were intentionally absent, so overall runtime readiness was expected false.
- Windows amd64 cross-build: PASS for core and Warden (PE32+ x86-64).
- Windows native execution: NOT TESTED.
- Full suite `-count=2 ./...`: NOT VERIFIED; exceeded execution window. This was also a known limitation of the C9.38 Model Fabric line. The isolated test where the second pass appeared to pause (`TestIncidentRegressionCorpusTamperFailsClosed`) passes twice when run alone, so no specific regression is proven yet.

## Benchmarks (Linux test host)
- External Warden lease+authorize: ~333,457 ns/op, 1,397 B/op, 24 allocs/op (300 iterations).
- Model Fabric task scoring: ~5.27 ns/op, 0 allocs/op.
- Loopback endpoint admission: ~300 ns/op, 144 B/op, 1 alloc/op.

## Security regression fixed during merge
The original Model Fabric C9.38 authorized `cfg.LlamaURL` before dynamically selecting an alternate local provider. C9.39 freezes the real selected route first and authorizes that exact endpoint before live execution. Replay hits stay offline and authorize only the logical local capability.

## Status
CANDIDATE. NOT PROMOTED. Stable untouched. Recurring cost target: EUR 0.
