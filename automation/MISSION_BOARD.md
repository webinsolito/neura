# NEURA H24 — Mission Board

- Generated: `2026-09-19T08:21:08.185518Z`
- Valid until: `2026-09-19T09:21:08.185518Z`
- Head: `7add889051e6a18e0d0595249cb4f968a8ad794c`
- Commits reviewed: **3**
- **FOCUS DELL'ORA: `DEPENDENCIES_LICENSE_SUPPLY_CHAIN`**

## File modificati nell'ultima finestra
- `automation/MISSION_BOARD.md`
- `automation/SCOUT_LEDGER.md`
- `automation/mission_board.json`
- `automation/state/supervisor_state.json`

## Missioni fino al prossimo controllo
1. **BASE-1 · CRITICAL · AUDIT** — Inventario commit/diff dell'ora precedente e classificazione PASS/FAIL/UNKNOWN.
2. **BASE-2 · CRITICAL · SAFETY** — Controllare che nessun commit autonomo abbia toccato stable/main direttamente o allargato capability protette.
3. **BASE-3 · HIGH · VERIFY** — Verificare stato test/evidence; un risultato mancante non puo diventare PASS.
4. **DEP-1 · HIGH · REPORT_ONLY** — Controllare nuove dipendenze, licenze, provenienza, version pinning e superficie supply-chain.
5. **DEP-2 · HIGH · VERIFY** — Verificare che ogni dipendenza nuova sia necessaria, offline-compatible e coerente con EUR 0.
6. **DEP-3 · HIGH · CODE_IF_EVIDENCE** — Rimuovere o sostituire soltanto dipendenze con rischio dimostrato; nessun upgrade cosmetico.

## Regola di stop
Se una missione produce FAIL di sicurezza/integrita/recovery, le successive missioni di CODE restano sospese fino al prossimo Supervisor.
