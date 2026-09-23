package main

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

func TestGovernedMemoryRejectsSecrets(t *testing.T) {
	s, err := NewStore(filepath.Join(t.TempDir(), "data"))
	if err != nil { t.Fatal(err) }
	_, _, err = s.SaveGovernedMemory(context.Background(), "api_key=abcdef1234567890", MemoryWriteMeta{Type:"fact", Provenance:"user-direct"}, nil)
	if err == nil { t.Fatal("secret was persisted") }
	if got := s.SearchMemory("abcdef1234567890", 10); len(got) != 0 { t.Fatalf("secret leaked: %+v", got) }
}

func TestGovernedMemoryRequiresTypedProvenance(t *testing.T) {
	s, _ := NewStore(filepath.Join(t.TempDir(), "data"))
	if _, _, err := s.SaveGovernedMemory(context.Background(), "hello", MemoryWriteMeta{Type:"unknown", Provenance:"user"}, nil); err == nil {
		t.Fatal("unknown type accepted")
	}
	if _, _, err := s.SaveGovernedMemory(context.Background(), "hello", MemoryWriteMeta{Type:"fact"}, nil); err == nil {
		t.Fatal("missing provenance accepted")
	}
}

func TestGovernedMemoryTemporalTruthHidesExpired(t *testing.T) {
	s, _ := NewStore(filepath.Join(t.TempDir(), "data"))
	exp := time.Now().UTC().Add(100 * time.Millisecond)
	if _, _, err := s.SaveGovernedMemory(context.Background(), "temporary observation alpha", MemoryWriteMeta{Type:"observation", Provenance:"tool:test", ValidUntil:&exp}, nil); err != nil {
		t.Fatal(err)
	}
	time.Sleep(150 * time.Millisecond)
	if got := s.SearchMemoryHybrid(context.Background(), "temporary alpha", 10, nil); len(got) != 0 {
		t.Fatalf("expired memory returned: %+v", got)
	}
}

func TestGovernedMemorySupersedesOlderTruth(t *testing.T) {
	s, _ := NewStore(filepath.Join(t.TempDir(), "data"))
	old, _, err := s.SaveGovernedMemory(context.Background(), "office is in room one", MemoryWriteMeta{Type:"fact", Provenance:"user-direct"}, nil)
	if err != nil { t.Fatal(err) }
	newer, _, err := s.SaveGovernedMemory(context.Background(), "office is in room two", MemoryWriteMeta{Type:"fact", Provenance:"user-direct", Supersedes:old.ID}, nil)
	if err != nil { t.Fatal(err) }
	got := s.SearchMemoryHybrid(context.Background(), "office room", 10, nil)
	for _, m := range got {
		if m.ID == old.ID { t.Fatalf("superseded memory returned: %+v", got) }
	}
	found := false
	for _, m := range got { if m.ID == newer.ID { found = true } }
	if !found { t.Fatalf("new truth missing: %+v", got) }
}

func TestGovernedMemoryMetadataPersistsAcrossRestart(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "data")
	s, _ := NewStore(dir)
	m, created, err := s.SaveGovernedMemory(context.Background(), "typed durable context", MemoryWriteMeta{Type:"decision", Provenance:"user-direct"}, nil)
	if err != nil || !created { t.Fatalf("created=%v err=%v", created, err) }
	if m.Type != "decision" || m.Provenance != "user-direct" || m.ValidFrom.IsZero() { t.Fatalf("%+v", m) }
	reloaded, err := NewStore(dir)
	if err != nil { t.Fatal(err) }
	got := reloaded.SearchMemoryHybrid(context.Background(), "typed durable", 10, nil)
	if len(got) != 1 || got[0].Type != "decision" || got[0].Provenance != "user-direct" { t.Fatalf("%+v", got) }
}

func TestDirectRememberUsesGovernedWrite(t *testing.T) {
	c, _, _ := testCore(t)
	r := c.Execute(context.Background(), "ricorda governance attiva")
	if r.Status != "ok" || len(r.Memory) != 1 { t.Fatalf("%+v", r) }
	if r.Memory[0].Type != "fact" || r.Memory[0].Provenance != "user-direct" { t.Fatalf("%+v", r.Memory[0]) }
}
