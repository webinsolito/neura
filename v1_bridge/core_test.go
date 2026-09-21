package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func testCore(t *testing.T) (*Core, *Store, *ToolRegistry) {
	t.Helper()
	dir := t.TempDir()
	store, err := NewStore(filepath.Join(dir, "data"))
	if err != nil { t.Fatal(err) }
	ws := filepath.Join(dir, "ws")
	if err := os.MkdirAll(ws, 0o700); err != nil { t.Fatal(err) }
	tools, err := NewToolRegistry(ws)
	if err != nil { t.Fatal(err) }
	return &Core{store: store, tools: tools}, store, tools
}
func TestMemorySaveDedupPersistence(t *testing.T) {
	c, s, _ := testCore(t)
	r := c.Execute(context.Background(), "ricorda compra il latte")
	if r.Status != "ok" { t.Fatalf("%+v", r) }
	r = c.Execute(context.Background(), "ricorda   compra il latte ")
	if r.Message != "duplicate memory already present" { t.Fatalf("%+v", r) }
	reloaded, err := NewStore(s.dir); if err != nil { t.Fatal(err) }
	if got := len(reloaded.SearchMemory("latte", 10)); got != 1 { t.Fatalf("want 1 got %d", got) }
}
func TestPathTraversalBlocked(t *testing.T) {
	c, _, _ := testCore(t)
	r := c.Execute(context.Background(), "leggi file ../../etc/passwd")
	if r.Status != "blocked" { t.Fatalf("expected blocked: %+v", r) }
}
func TestReadAndListVerified(t *testing.T) {
	c, _, tools := testCore(t)
	if err := os.WriteFile(filepath.Join(tools.workspace, "a.txt"), []byte("hello"), 0o600); err != nil { t.Fatal(err) }
	r := c.Execute(context.Background(), "lista file .")
	if r.Status != "ok" || !r.Results[0].Verified { t.Fatalf("%+v", r) }
	r = c.Execute(context.Background(), "leggi file a.txt")
	if r.Status != "ok" || r.Results[0].Data != "hello" { t.Fatalf("%+v", r) }
}
func TestUnsupportedGoalNoModelIsExplicit(t *testing.T) {
	c, _, _ := testCore(t)
	r := c.Execute(context.Background(), "scrivi una poesia")
	if r.Status != "unavailable" { t.Fatalf("%+v", r) }
}
func TestReceiptPersisted(t *testing.T) {
	c, s, _ := testCore(t); _ = c.Execute(context.Background(), "stato")
	if _, n := s.Counts(); n != 1 { t.Fatalf("receipt count=%d", n) }
	r := s.RecentReceipts(1)
	if len(r) != 1 || r[0].Status != "verified" { t.Fatalf("%+v", r) }
}

func TestCorruptPrimaryRecoversFromBackup(t *testing.T) {
	c, s, _ := testCore(t)
	r := c.Execute(context.Background(), "ricorda recovery reale")
	if r.Status != "ok" { t.Fatalf("%+v", r) }
	primary := filepath.Join(s.dir, "memory.jsonl")
	if _, err := os.Stat(primary+".bak"); err != nil { t.Fatalf("backup missing: %v", err) }
	if err := os.WriteFile(primary, []byte("{corrupt}\\n"), 0o600); err != nil { t.Fatal(err) }
	reloaded, err := NewStore(s.dir)
	if err != nil { t.Fatalf("recovery failed: %v", err) }
	if got := len(reloaded.SearchMemory("recovery", 10)); got != 1 { t.Fatalf("want recovered memory, got %d", got) }
}
