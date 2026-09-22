package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode"
)

type LocalEmbeddingProvider interface {
	Embed(context.Context, string) ([]float32, error)
}

type CommandEmbeddingProvider struct {
	Executable string
	Args       []string
	Timeout    time.Duration
}

const (
	maxEmbeddingDimensions = 65536
	maxEmbeddingInputBytes = 256 * 1024
	maxEmbeddingOutputBytes = 4 * 1024 * 1024
	maxEmbeddingTimeout = 60 * time.Second
)

func (p CommandEmbeddingProvider) Available() bool {
	if p.Executable == "" || !filepath.IsAbs(p.Executable) {
		return false
	}
	st, err := os.Stat(p.Executable)
	return err == nil && !st.IsDir()
}

func validateEmbedding(v []float32) bool {
	if len(v) == 0 || len(v) > maxEmbeddingDimensions {
		return false
	}
	for _, x := range v {
		f := float64(x)
		if math.IsNaN(f) || math.IsInf(f, 0) {
			return false
		}
	}
	return true
}

func effectiveEmbeddingTimeout(d time.Duration) time.Duration {
	if d <= 0 {
		return 8 * time.Second
	}
	if d > maxEmbeddingTimeout {
		return maxEmbeddingTimeout
	}
	return d
}

func (p CommandEmbeddingProvider) Embed(ctx context.Context, text string) ([]float32, error) {
	if !p.Available() {
		return nil, errors.New("embedding executable unavailable or not absolute")
	}
	if len([]byte(text)) > maxEmbeddingInputBytes {
		return nil, errors.New("embedding request too large")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	runCtx, cancel := context.WithTimeout(ctx, effectiveEmbeddingTimeout(p.Timeout))
	defer cancel()

	payload, err := json.Marshal(map[string]string{"text": text})
	if err != nil {
		return nil, err
	}
	cmd := exec.CommandContext(runCtx, p.Executable, p.Args...)
	cmd.Stdin = bytes.NewReader(payload)
	var out limitedBuffer
	out.limit = maxEmbeddingOutputBytes
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		if runCtx.Err() != nil {
			return nil, fmt.Errorf("embedding timeout or cancellation: %w", runCtx.Err())
		}
		return nil, fmt.Errorf("embedding process failed: %w", err)
	}
	if out.exceeded {
		return nil, errors.New("embedding output too large")
	}
	var resp struct {
		Embedding []float32 `json:"embedding"`
	}
	if err := json.Unmarshal(out.b, &resp); err != nil {
		return nil, fmt.Errorf("invalid embedding json: %w", err)
	}
	if !validateEmbedding(resp.Embedding) {
		return nil, errors.New("invalid embedding")
	}
	return resp.Embedding, nil
}

type MemoryQuery struct {
	Text      string
	Entities  []string
	Now       time.Time
	Embedding []float32
	TopK      int
}

type MemoryScore struct {
	Memory   Memory
	Score    float64
	BM25     float64
	Semantic float64
	Entity   float64
	Temporal float64
}

type RetrievalWeights struct {
	BM25     float64
	Semantic float64
	Entity   float64
	Temporal float64
}

func DefaultRetrievalWeights() RetrievalWeights {
	return RetrievalWeights{BM25: 0.45, Semantic: 0.30, Entity: 0.15, Temporal: 0.10}
}

func sanitizeRetrievalWeights(w RetrievalWeights) RetrievalWeights {
	vals := []*float64{&w.BM25, &w.Semantic, &w.Entity, &w.Temporal}
	sum := 0.0
	for _, v := range vals {
		if math.IsNaN(*v) || math.IsInf(*v, 0) || *v < 0 {
			*v = 0
		}
		sum += *v
	}
	if sum <= 0 {
		return DefaultRetrievalWeights()
	}
	for _, v := range vals {
		*v /= sum
	}
	return w
}

func memoryTerms(s string) []string {
	return strings.FieldsFunc(strings.ToLower(s), func(r rune) bool {
		return !(unicode.IsLetter(r) || unicode.IsDigit(r))
	})
}

func uniqueMemoryTerms(in []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(in))
	for _, t := range in {
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		out = append(out, t)
	}
	return out
}

