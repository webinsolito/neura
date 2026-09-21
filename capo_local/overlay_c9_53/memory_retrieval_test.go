package main

import (
	"math"
	"testing"
	"time"
)

func TestMemoryRetrievalHybridPrefersRelevantRecentEntityMatch(t *testing.T) {
	now := time.Date(2026, 9, 20, 8, 0, 0, 0, time.UTC)
	docs := []MemoryDocument{
		{ID: "old", Text: "NEURA memory retrieval architecture", Entities: []string{"NEURA"}, Timestamp: now.Add(-120 * 24 * time.Hour), Embedding: []float32{1, 0}},
		{ID: "best", Text: "NEURA memory retrieval fabric with temporal recall", Entities: []string{"NEURA", "memory"}, Timestamp: now.Add(-2 * time.Hour), Embedding: []float32{1, 0}},
		{ID: "noise", Text: "unrelated desktop theme", Entities: []string{"UI"}, Timestamp: now, Embedding: []float32{0, 1}},
	}
	r := RetrieveMemory(docs, MemoryQuery{Text: "NEURA memory retrieval", Entities: []string{"NEURA"}, Now: now, Embedding: []float32{1, 0}, TopK: 3}, DefaultRetrievalWeights())
	if len(r) != 3 || r[0].Document.ID != "best" {
		t.Fatalf("unexpected ranking: %#v", r)
	}
}

func TestMemoryRetrievalWorksWithoutEmbeddings(t *testing.T) {
	now := time.Now().UTC()
	docs := []MemoryDocument{{ID: "a", Text: "rollback stable candidate", Timestamp: now}, {ID: "b", Text: "pizza dinner", Timestamp: now}}
	r := RetrieveMemory(docs, MemoryQuery{Text: "stable rollback", Now: now, TopK: 1}, DefaultRetrievalWeights())
	if len(r) != 1 || r[0].Document.ID != "a" {
		t.Fatalf("lexical fallback failed: %#v", r)
	}
}

func TestMemoryRetrievalDeterministicTieBreak(t *testing.T) {
	now := time.Now().UTC()
	docs := []MemoryDocument{{ID: "b", Text: "same", Timestamp: now}, {ID: "a", Text: "same", Timestamp: now}}
	r := RetrieveMemory(docs, MemoryQuery{Text: "same", Now: now, TopK: 2}, DefaultRetrievalWeights())
	if r[0].Document.ID != "a" {
		t.Fatalf("tie break not deterministic: %#v", r)
	}
}

func TestMemoryRetrievalRejectsFutureTimestampBoost(t *testing.T) {
	now := time.Date(2026, 9, 21, 12, 0, 0, 0, time.UTC)
	docs := []MemoryDocument{
		{ID: "current", Text: "same", Timestamp: now.Add(-time.Hour)},
		{ID: "future", Text: "same", Timestamp: now.Add(time.Hour)},
	}
	weights := RetrievalWeights{Temporal: 1}
	r := RetrieveMemory(docs, MemoryQuery{Text: "same", Now: now, TopK: 2}, weights)
	if r[0].Document.ID != "current" {
		t.Fatalf("future-dated memory gained freshness advantage: %#v", r)
	}
	if r[1].Temporal != 0 {
		t.Fatalf("expected future temporal score 0, got %v", r[1].Temporal)
	}
}

func TestEntityOverlapNormalizesDuplicateAndBlankQueryEntities(t *testing.T) {
	got := entityOverlap([]string{"NEURA", " neura ", ""}, []string{"neura"})
	if got != 1 {
		t.Fatalf("expected normalized full overlap, got %v", got)
	}
}

func TestSanitizeRetrievalWeightsRejectsInvalidValues(t *testing.T) {
	got := sanitizeRetrievalWeights(RetrievalWeights{BM25: math.NaN(), Semantic: -1, Entity: math.Inf(1)})
	want := DefaultRetrievalWeights()
	if got != want {
		t.Fatalf("expected defaults for invalid zero-sum weights: got %#v want %#v", got, want)
	}
}

func TestCosineMemoryRejectsNonFiniteValues(t *testing.T) {
	if got := cosineMemory([]float32{1, float32(math.NaN())}, []float32{1, 1}); got != 0 {
		t.Fatalf("expected non-finite embedding to fail-soft to 0, got %v", got)
	}
}
