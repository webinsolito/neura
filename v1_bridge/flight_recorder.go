package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type FlightEvent struct {
	ID          string    `json:"id"`
	ParentID    string    `json:"parent_id,omitempty"`
	GoalHash    string    `json:"goal_hash,omitempty"`
	Action      string    `json:"action"`
	Phase       string    `json:"phase"`
	Status      string    `json:"status"`
	Error       string    `json:"error,omitempty"`
	CauseChain  []string  `json:"cause_chain,omitempty"`
	Replay      ReplayHint `json:"replay"`
	OccurredAt  time.Time `json:"occurred_at"`
}

type ReplayHint struct {
	Action string `json:"action"`
	Safe   bool   `json:"safe"`
}

type FlightRecorder struct {
	mu   sync.Mutex
	path string
	now  func() time.Time
}

func NewFlightRecorder(dataDir string) (*FlightRecorder, error) {
	if strings.TrimSpace(dataDir) == "" {
		return nil, errors.New("flight recorder data directory required")
	}
	abs, err := filepath.Abs(dataDir)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(abs, 0o700); err != nil {
		return nil, err
	}
	return &FlightRecorder{path: filepath.Join(abs, "flight.jsonl"), now: func() time.Time { return time.Now().UTC() }}, nil
}

func flightGoalHash(goal string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(goal)))
	return hex.EncodeToString(sum[:12])
}

func errorCauseChain(err error) []string {
	if err == nil {
		return nil
	}
	var out []string
	for err != nil {
		msg := strings.TrimSpace(err.Error())
		if msg != "" && (len(out) == 0 || out[len(out)-1] != msg) {
			out = append(out, msg)
		}
		err = errors.Unwrap(err)
	}
	return out
}

func (r *FlightRecorder) Record(parentID, goal, action, phase, status string, err error) (FlightEvent, error) {
	if r == nil {
		return FlightEvent{}, errors.New("flight recorder unavailable")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	at := r.now().UTC()
	event := FlightEvent{
		ID:         idFor(strings.Join([]string{parentID, action, phase, status, at.Format(time.RFC3339Nano)}, "|")),
		ParentID:   parentID,
		GoalHash:   flightGoalHash(goal),
		Action:     strings.TrimSpace(action),
		Phase:      strings.TrimSpace(phase),
		Status:     strings.TrimSpace(status),
		OccurredAt: at,
		Replay:     ReplayHint{Action: strings.TrimSpace(action), Safe: false},
	}
	if err != nil {
		event.Error = err.Error()
		event.CauseChain = errorCauseChain(err)
	}
	if event.Action == "" || event.Phase == "" || event.Status == "" {
		return FlightEvent{}, errors.New("flight event action, phase and status required")
	}
	b, err := json.Marshal(event)
	if err != nil {
		return FlightEvent{}, err
	}
	f, err := os.OpenFile(r.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return FlightEvent{}, err
	}
	defer f.Close()
	if _, err := f.Write(append(b, '\n')); err != nil {
		return FlightEvent{}, err
	}
	if err := f.Sync(); err != nil {
		return FlightEvent{}, err
	}
	return event, nil
}

func (r *FlightRecorder) Recent(limit int) ([]FlightEvent, error) {
	if r == nil {
		return nil, errors.New("flight recorder unavailable")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	f, err := os.Open(r.path)
	if errors.Is(err, os.ErrNotExist) {
		return []FlightEvent{}, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var all []FlightEvent
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	line := 0
	for sc.Scan() {
		line++
		if strings.TrimSpace(sc.Text()) == "" {
			continue
		}
		var e FlightEvent
		if err := json.Unmarshal(sc.Bytes(), &e); err != nil {
			return nil, fmt.Errorf("corrupt flight recorder line %d: %w", line, err)
		}
		all = append(all, e)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	if len(all) > limit {
		all = all[len(all)-limit:]
	}
	for i, j := 0, len(all)-1; i < j; i, j = i+1, j-1 {
		all[i], all[j] = all[j], all[i]
	}
	return all, nil
}
