# NEURA

NEURA is a local-first Windows AI agent focused on **reliability first, then speed, capability, and zero recurring cost**.

## Current verified baseline

The active integration line is `main`; `stable` remains a separate known-good release line. At `b64c33f7f6a5aa07a3b769d1ae74c4bfad51497d` the following GitHub Actions runs completed successfully:

- **NEURA V1 RC2 Pull Request Validation** — Linux and Windows native tests/builds, `go vet`, race detector, retrieval benchmark smoke, startup/health/memory smoke and Windows observe smoke.
- **NEURA Promotion Gate** — exact-SHA promotion validation.
- **NEURA Windows Package** — builds the Windows package, starts it, checks health, and uploads the `neura-windows-amd64` artifact.

GitHub branch protection is **not currently enabled** for `stable`; stable protection is therefore a release-process rule rather than a repository-enforced setting.

## Windows quick start

### Option A — verified CI package

1. Open the successful **NEURA Windows Package** workflow for the target commit.
2. Download the `neura-windows-amd64` artifact.
3. Extract it to a normal user-writable folder.
4. Start NEURA with the packaged launcher/executable.
5. Confirm the local health endpoint responds before relying on the agent for work.

The package workflow performs a real startup + health smoke test before publishing the artifact. A package from another SHA is not evidence for the current SHA.

### Option B — build locally

Requirements: Windows PowerShell and Go as required by the repository build.

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\build-windows.ps1
powershell -ExecutionPolicy Bypass -File .\scripts\run-windows.ps1
```

`build-windows.ps1` is the reproducible build entry point. `run-windows.ps1` starts the packaged application and can fall back to building when the binary is absent.

## Verification before release

For a candidate to be called verified, check the **candidate's exact SHA**, not merely a previous green run. Required evidence includes the applicable test workflow(s), regression checks and packaging/startup checks. Do not promote on the basis of source inspection alone.

## Release discipline

- `main`: current integration/reference branch.
- `stable`: known-good release line; do not write to it directly during autonomous work.
- `candidate/*`: isolated work for one macro-area at a time.
- `rollback/*`: recovery points retained before substantial transitions.
- Candidate work is not verified until its own applicable tests execute successfully.
- A successful test run on `main` does not certify a later candidate commit.
- Promotion to `stable` requires exact-SHA evidence and regression checks.

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

The supervisor must not treat elapsed time, the existence of test files, or an unexecuted workflow as proof of correctness.

## Cost target

Recurring target: **EUR 0**. Local models, local storage, open-source components and self-hosted execution are preferred. Paid APIs must remain optional rather than a V1 requirement.

## Known release limitation

Repository-level branch protection for `stable` is not enforced by GitHub at this time. Until that changes, promotion discipline and rollback branches are mandatory safeguards.

## V1 release rule

NEURA V1 is not considered verified until the essential agent loop, planning, tools, memory, recovery, safety, usable UI, packaging and documentation are integrated and the release candidate passes its required test and regression gates. Unknown is not PASS.
