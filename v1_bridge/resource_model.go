package main

import (
	"context"
	"errors"
	"runtime"
	"strings"
	"time"
)

type ResourceSnapshot struct {
	At             time.Time `json:"at"`
	HeapAllocBytes uint64    `json:"heap_alloc_bytes"`
	HeapSysBytes   uint64    `json:"heap_sys_bytes"`
	SysBytes       uint64    `json:"go_sys_bytes"`
	NumGC          uint32    `json:"num_gc"`
	Goroutines     int       `json:"goroutines"`
	GOMAXPROCS     int       `json:"gomaxprocs"`
	OS             string    `json:"os"`
	Arch           string    `json:"arch"`
}

type ModelRuntimeProfile struct {
	Backend          string `json:"backend"`
	Configured       bool   `json:"configured"`
	LoopbackOnly     bool   `json:"loopback_only"`
	PaidAPIRequired  bool   `json:"paid_api_required"`
	WarmStateKnown   bool   `json:"warm_state_known"`
	Note             string `json:"note,omitempty"`
}

type OperationMeasurement struct {
	Name       string           `json:"name"`
	StartedAt  time.Time        `json:"started_at"`
	FinishedAt time.Time        `json:"finished_at"`
	DurationMS float64          `json:"duration_ms"`
	Before     ResourceSnapshot `json:"before"`
	After      ResourceSnapshot `json:"after"`
}

func CaptureResources() ResourceSnapshot {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	return ResourceSnapshot{
		At:time.Now().UTC(),HeapAllocBytes:ms.HeapAlloc,HeapSysBytes:ms.HeapSys,SysBytes:ms.Sys,
		NumGC:ms.NumGC,Goroutines:runtime.NumGoroutine(),GOMAXPROCS:runtime.GOMAXPROCS(0),OS:runtime.GOOS,Arch:runtime.GOARCH,
	}
}

func MeasureOperation(ctx context.Context,name string,fn func(context.Context) error)(OperationMeasurement,error){
	name=strings.TrimSpace(name);if name==""{return OperationMeasurement{},errors.New("measurement name required")}
	if fn==nil{return OperationMeasurement{},errors.New("measurement function required")}
	m:=OperationMeasurement{Name:name,StartedAt:time.Now().UTC(),Before:CaptureResources()}
	err:=fn(ctx)
	m.FinishedAt=time.Now().UTC();m.After=CaptureResources();m.DurationMS=float64(m.FinishedAt.Sub(m.StartedAt).Nanoseconds())/1e6
	return m,err
}

func (m ModelAdapter) RuntimeProfile() ModelRuntimeProfile {
	if m.OllamaModel!="" {
		ok:=loopbackHTTP(m.OllamaURL)
		return ModelRuntimeProfile{Backend:"ollama-loopback",Configured:ok,LoopbackOnly:ok,PaidAPIRequired:false,WarmStateKnown:false,Note:"warm residency requires measurement against the locally configured model"}
	}
	if m.Executable!="" {
		ok:=m.Available()
		return ModelRuntimeProfile{Backend:"local-executable",Configured:ok,LoopbackOnly:true,PaidAPIRequired:false,WarmStateKnown:false,Note:"process residency is measured at runtime; no remote provider is required"}
	}
	return ModelRuntimeProfile{Backend:"none",Configured:false,LoopbackOnly:true,PaidAPIRequired:false,WarmStateKnown:false}
}
