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

// EntityStateStore is the single thread-safe source of truth for NEURA's visible
// state. Subscribers receive the current snapshot immediately and subsequent
// state changes without coupling animation code to the agent loop.
type EntityStateStore struct {
	mu          sync.RWMutex
	current     EntitySnapshot
	subscribers map[uint64]chan EntitySnapshot
	nextID      uint64
}

func NewEntityStateStore() *EntityStateStore {
	return &EntityStateStore{
		current:     EntitySnapshot{Mode: EntityIdle, UpdatedAt: time.Now().UTC()},
		subscribers: make(map[uint64]chan EntitySnapshot),
	}
}

func (s *EntityStateStore) Snapshot() EntitySnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.current
}

// Subscribe returns a bounded stream suitable for a UI renderer. Delivery is
// deliberately latest-value/non-blocking: a slow animation must never stall the
// agent. The cancel function is idempotent and closes the subscriber channel.
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
		// Keep only the freshest state. UI consumers do not need to replay stale
		// animation frames, and the core must never block on presentation work.
		select {
		case ch <- next:
		default:
			select {
			case <-ch:
			default:
			}
			select {
			case ch <- next:
			default:
			}
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
	if len(detail) > max {
		detail = detail[:max]
	}
	return detail
}
