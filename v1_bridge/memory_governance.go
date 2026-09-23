package main

import (
	"context"
	"errors"
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
	for _, x := range s.memories {
		if x.ID == m.ID {
			return x, false, nil
		}
	}
	if meta.Supersedes != "" {
		found := false
		for _, old := range s.memories {
			if old.ID == meta.Supersedes {
				found = true
				break
			}
		}
		if !found {
			return Memory{}, false, errors.New("superseded memory not found")
		}
	}
	if err := appendJSONL(filepath.Join(s.dir, "memory.jsonl"), m); err != nil {
		return Memory{}, false, err
	}
	s.memories = append(s.memories, m)
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
