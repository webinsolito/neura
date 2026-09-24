package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestExecutePlanPrevalidatesBeforeAnySideEffect(t *testing.T) {
	c, _, tools := testCore(t)
	p := Plan{Summary:"must reject whole plan", Steps:[]PlanStep{{Tool:"fs.write.workspace", Input:map[string]string{"path":"should-not-exist.txt","content":"x"}},{Tool:"shell.exec", Input:map[string]string{"command":"bad"}}}}
	r := c.ExecutePlan(context.Background(), "test prevalidation", p)
	if r.Status != "blocked" { t.Fatalf("%+v", r) }
	if _, err := os.Stat(filepath.Join(tools.workspace, "should-not-exist.txt")); !os.IsNotExist(err) { t.Fatal("side effect occurred before whole-plan validation") }
	if got := DefaultEntityState.Snapshot().Mode; got != EntityError { t.Fatalf("entity mode=%s", got) }
}

func TestExecutePlanVerifiedWorkspaceMutation(t *testing.T) {
	c, s, tools := testCore(t)
	p := Plan{Summary:"write verified file", Steps:[]PlanStep{{Tool:"fs.write.workspace", Input:map[string]string{"path":"verified.txt","content":"hello"}},{Tool:"fs.read", Input:map[string]string{"path":"verified.txt"}}}}
	r := c.ExecutePlan(context.Background(), "write then verify", p)
	if r.Status != "ok" || len(r.Results) != 2 { t.Fatalf("%+v", r) }
	if !r.Results[0].Verified || !r.Results[1].Verified || r.Results[1].Data != "hello" { t.Fatalf("%+v", r.Results) }
	b, err := os.ReadFile(filepath.Join(tools.workspace, "verified.txt")); if err != nil || string(b)!="hello" { t.Fatalf("%q %v", b, err) }
	if _, n := s.Counts(); n != 2 { t.Fatalf("receipts=%d", n) }
	if got := DefaultEntityState.Snapshot().Mode; got != EntitySuccess { t.Fatalf("entity mode=%s", got) }
}

func TestExecutePlanStopsAfterBlockedStep(t *testing.T) {
	c, _, tools := testCore(t)
	p := Plan{Summary:"stop on failure", Steps:[]PlanStep{{Tool:"fs.read", Input:map[string]string{"path":"missing.txt"}},{Tool:"fs.write.workspace", Input:map[string]string{"path":"must-not-run.txt","content":"bad"}}}}
	r := c.ExecutePlan(context.Background(), "stop on error", p)
	if r.Status != "partial" || len(r.Results) != 1 { t.Fatalf("%+v", r) }
	if _, err := os.Stat(filepath.Join(tools.workspace, "must-not-run.txt")); !os.IsNotExist(err) { t.Fatal("later step executed after failure") }
	if got := DefaultEntityState.Snapshot().Mode; got != EntityError { t.Fatalf("entity mode=%s", got) }
}

func TestExecutePlanPublishesWorkingBeforeCompletion(t *testing.T) {
	c, _, _ := testCore(t)
	updates, cancel := DefaultEntityState.Subscribe()
	defer cancel()
	<-updates
	p := Plan{Summary:"observe lifecycle", Steps:[]PlanStep{{Tool:"system.info", Input:map[string]string{}}}}
	r := c.ExecutePlan(context.Background(), "observe lifecycle", p)
	if r.Status != "ok" { t.Fatalf("%+v", r) }
	if got := DefaultEntityState.Snapshot().Mode; got != EntitySuccess { t.Fatalf("final entity mode=%s", got) }
}
