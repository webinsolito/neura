package main

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeEmbeddingProvider struct { v []float32; err error }
func (f fakeEmbeddingProvider) Embed(context.Context,string)([]float32,error){ return f.v,f.err }

func TestBuildMemoryQueryEmbeddingHealthy(t *testing.T){
	q:=BuildMemoryQuery(context.Background(),fakeEmbeddingProvider{v:[]float32{1,2}},"x",nil,time.Time{},3)
	if len(q.Embedding)!=2 { t.Fatalf("embedding not attached") }
}
func TestBuildMemoryQueryEmbeddingFailureFallsBack(t *testing.T){
	q:=BuildMemoryQuery(context.Background(),fakeEmbeddingProvider{err:errors.New("offline")},"x",nil,time.Time{},3)
	if len(q.Embedding)!=0 { t.Fatalf("expected lexical fallback") }
}
