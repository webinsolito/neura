package main

import (
	"strings"
	"sync"
	"testing"
	"time"
)

func TestEntityStateLifecycle(t *testing.T) {
	s := NewEntityStateStore()
	if got := s.Snapshot().Mode; got != EntityIdle {
		t.Fatalf("initial mode=%q", got)
	}

	modes := []EntityMode{EntityListening, EntityThinking, EntityWorking, EntitySuccess, EntityError, EntityIdle}
	for _, mode := range modes {
		if got := s.Set(mode, "ok").Mode; got != mode {
			t.Fatalf("mode=%q want=%q", got, mode)
		}
	}
}

func TestEntityStateRejectsUnknownAndSanitizesDetail(t *testing.T) {
	s := NewEntityStateStore()
	got := s.Set(EntityMode("dangerous-unknown"), strings.Repeat("x", 140)+"\nsecret")
	if got.Mode != EntityError {
		t.Fatalf("mode=%q", got.Mode)
	}
	if got.Detail != "invalid state" {
		t.Fatalf("detail=%q", got.Detail)
	}

	got = s.Set(EntityWorking, strings.Repeat("x", 140))
	if len(got.Detail) != 120 {
		t.Fatalf("detail length=%d", len(got.Detail))
	}
}

func TestEntityStateConcurrentAccess(t *testing.T) {
	s := NewEntityStateStore()
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.Set(EntityThinking, "planning")
			_ = s.Snapshot()
		}()
	}
	wg.Wait()
	if !validEntityMode(s.Snapshot().Mode) {
		t.Fatal("invalid final state")
	}
}

func TestEntityStateSubscriptionGetsInitialAndLatestWithoutBlocking(t *testing.T) {
	s := NewEntityStateStore()
	updates, cancel := s.Subscribe()
	defer cancel()

	select {
	case got := <-updates:
		if got.Mode != EntityIdle {
			t.Fatalf("initial subscription mode=%q", got.Mode)
		}
	case <-time.After(time.Second):
		t.Fatal("initial subscription snapshot timed out")
	}

	// Deliberately do not consume while publishing. Set must remain non-blocking
	// and the bounded stream should collapse stale animation states to the latest.
	s.Set(EntityListening, "listening")
	s.Set(EntityThinking, "planning")
	s.Set(EntityWorking, "executing")

	select {
	case got := <-updates:
		if got.Mode != EntityWorking || got.Detail != "executing" {
			t.Fatalf("latest subscription snapshot=%+v", got)
		}
	case <-time.After(time.Second):
		t.Fatal("latest subscription snapshot timed out")
	}
}

func TestEntityStateSubscriptionCancelIsIdempotent(t *testing.T) {
	s := NewEntityStateStore()
	updates, cancel := s.Subscribe()
	cancel()
	cancel()
	if _, ok := <-updates; ok {
		t.Fatal("subscription channel must be closed after cancel")
	}
}
