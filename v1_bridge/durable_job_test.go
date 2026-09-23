package main

import (
    "errors"
    "path/filepath"
    "testing"
)

func TestDurableJobHappyPathAndIdempotency(t *testing.T) {
    s, err := OpenJobStore(t.TempDir()); if err != nil { t.Fatal(err) }
    j, err := s.Enqueue("j1", "key-1", 1, 2); if err != nil { t.Fatal(err) }
    dup, err := s.Enqueue("other", "key-1", 9, 9); if err != nil { t.Fatal(err) }
    if dup.ID != j.ID { t.Fatal("idempotency key created duplicate job") }
    if _, err = s.Claim("j1", 1); err != nil { t.Fatal(err) }
    done, err := s.Complete("j1", 1, nil); if err != nil { t.Fatal(err) }
    if done.State != JobSucceeded { t.Fatalf("state=%s", done.State) }
}

func TestDurableJobGenerationFenceAndCancellation(t *testing.T) {
    s, _ := OpenJobStore(t.TempDir()); _, _ = s.Enqueue("j1", "key", 3, 2)
    if _, err := s.Claim("j1", 2); err == nil { t.Fatal("stale generation accepted") }
    if err := s.Cancel("j1", 3); err != nil { t.Fatal(err) }
    if _, err := s.Claim("j1", 3); err == nil { t.Fatal("cancelled job claimed") }
}

func TestDurableJobRetryBudget(t *testing.T) {
    s, _ := OpenJobStore(t.TempDir()); _, _ = s.Enqueue("j1", "key", 1, 2)
    _, _ = s.Claim("j1", 1)
    j, err := s.Complete("j1", 1, errors.New("boom")); if err != nil { t.Fatal(err) }
    if j.State != JobPending { t.Fatalf("expected retry pending, got %s", j.State) }
    _, _ = s.Claim("j1", 1)
    j, err = s.Complete("j1", 1, errors.New("boom again")); if err != nil { t.Fatal(err) }
    if j.State != JobFailed { t.Fatalf("expected failed, got %s", j.State) }
}

func TestDurableJobRestartRecovery(t *testing.T) {
    root := t.TempDir(); s, _ := OpenJobStore(root); _, _ = s.Enqueue("j1", "key", 4, 3); _, _ = s.Claim("j1", 4)
    reopened, err := OpenJobStore(root); if err != nil { t.Fatal(err) }
    if err := reopened.RecoverInterrupted(); err != nil { t.Fatal(err) }
    j, ok := reopened.Get("j1"); if !ok { t.Fatal("job lost across restart") }
    if j.State != JobPending || j.Generation != 5 { t.Fatalf("bad recovery: state=%s gen=%d", j.State, j.Generation) }
    if _, err := reopened.Claim("j1", 4); err == nil { t.Fatal("pre-crash worker generation not fenced") }
    if _, err := reopened.Claim("j1", 5); err != nil { t.Fatal(err) }
    if filepath.Base(reopened.path) != "jobs.json" { t.Fatal("unexpected persistence path") }
}
