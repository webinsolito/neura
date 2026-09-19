# NEURA H24 — Mission Board

- Generated: `2026-09-19T11:18:46.467936Z`
- Valid until: `2026-09-19T12:18:46.467936Z`
- Head: `82dec132932a4f36fd6099396485f9c6d67fe359`
- Commits reviewed: **2**
- **FOCUS DELL'ORA: `REPLAY_RECOVERY_ROLLBACK`**

## File modificati nell'ultima finestra
- `automation/MISSION_BOARD.md`
- `automation/SCOUT_LEDGER.md`
- `automation/mission_board.json`
- `automation/state/supervisor_state.json`

## Missioni fino al prossimo controllo
1. **BASE-1 · CRITICAL · AUDIT** — Inventario commit/diff dell'ora precedente e classificazione PASS/FAIL/UNKNOWN.
2. **BASE-2 · CRITICAL · SAFETY** — Controllare che nessun commit autonomo abbia toccato stable/main direttamente o allargato capability protette.
3. **BASE-3 · HIGH · VERIFY** — Verificare stato test/evidence; un risultato mancante non puo diventare PASS.
4. **RR-1 · HIGH · REPORT_ONLY** — Verificare replay, recovery, rollback e identita degli artefatti cambiati nell'ultima ora.
5. **RR-2 · HIGH · VERIFY** — Confrontare golden/replay evidence e cercare la prima divergenza rispetto alla baseline valida.
6. **RR-3 · HIGH · CODE_IF_EVIDENCE** — Se recovery o replay non sono riproducibili, fermare nuove missioni di codice e creare una candidate di ripristino.

## Regola di stop
Se una missione produce FAIL di sicurezza/integrita/recovery, le successive missioni di CODE restano sospese fino al prossimo Supervisor.
