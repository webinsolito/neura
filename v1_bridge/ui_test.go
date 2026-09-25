package main

import (
 "net/http/httptest"
 "strings"
 "testing"
)

func TestPremiumCommandLifecycleUI(t *testing.T){
 r:=httptest.NewRequest("GET","http://127.0.0.1:8765/",nil); w:=httptest.NewRecorder(); serveUI(w,r)
 if w.Code!=200 { t.Fatalf("status=%d",w.Code) }
 body:=w.Body.String()
 for _,want:=range []string{"PENSO","LAVORO","COMPLETATO","COSA È SUCCESSO","activityTitle","friendlyError","aria-live=\"polite\"","@media(max-width:720px)"}{if !strings.Contains(body,want){t.Fatalf("UI missing %q",want)}}
 if !strings.Contains(body,"send.disabled") {t.Fatal("submit concurrency guard missing")}
}

func TestUIKeepsLocalSecurityHeaders(t *testing.T){
 r:=httptest.NewRequest("GET","http://127.0.0.1:8765/",nil); w:=httptest.NewRecorder(); serveUI(w,r)
 csp:=w.Header().Get("Content-Security-Policy")
 for _,want:=range []string{"default-src 'self'","connect-src 'self'","object-src 'none'","frame-ancestors 'none'"}{if !strings.Contains(csp,want){t.Fatalf("CSP missing %q",want)}}
}
