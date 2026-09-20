package main

import (
	"testing"
	"time"
)

func TestMemoryRetrievalHybridPrefersRelevantRecentEntityMatch(t *testing.T) {
	now:=time.Date(2026,9,20,8,0,0,0,time.UTC)
	docs:=[]MemoryDocument{
		{ID:"old",Text:"NEURA memory retrieval architecture",Entities:[]string{"NEURA"},Timestamp:now.Add(-120*24*time.Hour),Embedding:[]float32{1,0}},
		{ID:"best",Text:"NEURA memory retrieval fabric with temporal recall",Entities:[]string{"NEURA","memory"},Timestamp:now.Add(-2*time.Hour),Embedding:[]float32{1,0}},
		{ID:"noise",Text:"unrelated desktop theme",Entities:[]string{"UI"},Timestamp:now,Embedding:[]float32{0,1}},
	}
	r:=RetrieveMemory(docs,MemoryQuery{Text:"NEURA memory retrieval",Entities:[]string{"NEURA"},Now:now,Embedding:[]float32{1,0},TopK:3},DefaultRetrievalWeights())
	if len(r)!=3 || r[0].Document.ID!="best" { t.Fatalf("unexpected ranking: %#v",r) }
}

func TestMemoryRetrievalWorksWithoutEmbeddings(t *testing.T) {
	now:=time.Now().UTC()
	docs:=[]MemoryDocument{{ID:"a",Text:"rollback stable candidate",Timestamp:now},{ID:"b",Text:"pizza dinner",Timestamp:now}}
	r:=RetrieveMemory(docs,MemoryQuery{Text:"stable rollback",Now:now,TopK:1},DefaultRetrievalWeights())
	if len(r)!=1 || r[0].Document.ID!="a" { t.Fatalf("lexical fallback failed: %#v",r) }
}

func TestMemoryRetrievalDeterministicTieBreak(t *testing.T) {
	now:=time.Now().UTC()
	docs:=[]MemoryDocument{{ID:"b",Text:"same",Timestamp:now},{ID:"a",Text:"same",Timestamp:now}}
	r:=RetrieveMemory(docs,MemoryQuery{Text:"same",Now:now,TopK:2},DefaultRetrievalWeights())
	if r[0].Document.ID!="a" { t.Fatalf("tie break not deterministic: %#v",r) }
}
