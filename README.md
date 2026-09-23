# NEURA

NEURA is a local-first Windows AI agent project focused on **reliability first, then speed, capability, and zero recurring cost**.

## Current verified baseline

The active integration line is `main`. The long-lived `stable` branch is a separate known-good line and must not be advanced without promotion evidence.

At commit `9d59d1ca993ea5ca5c33d5571c7c5f88a118d2ac`, the NEURA Promotion Gate completed successfully on Linux and Windows. The gate runs unit/integration tests, `go vet`, the Linux race detector, verified-action anti-regression tests, and promotion-policy unit tests.

Important: GitHub branch protection is **not currently enabled** for `stable`. Treat stable protection as a release-process requirement, not as an enforced repository setting, until GitHub reports protection enabled.

## Release discipline

- `main`: current integration/reference branch.
- `stable`: known-good release line; do not write to it directly during autonomous work.
- `candidate/*`: isolated work for one macro-area at a time.
- `rollback/*`: recovery points retained before substantial transitions.
- Candidate work must not be described as verified until its own applicable tests have executed successfully.
- A successful test run on `main` does not automatically certify a later candidate commit.
- Promotion to `stable` requires exact-SHA evidence and regression checks.

## Promotion gate

`.github/workflows/promotion-gate.yml` validates the repository on Ubuntu and Windows. On Linux it also executes the race detector. Pull requests targeting `main` or `stable` run the transition policy after the test jobs succeed. Pushes to `main` and `stable` record exact-SHA known-good evidence after the required jobs pass.

## Controlled evolution model

NEURA uses isolated candidate cycles:

```text
AUDIT CURRENT STATE
→ SELECT ONE MACRO-AREA
→ CREATE/CONFIRM RECOVERY POINT
→ IMPLEMENT ON CANDIDATE
→ RUN REAL TESTS
→ CHECK REGRESSIONS
→ KEEP OR ROLLBACK
→ PROMOTE ONLY WITH EVIDENCE
```

The supervisor must not treat the passage of time, the existence of test files, or an unexecuted workflow as proof of correctness.

## Cost target

Recurring target: **EUR 0**. Local models, local storage, open-source components and self-hosted execution are preferred. Paid APIs must remain optional rather than a V1 requirement.

## V1 release rule

NEURA V1 is not considered verified until the essential agent loop, planning, tools, memory, recovery, safety, usable UI, packaging and documentation are integrated and the release candidate passes its required test and regression gates. Unknown is not PASS.
