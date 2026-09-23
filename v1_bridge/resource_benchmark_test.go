package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkResourceSnapshot(b *testing.B) {
	for i:=0;i<b.N;i++ { _ = CaptureResources() }
}

func BenchmarkCoreStatusExecution(b *testing.B) {
	root:=b.TempDir()
	store,err:=NewStore(filepath.Join(root,"data"));if err!=nil{b.Fatal(err)}
	ws:=filepath.Join(root,"ws");if err:=os.MkdirAll(ws,0o700);err!=nil{b.Fatal(err)}
	tools,err:=NewToolRegistry(ws);if err!=nil{b.Fatal(err)}
	c:=&Core{store:store,tools:tools}
	ctx:=context.Background()
	b.ResetTimer()
	for i:=0;i<b.N;i++ { _ = c.Execute(ctx,"stato") }
}
