package main

import (
    "math"
    "sort"
    "strings"
    "time"
    "unicode"
)

type retrievalScore struct { Memory Memory; Score, Lexical, Semantic, Entity, Temporal float64 }

func memoryTerms(s string) []string { return strings.FieldsFunc(strings.ToLower(s), func(r rune) bool { return !(unicode.IsLetter(r) || unicode.IsDigit(r)) }) }
func termSet(s string) map[string]struct{} { out:=map[string]struct{}{}; for _,t:=range memoryTerms(s){if len([]rune(t))>1{out[t]=struct{}{}}}; return out }
func lexicalScore(q,d string) float64 { qs,ds:=termSet(q),termSet(d);if len(qs)==0{return 0};hit:=0;for t:=range qs{if _,ok:=ds[t];ok{hit++}};return float64(hit)/float64(len(qs)) }

// localSemanticVector is deterministic, offline and dependency-free. It is a
// bounded hashing projection used as a fuzzy semantic signal, never as truth.
func localSemanticVector(s string) [64]float64 { var v [64]float64;rs:=[]rune(strings.ToLower(s));for i:=0;i<len(rs);i++{end:=i+3;if end>len(rs){end=len(rs)};h:=uint64(1469598103934665603);for _,r:=range rs[i:end]{h^=uint64(r);h*=1099511628211};v[h%64]++};return v }
func semanticScore(a,b string) float64 { x,y:=localSemanticVector(a),localSemanticVector(b);var dot,xx,yy float64;for i:=range x{dot+=x[i]*y[i];xx+=x[i]*x[i];yy+=y[i]*y[i]};if xx==0||yy==0{return 0};z:=dot/(math.Sqrt(xx)*math.Sqrt(yy));if z<0{return 0};if z>1{return 1};return z }
func entityTerms(s string) map[string]struct{} { out:=map[string]struct{}{};for _,f:=range strings.Fields(s){r:=[]rune(strings.Trim(f,".,:;!?()[]{}\"'"));if len(r)>=2&&unicode.IsUpper(r[0]){out[strings.ToLower(string(r))]=struct{}{}}};return out }
func entityScore(q,d string) float64 { a,b:=entityTerms(q),entityTerms(d);if len(a)==0{return 0};hit:=0;for k:=range a{if _,ok:=b[k];ok{hit++}};return float64(hit)/float64(len(a)) }
func temporalScore(now,created time.Time) float64 { if created.IsZero()||created.After(now){return 0};return math.Exp(-(now.Sub(created).Hours()/24)/30) }
func rankMemories(memories []Memory,q string,limit int,now time.Time) []retrievalScore { if limit<=0||limit>50{limit=10};if now.IsZero(){now=time.Now().UTC()};out:=make([]retrievalScore,0,len(memories));for _,m:=range memories{l:=lexicalScore(q,m.Text);s:=semanticScore(q,m.Text);e:=entityScore(q,m.Text);t:=temporalScore(now,m.CreatedAt);score:=.45*l+.30*s+.15*e+.10*t;if strings.TrimSpace(q)==""{score=.10*t};if score>0{out=append(out,retrievalScore{m,score,l,s,e,t})}};sort.SliceStable(out,func(i,j int)bool{if math.Abs(out[i].Score-out[j].Score)<1e-12{return out[i].Memory.ID<out[j].Memory.ID};return out[i].Score>out[j].Score});if len(out)>limit{out=out[:limit]};return out }
func (s *Store)SearchMemoryHybrid(q string,limit int) []Memory { s.mu.Lock();cp:=append([]Memory(nil),s.memories...);s.mu.Unlock();ranked:=rankMemories(cp,q,limit,time.Now().UTC());out:=make([]Memory,len(ranked));for i,r:=range ranked{out[i]=r.Memory};return out }
