package main

import (
    "testing"
    "time"
)

func TestHybridRetrievalPrefersRelevantRecentEntityMatch(t *testing.T){now:=time.Date(2026,9,22,12,0,0,0,time.UTC);docs:=[]Memory{{ID:"old",Text:"NEURA memory retrieval architecture",CreatedAt:now.Add(-120*24*time.Hour)},{ID:"best",Text:"NEURA memory retrieval fabric temporal recall",CreatedAt:now.Add(-2*time.Hour)},{ID:"noise",Text:"desktop theme colors",CreatedAt:now}};r:=rankMemories(docs,"NEURA memory retrieval",3,now);if len(r)!=3||r[0].Memory.ID!="best"{t.Fatalf("unexpected ranking: %#v",r)}}
func TestHybridRetrievalUnicode(t *testing.T){now:=time.Now().UTC();docs:=[]Memory{{ID:"a",Text:"東京 記憶 NEURA",CreatedAt:now},{ID:"b",Text:"pizza dinner",CreatedAt:now}};r:=rankMemories(docs,"東京 NEURA",1,now);if len(r)!=1||r[0].Memory.ID!="a"{t.Fatalf("unicode retrieval failed: %#v",r)}}
func TestHybridRetrievalFutureTimestampGetsNoBoost(t *testing.T){now:=time.Now().UTC();if got:=temporalScore(now,now.Add(time.Hour));got!=0{t.Fatalf("future boost=%v",got)}}
func TestHybridRetrievalDeterministicTieBreak(t *testing.T){now:=time.Now().UTC();docs:=[]Memory{{ID:"b",Text:"same memory",CreatedAt:now},{ID:"a",Text:"same memory",CreatedAt:now}};r:=rankMemories(docs,"same memory",2,now);if r[0].Memory.ID!="a"{t.Fatalf("non deterministic: %#v",r)}}
func TestHybridRetrievalLimitBounded(t *testing.T){now:=time.Now().UTC();docs:=[]Memory{{ID:"a",Text:"x memory",CreatedAt:now},{ID:"b",Text:"x memory",CreatedAt:now},{ID:"c",Text:"x memory",CreatedAt:now}};if got:=len(rankMemories(docs,"memory",1,now));got!=1{t.Fatalf("limit=%d",got)}}
func TestHybridSearchStoreIntegration(t *testing.T){c,s,_:=testCore(t);_ = c; if _,_,err:=s.SaveMemory("NEURA rollback candidate");err!=nil{t.Fatal(err)};if _,_,err:=s.SaveMemory("pizza dinner");err!=nil{t.Fatal(err)};got:=s.SearchMemoryHybrid("NEURA rollback",1);if len(got)!=1||got[0].Text!="NEURA rollback candidate"{t.Fatalf("integration failed: %#v",got)}}
