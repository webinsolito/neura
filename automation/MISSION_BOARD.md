# NEURA H24 — Mission Board

- Generated: `2026-09-19T04:20:12.387264Z`
- Valid until: `2026-09-19T05:20:12.387264Z`
- Head: `b122c38181f28fee5d3ecd7480bdcf449b505224`
- Commits reviewed: **27**
- **FOCUS DELL'ORA: `SECURITY_WARDEN_SANDBOX`**

## File modificati nell'ultima finestra
- `.github/workflows/source-ci.yml`
- `.gitignore`
- `CANDIDATE_STATUS_C9_38.txt`
- `README.md`
- `RELEASE_IDENTITY.json`
- `automation/hourly_supervisor.py`
- `automation/mission_worker.ps1`
- `bootstrap/c9_38_source.tar.gz`
- `bootstrap/core.b64.part-00`
- `bootstrap/core.b64.part-01`
- `bootstrap/core.b64.part-02`
- `bootstrap/core.b64.part-03`
- `bootstrap/core.b64.part-04`
- `bootstrap/core.b64.part-05`
- `bootstrap/core.b64.part-06`
- `bootstrap/core.b64.part-07`
- `bootstrap/core.b64.part-08`
- `bootstrap/core.b64.part-09`
- `bootstrap/core.b64.part-10`
- `capo_local/CONSTITUTION.lock.json`
- `capo_local/capabilities.json`
- `research/BENCHMARK_MISSIONS_2026-09-19.md`
- `research/SCOUT_LEDGER.md`

## Missioni fino al prossimo controllo
1. **BASE-1 · CRITICAL · AUDIT** — Inventario commit/diff dell'ora precedente e classificazione PASS/FAIL/UNKNOWN.
2. **BASE-2 · CRITICAL · SAFETY** — Controllare che nessun commit autonomo abbia toccato stable/main direttamente o allargato capability protette.
3. **BASE-3 · HIGH · VERIFY** — Verificare stato test/evidence; un risultato mancante non puo diventare PASS.
4. **SEC-1 · HIGH · REPORT_ONLY** — Audit Warden, capability lease, policy hash e superfici di bypass introdotte dai diff recenti.
5. **SEC-2 · HIGH · VERIFY** — Eseguire test negativi fail-closed su permessi, tamper, lease scadute e process boundary pertinenti.
6. **SEC-3 · HIGH · CODE_IF_EVIDENCE** — Aprire una candidate solo se emerge un bypass dimostrabile; nessun allargamento capability automatico.

## Regola di stop
Se una missione produce FAIL di sicurezza/integrita/recovery, le successive missioni di CODE restano sospese fino al prossimo Supervisor.
