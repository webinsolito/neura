# NEURA H24 — Mission Board

- Generated: `2026-09-19T05:18:57.163225Z`
- Valid until: `2026-09-19T06:18:57.163225Z`
- Head: `484a74f419dd1bc29fcc7a881c942edfeba08234`
- Commits reviewed: **3**
- **FOCUS DELL'ORA: `MEMORY_TEMPORAL`**

## File modificati nell'ultima finestra
- `automation/MISSION_BOARD.md`
- `automation/mission_board.json`
- `automation/state/supervisor_state.json`
- `research/SCOUT_LEDGER.md`

## Missioni fino al prossimo controllo
1. **BASE-1 · CRITICAL · AUDIT** — Inventario commit/diff dell'ora precedente e classificazione PASS/FAIL/UNKNOWN.
2. **BASE-2 · CRITICAL · SAFETY** — Controllare che nessun commit autonomo abbia toccato stable/main direttamente o allargato capability protette.
3. **BASE-3 · HIGH · VERIFY** — Verificare stato test/evidence; un risultato mancante non puo diventare PASS.
4. **MEM-1 · HIGH · REPORT_ONLY** — Controllare provenance, supersedence, contaminazione cross-project e integrita memoria dopo i cambiamenti recenti.
5. **MEM-2 · HIGH · VERIFY** — Eseguire regression memory e verificare che memorie obsolete/revocate non guidino nuove decisioni.
6. **MEM-3 · HIGH · CODE_IF_EVIDENCE** — Creare candidate solo per un difetto riproducibile di memoria; evitare nuove feature non richieste.

## Regola di stop
Se una missione produce FAIL di sicurezza/integrita/recovery, le successive missioni di CODE restano sospese fino al prossimo Supervisor.
