package main

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type MemoryWriteMeta struct {
	Type       string
	Provenance string
	ValidFrom  time.Time
	ValidUntil *time.Time
	Supersedes string
}

var governedMemoryTypes = map[string]bool{
	"fact": true,
	"preference": true,
	"decision": true,
	"task": true,
	"observation": true,
}

var secretMemoryPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)-----BEGIN [A-Z ]*PRIVATE KEY-----`),
	regexp.MustCompile(`(?i)\b(password|passwd|secret|api[_-]?key|access[_-]?token|refresh[_-]?token)\s*[:=]\s*\S+`),
	regexp.MustCompile(`(?i)\bBearer\s+[A-Za-z0-9._~+/=-]{12,}`),
	regexp.MustCompile(`\bsk-[A-Za-z0-9_-]{16,}`),
}

func memorySecretBlocked(text string) bool {
	for _, re := range secretMemoryPatterns {
		if re.MatchString(text) {
			return true
		}
	}
	return false
}

func normalizeMemoryMeta(meta MemoryWriteMeta, now time.Time) (MemoryWriteMeta, error) {
	meta.Type = strings.ToLower(strings.TrimSpace(meta.Type))
	meta.Provenance = strings.TrimSpace(meta.Provenance)
	meta.Supersedes = strings.TrimSpace(meta.Supersedes)
	if !governedMemoryTypes[meta.Type] {
		return MemoryWriteMeta{}, errors.New("memory type not allowed")
	}
	if meta.Provenance == "" || len(meta.Provenance) > 200 {
		return MemoryWriteMeta{}, errors.New("memory provenance required")
	}
	if meta.ValidFrom.IsZero() {
		meta.ValidFrom = now.UTC()
	} else {
		meta.ValidFrom = meta.ValidFrom.UTC()
	}
	if meta.ValidFrom.After(now.UTC().Add(5 * time.Minute)) {
		return MemoryWriteMeta{}, errors.New("memory valid_from too far in future")
	}
	if meta.ValidUntil != nil {
		v := meta.ValidUntil.UTC()
		if !v.After(meta.ValidFrom) {
			return MemoryWriteMeta{}, errors.New("memory valid_until must be after valid_from")
		}
		meta.ValidUntil = &v
	}
	return meta, nil
}

func validateMemoryGraph(docs []Memory) error {
	byID := map[string]Memory{}
	successor := map[string]string{}
	for _, m := range docs {
		if strings.TrimSpace(m.ID) == "" { return errors.New("memory graph contains empty id") }
		if _, exists := byID[m.ID]; exists { return fmt.Errorf("memory graph duplicate id %s", m.ID) }
		byID[m.ID] = m
	}
	for _, m := range docs {
		if m.Supersedes == "" { continue }
		if m.Supersedes == m.ID { return fmt.Errorf("memory graph self-supersession %s", m.ID) }
		parent, ok := byID[m.Supersedes]
		if !ok { return fmt.Errorf("memory graph missing predecessor %s", m.Supersedes) }
		if parent.Type != "" && m.Type != "" && parent.Type != m.Type { return fmt.Errorf("memory graph type mismatch %s -> %s", parent.Type, m.Type) }
		if prior, ok := successor[m.Supersedes]; ok && prior != m.ID { return fmt.Errorf("memory graph conflict: %s has successors %s and %s", m.Supersedes, prior, m.ID) }
		successor[m.Supersedes] = m.ID
	}
	for id := range byID {
		seen := map[string]bool{}
		cur := id
		for cur != "" {
			if seen[cur] { return fmt.Errorf("memory graph cycle at %s", cur) }
			seen[cur] = true
			m := byID[cur]
			cur = m.Supersedes
		}
	}
	return nil
}

func (s *Store) MemoryLineage(id string) ([]Memory, error) {
	id = strings.TrimSpace(id)
	if id == "" { return nil, errors.New("memory id required") }
	s.mu.Lock()
	cp := append([]Memory(nil), s.memories...)
	s.mu.Unlock()
	if err := validateMemoryGraph(cp); err != nil { return nil, err }
	byID := map[string]Memory{}
	for _, m := range cp { byID[m.ID] = m }
	cur, ok := byID[id]
	if !ok { return nil, errors.New("memory not found") }
	out := []Memory{}
	for {
		out = append(out, cur)
		if cur.Supersedes == "" { break }
		cur = byID[cur.Supersedes]
	}
	return out, nil
}

func (s *Store) SaveGovernedMemory(ctx context.Context, text string, meta MemoryWriteMeta, p LocalEmbeddingProvider) (Memory, bool, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return Memory{}, false, errors.New("empty memory")
	}
	if len([]byte(text)) > 32*1024 {
		return Memory{}, false, errors.New("memory too large")
	}
	if memorySecretBlocked(text) {
		return Memory{}, false, errors.New("memory blocked by secret firewall")
	}
	now := time.Now().UTC()
	meta, err := normalizeMemoryMeta(meta, now)
	if err != nil {
		return Memory{}, false, err
	}

	m := Memory{
		ID:         idFor(text),
		Text:       text,
		Persistent: true,
		CreatedAt:  now,
		Entities:   extractMemoryEntities(text),
		Type:       meta.Type,
		Provenance: meta.Provenance,
		ValidFrom:  meta.ValidFrom,
		ValidUntil: meta.ValidUntil,
		Supersedes: meta.Supersedes,
	}
	if p != nil {
		if emb, err := p.Embed(ctx, text); err == nil && validateEmbedding(emb) {
			m.Embedding = emb
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if err := validateMemoryGraph(s.memories); err != nil { return Memory{}, false, fmt.Errorf("memory store invalid: %w", err) }
	if meta.Supersedes == m.ID && meta.Supersedes != "" {
		return Memory{}, false, errors.New("memory cannot supersede itself")
	}
	if meta.Supersedes != "" {
		var parent Memory
		found := false
		for _, old := range s.memories {
			if old.ID == meta.Supersedes { parent = old; found = true; break }
		}
		if !found { return Memory{}, false, errors.New("superseded memory not found") }
		if parent.Type != "" && parent.Type != meta.Type {
			return Memory{}, false, errors.New("superseded memory type mismatch")
		}
		for _, existing := range s.memories {
			if existing.Supersedes == meta.Supersedes && existing.ID != m.ID {
				return Memory{}, false, fmt.Errorf("memory conflict: predecessor already superseded by %s", existing.ID)
			}
		}
	}
	for _, x := range s.memories {
		if x.ID == m.ID {
			if x.Supersedes != m.Supersedes || (x.Type != "" && x.Type != m.Type) {
				return Memory{}, false, errors.New("duplicate memory metadata conflict")
			}
			return x, false, nil
		}
	}
	next := append(append([]Memory(nil), s.memories...), m)
	if err := validateMemoryGraph(next); err != nil { return Memory{}, false, err }
	if err := appendJSONL(filepath.Join(s.dir, "memory.jsonl"), m); err != nil {
		return Memory{}, false, err
	}
	s.memories = next
	return m, true, nil
}

func filterGovernedMemories(docs []Memory, now time.Time) []Memory {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	superseded := map[string]bool{}
	for _, m := range docs {
		if m.Supersedes == "" {
			continue
		}
		if !m.ValidFrom.IsZero() && m.ValidFrom.After(now) {
			continue
		}
		if m.ValidUntil != nil && !m.ValidUntil.After(now) {
			continue
		}
		superseded[m.Supersedes] = true
	}
	out := make([]Memory, 0, len(docs))
	for _, m := range docs {
		if superseded[m.ID] {
			continue
		}
		if !m.ValidFrom.IsZero() && m.ValidFrom.After(now) {
			continue
		}
		if m.ValidUntil != nil && !m.ValidUntil.After(now) {
			continue
		}
		out = append(out, m)
	}
	return out
}
