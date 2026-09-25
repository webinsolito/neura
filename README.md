# NEURA

NEURA is a local-first Windows AI agent focused on **reliability first, then speed, capability, and zero recurring cost**.

## Current verified baseline

The active integration line is `main`; `stable` remains a separate known-good release line. The exact current verified evidence is recorded in `DIRECTOR_STATE.json`; do not rely on a stale SHA copied into documentation. The required GitHub Actions evidence includes:

- **NEURA V1 RC2 Pull Request Validation** — Linux and Windows native tests/builds, `go vet`, race detector, retrieval benchmark smoke, startup/health/memory smoke and Windows observe smoke.
- **NEURA Promotion Gate** — exact-SHA promotion validation.
- **NEURA Windows Package** — builds the Windows package, starts it, checks health, and uploads the `neura-windows-amd64` artifact.

GitHub branch protection is **not currently enabled** for `stable`; stable protection is therefore a release-process rule rather than a repository-enforced setting.

## Windows quick start

### Option A — verified CI package

1. Open the successful **NEURA Windows Package** workflow for the target commit.
2. Download the `NEURA-Windows-amd64` artifact.
3. Extract the complete `NEURA-Windows` folder.
4. Double-click **`Launch-NEURA.cmd`**.
5. The launcher checks the executable, writable data/workspace folders, the local port and the health endpoint; when ready it opens the local NEURA interface.

The published package does **not** require Go or a paid cloud/API dependency at runtime. Data and logs stay under the local Windows profile by default. If startup fails, the launcher shows an actionable message and records details under `%LOCALAPPDATA%\NEURA\logs`.

The package workflow verifies package contents and SHA256, a successful preflight, the missing-executable failure UX, a real launcher startup/health check and clean shutdown before publishing the artifact. A package from another SHA is not evidence for the current SHA.

### Option B — build locally

Requirements: Windows PowerShell and Go as required by the repository build.

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\build-windows.ps1
powershell -ExecutionPolicy Bypass -File .\scripts\run-windows.ps1
```

`build-windows.ps1` is the reproducible developer build entry point and creates `dist\NEURA-Windows` with the executable, double-click launcher, preflight launcher, instructions and SHA256 manifest. `run-windows.ps1` builds only when needed, then runs through the same packaged launcher path used by users.

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
