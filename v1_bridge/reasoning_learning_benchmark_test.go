package main

import (
	"context"
	"fmt"
	"testing"
)

// TestReasoningLearningV1Benchmark is a deterministic V1 capability gate.
// It measures the verified-action loop and the governed learning gate without
// requiring a paid model or network service.
func TestReasoningLearningV1Benchmark(t *testing.T) {
	workspace := t.TempDir()
	core, _, _ := testCore(t)

	cases := []Plan{
		{Summary: "empty reasoning plan"},
		{Summary: "inspect workspace", Steps: []PlanStep{{Tool: "fs.list", Input: map[string]string{"path": "."}}}},
	}

	completed := 0
	verified := 0
	for i, plan := range cases {
		res := core.ExecutePlan(context.Background(), fmt.Sprintf("benchmark-%d", i), plan)
		if res.Status == "ok" {
			completed++
		}
		for _, r := range res.Results {
			if r.Verified {
				verified++
			}
		}
	}
	if completed != len(cases) {
		t.Fatalf("task completion=%d/%d", completed, len(cases))
	}
	if verified != 1 {
		t.Fatalf("verified tool selection/execution=%d want 1", verified)
	}

	mgr, err := NewImprovementManager(workspace)
	if err != nil {
		t.Fatal(err)
	}
	candidate, err := mgr.Stage("reasoning-learning-gate", "reasoning-learning", []ImprovementChange{{Path: "learned.txt", Content: "verified procedure\n"}})
	if err != nil {
		t.Fatal(err)
	}
	if !candidate.RollbackReady || candidate.AutoPromotion {
		t.Fatalf("unsafe learning candidate: rollback=%v auto=%v", candidate.RollbackReady, candidate.AutoPromotion)
	}

	candidate, err = mgr.RecordEvaluation(candidate.ID, ImprovementEvaluation{
		TestsPassed:       true,
		BenchmarkMeasured: true,
		BenchmarkNoWorse:  true,
		Evidence:          []string{"deterministic reasoning-learning benchmark"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !candidate.PromotionReady {
		t.Fatal("verified candidate should be promotion-ready")
	}
	ok, _, err := mgr.PromotionDecision(candidate.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("governed learning gate rejected verified candidate")
	}
}
