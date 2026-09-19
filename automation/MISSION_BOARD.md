# NEURA H24 — Mission Board

- Generated: `2026-09-19T10:19:39.409672Z`
- Valid until: `2026-09-19T11:19:39.409672Z`
- Head: `1a5236e87bf5ec8661f836fecc6829bce944580d`
- Commits reviewed: **2**
- **FOCUS DELL'ORA: `PERFORMANCE_RESOURCES`**

## File modificati nell'ultima finestra
- `automation/MISSION_BOARD.md`
- `automation/SCOUT_LEDGER.md`
- `automation/mission_board.json`
- `automation/state/supervisor_state.json`

## Missioni fino al prossimo controllo
1. **BASE-1 · CRITICAL · AUDIT** — Inventario commit/diff dell'ora precedente e classificazione PASS/FAIL/UNKNOWN.
2. **BASE-2 · CRITICAL · SAFETY** — Controllare che nessun commit autonomo abbia toccato stable/main direttamente o allargato capability protette.
3. **BASE-3 · HIGH · VERIFY** — Verificare stato test/evidence; un risultato mancante non puo diventare PASS.
4. **PERF-1 · HIGH · REPORT_ONLY** — Misurare solo metriche pertinenti alle modifiche: tempo, RAM, allocazioni, dimensione binario o startup.
5. **PERF-2 · HIGH · VERIFY** — Confrontare candidate vs baseline usando lo stesso benchmark; nessun punteggio inventato.
6. **PERF-3 · HIGH · CODE_IF_EVIDENCE** — Ottimizzare soltanto regressioni misurate e senza ridurre affidabilita.

## Regola di stop
Se una missione produce FAIL di sicurezza/integrita/recovery, le successive missioni di CODE restano sospese fino al prossimo Supervisor.
