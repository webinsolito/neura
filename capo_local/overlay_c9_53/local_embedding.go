package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os/exec"
	"time"
)

type LocalEmbeddingProvider interface {
	Embed(context.Context, string) ([]float32, error)
}

// BitNetCommandEmbeddingProvider is an optional local-only adapter.
// NEURA never downloads a model and never calls a remote endpoint here.
// The configured executable must already exist locally and accept JSON on stdin.
type BitNetCommandEmbeddingProvider struct {
	Executable string
	Args       []string
	Timeout    time.Duration
}

func (p BitNetCommandEmbeddingProvider) Embed(ctx context.Context, text string) ([]float32, error) {
	if p.Executable == "" { return nil, errors.New("bitnet embedding executable not configured") }
	timeout:=p.Timeout; if timeout<=0 { timeout=8*time.Second }
	ctx,cancel:=context.WithTimeout(ctx,timeout); defer cancel()
	in,_:=json.Marshal(map[string]string{"text":text})
	cmd:=exec.CommandContext(ctx,p.Executable,p.Args...)
	cmd.Stdin=bytes.NewReader(in)
	var out bytes.Buffer
	cmd.Stdout=&out
	if err:=cmd.Run(); err!=nil { return nil,err }
	var resp struct{ Embedding []float32 `json:"embedding"` }
	if err:=json.Unmarshal(out.Bytes(),&resp); err!=nil { return nil,err }
	if len(resp.Embedding)==0 { return nil,errors.New("empty local embedding") }
	return resp.Embedding,nil
}

// BuildMemoryQuery adds semantic retrieval when the local provider is healthy.
// Failure is deliberately non-fatal: BM25/entity/temporal retrieval stays active.
func BuildMemoryQuery(ctx context.Context, p LocalEmbeddingProvider, text string, entities []string, now time.Time, topK int) MemoryQuery {
	q:=MemoryQuery{Text:text,Entities:entities,Now:now,TopK:topK}
	if p!=nil {
		if e,err:=p.Embed(ctx,text); err==nil { q.Embedding=e }
	}
	return q
}
