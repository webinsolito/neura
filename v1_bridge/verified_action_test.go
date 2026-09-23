package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func writerFixture(t *testing.T) (*WorkspaceWriter, *CapabilityGate, string) {
	t.Helper()
	now := time.Date(2026, 9, 23, 12, 0, 0, 0, time.UTC)
	g, err := newCapabilityGateForTest(bytes.Repeat([]byte{7}, 32), func() time.Time { return now })
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	w, err := NewWorkspaceWriter(root, g)
	if err != nil {
		t.Fatal(err)
	}
	return w, g, root
}
func writePassport(t *testing.T, g *CapabilityGate, purpose string, cap string) CapabilityPassport {
	t.Helper()
	p, err := g.Issue("test", purpose, []string{cap}, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	return p
}
func TestVerifiedWriteHappyPathAndRollback(t *testing.T) {
	w, g, root := writerFixture(t)
	target := filepath.Join(root, "notes.txt")
	if err := os.WriteFile(target, []byte("before"), 0o600); err != nil {
		t.Fatal(err)
	}
	p := writePassport(t, g, "edit notes", capFSWrite)
	r, err := w.Write("notes.txt", []byte("after"), p, "edit notes")
	if err != nil {
		t.Fatal(err)
	}
	if !r.Verified {
		t.Fatal("write not verified")
	}
	b, _ := os.ReadFile(target)
	if string(b) != "after" {
		t.Fatalf("got %q", b)
	}
	if err := w.Rollback(r); err != nil {
		t.Fatal(err)
	}
	b, _ = os.ReadFile(target)
	if string(b) != "before" {
		t.Fatalf("rollback got %q", b)
	}
}
func TestVerifiedWriteCapabilityDenied(t *testing.T) {
	w, g, _ := writerFixture(t)
	p := writePassport(t, g, "edit", "fs.read")
	if _, err := w.Write("x.txt", []byte("x"), p, "edit"); err == nil || !strings.Contains(err.Error(), "capability denied") {
		t.Fatalf("expected capability denial, got %v", err)
	}
}
func TestVerifiedWriteWrongPurpose(t *testing.T) {
	w, g, _ := writerFixture(t)
	p := writePassport(t, g, "one", capFSWrite)
	if _, err := w.Write("x.txt", []byte("x"), p, "two"); err == nil {
		t.Fatal("expected purpose denial")
	}
}
func TestVerifiedWriteExpiredPassport(t *testing.T) {
	now := time.Date(2026, 9, 23, 12, 0, 0, 0, time.UTC)
	clock := now
	g, _ := newCapabilityGateForTest(bytes.Repeat([]byte{8}, 32), func() time.Time { return clock })
	w, _ := NewWorkspaceWriter(t.TempDir(), g)
	p, _ := g.Issue("test", "edit", []string{capFSWrite}, time.Second)
	clock = now.Add(2 * time.Second)
	if _, err := w.Write("x.txt", []byte("x"), p, "edit"); err == nil {
		t.Fatal("expected expiry denial")
	}
}
func TestVerifiedWriteTamperedPassport(t *testing.T) {
	w, g, _ := writerFixture(t)
	p := writePassport(t, g, "edit", capFSWrite)
	p.Signature = "00" + p.Signature[2:]
	if _, err := w.Write("x.txt", []byte("x"), p, "edit"); err == nil {
		t.Fatal("expected tamper denial")
	}
}
func TestVerifiedWritePathEscape(t *testing.T) {
	w, g, _ := writerFixture(t)
	p := writePassport(t, g, "edit", capFSWrite)
	if _, err := w.Write(filepath.Join("..", "escape.txt"), []byte("x"), p, "edit"); err == nil || !strings.Contains(err.Error(), "escapes workspace") {
		t.Fatalf("expected escape denial, got %v", err)
	}
}
func TestVerifiedWriteSensitivePath(t *testing.T) {
	w, g, _ := writerFixture(t)
	p := writePassport(t, g, "edit", capFSWrite)
	if _, err := w.Write(".env", []byte("SECRET=x"), p, "edit"); err == nil || !strings.Contains(err.Error(), "secret firewall") {
		t.Fatalf("expected secret denial, got %v", err)
	}
}
func TestVerifiedWriteSizeLimit(t *testing.T) {
	w, g, _ := writerFixture(t)
	p := writePassport(t, g, "edit", capFSWrite)
	if _, err := w.Write("big.bin", make([]byte, maxWorkspaceWrite+1), p, "edit"); err == nil {
		t.Fatal("expected size denial")
	}
}
func TestRollbackNewFileRemovesIt(t *testing.T) {
	w, g, root := writerFixture(t)
	p := writePassport(t, g, "create", capFSWrite)
	r, err := w.Write("new.txt", []byte("new"), p, "create")
	if err != nil {
		t.Fatal(err)
	}
	if err := w.Rollback(r); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "new.txt")); !os.IsNotExist(err) {
		t.Fatalf("expected removal, got %v", err)
	}
}

func TestVerifiedWriteRecoversFromInterruptedTemp(t *testing.T) {
	w, g, root := writerFixture(t)
	target := filepath.Join(root, "recover.txt")
	if err := os.WriteFile(target, []byte("before"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target+".neura.tmp", []byte("partial-from-interrupted-write"), 0o600); err != nil {
		t.Fatal(err)
	}
	p := writePassport(t, g, "recover", capFSWrite)
	r, err := w.Write("recover.txt", []byte("after"), p, "recover")
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "after" {
		t.Fatalf("unexpected content %q", b)
	}
	if _, err := os.Stat(target + ".neura.tmp"); !os.IsNotExist(err) {
		t.Fatalf("stale temp survived verified write: %v", err)
	}
	if err := w.Rollback(r); err != nil {
		t.Fatal(err)
	}
	b, err = os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "before" {
		t.Fatalf("rollback got %q", b)
	}
}
func TestRollbackRejectsTamperedBackup(t *testing.T) {
	w, g, root := writerFixture(t)
	target := filepath.Join(root, "tamper.txt")
	if err := os.WriteFile(target, []byte("before"), 0o600); err != nil {
		t.Fatal(err)
	}
	p := writePassport(t, g, "tamper", capFSWrite)
	r, err := w.Write("tamper.txt", []byte("after"), p, "tamper")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(r.BackupPath, []byte("tampered"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := w.Rollback(r); err == nil || !strings.Contains(err.Error(), "backup integrity mismatch") {
		t.Fatalf("expected backup integrity failure, got %v", err)
	}
}
