# C9.53 Memory Retrieval Fabric

Priority remains **reliability > speed > power > zero recurring cost**.

This candidate extracts only the useful architectural ideas: hybrid retrieval (lexical, semantic, entity and temporal signals) and an optional local embedding adapter. It does not import Mem0 or BitNet as a framework and introduces no Go module dependency.

## Safety properties
- Missing or failed embeddings do not disable memory retrieval.
- No network endpoint exists in the embedding adapter.
- Ranking is deterministic for equal scores.
- Stable is untouched and a pre-C9.53 rollback branch exists.
- CI remains manual via workflow_dispatch.

## Candidate test plan
Focused ranking/fallback tests, gofmt, go vet, full Go anti-regression, focused race test, and Windows amd64 cross-build are encoded in the manual workflow. These checks must not be reported as PASS until a real run completes.

## Deferred
Real BitNet embedding quality/latency/RAM benchmark and native Windows execution remain NOT TESTED. Capability/profile isolation is deliberately deferred to a separate C9.54 cycle so one macro-area is changed per candidate.
