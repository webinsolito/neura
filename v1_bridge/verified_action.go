package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	capFSWrite        = "fs.write.workspace"
	maxWorkspaceWrite = 256 * 1024
)

type WriteReceipt struct {
	Path         string `json:"path"`
	BeforeSHA256 string `json:"before_sha256,omitempty"`
	AfterSHA256  string `json:"after_sha256"`
	Bytes        int    `json:"bytes"`
	Verified     bool   `json:"verified"`
	BackupPath   string `json:"backup_path,omitempty"`
}
type WorkspaceWriter struct {
	root string
	gate *CapabilityGate
}

func NewWorkspaceWriter(root string, gate *CapabilityGate) (*WorkspaceWriter, error) {
	if gate == nil {
		return nil, errors.New("capability gate required")
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(abs, 0o700); err != nil {
		return nil, err
	}
	real, err := filepath.EvalSymlinks(abs)
	if err == nil {
		abs = real
	}
	return &WorkspaceWriter{root: abs, gate: gate}, nil
}
func (w *WorkspaceWriter) resolve(path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", errors.New("path required")
	}
	p := path
	if !filepath.IsAbs(p) {
		p = filepath.Join(w.root, p)
	}
	p = filepath.Clean(p)
	parent := filepath.Dir(p)
	realParent, err := filepath.EvalSymlinks(parent)
	if err == nil {
		p = filepath.Join(realParent, filepath.Base(p))
	}
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(w.root, abs)
	if err != nil {
		return "", err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return "", errors.New("path escapes workspace")
	}
	if sensitivePath(abs) {
		return "", errors.New("sensitive file blocked by secret firewall")
	}
	return abs, nil
}
func digestBytes(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func (w *WorkspaceWriter) Write(path string, data []byte, passport CapabilityPassport, purpose string) (WriteReceipt, error) {
	if len(data) > maxWorkspaceWrite {
		return WriteReceipt{}, errors.New("write exceeds 256 KiB limit")
	}
	if err := w.gate.AuthorizeOnce(passport, capFSWrite, purpose); err != nil {
		return WriteReceipt{}, fmt.Errorf("capability denied: %w", err)
	}
	target, err := w.resolve(path)
	if err != nil {
		return WriteReceipt{}, err
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
		return WriteReceipt{}, err
	}
	r := WriteReceipt{Path: target, Bytes: len(data), AfterSHA256: digestBytes(data)}
	if old, err := os.ReadFile(target); err == nil {
		r.BeforeSHA256 = digestBytes(old)
		r.BackupPath = target + ".neura.bak"
		if err := writeAtomic(r.BackupPath, old); err != nil {
			return WriteReceipt{}, fmt.Errorf("backup failed: %w", err)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return WriteReceipt{}, err
	}
	if err := writeAtomic(target, data); err != nil {
		return WriteReceipt{}, err
	}
	got, err := os.ReadFile(target)
	if err != nil {
		_ = w.Rollback(r)
		return WriteReceipt{}, fmt.Errorf("post-read failed: %w", err)
	}
	if digestBytes(got) != r.AfterSHA256 {
		_ = w.Rollback(r)
		return WriteReceipt{}, errors.New("postcondition verification failed")
	}
	r.Verified = true
	return r, nil
}
func writeAtomic(path string, data []byte) error {
	tmp := path + ".neura.tmp"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	ok := false
	defer func() {
		_ = f.Close()
		if !ok {
			_ = os.Remove(tmp)
		}
	}()
	if _, err = f.Write(data); err != nil {
		return err
	}
	if err = f.Sync(); err != nil {
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	if err = os.Rename(tmp, path); err != nil {
		return err
	}
	ok = true
	return nil
}
func (w *WorkspaceWriter) Rollback(r WriteReceipt) error {
	target, err := w.resolve(r.Path)
	if err != nil {
		return err
	}
	if r.BackupPath == "" {
		if err := os.Remove(target); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		if _, err := os.Stat(target); !errors.Is(err, os.ErrNotExist) {
			if err == nil {
				return errors.New("rollback verification failed: new file still exists")
			}
			return fmt.Errorf("rollback verification failed: %w", err)
		}
		return nil
	}
	backup, err := w.resolve(r.BackupPath)
	if err != nil {
		return err
	}
	b, err := os.ReadFile(backup)
	if err != nil {
		return err
	}
	if r.BeforeSHA256 != "" && digestBytes(b) != r.BeforeSHA256 {
		return errors.New("backup integrity mismatch")
	}
	if err := writeAtomic(target, b); err != nil {
		return err
	}
	restored, err := os.ReadFile(target)
	if err != nil {
		return fmt.Errorf("rollback verification read failed: %w", err)
	}
	if r.BeforeSHA256 != "" && digestBytes(restored) != r.BeforeSHA256 {
		return errors.New("rollback verification failed: restored digest mismatch")
	}
	return nil
}
