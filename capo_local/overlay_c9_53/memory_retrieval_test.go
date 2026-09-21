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
	docs := []MemoryDocument{{ID: "current", Text: "same", Timestamp: now.Add(-time.Hour)}, {ID: "future", Text: "same", Timestamp: now.Add(time.Hour)}}
	r := RetrieveMemory(docs, MemoryQuery{Text: "same", Now: now, TopK: 2}, RetrievalWeights{Temporal: 1})
	if r[0].Document.ID != "current" || r[1].Temporal != 0 {
		t.Fatalf("future boost not rejected: %#v", r)
	}
}

func TestRepeatedQueryTermsDoNotInflateBM25(t *testing.T) {
	now := time.Now().UTC()
	docs := []MemoryDocument{{ID: "a", Text: "memory retrieval"}, {ID: "b", Text: "other"}}
	one := RetrieveMemory(docs, MemoryQuery{Text: "memory", Now: now, TopK: 2}, RetrievalWeights{BM25: 1})
	repeated := RetrieveMemory(docs, MemoryQuery{Text: "memory memory memory memory", Now: now, TopK: 2}, RetrievalWeights{BM25: 1})
	if math.Abs(one[0].BM25-repeated[0].BM25) > 1e-12 {
		t.Fatalf("query repetition inflated BM25: one=%v repeated=%v", one[0].BM25, repeated[0].BM25)
	}
}

func TestUnicodeTokenizationPreservesNonLatinTerms(t *testing.T) {
	tokens := tokenizeMemory("NEURA 東京 память 123")
	want := map[string]bool{"neura": true, "東京": true, "память": true, "123": true}
	for _, token := range tokens {
		delete(want, token)
	}
	if len(want) != 0 {
		t.Fatalf("missing unicode tokens: %#v; got %#v", want, tokens)
	}
}

func TestEntityOverlapNormalizesDuplicateAndBlankQueryEntities(t *testing.T) {
	if got := entityOverlap([]string{"NEURA", " neura ", ""}, []string{"neura"}); got != 1 {
		t.Fatalf("expected normalized full overlap, got %v", got)
	}
}

func TestSanitizeRetrievalWeightsRejectsInvalidValues(t *testing.T) {
	got := sanitizeRetrievalWeights(RetrievalWeights{BM25: math.NaN(), Semantic: -1, Entity: math.Inf(1)})
	if got != DefaultRetrievalWeights() {
		t.Fatalf("unexpected sanitized weights: %#v", got)
	}
}

func TestCosineMemoryRejectsNonFiniteValues(t *testing.T) {
	if got := cosineMemory([]float32{1, float32(math.NaN())}, []float32{1, 1}); got != 0 {
		t.Fatalf("expected non-finite embedding to fail-soft to 0, got %v", got)
	}
}
