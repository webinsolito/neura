# NEURA H24 — Mission Board

- Generated: `2026-09-19T07:19:00.768809Z`
- Valid until: `2026-09-19T08:19:00.768809Z`
- Head: `df83d103e4d57ceeaf6908f6c6d4eeed442fcef1`
- Commits reviewed: **2**
- **FOCUS DELL'ORA: `VOICE_UX_COMPUTER_USE`**

## File modificati nell'ultima finestra
- `automation/MISSION_BOARD.md`
- `automation/SCOUT_LEDGER.md`
- `automation/mission_board.json`
- `automation/state/supervisor_state.json`

## Missioni fino al prossimo controllo
1. **BASE-1 · CRITICAL · AUDIT** — Inventario commit/diff dell'ora precedente e classificazione PASS/FAIL/UNKNOWN.
2. **BASE-2 · CRITICAL · SAFETY** — Controllare che nessun commit autonomo abbia toccato stable/main direttamente o allargato capability protette.
3. **BASE-3 · HIGH · VERIFY** — Verificare stato test/evidence; un risultato mancante non puo diventare PASS.
4. **UX-1 · HIGH · REPORT_ONLY** — Controllare latenza, cancellazione, accessibilita e sicurezza delle interazioni/UI/computer-use toccate.
5. **UX-2 · HIGH · VERIFY** — Verificare che UX e computer-use non aggirino Warden, audit o post-action verification.
6. **UX-3 · HIGH · CODE_IF_EVIDENCE** — Modificare UX/voice soltanto con criterio misurabile e senza compromettere safety/performance.

## Regola di stop
Se una missione produce FAIL di sicurezza/integrita/recovery, le successive missioni di CODE restano sospese fino al prossimo Supervisor.
