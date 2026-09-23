package main

import (
	"context"
	"strings"
	"testing"
	"time"
)

func fixedCapabilityGate(t *testing.T) (*CapabilityGate, *time.Time) {
	t.Helper()
	now := time.Date(2026, 9, 22, 15, 30, 0, 0, time.UTC)
	g, err := newCapabilityGateForTest([]byte("0123456789abcdef0123456789abcdef"), func() time.Time { return now })
	if err != nil {
		t.Fatal(err)
	}
	return g, &now
}

func TestCapabilityPassportValidAuthorization(t *testing.T) {
	g, _ := fixedCapabilityGate(t)
	p, err := g.Issue("core-direct", "stato", []string{capSystemObserve}, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if err := g.Authorize(p, capSystemObserve, "stato"); err != nil {
		t.Fatalf("valid passport denied: %v", err)
	}
}

func TestCapabilityPassportRejectsWrongCapability(t *testing.T) {
	g, _ := fixedCapabilityGate(t)
	p, err := g.Issue("core-direct", "leggi file a.txt", []string{capFSList}, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if err := g.Authorize(p, capFSRead, "leggi file a.txt"); err == nil {
		t.Fatal("wrong capability accepted")
	}
}

func TestCapabilityPassportRejectsTampering(t *testing.T) {
	g, _ := fixedCapabilityGate(t)
	p, err := g.Issue("core-direct", "stato", []string{capSystemObserve}, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	p.Capabilities = []string{capFSRead}
	if err := g.Authorize(p, capFSRead, "stato"); err == nil || !strings.Contains(err.Error(), "signature") {
		t.Fatalf("tampered passport accepted or wrong error: %v", err)
	}
}

func TestCapabilityPassportRejectsExpired(t *testing.T) {
	g, now := fixedCapabilityGate(t)
	p, err := g.Issue("core-direct", "stato", []string{capSystemObserve}, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	*now = now.Add(2 * time.Second)
	if err := g.Authorize(p, capSystemObserve, "stato"); err == nil || !strings.Contains(err.Error(), "expired") {
		t.Fatalf("expired passport accepted: %v", err)
	}
}

func TestCapabilityPassportRejectsPurposeMismatch(t *testing.T) {
	g, _ := fixedCapabilityGate(t)
	p, err := g.Issue("core-direct", "stato", []string{capSystemObserve}, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if err := g.Authorize(p, capSystemObserve, "leggi file a.txt"); err == nil || !strings.Contains(err.Error(), "purpose") {
		t.Fatalf("purpose mismatch accepted: %v", err)
	}
}

func TestCapabilityPassportRejectsWildcard(t *testing.T) {
	g, _ := fixedCapabilityGate(t)
	if _, err := g.Issue("core-direct", "stato", []string{"*"}, time.Minute); err == nil {
		t.Fatal("wildcard capability accepted")
	}
}

func TestCapabilityPassportRejectsExcessiveTTL(t *testing.T) {
	g, _ := fixedCapabilityGate(t)
	if _, err := g.Issue("core-direct", "stato", []string{capSystemObserve}, maxPassportTTL+time.Second); err == nil {
		t.Fatal("excessive passport ttl accepted")
	}
}

func TestCapabilityToolRegistryFailsClosedWithoutPassport(t *testing.T) {
	c, _, tools := testCore(t)
	_ = c
	r := tools.Run(context.Background(), "system.info", nil, CapabilityPassport{}, "stato")
	if r.Error == "" || !strings.Contains(r.Error, "capability denied") {
		t.Fatalf("missing passport did not fail closed: %+v", r)
	}
}

func TestCapabilityToolRegistryRejectsPassportForDifferentTool(t *testing.T) {
	_, _, tools := testCore(t)
	p, err := tools.gate.Issue("test", "read", []string{capFSList}, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	r := tools.Run(context.Background(), "fs.read", map[string]string{"path": "a.txt"}, p, "read")
	if r.Error == "" || !strings.Contains(r.Error, "denied") {
		t.Fatalf("cross-capability passport accepted: %+v", r)
	}
}

func TestCapabilityCoreDirectFlowStillWorks(t *testing.T) {
	c, _, _ := testCore(t)
	r := c.Execute(context.Background(), "stato")
	if r.Status != "ok" || len(r.Results) != 1 || !r.Results[0].Verified {
		t.Fatalf("authorized direct flow failed: %+v", r)
	}
}

func TestCapabilityModelPlanCannotEscalateBeyondToolAllowlist(t *testing.T) {
	p := Plan{Summary: "bad", Steps: []PlanStep{{Tool: "shell.exec", Input: map[string]string{}}}}
	if _, err := validatePlan(p, []string{"system.info"}); err == nil {
		t.Fatal("model plan escalation accepted")
	}
}

func TestCapabilityCatalogHasExactKnownTools(t *testing.T) {
	got := capabilityCatalog()
	want := map[string]string{
		"system.info":       capSystemObserve,
		"fs.list":           capFSList,
		"fs.read":           capFSRead,
		"fs.write.workspace": capFSWrite,
		"windows.processes": capWindowsObserve,
	}
	if len(got) != len(want) {
		t.Fatalf("unexpected catalog size: %#v", got)
	}
	for tool, cap := range want {
		if got[tool] != cap {
			t.Fatalf("tool %s capability=%q want=%q", tool, got[tool], cap)
		}
	}
}

func TestCapabilityToolRegistryVerifiedWorkspaceWrite(t *testing.T) {
	_, _, tools := testCore(t)
	p, err := tools.gate.Issue("test", "write", []string{capFSWrite}, time.Minute)
	if err != nil { t.Fatal(err) }
	r := tools.Run(context.Background(), "fs.write.workspace", map[string]string{"path":"tool.txt","content":"ok"}, p, "write")
	if r.Error != "" || !r.Verified { t.Fatalf("%+v", r) }
}
