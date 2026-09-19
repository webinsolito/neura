#!/usr/bin/env python3
from __future__ import annotations
import datetime as dt
import fnmatch
import json
import os
import pathlib
import subprocess
from typing import Dict, List

ROOT = pathlib.Path(__file__).resolve().parents[1]
AUTO = ROOT / "automation"
STATE_DIR = AUTO / "state"
STATE_DIR.mkdir(parents=True, exist_ok=True)
STATE_PATH = STATE_DIR / "supervisor_state.json"
BOARD_JSON = AUTO / "mission_board.json"
BOARD_MD = AUTO / "MISSION_BOARD.md"

FOCI = [
    "CODE_TEST",
    "SECURITY_WARDEN_SANDBOX",
    "MEMORY_TEMPORAL",
    "REPLAY_RECOVERY_ROLLBACK",
    "PERFORMANCE_RESOURCES",
    "WINDOWS_COMPATIBILITY",
    "DEPENDENCIES_LICENSE_SUPPLY_CHAIN",
    "VOICE_UX_COMPUTER_USE",
]

PATTERNS: Dict[str, List[str]] = {
    "CODE_TEST": ["capo_local/src/*.go", "capo_local/src/**/*_test.go", "regression_corpus/*"],
    "SECURITY_WARDEN_SANDBOX": ["capo_local/src/security.go", "capo_local/src/trust.go", "capo_local/src/external_warden*", "capo_local/src/warden/*", "capo_local/src/skill_admission*", "capo_local/warden_policy*", ".github/*"],
    "MEMORY_TEMPORAL": ["capo_local/src/memory*", "capo_local/src/trajectory*", "capo_local/src/skills*"],
    "REPLAY_RECOVERY_ROLLBACK": ["capo_local/src/replay*", "capo_local/src/golden_replay*", "capo_local/src/recovery*", "capo_local/src/release_identity*", "capo_local/src/incident_regression*", "runtime/recovery/*"],
    "PERFORMANCE_RESOURCES": ["capo_local/src/*benchmark*", "capo_local/src/runtime_supervisor*", "capo_local/src/model_fabric*", "capo_local/src/router*"],
    "WINDOWS_COMPATIBILITY": ["**/*windows*", "*.ps1", "*.cmd", "capo_local/src/hardware_windows.go", "capo_local/src/secret_store_windows*"],
    "DEPENDENCIES_LICENSE_SUPPLY_CHAIN": ["**/go.mod", "**/go.sum", "LICENSE*", "NOTICE*", ".github/dependabot.yml", "capo_local/src/tool_definition_pin*", "capo_local/src/skill_admission*"],
    "VOICE_UX_COMPUTER_USE": ["**/*voice*", "**/*audio*", "**/*ui*", "**/*.html", "**/*computer*", "**/*accessibility*"],
}

