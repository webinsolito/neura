package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSafeWorkspaceMutationLifecycle(t *testing.T) {
	c, _, tools := testCore(t)
	mkdir := c.ExecutePlan(context.Background(), "mkdir", Plan{Steps:[]PlanStep{{Tool:"fs.mkdir.workspace",Input:map[string]string{"path":"docs"}}}})
	if mkdir.Status!="ok" || !mkdir.Results[0].Verified { t.Fatalf("%+v",mkdir) }
	if err:=os.WriteFile(filepath.Join(tools.workspace,"docs","a.txt"),[]byte("hello"),0o600);err!=nil{t.Fatal(err)}

	move := c.ExecutePlan(context.Background(), "move", Plan{Steps:[]PlanStep{{Tool:"fs.move.workspace",Input:map[string]string{"from":"docs/a.txt","to":"docs/b.txt"}}}})
	if move.Status!="ok" || !move.Results[0].Verified { t.Fatalf("%+v",move) }
	if _,err:=os.Stat(filepath.Join(tools.workspace,"docs","a.txt"));!os.IsNotExist(err){t.Fatal("source still exists")}
	if b,err:=os.ReadFile(filepath.Join(tools.workspace,"docs","b.txt"));err!=nil||string(b)!="hello"{t.Fatalf("%q %v",b,err)}

	del := c.ExecutePlan(context.Background(), "delete", Plan{Steps:[]PlanStep{{Tool:"fs.delete.workspace",Input:map[string]string{"path":"docs/b.txt"}}}})
	if del.Status!="ok" || !del.Results[0].Verified { t.Fatalf("%+v",del) }
	rec,ok:=del.Results[0].Data.(DeleteReceipt);if !ok{t.Fatalf("unexpected receipt %#v",del.Results[0].Data)}
	if _,err:=os.Stat(rec.Original);!os.IsNotExist(err){t.Fatal("original survived quarantine")}
	if _,err:=os.Stat(rec.Quarantine);err!=nil{t.Fatal("quarantine missing")}

	restore := c.ExecutePlan(context.Background(), "restore", Plan{Steps:[]PlanStep{{Tool:"fs.restore.workspace",Input:map[string]string{"quarantine":rec.Quarantine,"original":rec.Original}}}})
	if restore.Status!="ok" || !restore.Results[0].Verified { t.Fatalf("%+v",restore) }
	if b,err:=os.ReadFile(rec.Original);err!=nil||string(b)!="hello"{t.Fatalf("%q %v",b,err)}
}

func TestSafeWorkspaceToolsBlockPathEscapeAndInternalState(t *testing.T) {
	c,_,_:=testCore(t)
	for _,step:=range []PlanStep{
		{Tool:"fs.mkdir.workspace",Input:map[string]string{"path":"../../escape"}},
		{Tool:"fs.delete.workspace",Input:map[string]string{"path":".neura/jobs.json"}},
		{Tool:"fs.move.workspace",Input:map[string]string{"from":".neura/jobs.json","to":"x"}},
	}{
		r:=c.ExecutePlan(context.Background(),"blocked",Plan{Steps:[]PlanStep{step}})
		if r.Status=="ok"{t.Fatalf("unsafe mutation accepted: %+v",step)}
	}
}

func TestSafeWorkspaceWrongCapabilityDenied(t *testing.T) {
	_,_,tools:=testCore(t)
	p,err:=tools.gate.Issue("test","delete",[]string{capFSRead},time.Minute);if err!=nil{t.Fatal(err)}
	r:=tools.Run(context.Background(),"fs.delete.workspace",map[string]string{"path":"x"},p,"delete")
	if r.Error==""{t.Fatal("wrong capability accepted")}
}

func TestMkdirRollbackOnlyRemovesCreatedEmptyDirectory(t *testing.T) {
	_,_,tools:=testCore(t)
	rec,err:=tools.mkdirWorkspace("tempdir");if err!=nil{t.Fatal(err)}
	if err:=tools.rollbackMkdir(rec);err!=nil{t.Fatal(err)}
	if _,err:=os.Stat(rec.Path);!os.IsNotExist(err){t.Fatal("rollback did not remove directory")}
}
