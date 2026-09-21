package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os/exec"
	"time"
)

type LocalEmbeddingProvider interface {
	Embed(context.Context, string) ([]float32, error)
}

type BitNetCommandEmbeddingProvider struct {
	Executable string
	Args       []string
	Timeout    time.Duration
}

const (
	maxLocalEmbeddingDimensions   = 65536
	maxLocalEmbeddingRequestBytes = 256 * 1024
	maxLocalEmbeddingOutputBytes  = 4 * 1024 * 1024
	maxLocalEmbeddingStderrBytes  = 16 * 1024
	maxLocalEmbeddingTimeout      = 60 * time.Second
)

type cappedBuffer struct {
	buf      bytes.Buffer
	max      int
	exceeded bool
}

func (w *cappedBuffer) Write(p []byte) (int, error) {
	if w.max <= 0 {
		w.exceeded = len(p) > 0
		return len(p), nil
	}
	remaining := w.max - w.buf.Len()
	if remaining <= 0 {
		if len(p) > 0 {
			w.exceeded = true
		}
		return len(p), nil
	}
	if len(p) > remaining {
		_, _ = w.buf.Write(p[:remaining])
		w.exceeded = true
		return len(p), nil
	}
	_, _ = w.buf.Write(p)
	return len(p), nil
}

func (w *cappedBuffer) Bytes() []byte {
	return w.buf.Bytes()
}

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

func effectiveEmbeddingTimeout(timeout time.Duration) time.Duration {
	if timeout <= 0 {
		return 8 * time.Second
	}
	if timeout > maxLocalEmbeddingTimeout {
		return maxLocalEmbeddingTimeout
	}
	return timeout
}

func (p BitNetCommandEmbeddingProvider) Embed(ctx context.Context, text string) ([]float32, error) {
	if p.Executable == "" {
		return nil, errors.New("bitnet embedding executable not configured")
	}
	if len([]byte(text)) > maxLocalEmbeddingRequestBytes {
		return nil, errors.New("local embedding request too large")
	}
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("local embedding context unavailable: %w", err)
	}

	timeout := effectiveEmbeddingTimeout(p.Timeout)
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	in, err := json.Marshal(map[string]string{"text": text})
	if err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(runCtx, p.Executable, p.Args...)
	cmd.WaitDelay = time.Second
	cmd.Stdin = bytes.NewReader(in)

	stdout := &cappedBuffer{max: maxLocalEmbeddingOutputBytes}
	stderr := &cappedBuffer{max: maxLocalEmbeddingStderrBytes}
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err = cmd.Run()
	if runCtx.Err() != nil {
		return nil, fmt.Errorf("local embedding execution canceled or timed out: %w", runCtx.Err())
	}
	if err != nil {
		return nil, fmt.Errorf("local embedding process failed: %w", err)
	}
	if stdout.exceeded {
		return nil, errors.New("local embedding output too large")
	}

	var resp map[string][]float32
	if err := json.Unmarshal(stdout.Bytes(), &resp); err != nil {
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
