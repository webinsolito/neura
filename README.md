# NEURA

NEURA is a local-first Windows AI agent project focused on **reliability first, then speed, capability, and zero recurring cost**.

## Current baseline
- C9.38 — External Warden & Capability Leases
- Stable is never modified automatically.
- Autonomous work happens only in isolated candidate runs.
- One macro-area = one candidate/run.
- Every promotion requires deterministic evidence, tests, anti-regression and rollback.

## H24 evolution model
NEURA uses an hourly supervisor cycle:

```
AUDIT PREVIOUS HOUR
→ VERIFY EVIDENCE
→ CHOOSE ROTATING FOCUS
→ BUILD MISSION BOARD
→ EXECUTE MISSIONS ONE-BY-ONE
→ STOP NEW WORK BEFORE NEXT HOURLY REVIEW
→ REPEAT
```

The hourly supervisor does **not** blindly change code. It first checks what changed, what passed, what failed, and what must be investigated next.

### Rotating hourly focus
The controlled activity changes according to risk and recent modifications:
- CODE / TEST
- SECURITY / WARDEN / SANDBOX
- MEMORY / TEMPORAL CONSISTENCY
- REPLAY / RECOVERY / ROLLBACK
- PERFORMANCE / RESOURCE USE
- WINDOWS COMPATIBILITY
- DEPENDENCIES / LICENSE / SUPPLY CHAIN
- VOICE / UX / COMPUTER-USE

A minimum safety gate is always executed regardless of focus.

## Branch policy
- `main`: human-controlled reference branch.
- `stable`: protected known-good line.
- `candidate/*`: autonomous work only.
- No autonomous direct merge into `stable`.

## Cost
Recurring target: EUR 0. Local models and self-hosted execution are preferred.

> Status and PASS claims must be backed by real test evidence. Unknown is not PASS.
