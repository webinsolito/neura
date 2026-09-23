package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type RecoveryAction struct {
	Path     string `json:"path"`
	Decision string `json:"decision"`
	Verified bool   `json:"verified"`
	Detail   string `json:"detail,omitempty"`
}

type RecoveryReport struct {
	StartedAt       time.Time        `json:"started_at"`
	FinishedAt      time.Time        `json:"finished_at"`
	WorkspaceActions []RecoveryAction `json:"workspace_actions"`
	JobsRecovered   bool             `json:"jobs_recovered"`
	ManualReview    bool             `json:"manual_review"`
}

type RecoveryManager struct {
	workspace string
}

func NewRecoveryManager(workspace string) (*RecoveryManager, error) {
	abs, err := filepath.Abs(workspace)
	if err != nil { return nil, err }
	real, err := filepath.EvalSymlinks(abs)
	if err == nil { abs = real }
	return &RecoveryManager{workspace:abs}, nil
}

func (r *RecoveryManager) recoverAtomicResidues() ([]RecoveryAction, error) {
	var actions []RecoveryAction
	seen := 0
	err := filepath.WalkDir(r.workspace, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil { return walkErr }
		if d.IsDir() {
			if d.Name()==".git" { return filepath.SkipDir }
			return nil
		}
		seen++
		if seen > 5000 { return errors.New("recovery scan limit exceeded") }
		if !strings.HasSuffix(d.Name(), ".neura.tmp") { return nil }

		target := strings.TrimSuffix(path, ".neura.tmp")
		backup := target + ".neura.bak"

		if _, err := os.Stat(target); err == nil {
			if err := os.Remove(path); err != nil { return err }
			if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
				return errors.New("stale temp cleanup verification failed")
			}
			actions = append(actions, RecoveryAction{Path:target,Decision:"keep_target_remove_stale_temp",Verified:true})
			return nil
		} else if !errors.Is(err, os.ErrNotExist) {
			return err
		}

		if b, err := os.ReadFile(backup); err == nil {
			if err := writeAtomic(target, b); err != nil { return err }
			if restored, err := os.ReadFile(target); err != nil || digestBytes(restored)!=digestBytes(b) {
				return errors.New("backup restoration verification failed")
			}
			if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) { return err }
			actions = append(actions, RecoveryAction{Path:target,Decision:"rollback_from_backup",Verified:true})
			return nil
		} else if !errors.Is(err, os.ErrNotExist) {
			return err
		}

		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) { return err }
		if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
			return errors.New("incomplete creation cancellation verification failed")
		}
		actions = append(actions, RecoveryAction{Path:target,Decision:"cancel_incomplete_creation",Verified:true})
		return nil
	})
	return actions, err
}

func (r *RecoveryManager) Recover() (RecoveryReport, error) {
	rep := RecoveryReport{StartedAt:time.Now().UTC()}
	actions, err := r.recoverAtomicResidues()
	if err != nil {
		rep.ManualReview = true
		rep.FinishedAt = time.Now().UTC()
		return rep, fmt.Errorf("workspace recovery failed: %w", err)
	}
	rep.WorkspaceActions = actions

	jobStore, err := OpenJobStore(r.workspace)
	if err != nil {
		rep.ManualReview = true
		rep.FinishedAt = time.Now().UTC()
		return rep, fmt.Errorf("job recovery failed closed: %w", err)
	}
	beforeRunning := false
	for _, j := range jobStore.jobs {
		if j.State == JobRunning { beforeRunning = true; break }
	}
	if err := jobStore.RecoverInterrupted(); err != nil {
		rep.ManualReview = true
		rep.FinishedAt = time.Now().UTC()
		return rep, fmt.Errorf("job recovery failed: %w", err)
	}
	rep.JobsRecovered = beforeRunning
	rep.FinishedAt = time.Now().UTC()
	return rep, nil
}
