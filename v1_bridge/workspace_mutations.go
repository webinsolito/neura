package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type DirMutationReceipt struct {
	Path     string `json:"path"`
	Created  bool   `json:"created"`
	Verified bool   `json:"verified"`
}

type MoveReceipt struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Verified bool   `json:"verified"`
}

type DeleteReceipt struct {
	Original   string `json:"original"`
	Quarantine string `json:"quarantine"`
	Verified   bool   `json:"verified"`
}

func (r *ToolRegistry) safeMutationPath(user string) (string, error) {
	if r == nil || r.writer == nil {
		return "", errors.New("workspace mutation unavailable")
	}
	p, err := r.writer.resolve(user)
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(r.workspace, p)
	if err != nil {
		return "", err
	}
	if rel == ".neura" || strings.HasPrefix(rel, ".neura"+string(os.PathSeparator)) {
		return "", errors.New("NEURA internal state is protected")
	}
	return p, nil
}

func (r *ToolRegistry) mkdirWorkspace(path string) (DirMutationReceipt, error) {
	p, err := r.safeMutationPath(path)
	if err != nil { return DirMutationReceipt{}, err }
	if st, err := os.Stat(p); err == nil {
		if !st.IsDir() { return DirMutationReceipt{}, errors.New("path exists and is not a directory") }
		return DirMutationReceipt{Path:p, Created:false, Verified:true}, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return DirMutationReceipt{}, err
	}
	if err := os.MkdirAll(p, 0o700); err != nil { return DirMutationReceipt{}, err }
	st, err := os.Stat(p)
	if err != nil || !st.IsDir() {
		return DirMutationReceipt{}, errors.New("mkdir postcondition verification failed")
	}
	return DirMutationReceipt{Path:p, Created:true, Verified:true}, nil
}

func (r *ToolRegistry) rollbackMkdir(rec DirMutationReceipt) error {
	if !rec.Created { return nil }
	p, err := r.safeMutationPath(rec.Path)
	if err != nil { return err }
	if err := os.Remove(p); err != nil {
		return fmt.Errorf("mkdir rollback requires empty directory: %w", err)
	}
	if _, err := os.Stat(p); !errors.Is(err, os.ErrNotExist) {
		return errors.New("mkdir rollback verification failed")
	}
	return nil
}

func (r *ToolRegistry) moveWorkspace(from, to string) (MoveReceipt, error) {
	src, err := r.safeMutationPath(from)
	if err != nil { return MoveReceipt{}, err }
	dst, err := r.safeMutationPath(to)
	if err != nil { return MoveReceipt{}, err }
	if src == dst { return MoveReceipt{}, errors.New("source and destination are identical") }
	if _, err := os.Stat(src); err != nil { return MoveReceipt{}, err }
	if _, err := os.Stat(dst); err == nil { return MoveReceipt{}, errors.New("destination already exists") } else if !errors.Is(err, os.ErrNotExist) { return MoveReceipt{}, err }
	if err := os.MkdirAll(filepath.Dir(dst), 0o700); err != nil { return MoveReceipt{}, err }
	if err := os.Rename(src, dst); err != nil { return MoveReceipt{}, err }
	if _, err := os.Stat(src); !errors.Is(err, os.ErrNotExist) {
		_ = os.Rename(dst, src)
		return MoveReceipt{}, errors.New("move verification failed: source still exists")
	}
	if _, err := os.Stat(dst); err != nil {
		_ = os.Rename(dst, src)
		return MoveReceipt{}, errors.New("move verification failed: destination missing")
	}
	return MoveReceipt{From:src, To:dst, Verified:true}, nil
}

func (r *ToolRegistry) rollbackMove(rec MoveReceipt) error {
	dst, err := r.safeMutationPath(rec.To)
	if err != nil { return err }
	src, err := r.safeMutationPath(rec.From)
	if err != nil { return err }
	if _, err := os.Stat(src); err == nil { return errors.New("move rollback blocked: original path occupied") }
	if err := os.Rename(dst, src); err != nil { return err }
	if _, err := os.Stat(src); err != nil { return errors.New("move rollback verification failed") }
	if _, err := os.Stat(dst); !errors.Is(err, os.ErrNotExist) { return errors.New("move rollback verification failed: destination remains") }
	return nil
}

func (r *ToolRegistry) quarantineDelete(path string) (DeleteReceipt, error) {
	src, err := r.safeMutationPath(path)
	if err != nil { return DeleteReceipt{}, err }
	if _, err := os.Stat(src); err != nil { return DeleteReceipt{}, err }
	trashRoot := filepath.Join(r.workspace, ".neura", "trash")
	if err := os.MkdirAll(trashRoot, 0o700); err != nil { return DeleteReceipt{}, err }
	name := fmt.Sprintf("%d-%s", time.Now().UTC().UnixNano(), filepath.Base(src))
	dst := filepath.Join(trashRoot, name)
	if err := os.Rename(src, dst); err != nil { return DeleteReceipt{}, err }
	if _, err := os.Stat(src); !errors.Is(err, os.ErrNotExist) {
		_ = os.Rename(dst, src)
		return DeleteReceipt{}, errors.New("delete verification failed: original remains")
	}
	if _, err := os.Stat(dst); err != nil {
		_ = os.Rename(dst, src)
		return DeleteReceipt{}, errors.New("delete verification failed: quarantine missing")
	}
	return DeleteReceipt{Original:src, Quarantine:dst, Verified:true}, nil
}

func (r *ToolRegistry) restoreDeleted(quarantine, original string) (DeleteReceipt, error) {
	qAbs, err := filepath.Abs(quarantine)
	if err != nil { return DeleteReceipt{}, err }
	trashRoot := filepath.Join(r.workspace, ".neura", "trash")
	rel, err := filepath.Rel(trashRoot, qAbs)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return DeleteReceipt{}, errors.New("restore source is outside NEURA quarantine")
	}
	orig, err := r.safeMutationPath(original)
	if err != nil { return DeleteReceipt{}, err }
	if _, err := os.Stat(qAbs); err != nil { return DeleteReceipt{}, err }
	if _, err := os.Stat(orig); err == nil { return DeleteReceipt{}, errors.New("restore destination already exists") } else if !errors.Is(err, os.ErrNotExist) { return DeleteReceipt{}, err }
	if err := os.MkdirAll(filepath.Dir(orig), 0o700); err != nil { return DeleteReceipt{}, err }
	if err := os.Rename(qAbs, orig); err != nil { return DeleteReceipt{}, err }
	if _, err := os.Stat(orig); err != nil { return DeleteReceipt{}, errors.New("restore verification failed") }
	if _, err := os.Stat(qAbs); !errors.Is(err, os.ErrNotExist) { return DeleteReceipt{}, errors.New("restore verification failed: quarantine remains") }
	return DeleteReceipt{Original:orig, Quarantine:qAbs, Verified:true}, nil
}
