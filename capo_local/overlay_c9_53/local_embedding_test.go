package main

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"strings"
	"testing"
	"time"
)

type fakeEmbeddingProvider struct {
	v   []float32
	err error
}

func (f fakeEmbeddingProvider) Embed(context.Context, string) ([]float32, error) {
	return f.v, f.err
}

func TestBuildMemoryQueryEmbeddingHealthy(t *testing.T) {
	q := BuildMemoryQuery(context.Background(), fakeEmbeddingProvider{v: []float32{1, 2}}, "x", nil, time.Time{}, 3)
	if len(q.Embedding) != 2 {
		t.Fatalf("embedding not attached")
	}
}

func TestBuildMemoryQueryEmbeddingFailureFallsBack(t *testing.T) {
	q := BuildMemoryQuery(context.Background(), fakeEmbeddingProvider{err: errors.New("offline")}, "x", nil, time.Time{}, 3)
	if len(q.Embedding) != 0 {
		t.Fatalf("expected lexical fallback")
	}
}

func TestBuildMemoryQueryRejectsNonFiniteProviderEmbedding(t *testing.T) {
	q := BuildMemoryQuery(context.Background(), fakeEmbeddingProvider{v: []float32{1, float32(math.NaN())}}, "x", nil, time.Time{}, 3)
	if len(q.Embedding) != 0 {
		t.Fatalf("poisoned embedding attached: %#v", q.Embedding)
	}
}

func TestBuildMemoryQueryRejectsOversizedProviderEmbedding(t *testing.T) {
	huge := make([]float32, maxLocalEmbeddingDimensions+1)
	q := BuildMemoryQuery(context.Background(), fakeEmbeddingProvider{v: huge}, "x", nil, time.Time{}, 3)
	if len(q.Embedding) != 0 {
		t.Fatalf("oversized embedding attached: %d", len(q.Embedding))
	}
}

func TestBitNetCommandEmbeddingProviderValidProcess(t *testing.T) {
	p := BitNetCommandEmbeddingProvider{
		Executable: os.Args[0],
		Args:       []string{"-test.run=TestEmbeddingHelperProcess", "--", "valid"},
		Timeout:    10 * time.Second,
	}
	e, err := p.Embed(context.Background(), "hello")
	if err != nil || len(e) != 2 || e[0] != 1 || e[1] != 2 {
		t.Fatalf("unexpected result: %v %v", e, err)
	}
}

func TestBitNetCommandEmbeddingProviderTimeoutFallsBack(t *testing.T) {
	p := BitNetCommandEmbeddingProvider{
		Executable: os.Args[0],
		Args:       []string{"-test.run=TestEmbeddingHelperProcess", "--", "sleep"},
		Timeout:    20 * time.Millisecond,
	}
	if _, err := p.Embed(context.Background(), "hello"); err == nil || !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("expected timeout, got %v", err)
	}
}

func TestBitNetCommandEmbeddingProviderRejectsOversizedStdout(t *testing.T) {
	p := BitNetCommandEmbeddingProvider{
		Executable: os.Args[0],
		Args:       []string{"-test.run=TestEmbeddingHelperProcess", "--", "huge"},
		Timeout:    10 * time.Second,
	}
	if _, err := p.Embed(context.Background(), "hello"); err == nil || !strings.Contains(err.Error(), "output too large") {
		t.Fatalf("expected output cap error, got %v", err)
	}
}

func TestBitNetCommandEmbeddingProviderRejectsOversizedInputBeforeExec(t *testing.T) {
	p := BitNetCommandEmbeddingProvider{
		Executable: "definitely-not-a-real-executable",
		Timeout:    time.Second,
	}
	if _, err := p.Embed(context.Background(), strings.Repeat("x", maxLocalEmbeddingRequestBytes+1)); err == nil || !strings.Contains(err.Error(), "request too large") {
		t.Fatalf("expected request cap error, got %v", err)
	}
}

func TestEffectiveEmbeddingTimeoutIsCapped(t *testing.T) {
	if got := effectiveEmbeddingTimeout(10 * maxLocalEmbeddingTimeout); got != maxLocalEmbeddingTimeout {
		t.Fatalf("timeout not capped: %v", got)
	}
}

func TestEmbeddingHelperProcess(t *testing.T) {
	idx := -1
	for i, a := range os.Args {
		if a == "--" {
			idx = i
			break
		}
	}
	if idx < 0 || idx+1 >= len(os.Args) {
		return
	}

	switch os.Args[idx+1] {
	case "valid":
		fmt.Print("{\"embedding\":[1,2]}")
	case "sleep":
		time.Sleep(250 * time.Millisecond)
		fmt.Print("{\"embedding\":[1,2]}")
	case "huge":
		fmt.Print(strings.Repeat("x", maxLocalEmbeddingOutputBytes+1024))
	}
	os.Exit(0)
}
