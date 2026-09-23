package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRecoveryKeepsCommittedTargetAndRemovesStaleTemp(t *testing.T) {
	root:=t.TempDir()
	target:=filepath.Join(root,"a.txt")
	if err:=os.WriteFile(target,[]byte("old"),0o600);err!=nil{t.Fatal(err)}
	if err:=os.WriteFile(target+".neura.tmp",[]byte("new-partial"),0o600);err!=nil{t.Fatal(err)}
	rm,_:=NewRecoveryManager(root)
	rep,err:=rm.Recover();if err!=nil{t.Fatal(err)}
	if len(rep.WorkspaceActions)!=1 || rep.WorkspaceActions[0].Decision!="keep_target_remove_stale_temp" || !rep.WorkspaceActions[0].Verified {t.Fatalf("%+v",rep)}
	b,_:=os.ReadFile(target);if string(b)!="old"{t.Fatalf("target changed: %q",b)}
	if _,err:=os.Stat(target+".neura.tmp");!os.IsNotExist(err){t.Fatal("stale temp remains")}
}

func TestRecoveryRestoresBackupWhenTargetMissing(t *testing.T) {
	root:=t.TempDir();target:=filepath.Join(root,"a.txt")
	if err:=os.WriteFile(target+".neura.bak",[]byte("before"),0o600);err!=nil{t.Fatal(err)}
	if err:=os.WriteFile(target+".neura.tmp",[]byte("partial"),0o600);err!=nil{t.Fatal(err)}
	rm,_:=NewRecoveryManager(root)
	rep,err:=rm.Recover();if err!=nil{t.Fatal(err)}
	if rep.WorkspaceActions[0].Decision!="rollback_from_backup"{t.Fatalf("%+v",rep)}
	b,_:=os.ReadFile(target);if string(b)!="before"{t.Fatalf("%q",b)}
}

func TestRecoveryCancelsIncompleteNewFile(t *testing.T) {
	root:=t.TempDir();target:=filepath.Join(root,"new.txt")
	if err:=os.WriteFile(target+".neura.tmp",[]byte("partial"),0o600);err!=nil{t.Fatal(err)}
	rm,_:=NewRecoveryManager(root)
	rep,err:=rm.Recover();if err!=nil{t.Fatal(err)}
	if rep.WorkspaceActions[0].Decision!="cancel_incomplete_creation"{t.Fatalf("%+v",rep)}
	if _,err:=os.Stat(target);!os.IsNotExist(err){t.Fatal("incomplete target materialized")}
}

func TestRecoveryRecoversInterruptedDurableJob(t *testing.T) {
	root:=t.TempDir();s,err:=OpenJobStore(root);if err!=nil{t.Fatal(err)}
	_,_=s.Enqueue("j1","k1",7,3);_,_=s.Claim("j1",7)
	rm,_:=NewRecoveryManager(root)
	rep,err:=rm.Recover();if err!=nil{t.Fatal(err)}
	if !rep.JobsRecovered {t.Fatalf("%+v",rep)}
	reopened,_:=OpenJobStore(root);j,_:=reopened.Get("j1")
	if j.State!=JobPending || j.Generation!=8 {t.Fatalf("%+v",j)}
}

func TestRecoveryCorruptJobStateFailsClosed(t *testing.T) {
	root:=t.TempDir();dir:=filepath.Join(root,".neura");_ = os.MkdirAll(dir,0o700)
	_ = os.WriteFile(filepath.Join(dir,"jobs.json"),[]byte("{broken"),0o600)
	rm,_:=NewRecoveryManager(root)
	rep,err:=rm.Recover()
	if err==nil || !rep.ManualReview {t.Fatalf("rep=%+v err=%v",rep,err)}
}
