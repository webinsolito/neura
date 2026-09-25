package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func testServer(t *testing.T) *Server {t.Helper(); c, s, tools := testCore(t);tasks, err := NewTaskRunner(c, tools.workspace); if err != nil { t.Fatal(err) };return &Server{core: c, store: s, tools: tools, allowedOrigin: "https://webinsolito.github.io", tasks: tasks}}
func doReq(t *testing.T, s *Server, method, path string, body any, origin string) *httptest.ResponseRecorder {t.Helper(); var b bytes.Buffer;if body != nil { _ = json.NewEncoder(&b).Encode(body) };req := httptest.NewRequest(method, "http://127.0.0.1"+path, &b); req.RemoteAddr = "127.0.0.1:12345";if origin != "" { req.Header.Set("Origin", origin) };rr := httptest.NewRecorder(); s.routes().ServeHTTP(rr, req); return rr}
func TestLocalUIAndAPIWiring(t *testing.T) {s := testServer(t);rr := doReq(t, s, "GET", "/", nil, "");if rr.Code != 200 { t.Fatalf("ui code=%d body=%s", rr.Code, rr.Body.String()) };if ct:=rr.Header().Get("Content-Type"); ct!="text/html; charset=utf-8" { t.Fatalf("content-type=%q",ct) };body:=rr.Body.Bytes();for _, token := range []string{"NEURA", "data-state=\"idle\"", "fetch('/status')", "fetch('/command'", "fetch('/tasks?limit=6')", "data-action=\"retry\"", "RECOVERY TASK", "MEMORIA ATTIVA", "fetch('/memory?q='", "fetch('/memory/history?id='", "Content-Security-Policy"} { if !bytes.Contains(body,[]byte(token)) && token!="Content-Security-Policy" { t.Fatalf("UI missing %q",token) } };if rr.Header().Get("Content-Security-Policy")=="" { t.Fatal("missing CSP") };rr = doReq(t,s,"GET","/status",nil,""); if rr.Code!=200 || !bytes.Contains(rr.Body.Bytes(),[]byte(`"core":"online"`)) { t.Fatalf("status=%d %s",rr.Code,rr.Body.String()) };rr = doReq(t,s,"POST","/command",map[string]string{"goal":"ricorda ui wiring reale"},""); if rr.Code!=200 { t.Fatalf("command=%d %s",rr.Code,rr.Body.String()) }}
func TestCommandRejectsEmptyGoal(t *testing.T){s:=testServer(t);rr:=doReq(t,s,"POST","/command",map[string]string{"goal":"   "},"");if rr.Code!=http.StatusBadRequest||!strings.Contains(rr.Body.String(),"goal is required"){t.Fatalf("code=%d body=%s",rr.Code,rr.Body.String())}}
func TestCommandRejectsOversizedGoal(t *testing.T){s:=testServer(t);rr:=doReq(t,s,"POST","/command",map[string]string{"goal":strings.Repeat("x",maxCommandRunes+1)},"");if rr.Code!=http.StatusBadRequest||!strings.Contains(rr.Body.String(),"goal exceeds"){t.Fatalf("code=%d body=%s",rr.Code,rr.Body.String())}}
func TestCommandTrimsMessageFallback(t *testing.T){s:=testServer(t);rr:=doReq(t,s,"POST","/command",map[string]string{"message":"  ricorda fallback pulito  "},"");if rr.Code!=http.StatusOK{t.Fatalf("code=%d body=%s",rr.Code,rr.Body.String())}}
func TestUIUnknownPathDoesNotMasqueradeAsApp(t *testing.T){ s:=testServer(t); rr:=doReq(t,s,"GET","/missing-ui-route",nil,""); if rr.Code!=http.StatusNotFound { t.Fatalf("code=%d",rr.Code) } }
func TestHealthAndCommand(t *testing.T) {s := testServer(t);rr := doReq(t, s, "GET", "/health", nil, ""); if rr.Code != 200 { t.Fatal(rr.Code, rr.Body.String()) };rr = doReq(t, s, "POST", "/command", map[string]string{"goal": "ricorda test reale"}, "https://webinsolito.github.io");if rr.Code != 200 { t.Fatal(rr.Code, rr.Body.String()) }}
func TestCorsDenied(t *testing.T) {s := testServer(t); rr := doReq(t, s, "GET", "/health", nil, "https://evil.example");if rr.Code != http.StatusForbidden { t.Fatalf("code=%d", rr.Code) }}
func TestNonLoopbackDenied(t *testing.T) {s := testServer(t); req := httptest.NewRequest("GET", "http://127.0.0.1/health", nil); req.RemoteAddr = "192.168.1.4:5555";rr := httptest.NewRecorder(); s.routes().ServeHTTP(rr, req);if rr.Code != http.StatusForbidden { t.Fatalf("code=%d", rr.Code) }}
func TestUnknownJSONFieldRejected(t *testing.T) {s := testServer(t); rr := doReq(t, s, "POST", "/command", map[string]any{"goal": "stato", "admin": true}, "");if rr.Code != 400 { t.Fatalf("code=%d body=%s", rr.Code, rr.Body.String()) }}
func TestStatusTruthfulWindowsUnavailable(t *testing.T) {s := testServer(t); rr := doReq(t, s, "GET", "/windows/status", nil, "");if rr.Code != 200 { t.Fatal(rr.Code) };if !bytes.Contains(rr.Body.Bytes(), []byte("allowlisted_launch_v1")) { t.Fatalf("%s", rr.Body.String()) }}
func TestStoreCorruptionFailsClosed(t *testing.T) {dir := t.TempDir(); _ = os.MkdirAll(filepath.Join(dir, "data"), 0o700);_ = os.WriteFile(filepath.Join(dir, "data", "memory.jsonl"), []byte("{bad}\n"), 0o600);if _, err := NewStore(filepath.Join(dir, "data")); err == nil { t.Fatal("expected corruption error") }}
func TestOllamaPlannerStrictAllowedTools(t *testing.T) {ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){ _ = json.NewEncoder(w).Encode(map[string]string{"response": `{"summary":"ok","steps":[{"tool":"system.info","input":{}}]}`}) })); defer ts.Close();m := ModelAdapter{OllamaURL:ts.URL,OllamaModel:"test",Timeout:time.Second};p,err:=m.Plan(context.Background(),"status",[]string{"system.info"}); if err!=nil { t.Fatal(err) }; if len(p.Steps)!=1 || p.Steps[0].Tool!="system.info" { t.Fatalf("%+v",p) }}
func TestOllamaPlannerRejectsDisallowedTool(t *testing.T) {ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){ _ = json.NewEncoder(w).Encode(map[string]string{"response": `{"summary":"bad","steps":[{"tool":"shell.exec","input":{}}]}`}) })); defer ts.Close();m := ModelAdapter{OllamaURL:ts.URL,OllamaModel:"test",Timeout:time.Second};if _,err:=m.Plan(context.Background(),"bad",[]string{"system.info"}); err==nil { t.Fatal("disallowed tool accepted") }}
func TestTrailingInvalidJSONRejected(t *testing.T) {s:=testServer(t);req:=httptest.NewRequest("POST","http://127.0.0.1/command",bytes.NewBufferString(`{"goal":"stato"} garbage`));req.RemoteAddr="127.0.0.1:12345";rr:=httptest.NewRecorder();s.routes().ServeHTTP(rr,req);if rr.Code!=http.StatusBadRequest { t.Fatalf("code=%d body=%s",rr.Code,rr.Body.String()) }}
