package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func testServer(t *testing.T) *Server {
	t.Helper(); c, s, tools := testCore(t)
	return &Server{core: c, store: s, tools: tools, allowedOrigin: "https://webinsolito.github.io"}
}
func doReq(t *testing.T, s *Server, method, path string, body any, origin string) *httptest.ResponseRecorder {
	t.Helper(); var b bytes.Buffer
	if body != nil { _ = json.NewEncoder(&b).Encode(body) }
	req := httptest.NewRequest(method, "http://127.0.0.1"+path, &b); req.RemoteAddr = "127.0.0.1:12345"
	if origin != "" { req.Header.Set("Origin", origin) }
	rr := httptest.NewRecorder(); s.routes().ServeHTTP(rr, req); return rr
}
func TestHealthAndCommand(t *testing.T) {
	s := testServer(t)
	rr := doReq(t, s, "GET", "/health", nil, ""); if rr.Code != 200 { t.Fatal(rr.Code, rr.Body.String()) }
	rr = doReq(t, s, "POST", "/command", map[string]string{"goal": "ricorda test reale"}, "https://webinsolito.github.io")
	if rr.Code != 200 { t.Fatal(rr.Code, rr.Body.String()) }
}
func TestCorsDenied(t *testing.T) {
	s := testServer(t); rr := doReq(t, s, "GET", "/health", nil, "https://evil.example")
	if rr.Code != http.StatusForbidden { t.Fatalf("code=%d", rr.Code) }
}
func TestNonLoopbackDenied(t *testing.T) {
	s := testServer(t); req := httptest.NewRequest("GET", "http://127.0.0.1/health", nil); req.RemoteAddr = "192.168.1.4:5555"
	rr := httptest.NewRecorder(); s.routes().ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden { t.Fatalf("code=%d", rr.Code) }
}
func TestUnknownJSONFieldRejected(t *testing.T) {
	s := testServer(t); rr := doReq(t, s, "POST", "/command", map[string]any{"goal": "stato", "admin": true}, "")
	if rr.Code != 400 { t.Fatalf("code=%d body=%s", rr.Code, rr.Body.String()) }
}
func TestStatusTruthfulWindowsUnavailable(t *testing.T) {
	s := testServer(t); rr := doReq(t, s, "GET", "/windows/status", nil, "")
	if rr.Code != 200 { t.Fatal(rr.Code) }
	if !bytes.Contains(rr.Body.Bytes(), []byte("not_implemented_in_rc1")) { t.Fatalf("%s", rr.Body.String()) }
}
func TestStoreCorruptionFailsClosed(t *testing.T) {
	dir := t.TempDir(); _ = os.MkdirAll(filepath.Join(dir, "data"), 0o700)
	_ = os.WriteFile(filepath.Join(dir, "data", "memory.jsonl"), []byte("{bad}\n"), 0o600)
	if _, err := NewStore(filepath.Join(dir, "data")); err == nil { t.Fatal("expected corruption error") }
}

func TestOllamaPlannerStrictAllowedTools(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){ _ = json.NewEncoder(w).Encode(map[string]string{"response": `{"summary":"ok","steps":[{"tool":"system.info","input":{}}]}`}) })); defer ts.Close()
	m := ModelAdapter{OllamaURL:ts.URL,OllamaModel:"test",Timeout:time.Second}
	p,err:=m.Plan(context.Background(),"status",[]string{"system.info"}); if err!=nil { t.Fatal(err) }; if len(p.Steps)!=1 || p.Steps[0].Tool!="system.info" { t.Fatalf("%+v",p) }
}
func TestOllamaPlannerRejectsDisallowedTool(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){ _ = json.NewEncoder(w).Encode(map[string]string{"response": `{"summary":"bad","steps":[{"tool":"shell.exec","input":{}}]}`}) })); defer ts.Close()
	m := ModelAdapter{OllamaURL:ts.URL,OllamaModel:"test",Timeout:time.Second}
	if _,err:=m.Plan(context.Background(),"bad",[]string{"system.info"}); err==nil { t.Fatal("disallowed tool accepted") }
}
