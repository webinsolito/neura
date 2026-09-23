package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestTaskRunnerMultiStepPersistsProgress(t *testing.T) {
	c,_,tools:=testCore(t)
	runner,err:=NewTaskRunner(c,tools.workspace);if err!=nil{t.Fatal(err)}
	plan:=Plan{Summary:"create and verify",Steps:[]PlanStep{
		{Tool:"fs.write.workspace",Input:map[string]string{"path":"task.txt","content":"hello"}},
		{Tool:"fs.read",Input:map[string]string{"path":"task.txt"}},
	}}
	task,err:=runner.store.Create("t1","write task file",plan,tools.List());if err!=nil{t.Fatal(err)}
	task,res,err:=runner.RunNext(context.Background(),"t1",task.Generation);if err!=nil{t.Fatal(err)}
	if task.State!=TaskPending||task.NextStep!=1||res.Status!="ok"{t.Fatalf("task=%+v res=%+v",task,res)}
	reopened,err:=OpenTaskStore(tools.workspace);if err!=nil{t.Fatal(err)}
	persisted,_:=reopened.Get("t1");if persisted.NextStep!=1{t.Fatalf("%+v",persisted)}
	runner.store=reopened
	task,res,err=runner.RunNext(context.Background(),"t1",persisted.Generation);if err!=nil{t.Fatal(err)}
	if task.State!=TaskSucceeded||task.NextStep!=2||res.Status!="ok"{t.Fatalf("task=%+v res=%+v",task,res)}
}

func TestTaskRunnerStopsOnFailedStep(t *testing.T) {
	c,_,tools:=testCore(t);runner,_:=NewTaskRunner(c,tools.workspace)
	plan:=Plan{Steps:[]PlanStep{
		{Tool:"fs.read",Input:map[string]string{"path":"missing.txt"}},
		{Tool:"fs.write.workspace",Input:map[string]string{"path":"must-not-run.txt","content":"bad"}},
	}}
	task,_:=runner.store.Create("t2","fail safely",plan,tools.List())
	task,res,err:=runner.RunNext(context.Background(),"t2",task.Generation);if err!=nil{t.Fatal(err)}
	if task.State!=TaskBlocked||res.Status!="partial"{t.Fatalf("task=%+v res=%+v",task,res)}
	if _,err:=os.Stat(filepath.Join(tools.workspace,"must-not-run.txt"));!os.IsNotExist(err){t.Fatal("later step ran after failure")}
}

func TestTaskRunnerInterruptedMutationNeedsManualReview(t *testing.T) {
	c,_,tools:=testCore(t);runner,_:=NewTaskRunner(c,tools.workspace)
	task,_:=runner.store.Create("t3","mutation",Plan{Steps:[]PlanStep{{Tool:"fs.write.workspace",Input:map[string]string{"path":"x.txt","content":"x"}}}},tools.List())
	_,_,err:=runner.store.beginStep("t3",task.Generation);if err!=nil{t.Fatal(err)}
	reopened,_:=OpenTaskStore(tools.workspace)
	if err:=reopened.RecoverInterrupted();err!=nil{t.Fatal(err)}
	got,_:=reopened.Get("t3")
	if got.State!=TaskManualReview||got.Generation!=task.Generation+1{t.Fatalf("%+v",got)}
	if _,err:=reopened.ResolveManualReview("t3",task.Generation,false);err==nil{t.Fatal("stale generation accepted")}
	got,err=reopened.ResolveManualReview("t3",got.Generation,false);if err!=nil{t.Fatal(err)}
	if got.State!=TaskPending||got.NextStep!=0{t.Fatalf("%+v",got)}
}

func TestTaskRunnerInterruptedReadOnlyStepSafeToRetry(t *testing.T) {
	c,_,tools:=testCore(t);_ = os.WriteFile(filepath.Join(tools.workspace,"a.txt"),[]byte("ok"),0o600)
	runner,_:=NewTaskRunner(c,tools.workspace)
	task,_:=runner.store.Create("t4","read",Plan{Steps:[]PlanStep{{Tool:"fs.read",Input:map[string]string{"path":"a.txt"}}}},tools.List())
	_,_,err:=runner.store.beginStep("t4",task.Generation);if err!=nil{t.Fatal(err)}
	reopened,_:=OpenTaskStore(tools.workspace);_ = reopened.RecoverInterrupted()
	got,_:=reopened.Get("t4")
	if got.State!=TaskPending||got.Generation!=task.Generation+1||got.NextStep!=0{t.Fatalf("%+v",got)}
}

func TestTaskStoreRejectsSecretBearingPersistentPlan(t *testing.T) {
	c,_,tools:=testCore(t);runner,_:=NewTaskRunner(c,tools.workspace)
	_,err:=runner.store.Create("t5","secret",Plan{Steps:[]PlanStep{{Tool:"fs.write.workspace",Input:map[string]string{"path":"a.txt","content":"api_key=abcdef1234567890"}}}},tools.List())
	if err==nil{t.Fatal("secret-bearing plan persisted")}
}
