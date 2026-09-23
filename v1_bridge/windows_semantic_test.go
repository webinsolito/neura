package main

import (
	"context"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestWindowsSemanticToolHiddenOffWindows(t *testing.T) {
	if runtime.GOOS=="windows" { t.Skip("non-Windows contract") }
	_,_,tools:=testCore(t)
	for _,name:=range tools.List(){ if name=="windows.app.launch" { t.Fatal("Windows mutating tool exposed off Windows") } }
}

func TestWindowsSemanticWrongCapabilityDenied(t *testing.T) {
	_,_,tools:=testCore(t)
	p,err:=tools.gate.Issue("test","launch",[]string{capWindowsObserve},time.Minute);if err!=nil{t.Fatal(err)}
	r:=tools.Run(context.Background(),"windows.app.launch",map[string]string{"app":"notepad"},p,"launch")
	if !strings.Contains(r.Error,"capability denied"){t.Fatalf("%+v",r)}
}

func TestWindowsSemanticLaunchNotepadNative(t *testing.T) {
	if runtime.GOOS!="windows" { t.Skip("requires Windows native") }
	_,_,tools:=testCore(t)
	p,err:=tools.gate.Issue("test","launch notepad",[]string{capWindowsLaunch},time.Minute);if err!=nil{t.Fatal(err)}
	r:=tools.Run(context.Background(),"windows.app.launch",map[string]string{"app":"notepad"},p,"launch notepad")
	if r.Error!="" || !r.Verified { t.Fatalf("%+v",r) }
	rec,ok:=r.Data.(WindowsLaunchReceipt);if !ok || rec.PID<=0 || rec.Executable=="" { t.Fatalf("%#v",r.Data) }
	proc,err:=os.FindProcess(rec.PID);if err==nil{_ = proc.Kill()}
}

func TestWindowsSemanticRejectsUnallowlistedAppNative(t *testing.T) {
	if runtime.GOOS!="windows" { t.Skip("requires Windows native") }
	if _,err:=resolveAllowedWindowsApp("powershell");err==nil{t.Fatal("unallowlisted app accepted")}
	if _,err:=resolveAllowedWindowsApp("cmd");err==nil{t.Fatal("shell app accepted")}
}
