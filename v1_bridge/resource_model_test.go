package main

import (
	"context"
	"testing"
	"time"
)

func TestResourceSnapshotReportsRealRuntimeState(t *testing.T) {
	s:=CaptureResources()
	if s.At.IsZero()||s.SysBytes==0||s.Goroutines<1||s.GOMAXPROCS<1||s.OS==""||s.Arch==""{t.Fatalf("%+v",s)}
}

func TestMeasureOperationUsesObservedElapsedTime(t *testing.T) {
	m,err:=MeasureOperation(context.Background(),"sleep",func(context.Context)error{time.Sleep(5*time.Millisecond);return nil})
	if err!=nil{t.Fatal(err)}
	if m.DurationMS<=0||m.FinishedAt.Before(m.StartedAt){t.Fatalf("%+v",m)}
}

func TestModelProfileNeverRequiresPaidAPI(t *testing.T) {
	cases:=[]ModelAdapter{
		{},
		{OllamaURL:"http://127.0.0.1:11434",OllamaModel:"local"},
	}
	for _,m:=range cases {
		p:=m.RuntimeProfile()
		if p.PaidAPIRequired {t.Fatalf("%+v",p)}
		if p.Backend=="ollama-loopback"&&!p.LoopbackOnly{t.Fatalf("%+v",p)}
	}
}

func TestModelProfileRejectsRemoteOllamaAsConfigured(t *testing.T) {
	p:=(ModelAdapter{OllamaURL:"http://192.168.1.5:11434",OllamaModel:"x"}).RuntimeProfile()
	if p.Configured||p.LoopbackOnly{t.Fatalf("%+v",p)}
}
