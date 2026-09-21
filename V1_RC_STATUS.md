# NEURA V1 Integration RC1 — Real Status

Branch: `candidate/v1-integration-rc1`
Base: `candidate/c9-53-memory-retrieval-fabric`
Rollback: `rollback/pre-v1-integration-rc1-2026-09-21`
Stable: unchanged

## Implemented and locally verified

| Function | Implemented | Tested | Limit | State |
|---|---|---|---|---|
| Local bridge HTTP | yes | unit/integration/race + Windows cross-build | Windows native launch not tested here | GREEN-local |
| CAPO command ingress | yes | HTTP integration | unsupported goals need local model | GREEN-local |
| Persistent memory save/search | yes | persistence + dedupe tests | V1 lexical retrieval in bridge; C9.53 advanced retrieval remains inherited candidate | GREEN-local |
| Receipt/log persistence | yes | persistence test | JSONL local store | GREEN-local |
| Recovery from corrupt primary store | yes | corruption recovery test | one rolling .bak copy | GREEN-local |
| Tool allowlist | yes | read/list + denied-tool tests | V1 only system.info/fs.list/fs.read/windows.processes | GREEN-local |
| Path traversal protection | yes | negative test | workspace-scoped | GREEN-local |
| Secret file firewall | yes | negative test | filename/extension based V1 policy | GREEN-local |
| Local model executable adapter | yes | validation tests | real runtime binary not benchmarked | YELLOW |
| Local Ollama planner adapter | yes | mocked loopback integration + disallowed tool rejection | real installed model not benchmarked | YELLOW |
| Windows observe | yes | non-Windows fallback + Windows cross-build | Windows native tasklist execution not run in this environment | YELLOW |
| Windows action | no | n/a | explicitly not_implemented_in_rc1 | YELLOW/BACKLOG |
| UI → bridge | yes | JS syntax checked | browser-to-localhost runtime connectivity needs real Windows/browser test | YELLOW |
| Candidate/stable separation | yes | branch/rollback present | inherited C9.53 GitHub gate still unexecuted | YELLOW |
| Manual V1 RC CI | yes | workflow installed on main | first workflow_dispatch run still required | YELLOW |

## Local test evidence

- `go vet ./...`: PASS
- `go test -count=1 ./...`: PASS
- `go test -race -count=1 ./...`: PASS
- Windows amd64 cross-build: PASS
- Last locally built RC1 bridge SHA-256 after Windows-observe step: `ac67a7950c4cc40341ecdf5355fb108ec1972e160eab28e6268cd1ba45fcc5b6`

## Not claimed as PASS

- Windows native execution
- Real local model benchmark
- GUI computer-use actions
- Inherited C9.53 full GitHub Action
- V1 RC manual GitHub Action
- Stable promotion