func extractMemoryEntities(s string) []string {
	raw := strings.FieldsFunc(s, func(r rune) bool {
		return !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_')
	})
	seen := map[string]struct{}{}
	var out []string
	for _, token := range raw {
		runes := []rune(token)
		if len(runes) < 2 {
			continue
		}
		hasDigit := false
		letters := 0
		upper := 0
		for _, r := range runes {
			if unicode.IsDigit(r) {
				hasDigit = true
			}
			if unicode.IsLetter(r) {
				letters++
				if unicode.IsUpper(r) {
					upper++
				}
			}
		}
		isEntity := hasDigit || (letters >= 2 && upper == letters) || (unicode.IsUpper(runes[0]) && len(runes) >= 3)
		if !isEntity {
			continue
		}
		n := strings.ToLower(token)
		if _, ok := seen[n]; ok {
			continue
		}
		seen[n] = struct{}{}
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}

func normalizeEntitySet(in []string) map[string]struct{} {
	out := map[string]struct{}{}
	for _, x := range in {
		x = strings.ToLower(strings.TrimSpace(x))
		if x != "" {
			out[x] = struct{}{}
		}
	}
	return out
}

func entityOverlap(query, doc []string) float64 {
	q := normalizeEntitySet(query)
	if len(q) == 0 {
		return 0
	}
	d := normalizeEntitySet(doc)
	if len(d) == 0 {
		return 0
	}
	hit := 0
	for k := range q {
		if _, ok := d[k]; ok {
			hit++
		}
	}
	return float64(hit) / float64(len(q))
}

func cosineMemory(a, b []float32) float64 {
	if len(a) == 0 || len(a) != len(b) {
		return 0
	}
	var dot, aa, bb float64
	for i := range a {
		x, y := float64(a[i]), float64(b[i])
		if math.IsNaN(x) || math.IsNaN(y) || math.IsInf(x, 0) || math.IsInf(y, 0) {
			return 0
		}
		dot += x * y
		aa += x * x
		bb += y * y
	}
	if aa == 0 || bb == 0 {
		return 0
	}
	v := dot / (math.Sqrt(aa) * math.Sqrt(bb))
	if v <= 0 || math.IsNaN(v) || math.IsInf(v, 0) {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func temporalMemory(now, created time.Time) float64 {
	if now.IsZero() || created.IsZero() || created.After(now) {
		return 0
	}
	days := now.Sub(created).Hours() / 24
	return math.Exp(-days / 30.0)
}

func bm25Memory(doc, query []string, df map[string]int, n int, avgdl float64) float64 {
	if len(doc) == 0 || len(query) == 0 || n == 0 {
		return 0
	}
	tf := map[string]int{}
	for _, t := range doc {
		tf[t]++
	}
	const k1 = 1.2
	const b = 0.75
	raw := 0.0
	for _, t := range query {
		f := float64(tf[t])
		if f == 0 {
			continue
		}
		idf := math.Log(1 + (float64(n-df[t])+0.5)/(float64(df[t])+0.5))
		raw += idf * (f * (k1 + 1)) / (f + k1*(1-b+b*float64(len(doc))/avgdl))
	}
	return raw / (1 + raw)
}

func RetrieveMemory(docs []Memory, q MemoryQuery, w RetrievalWeights) []MemoryScore {
	if q.TopK <= 0 || q.TopK > 50 {
		q.TopK = 10
	}
	if q.Now.IsZero() {
		q.Now = time.Now().UTC()
	}
	if len(q.Entities) == 0 {
		q.Entities = extractMemoryEntities(q.Text)
	}
	w = sanitizeRetrievalWeights(w)
	qterms := uniqueMemoryTerms(memoryTerms(q.Text))

	docTerms := make([][]string, len(docs))
	df := map[string]int{}
	totalLen := 0
	for i, d := range docs {
		docTerms[i] = memoryTerms(d.Text)
		totalLen += len(docTerms[i])
		seen := map[string]struct{}{}
		for _, term := range docTerms[i] {
			if _, ok := seen[term]; ok {
				continue
			}
			seen[term] = struct{}{}
			df[term]++
		}
	}
	avgdl := 1.0
	if len(docs) > 0 {
		avgdl = math.Max(1, float64(totalLen)/float64(len(docs)))
	}

	out := make([]MemoryScore, 0, len(docs))
	for i, d := range docs {
		entities := d.Entities
		if len(entities) == 0 {
			entities = extractMemoryEntities(d.Text)
		}
		bm := bm25Memory(docTerms[i], qterms, df, len(docs), avgdl)
		sem := cosineMemory(q.Embedding, d.Embedding)
		ent := entityOverlap(q.Entities, entities)
		tmp := temporalMemory(q.Now, d.CreatedAt)
		relevant := len(qterms) == 0 || bm > 0 || sem > 0 || ent > 0
		if !relevant {
			continue
		}
		score := w.BM25*bm + w.Semantic*sem + w.Entity*ent + w.Temporal*tmp
		out = append(out, MemoryScore{Memory: d, Score: score, BM25: bm, Semantic: sem, Entity: ent, Temporal: tmp})
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Score == out[j].Score {
			if out[i].Memory.CreatedAt.Equal(out[j].Memory.CreatedAt) {
				return out[i].Memory.ID < out[j].Memory.ID
			}
			return out[i].Memory.CreatedAt.After(out[j].Memory.CreatedAt)
		}
		return out[i].Score > out[j].Score
	})
	if len(out) > q.TopK {
		out = out[:q.TopK]
	}
	return out
}

func BuildMemoryQuery(ctx context.Context, p LocalEmbeddingProvider, text string, topK int) MemoryQuery {
	q := MemoryQuery{Text: text, Entities: extractMemoryEntities(text), Now: time.Now().UTC(), TopK: topK}
	if p != nil {
		if emb, err := p.Embed(ctx, text); err == nil && validateEmbedding(emb) {
			q.Embedding = emb
		}
	}
	return q
}

func (s *Store) SaveMemoryWithEmbedding(ctx context.Context, text string, p LocalEmbeddingProvider) (Memory, bool, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return Memory{}, false, errors.New("empty memory")
	}
	if len([]byte(text)) > 32*1024 {
		return Memory{}, false, errors.New("memory too large")
	}
	m := Memory{
		ID:         idFor(text),
		Text:       text,
		Persistent: true,
		CreatedAt:  time.Now().UTC(),
		Entities:   extractMemoryEntities(text),
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
	if err := appendJSONL(filepath.Join(s.dir, "memory.jsonl"), m); err != nil {
		return Memory{}, false, err
	}
	s.memories = append(s.memories, m)
	return m, true, nil
}

func (s *Store) SearchMemoryHybrid(ctx context.Context, q string, limit int, p LocalEmbeddingProvider) []Memory {
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	s.mu.Lock()
	cp := append([]Memory(nil), s.memories...)
	s.mu.Unlock()
	query := BuildMemoryQuery(ctx, p, q, limit)
	scored := RetrieveMemory(cp, query, DefaultRetrievalWeights())
	out := make([]Memory, len(scored))
	for i, x := range scored {
		out[i] = x.Memory
	}
	return out
}
