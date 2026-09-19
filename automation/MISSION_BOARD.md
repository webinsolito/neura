# NEURA H24 — Mission Board

- Generated: `2026-09-19T06:24:25.124058Z`
- Valid until: `2026-09-19T07:24:25.124058Z`
- Head: `60f9022c2e404dd2286df97da4d67e1d9b2b3855`
- Commits reviewed: **1**
- **FOCUS DELL'ORA: `CODE_TEST`**

## File modificati nell'ultima finestra
- `automation/MISSION_BOARD.md`
- `automation/mission_board.json`
- `automation/state/supervisor_state.json`

## Missioni fino al prossimo controllo
1. **BASE-1 · CRITICAL · AUDIT** — Inventario commit/diff dell'ora precedente e classificazione PASS/FAIL/UNKNOWN.
2. **BASE-2 · CRITICAL · SAFETY** — Controllare che nessun commit autonomo abbia toccato stable/main direttamente o allargato capability protette.
3. **BASE-3 · HIGH · VERIFY** — Verificare stato test/evidence; un risultato mancante non puo diventare PASS.
4. **CT-1 · HIGH · REPORT_ONLY** — Controllare diff e test toccati nell'ultima ora; cercare test deboli, regressioni o codice duplicato.
5. **CT-2 · HIGH · VERIFY** — Eseguire source gate mirato sulle aree modificate e classificare PASS/FAIL/UNKNOWN senza reinterpretazioni.
6. **CT-3 · HIGH · CODE_IF_EVIDENCE** — Se esiste un FAIL riproducibile, preparare una sola candidate di correzione con rollback; altrimenti nessuna modifica.

## Regola di stop
Se una missione produce FAIL di sicurezza/integrita/recovery, le successive missioni di CODE restano sospese fino al prossimo Supervisor.
