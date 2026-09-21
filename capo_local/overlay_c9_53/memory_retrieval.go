package main

import (
	"math"
	"sort"
	"strings"
	"time"
	"unicode"
)

type MemoryDocument struct {
	ID string
	Text string
	Entities []string
	Timestamp time.Time
	Embedding []float32
}

type MemoryQuery struct {
	Text string
	Entities []string
	Now time.Time
	Embedding []float32
	TopK int
}

type MemoryScore struct {
	Document MemoryDocument
	Score float64
	BM25 float64
	Semantic float64
	Entity float64
	Temporal float64
}

type RetrievalWeights struct {
	BM25, Semantic, Entity, Temporal float64
}

func DefaultRetrievalWeights() RetrievalWeights {
	return RetrievalWeights{BM25: 0.40, Semantic: 0.35, Entity: 0.15, Temporal: 0.10}
}

func RetrieveMemory(docs []MemoryDocument, q MemoryQuery, w RetrievalWeights) []MemoryScore {
	if q.TopK <= 0 {
		q.TopK = 8
	}
	if q.Now.IsZero() {
		q.Now = time.Now().UTC()
	}
	w = sanitizeRetrievalWeights(w)
	qterms := uniqueTerms(tokenizeMemory(q.Text))
	df := map[string]int{}
	docTerms := make([][]string, len(docs))
	totalLen := 0
	for i, d := range docs {
		docTerms[i] = tokenizeMemory(d.Text)
		totalLen += len(docTerms[i])
		seen := map[string]bool{}
		for _, t := range docTerms[i] {
			if !seen[t] {
				df[t]++
				seen[t] = true
			}
		}
	}
	avgdl := 1.0
	if len(docs) > 0 {
		avgdl = math.Max(1, float64(totalLen)/float64(len(docs)))
	}
	out := make([]MemoryScore, 0, len(docs))
	for i, d := range docs {
		bm := bm25Score(docTerms[i], qterms, df, len(docs), avgdl)
		sem := cosineMemory(q.Embedding, d.Embedding)
		ent := entityOverlap(q.Entities, d.Entities)
		tmp := temporalMemory(q.Now, d.Timestamp)
		s := w.BM25*bm + w.Semantic*sem + w.Entity*ent + w.Temporal*tmp
		out = append(out, MemoryScore{Document: d, Score: s, BM25: bm, Semantic: sem, Entity: ent, Temporal: tmp})
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Score == out[j].Score {
			return out[i].Document.ID < out[j].Document.ID
		}
		return out[i].Score > out[j].Score
	})
	if len(out) > q.TopK {
		out = out[:q.TopK]
	}
	return out
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

func tokenizeMemory(s string) []string {
	return strings.FieldsFunc(strings.ToLower(s), func(r rune) bool {
		return !(unicode.IsLetter(r) || unicode.IsDigit(r))
	})
}

func uniqueTerms(in []string) []string {
	if len(in) < 2 {
		return in
	}
	out := make([]string, 0, len(in))
	seen := make(map[string]struct{}, len(in))
	for _, t := range in {
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

func bm25Score(doc, query []string, df map[string]int, n int, avgdl float64) float64 {
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
	if math.IsNaN(v) || math.IsInf(v, 0) || v <= 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func normalizeEntities(in []string) map[string]struct{} {
	out := make(map[string]struct{}, len(in))
	for _, x := range in {
		k := strings.ToLower(strings.TrimSpace(x))
		if k != "" {
			out[k] = struct{}{}
		}
	}
	return out
}

func entityOverlap(a, b []string) float64 {
	qa := normalizeEntities(a)
	if len(qa) == 0 {
		return 0
	}
	db := normalizeEntities(b)
	if len(db) == 0 {
		return 0
	}
	hits := 0
	for k := range qa {
		if _, ok := db[k]; ok {
			hits++
		}
	}
	return float64(hits) / float64(len(qa))
}

func temporalMemory(now, ts time.Time) float64 {
	if now.IsZero() || ts.IsZero() {
		return 0
	}
	if ts.After(now) {
		return 0
	}
	days := now.Sub(ts).Hours() / 24
	return math.Exp(-days / 30.0)
}
