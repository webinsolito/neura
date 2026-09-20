# NEURA C9.52 — Recovery + Release Integrity

## Goal
Close the main promotion blocker left by C9.51: stale static recovery and weak parent-artifact verification.

## Implemented
- Release identity bumped to schema 2 and version 0.40.52-c9.52-recovery-release-integrity.
- Release parent is now a real local artifact: bootstrap/c9_51_context_budget_core.patch.gz.b64.
- Startup verification now checks the parent file exists and its SHA-256 matches.
- Absolute paths, volume-qualified paths and parent traversal are rejected fail-closed.
- C9.51 native Context Budget Core remains cumulative.
- Static recovery was rebuilt from the current candidate with 216 critical files.
- New C9.52 Windows amd64 core and Warden binaries were rebuilt.
- Main and stable were not modified.

## Bugs found by real testing
The first full suite after schema 2 failed five autonomous-evolution E2E tests. Their test fixture still generated schema-1 release identities with a fake parent hash. The fixture was upgraded to create a real local parent artifact and matching SHA-256. The suite was then rerun.

## Real local evidence
- Release identity + parent drift + traversal tests: PASS.
- Static recovery current-files ↔ manifest ↔ ZIP: PASS with no recovery skip.
- Full go test -count=1 ./...: PASS (~14.7 s).
- go vet ./...: PASS.
- Focused race on changed areas (release identity, evolution fixture, context governor, llama integration): PASS (~3.5 s).
- Warden race: PASS (~1.35 s).
- Full core race: NOT VERIFIED; the combined command exceeded the execution window before completion.
- Windows amd64 cross-build: PASS for core and Warden (PE32+ x86-64).
- Windows native execution: NOT TESTED.

## Recovery evidence
- Critical files: 216
- critical_manifest.json SHA-256: a29fdcd90ff45f09c2d3d147b9d72da67a2456de7091b11353b964de9a8fa920
- NEURA_CORE_RECOVERY.zip SHA-256: 0ef27f9af424d2a91b59fd6728e977d74148ad5cf354b2b9041a204328a06b31
- Recovery ZIP size: 17,316,492 bytes

## Windows binaries
- Core SHA-256: f581bf9e8e50db18722b129459b2b2bc15ce3042c6352b616617748c36993416
- Warden SHA-256: f73d2ff4e0fff9dfc67a53de9858ba6ae06c087cbea1feaeeb47acebb4285c64

## GitHub transport
- C9.52 source delta SHA-256: d6898f851664d90e02184bf005860b94ce3fb8d4d9d359593def1691c516409d
- Compressed delta SHA-256: 08e1e1a4fee85d381d9fd26811e5dc7cf83f36e064c44a40d8cb2dad0d6d8ac0
- GitHub workflow remains workflow_dispatch only.
- GitHub CI has NOT been dispatched, therefore GitHub CI is NOT claimed PASS.
- Static package recovery is locally verified; the source-only GitHub CI does not recreate the entire 216-file shipped package and explicitly reports that limitation.

## Status
CANDIDATE. NOT PROMOTED. EUR 0 recurring cost. Main/stable untouched.
