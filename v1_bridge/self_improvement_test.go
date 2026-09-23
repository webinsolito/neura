package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestImprovementStagesCandidateAndRollbackWithoutTouchingLive(t *testing.T){
	root:=t.TempDir();live:=filepath.Join(root,"core.txt");_ = os.WriteFile(live,[]byte("old"),0o600)
	m,err:=NewImprovementManager(root);if err!=nil{t.Fatal(err)}
	c,err:=m.Stage("c1","memory",[]ImprovementChange{{Path:"core.txt",Content:"new"}});if err!=nil{t.Fatal(err)}
	if !c.RollbackReady||c.State!="staged"||c.AutoPromotion{t.Fatalf("%+v",c)}
	b,_:=os.ReadFile(live);if string(b)!="old"{t.Fatalf("live mutated during staging: %q",b)}
	cb,_:=os.ReadFile(filepath.Join(root,".neura","improvements","c1","candidate","core.txt"));if string(cb)!="new"{t.Fatal("candidate content missing")}
	rb,_:=os.ReadFile(filepath.Join(root,".neura","improvements","c1","rollback","core.txt"));if string(rb)!="old"{t.Fatal("rollback content missing")}
}

func TestImprovementBlocksEscapeSecretsAndGitMetadata(t *testing.T){
	m,_:=NewImprovementManager(t.TempDir())
	for _,ch:=range []ImprovementChange{
		{Path:"../escape.txt",Content:"x"},
		{Path:".git/config",Content:"x"},
		{Path:"x.txt",Content:"api_key=abcdef1234567890"},
	}{
		if _,err:=m.Stage("x"+idFor(ch.Path)[:4],"security",[]ImprovementChange{ch});err==nil{t.Fatalf("unsafe change accepted: %+v",ch)}
	}
}

func TestImprovementPromotionGateRequiresTestsAndBenchmark(t *testing.T){
	root:=t.TempDir();_ = os.WriteFile(filepath.Join(root,"a.txt"),[]byte("a"),0o600);m,_:=NewImprovementManager(root)
	_,err:=m.Stage("c2","planner",[]ImprovementChange{{Path:"a.txt",Content:"b"}});if err!=nil{t.Fatal(err)}
	ok,_,err:=m.PromotionDecision("c2");if err!=nil||ok{t.Fatalf("ok=%v err=%v",ok,err)}
	c,err:=m.RecordEvaluation("c2",ImprovementEvaluation{TestsPassed:true,BenchmarkMeasured:true,BenchmarkNoWorse:true,Evidence:[]string{"linux-pass","windows-pass","bench-no-regression"}});if err!=nil{t.Fatal(err)}
	if !c.PromotionReady{t.Fatalf("%+v",c)}
	ok,msg,err:=m.PromotionDecision("c2");if err!=nil||!ok||msg==""{t.Fatalf("ok=%v msg=%q err=%v",ok,msg,err)}
}

func TestImprovementBenchmarkRegressionBlocksPromotion(t *testing.T){
	root:=t.TempDir();_ = os.WriteFile(filepath.Join(root,"a.txt"),[]byte("a"),0o600);m,_:=NewImprovementManager(root)
	_,_=m.Stage("c3","resources",[]ImprovementChange{{Path:"a.txt",Content:"b"}})
	c,_:=m.RecordEvaluation("c3",ImprovementEvaluation{TestsPassed:true,BenchmarkMeasured:true,BenchmarkNoWorse:false})
	if c.PromotionReady{t.Fatal("benchmark regression candidate became promotable")}
}
