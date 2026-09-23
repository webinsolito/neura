package main

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestFlightRecorderPersistsStructuredCauseChain(t *testing.T) {
	r, err := NewFlightRecorder(t.TempDir())
	if err != nil { t.Fatal(err) }
	base := errors.New("disk full")
	wrapped := fmt.Errorf("persist job failed: %w", base)
	ev, err := r.Record("", "write invoice", "job.persist", "execute", "failed", wrapped)
	if err != nil { t.Fatal(err) }
	if ev.Error == "" || len(ev.CauseChain) != 2 { t.Fatalf("%+v", ev) }
	if ev.CauseChain[0] != "persist job failed: disk full" || ev.CauseChain[1] != "disk full" { t.Fatalf("%+v", ev.CauseChain) }
	recent, err := r.Recent(10)
	if err != nil { t.Fatal(err) }
	if len(recent) != 1 || recent[0].ID != ev.ID { t.Fatalf("%+v", recent) }
}

func TestFlightRecorderDoesNotStoreRawGoal(t *testing.T) {
	r, _ := NewFlightRecorder(t.TempDir())
	goal := "TOKEN=super-secret-value"
	ev, err := r.Record("", goal, "fs.write.workspace", "authorize", "blocked", errors.New("capability denied"))
	if err != nil { t.Fatal(err) }
	if ev.GoalHash == "" || strings.Contains(ev.GoalHash, "secret") { t.Fatalf("%+v", ev) }
	recent, _ := r.Recent(1)
	if strings.Contains(fmt.Sprintf("%+v", recent), goal) { t.Fatal("raw goal leaked into flight recorder") }
}

func TestFlightRecorderCausalParentAndStableOrdering(t *testing.T) {
	r, _ := NewFlightRecorder(t.TempDir())
	now := time.Date(2026, 9, 23, 21, 0, 0, 0, time.UTC)
	r.now = func() time.Time { now }
	first, err := r.Record("", "goal", "plan", "plan", "ok", nil)
	if err != nil { t.Fatal(err) }
	now = now.Add(time.Second)
	second, err := r.Record(first.ID, "goal", "tool", "execute", "failed", errors.New("boom"))
	if err != nil { t.Fatal(err) }
	if second.ParentID != first.ID { t.Fatalf("%+v", second) }
	recent, _ := r.Recent(10)
	if len(recent) != 2 || recent[0].ID != second.ID || recent[1].ID != first.ID { t.Fatalf("%+v", recent) }
}

func TestFlightRecorderReplayIsFailClosed(t *testing.T) {
	r, _ := NewFlightRecorder(t.TempDir())
	ev, err := r.Record("", "goal", "fs.write.workspace", "execute", "failed", errors.New("boom"))
	if err != nil { t.Fatal(err) }
	if ev.Replay.Safe { t.Fatal("flight recorder must never mark replay safe automatically") }
}
