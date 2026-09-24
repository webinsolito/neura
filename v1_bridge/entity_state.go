package main

import (
	"strings"
	"sync"
	"time"
)

// EntityMode is the small, stable vocabulary exposed to the presentation layer.
// Keeping it in the core avoids UI implementations inventing incompatible states.
type EntityMode string

const (
	EntityIdle      EntityMode = "idle"
	EntityListening EntityMode = "listening"
	EntityThinking  EntityMode = "thinking"
	EntityWorking   EntityMode = "working"
	EntitySuccess   EntityMode = "success"
	EntityError     EntityMode = "error"
)

// EntitySnapshot is presentation-safe state. Detail must never contain secrets,
// raw tool payloads, paths, or stack traces; callers should provide short labels.
type EntitySnapshot struct {
	Mode      EntityMode `json:"mode"`
	Detail    string     `json:"detail,omitempty"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// EntityStateStore gives every future UI (desktop/web/native) one thread-safe
// source of truth for NEURA's visible state without coupling animation to agent logic.
type EntityStateStore struct {
	mu      sync.RWMutex
	current EntitySnapshot
}

func NewEntityStateStore() *EntityStateStore {
	return &EntityStateStore{current: EntitySnapshot{Mode: EntityIdle, UpdatedAt: time.Now().UTC()}}
}

func (s *EntityStateStore) Snapshot() EntitySnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.current
}

func (s *EntityStateStore) Set(mode EntityMode, detail string) EntitySnapshot {
	if !validEntityMode(mode) {
		mode = EntityError
		detail = "invalid state"
	}
	detail = sanitizeEntityDetail(detail)
	next := EntitySnapshot{Mode: mode, Detail: detail, UpdatedAt: time.Now().UTC()}
	s.mu.Lock()
	s.current = next
	s.mu.Unlock()
	return next
}

func validEntityMode(mode EntityMode) bool {
	switch mode {
	case EntityIdle, EntityListening, EntityThinking, EntityWorking, EntitySuccess, EntityError:
		return true
	default:
		return false
	}
}

func sanitizeEntityDetail(detail string) string {
	detail = strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(detail, "\r", " "), "\n", " "))
	const max = 120
	if len(detail) > max {
		detail = detail[:max]
	}
	return detail
}
