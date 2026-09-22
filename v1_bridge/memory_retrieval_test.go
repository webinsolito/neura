package main

import (
	"context"
	"errors"
	"math"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type fakeEmbeddingProvider struct {
	fail bool
}

func (p fakeEmbeddingProvider) Embed(_ context.Context, text string) ([]float32, error) {
	if p.fail {
		return nil, errors.New("embedding unavailable")
	}
	if strings.Contains(strings.ToLower(text), "neura") {
		return []float32{1, 0}, nil
	}
	return []float32{0, 1}, nil
}

func TestC953HybridRankingPrefersRelevantRecentEntityMatch(t *testing.T) {
	now := time.Date(2026, 9, 22, 12, 0, 0, 0, time.UTC)
	docs := []Memory{
		{ID: "old", Text: "NEURA memory retrieval architecture", Entities: []string{"neura"}, CreatedAt: now.Add(-120 * 24 * time.Hour), Embedding: []float32{1, 0}},
		{ID: "best", Text: "NEURA memory retrieval fabric with temporal recall", Entities: []string{"neura"}, CreatedAt: now.Add(-2 * time.Hour), Embedding: []float32{1, 0}},
		{ID: "noise", Text: "unrelated desktop theme", Entities: []string{"ui"}, CreatedAt: now, Embedding: []float32{0, 1}},
	}
	got := RetrieveMemory(docs, MemoryQuery{
		Text: "NEURA memory retrieval", Entities: []string{"neura"}, Now: now,
		Embedding: []float32{1, 0}, TopK: 3,
	}, DefaultRetrievalWeights())
	if len(got) < 2 || got[0].Memory.ID != "best" {
		t.Fatalf("unexpected ranking: %#v", got)
	}
	for _, item := range got {
		if item.Memory.ID == "noise" {
			t.Fatalf("irrelevant recent memory admitted only by recency: %#v", got)
		}
	}
}

func TestC953LexicalFallbackWithoutEmbedding(t *testing.T) {
	now := time.Now().UTC()
	docs := []Memory{
		{ID: "a", Text: "rollback stable candidate", CreatedAt: now},
		{ID: "b", Text: "pizza dinner", CreatedAt: now},
	}
	got := RetrieveMemory(docs, MemoryQuery{Text: "stable rollback", Now: now, TopK: 1}, DefaultRetrievalWeights())
	if len(got) != 1 || got[0].Memory.ID != "a" {
		t.Fatalf("lexical fallback failed: %#v", got)
	}
}

func TestC953DeterministicTieBreak(t *testing.T) {
	now := time.Now().UTC()
	docs := []Memory{
		{ID: "b", Text: "same", CreatedAt: now},
		{ID: "a", Text: "same", CreatedAt: now},
	}
	got := RetrieveMemory(docs, MemoryQuery{Text: "same", Now: now, TopK: 2}, RetrievalWeights{BM25: 1})
	if len(got) != 2 || got[0].Memory.ID != "a" {
		t.Fatalf("tie break not deterministic: %#v", got)
	}
}

func TestC953FutureTimestampCannotBoost(t *testing.T) {
	now := time.Date(2026, 9, 22, 12, 0, 0, 0, time.UTC)
	docs := []Memory{
		{ID: "current", Text: "same", CreatedAt: now.Add(-time.Hour)},
		{ID: "future", Text: "same", CreatedAt: now.Add(time.Hour)},
	}
	got := RetrieveMemory(docs, MemoryQuery{Text: "same", Now: now, TopK: 2}, RetrievalWeights{BM25: 0.5, Temporal: 0.5})
	if len(got) != 2 || got[0].Memory.ID != "current" || got[1].Temporal != 0 {
		t.Fatalf("future timestamp boosted: %#v", got)
	}
}

func TestC953RepeatedTermsDoNotInflateBM25(t *testing.T) {
	now := time.Now().UTC()
	docs := []Memory{{ID: "a", Text: "memory retrieval", CreatedAt: now}, {ID: "b", Text: "other", CreatedAt: now}}
	one := RetrieveMemory(docs, MemoryQuery{Text: "memory", Now: now, TopK: 2}, RetrievalWeights{BM25: 1})
	many := RetrieveMemory(docs, MemoryQuery{Text: "memory memory memory memory", Now: now, TopK: 2}, RetrievalWeights{BM25: 1})
	if len(one) == 0 || len(many) == 0 || math.Abs(one[0].BM25-many[0].BM25) > 1e-12 {
		t.Fatalf("query repetition changed BM25: one=%#v many=%#v", one, many)
	}
}

func TestC953UnicodeTokenization(t *testing.T) {
	got := uniqueMemoryTerms(memoryTerms("NEURA 東京 память 123"))
	want := map[string]bool{"neura": true, "東京": true, "память": true, "123": true}
	for _, token := range got {
		delete(want, token)
	}
	if len(want) != 0 {
		t.Fatalf("missing unicode tokens: %#v got=%#v", want, got)
	}
}

func TestC953EntityExtractionAndNormalization(t *testing.T) {
	got := extractMemoryEntities("NEURA incontra Mario su PC-2026")
	set := normalizeEntitySet(got)
	for _, want := range []string{"neura", "mario", "pc-2026"} {
		if _, ok := set[want]; !ok {
			t.Fatalf("missing entity %q in %#v", want, got)
		}
	}
	if score := entityOverlap([]string{"NEURA", " neura ", ""}, []string{"neura"}); score != 1 {
		t.Fatalf("entity normalization failed: %v", score)
	}
}

func TestC953InvalidWeightsFailSafe(t *testing.T) {
	got := sanitizeRetrievalWeights(RetrievalWeights{BM25: math.NaN(), Semantic: -1, Entity: math.Inf(1)})
	if got != DefaultRetrievalWeights() {
		t.Fatalf("unexpected sanitized weights: %#v", got)
	}
	if cosineMemory([]float32{1, float32(math.NaN())}, []float32{1, 1}) != 0 {
		t.Fatal("non-finite embedding accepted")
	}
}

func TestC953EmbeddingPersistenceAndHybridSearch(t *testing.T) {
	root := t.TempDir()
	s, err := NewStore(filepath.Join(root, "data"))
	if err != nil {
		t.Fatal(err)
	}
	p := fakeEmbeddingProvider{}
	if _, _, err := s.SaveMemoryWithEmbedding(context.Background(), "NEURA progetto memoria", p); err != nil {
		t.Fatal(err)
	}
	if _, _, err := s.SaveMemoryWithEmbedding(context.Background(), "cena pizza", p); err != nil {
		t.Fatal(err)
	}
	reloaded, err := NewStore(s.dir)
	if err != nil {
		t.Fatal(err)
	}
	got := reloaded.SearchMemoryHybrid(context.Background(), "NEURA", 1, p)
	if len(got) != 1 || !strings.Contains(got[0].Text, "NEURA") || len(got[0].Embedding) == 0 {
		t.Fatalf("embedding persistence/search failed: %#v", got)
	}
}

func TestC953EmbeddingFailureFallsBackCleanly(t *testing.T) {
	root := t.TempDir()
	s, err := NewStore(filepath.Join(root, "data"))
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := s.SaveMemoryWithEmbedding(context.Background(), "rollback candidate stabile", fakeEmbeddingProvider{fail: true}); err != nil {
		t.Fatal(err)
	}
	got := s.SearchMemoryHybrid(context.Background(), "rollback stabile", 10, fakeEmbeddingProvider{fail: true})
	if len(got) != 1 {
		t.Fatalf("fallback search failed: %#v", got)
	}
}

func TestC953CommandEmbeddingProviderRejectsRelativeExecutable(t *testing.T) {
	p := CommandEmbeddingProvider{Executable: "relative-embedder"}
	if p.Available() {
		t.Fatal("relative embedding executable accepted")
	}
	if _, err := p.Embed(context.Background(), "test"); err == nil {
		t.Fatal("relative embedding executable executed")
	}
}