FOCUS_MISSIONS = {
    "CODE_TEST": [
        ("CT-1", "Controllare diff e test toccati nell'ultima ora; cercare test deboli, regressioni o codice duplicato.", "REPORT_ONLY"),
        ("CT-2", "Eseguire source gate mirato sulle aree modificate e classificare PASS/FAIL/UNKNOWN senza reinterpretazioni.", "VERIFY"),
        ("CT-3", "Se esiste un FAIL riproducibile, preparare una sola candidate di correzione con rollback; altrimenti nessuna modifica.", "CODE_IF_EVIDENCE"),
    ],
    "SECURITY_WARDEN_SANDBOX": [
        ("SEC-1", "Audit Warden, capability lease, policy hash e superfici di bypass introdotte dai diff recenti.", "REPORT_ONLY"),
        ("SEC-2", "Eseguire test negativi fail-closed su permessi, tamper, lease scadute e process boundary pertinenti.", "VERIFY"),
        ("SEC-3", "Aprire una candidate solo se emerge un bypass dimostrabile; nessun allargamento capability automatico.", "CODE_IF_EVIDENCE"),
    ],
    "MEMORY_TEMPORAL": [
        ("MEM-1", "Controllare provenance, supersedence, contaminazione cross-project e integrita memoria dopo i cambiamenti recenti.", "REPORT_ONLY"),
        ("MEM-2", "Eseguire regression memory e verificare che memorie obsolete/revocate non guidino nuove decisioni.", "VERIFY"),
        ("MEM-3", "Creare candidate solo per un difetto riproducibile di memoria; evitare nuove feature non richieste.", "CODE_IF_EVIDENCE"),
    ],
    "REPLAY_RECOVERY_ROLLBACK": [
        ("RR-1", "Verificare replay, recovery, rollback e identita degli artefatti cambiati nell'ultima ora.", "REPORT_ONLY"),
        ("RR-2", "Confrontare golden/replay evidence e cercare la prima divergenza rispetto alla baseline valida.", "VERIFY"),
        ("RR-3", "Se recovery o replay non sono riproducibili, fermare nuove missioni di codice e creare una candidate di ripristino.", "CODE_IF_EVIDENCE"),
    ],
    "PERFORMANCE_RESOURCES": [
        ("PERF-1", "Misurare solo metriche pertinenti alle modifiche: tempo, RAM, allocazioni, dimensione binario o startup.", "REPORT_ONLY"),
        ("PERF-2", "Confrontare candidate vs baseline usando lo stesso benchmark; nessun punteggio inventato.", "VERIFY"),
        ("PERF-3", "Ottimizzare soltanto regressioni misurate e senza ridurre affidabilita.", "CODE_IF_EVIDENCE"),
    ],
    "WINDOWS_COMPATIBILITY": [
        ("WIN-1", "Controllare path, processi, ACL/DPAPI, build tags e assunzioni Windows introdotte dai diff recenti.", "REPORT_ONLY"),
        ("WIN-2", "Eseguire cross-build Windows e, sul runner Windows, smoke/test nativi pertinenti.", "VERIFY"),
        ("WIN-3", "Correggere solo incompatibilita riproducibili; UNKNOWN Windows resta UNKNOWN finche non testato nativamente.", "CODE_IF_EVIDENCE"),
    ],
    "DEPENDENCIES_LICENSE_SUPPLY_CHAIN": [
        ("DEP-1", "Controllare nuove dipendenze, licenze, provenienza, version pinning e superficie supply-chain.", "REPORT_ONLY"),
        ("DEP-2", "Verificare che ogni dipendenza nuova sia necessaria, offline-compatible e coerente con EUR 0.", "VERIFY"),
        ("DEP-3", "Rimuovere o sostituire soltanto dipendenze con rischio dimostrato; nessun upgrade cosmetico.", "CODE_IF_EVIDENCE"),
    ],
    "VOICE_UX_COMPUTER_USE": [
        ("UX-1", "Controllare latenza, cancellazione, accessibilita e sicurezza delle interazioni/UI/computer-use toccate.", "REPORT_ONLY"),
        ("UX-2", "Verificare che UX e computer-use non aggirino Warden, audit o post-action verification.", "VERIFY"),
        ("UX-3", "Modificare UX/voice soltanto con criterio misurabile e senza compromettere safety/performance.", "CODE_IF_EVIDENCE"),
    ],
}

MANDATORY = [
    {"id": "BASE-1", "priority": "CRITICAL", "objective": "Inventario commit/diff dell'ora precedente e classificazione PASS/FAIL/UNKNOWN.", "kind": "AUDIT", "status": "PENDING"},
    {"id": "BASE-2", "priority": "CRITICAL", "objective": "Controllare che nessun commit autonomo abbia toccato stable/main direttamente o allargato capability protette.", "kind": "SAFETY", "status": "PENDING"},
    {"id": "BASE-3", "priority": "HIGH", "objective": "Verificare stato test/evidence; un risultato mancante non puo diventare PASS.", "kind": "VERIFY", "status": "PENDING"},
]

def sh(*args: str) -> str:
    return subprocess.check_output(args, cwd=ROOT, text=True, stderr=subprocess.DEVNULL).strip()

def changed_files(minutes: int) -> List[str]:
    try:
        out = sh("git", "log", f"--since={minutes} minutes ago", "--name-only", "--pretty=format:")
    except Exception:
        return []
    return sorted({line.strip().replace('\\', '/') for line in out.splitlines() if line.strip()})

def commit_count(minutes: int) -> int:
    try:
        return int(sh("git", "rev-list", "--count", f"--since={minutes} minutes ago", "HEAD") or "0")
    except Exception:
        return 0

def matches(path: str, pats: List[str]) -> bool:
    p = path.replace('\\', '/')
    return any(fnmatch.fnmatch(p, pat) or fnmatch.fnmatch(p.split('/')[-1], pat) for pat in pats)

