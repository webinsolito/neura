package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"math"
	"os/exec"
	"time"
)

type LocalEmbeddingProvider interface {
	Embed(context.Context, string) ([]float32, error)
}

type BitNetCommandEmbeddingProvider struct {
	Executable string
	Args []string
	Timeout time.Duration
}

const maxLocalEmbeddingDimensions = 65536

func validateLocalEmbedding(e []float32) bool {
	if len(e) == 0 || len(e) > maxLocalEmbeddingDimensions {
		return false
	}
	for _, v := range e {
		f := float64(v)
		if math.IsNaN(f) || math.IsInf(f, 0) {
			return false
		}
	}
	return true
}

func (p BitNetCommandEmbeddingProvider) Embed(ctx context.Context, text string) ([]float32, error) {
	if p.Executable == "" {
		return nil, errors.New("bitnet embedding executable not configured")
	}
	timeout := p.Timeout
	if timeout <= 0 {
		timeout = 8 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	in, _ := json.Marshal(map[string]string{"text": text})
	cmd := exec.CommandContext(ctx, p.Executable, p.Args...)
	cmd.Stdin = bytes.NewReader(in)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	var resp map[string][]float32
	if err := json.Unmarshal(out.Bytes(), &resp); err != nil {
		return nil, err
	}
	e := resp["embedding"]
	if !validateLocalEmbedding(e) {
		return nil, errors.New("invalid local embedding")
	}
	return e, nil
}

func BuildMemoryQuery(ctx context.Context, p LocalEmbeddingProvider, text string, entities []string, now time.Time, topK int) MemoryQuery {
	q := MemoryQuery{Text: text, Entities: entities, Now: now, TopK: topK}
	if p != nil {
		if e, err := p.Embed(ctx, text); err == nil && validateLocalEmbedding(e) {
			q.Embedding = e
		}
	}
	return q
}
