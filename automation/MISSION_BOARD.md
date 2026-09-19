# NEURA H24 — Mission Board

- Generated: `2026-09-19T09:20:03.108187Z`
- Valid until: `2026-09-19T10:20:03.108187Z`
- Head: `f16d38d971ba72644c74ba278d14897203f611e9`
- Commits reviewed: **1**
- **FOCUS DELL'ORA: `WINDOWS_COMPATIBILITY`**

## File modificati nell'ultima finestra
- `automation/MISSION_BOARD.md`
- `automation/mission_board.json`
- `automation/state/supervisor_state.json`

## Missioni fino al prossimo controllo
1. **BASE-1 · CRITICAL · AUDIT** — Inventario commit/diff dell'ora precedente e classificazione PASS/FAIL/UNKNOWN.
2. **BASE-2 · CRITICAL · SAFETY** — Controllare che nessun commit autonomo abbia toccato stable/main direttamente o allargato capability protette.
3. **BASE-3 · HIGH · VERIFY** — Verificare stato test/evidence; un risultato mancante non puo diventare PASS.
4. **WIN-1 · HIGH · REPORT_ONLY** — Controllare path, processi, ACL/DPAPI, build tags e assunzioni Windows introdotte dai diff recenti.
5. **WIN-2 · HIGH · VERIFY** — Eseguire cross-build Windows e, sul runner Windows, smoke/test nativi pertinenti.
6. **WIN-3 · HIGH · CODE_IF_EVIDENCE** — Correggere solo incompatibilita riproducibili; UNKNOWN Windows resta UNKNOWN finche non testato nativamente.

## Regola di stop
Se una missione produce FAIL di sicurezza/integrita/recovery, le successive missioni di CODE restano sospese fino al prossimo Supervisor.
