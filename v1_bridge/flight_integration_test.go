package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
)

func TestCoreReceiptAlsoCreatesFlightEvent(t *testing.T) {
	c, _, _ := testCore(t)
	rec, err := NewFlightRecorder(filepath.Join(t.TempDir(), "flight"))
	if err != nil { t.Fatal(err) }
	c.recorder = rec
	res := c.Execute(context.Background(), "leggi file missing.txt")
	if res.Status != "blocked" { t.Fatalf("%+v", res) }
	events, err := rec.Recent(10)
	if err != nil { t.Fatal(err) }
	if len(events) != 1 { t.Fatalf("events=%d", len(events)) }
	if events[0].Action != "fs.read" || events[0].Status != "blocked" || len(events[0].CauseChain) == 0 {
		t.Fatalf("%+v", events[0])
	}
}

func TestFlightEndpointReturnsStructuredEvents(t *testing.T) {
	s := testServer(t)
	rec, err := NewFlightRecorder(filepath.Join(t.TempDir(), "flight"))
	if err != nil { t.Fatal(err) }
	s.core.recorder = rec
	_, err = rec.Record("", "sensitive goal text", "job.execute", "execute", "failed", errors.New("verified failure"))
	if err != nil { t.Fatal(err) }
	rr := doReq(t, s, http.MethodGet, "/flight?limit=10", nil, "")
	if rr.Code != http.StatusOK { t.Fatalf("code=%d body=%s", rr.Code, rr.Body.String()) }
	var body struct{ Items []FlightEvent `json:"items"` }
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil { t.Fatal(err) }
	if len(body.Items) != 1 || body.Items[0].Action != "job.execute" { t.Fatalf("%+v", body.Items) }
	if strings.Contains(rr.Body.String(), "sensitive goal text") { t.Fatal("raw goal leaked through flight endpoint") }
}