def load_state() -> dict:
    if not STATE_PATH.exists():
        return {"last_focus": None, "last_checked": {}, "cycles": 0}
    try:
        return json.loads(STATE_PATH.read_text(encoding="utf-8"))
    except Exception:
        return {"last_focus": None, "last_checked": {}, "cycles": 0}

def choose_focus(files: List[str], state: dict, now: dt.datetime) -> tuple[str, dict]:
    scores = {}
    last_checked = state.get("last_checked", {})
    for idx, focus in enumerate(FOCI):
        touch = sum(1 for f in files if matches(f, PATTERNS[focus]))
        last = last_checked.get(focus)
        if last:
            try:
                age = max(0.0, (now - dt.datetime.fromisoformat(last.replace('Z', '+00:00'))).total_seconds() / 3600)
            except Exception:
                age = 24.0
        else:
            age = 24.0
        staleness = min(age, 24.0) / 3.0
        rotation = 0.01 * ((now.hour + idx) % len(FOCI))
        score = touch * 6.0 + staleness + rotation
        if focus == state.get("last_focus") and touch == 0:
            score -= 3.0
        scores[focus] = round(score, 3)
    return max(FOCI, key=lambda x: scores[x]), scores

def main() -> int:
    minutes = int(os.environ.get("NEURA_SUPERVISOR_WINDOW_MINUTES", "75"))
    now = dt.datetime.now(dt.timezone.utc)
    state = load_state()
    files = changed_files(minutes)
    focus, scores = choose_focus(files, state, now)
    head = sh("git", "rev-parse", "HEAD")
    commits = commit_count(minutes)
    focus_missions = []
    for mid, objective, kind in FOCUS_MISSIONS[focus]:
        focus_missions.append({"id": mid, "priority": "HIGH", "objective": objective, "kind": kind, "status": "PENDING"})
    board = {
        "schema": 1,
        "generated_at_utc": now.isoformat().replace('+00:00', 'Z'),
        "valid_until_utc": (now + dt.timedelta(hours=1)).isoformat().replace('+00:00', 'Z'),
        "branch": os.environ.get("GITHUB_REF_NAME", "candidate/autonomy"),
        "head_sha": head,
        "window_minutes": minutes,
        "commits_in_window": commits,
        "changed_files": files,
        "activity_controlled": focus,
        "focus_scores": scores,
        "rules": {
            "one_macro_area_per_candidate_run": True,
            "stable_direct_write": False,
            "unknown_is_pass": False,
            "new_capability_without_human_approval": False,
            "max_worker_missions_before_next_supervisor": 3,
        },
        "missions": MANDATORY + focus_missions,
    }
    BOARD_JSON.write_text(json.dumps(board, indent=2, ensure_ascii=False) + "\n", encoding="utf-8")
    lines = [
        "# NEURA H24 — Mission Board",
        "",
        f"- Generated: `{board['generated_at_utc']}`",
        f"- Valid until: `{board['valid_until_utc']}`",
        f"- Head: `{head}`",
        f"- Commits reviewed: **{commits}**",
        f"- **FOCUS DELL'ORA: `{focus}`**",
        "",
        "## File modificati nell'ultima finestra",
    ]
    lines += [f"- `{f}`" for f in files] if files else ["- Nessun file modificato nella finestra."]
    lines += ["", "## Missioni fino al prossimo controllo"]
    for i, m in enumerate(board["missions"], 1):
        lines.append(f"{i}. **{m['id']} · {m['priority']} · {m['kind']}** — {m['objective']}")
    lines += [
        "",
        "## Regola di stop",
        "Se una missione produce FAIL di sicurezza/integrita/recovery, le successive missioni di CODE restano sospese fino al prossimo Supervisor.",
    ]
    BOARD_MD.write_text("\n".join(lines) + "\n", encoding="utf-8")
    state.setdefault("last_checked", {})[focus] = board["generated_at_utc"]
    state["last_focus"] = focus
    state["cycles"] = int(state.get("cycles", 0)) + 1
    state["last_head_sha"] = head
    STATE_PATH.write_text(json.dumps(state, indent=2) + "\n", encoding="utf-8")
    print(json.dumps({"focus": focus, "commits": commits, "changed_files": len(files), "board": str(BOARD_JSON.relative_to(ROOT))}))
    return 0

if __name__ == "__main__":
    raise SystemExit(main())
