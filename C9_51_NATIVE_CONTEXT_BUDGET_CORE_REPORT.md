# NEURA C9.51 — Native Context Budget Core

## Goal
Turn the C9.50 Python prototype into a real Go core component on top of the cumulative C9.39 security + model-fabric baseline.

## Implemented
- Shared deterministic context governor for llama.cpp and alternate local OpenAI-compatible model routes.
- Prompt capacity derives from context window minus completion budget and a protected reserve.
- Mandatory truth (system + current user + backend result) is never silently truncated. If it cannot fit, inference fails closed.
- Optional context is bounded deterministically.
- Verified memory and skills are retained ahead of bulk runtime diagnostics.
- Conversation history is admitted newest-first as complete turns; older turns are dropped first.
- Runtime diagnostics use UTF-8-safe head/tail trimming.
- /no_think safety suffix is included in the budget so it cannot create an off-by-one overflow.
- Non-KEEP decisions and hard stops are recorded in the existing tamper-evident Flight Recorder.
- No new runtime dependency and no network/cloud requirement were added.

## Bugs found while integrating
1. C9.50 originally diverged from C9.39 instead of extending it. C9.50B corrected the lineage.
2. The first native ordering let an oversized runtime block consume all optional capacity and starve the newest conversation turn. Priority was corrected.
3. The first llama.cpp integration appended /no_think after budgeting, allowing a small post-budget overflow. The suffix is now reserved inside the budget.

## Real local test evidence
- Focused Context Governor + llama.cpp integration tests: PASS.
- Full go test -count=1 ./... with only TestStaticRecoveryBundleMatchesCurrentRelease skipped: PASS (~13.6 s core).
- go vet ./...: PASS.
- Core go test -race with only the stale recovery-bundle test skipped: PASS (~15.6 s).
- Warden go test -race: PASS (~1.0 s).
- Windows amd64 cross-build: PASS for core and Warden; both produced PE32+ x86-64 executables.
- Windows native execution: NOT TESTED.

## Benchmark evidence on Linux test host
Three runs:
- Small context: 1.043 us/op, 1.023 us/op, 0.995 us/op; 1,768 B/op; 12 allocs/op.
- Pressure context: 38.353 us/op, 52.850 us/op, 36.204 us/op; ~231 KB/op; 18 allocs/op.

The pressure case intentionally builds very large synthetic runtime/history strings; no claim is made that these timings represent end-to-end LLM latency.

## Remaining promotion blocker
The static recovery bundle is correctly stale because critical Go source files changed. It has NOT been rebuilt for C9.51. Therefore C9.51 is not promotion-ready.

The manual GitHub source CI was updated to:
- verify C9.38 archive and C9.39 delta,
- verify/decode/apply the C9.51 delta,
- run formatting, vet, focused tests, full anti-regression, race checks, benchmark smoke, and Windows amd64 cross-build.

That GitHub workflow has NOT been dispatched, so GitHub CI is NOT claimed PASS.

## Safety
Stable and main remain untouched. Recurring cost remains EUR 0.
