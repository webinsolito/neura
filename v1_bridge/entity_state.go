package main

import (
	"strings"
	"sync"
	"time"
)

type EntityMode string

const (
	EntityIdle      EntityMode = "idle"
	EntityListening EntityMode = "listening"
	EntityThinking  EntityMode = "thinking"
	EntityWorking   EntityMode = "working"
	EntitySuccess   EntityMode = "success"
	EntityError     EntityMode = "error"
)

type EntitySnapshot struct {
	Mode      EntityMode `json:"mode"`
	Detail    string     `json:"detail,omitempty"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type EntityStateStore struct {
	mu          sync.RWMutex
	current     EntitySnapshot
	subscribers map[uint64]chan EntitySnapshot
	nextID      uint64
}

func NewEntityStateStore() *EntityStateStore {
	return &EntityStateStore{current: EntitySnapshot{Mode: EntityIdle, UpdatedAt: time.Now().UTC()}, subscribers: make(map[uint64]chan EntitySnapshot)}
}

// DefaultEntityState is the process-wide presentation state used by the V1 core.
// It deliberately contains only sanitized presentation data and no tool payloads.
var DefaultEntityState = NewEntityStateStore()

func (s *EntityStateStore) Snapshot() EntitySnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.current
}

func (s *EntityStateStore) Subscribe() (<-chan EntitySnapshot, func()) {
	s.mu.Lock()
	id := s.nextID
	s.nextID++
	ch := make(chan EntitySnapshot, 1)
	ch <- s.current
	s.subscribers[id] = ch
	s.mu.Unlock()

	var once sync.Once
	cancel := func() {
		once.Do(func() {
			s.mu.Lock()
			if owned, ok := s.subscribers[id]; ok {
				delete(s.subscribers, id)
				select { case <-owned: default: }
				close(owned)
			}
			s.mu.Unlock()
		})
	}
	return ch, cancel
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
	for _, ch := range s.subscribers {
		select {
		case ch <- next:
		default:
			select { case <-ch: default: }
			select { case ch <- next: default: }
		}
	}
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
	if len(detail) > max { detail = detail[:max] }
	return detail
}
