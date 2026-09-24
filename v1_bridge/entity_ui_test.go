package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestEntityUIIsLocalDependencyFreeAndHardened(t *testing.T) {
	s := testServer(t)
	rr := doReq(t, s, http.MethodGet, "/", nil, "")
	if rr.Code != http.StatusOK { t.Fatalf("code=%d", rr.Code) }
	body := rr.Body.String()
	for _, want := range []string{"NEURA", "EventSource('/entity/events", "prefers-reduced-motion", "class=\"entity\""} {
		if !strings.Contains(body, want) { t.Fatalf("missing %q", want) }
	}
	for _, forbidden := range []string{"https://cdn", "http://cdn", "<script src=", "<link rel=\"stylesheet\""} {
		if strings.Contains(body, forbidden) { t.Fatalf("external dependency found: %s", forbidden) }
	}
	if got := rr.Header().Get("Content-Security-Policy"); got == "" || !strings.Contains(got, "default-src 'none'") {
		t.Fatalf("missing hardened CSP: %q", got)
	}
}

func TestEntitySnapshotReflectsRealStateAndReducedMotion(t *testing.T) {
	s := testServer(t)
	DefaultEntityState.Set(EntityWorking, "executing verified step")
	t.Cleanup(func(){ DefaultEntityState.Set(EntityIdle, "") })
	rr := doReq(t, s, http.MethodGet, "/entity?reduced_motion=1", nil, "")
	if rr.Code != http.StatusOK { t.Fatalf("code=%d body=%s", rr.Code, rr.Body.String()) }
	var p EntityPresentation
	if err := json.Unmarshal(rr.Body.Bytes(), &p); err != nil { t.Fatal(err) }
	if p.Snapshot.Mode != EntityWorking || p.Visual.Motion != "flow" { t.Fatalf("%+v", p) }
	if p.Frame.Motion != "still" || p.Frame.ParticleRate != 0 || p.Frame.RotateXDeg != 0 || p.Frame.RotateYDeg != 0 {
		t.Fatalf("reduced motion not enforced: %+v", p.Frame)
	}
}

func TestEntitySSEStreamsInitialPresentation(t *testing.T) {
	s := testServer(t)
	DefaultEntityState.Set(EntityThinking, "planning task")
	t.Cleanup(func(){ DefaultEntityState.Set(EntityIdle, "") })

	ctx, cancel := context.WithCancel(context.Background())
	req := httptest.NewRequest(http.MethodGet, "http://127.0.0.1/entity/events", nil).WithContext(ctx)
	req.RemoteAddr = "127.0.0.1:12345"
	rr := httptest.NewRecorder()
	done := make(chan struct{})
	go func(){ s.routes().ServeHTTP(rr, req); close(done) }()

	deadline := time.Now().Add(500 * time.Millisecond)
	for !strings.Contains(rr.Body.String(), "event: entity") && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("SSE handler did not stop after context cancellation")
	}
	body := rr.Body.String()
	if !strings.Contains(body, "event: entity") || !strings.Contains(body, "\"mode\":\"thinking\"") {
		t.Fatalf("unexpected SSE body: %s", body)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "text/event-stream" {
		t.Fatalf("content-type=%q", ct)
	}
}
