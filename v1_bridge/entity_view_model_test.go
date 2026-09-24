package main

import "testing"

func TestVisualStateMapsLifecycle(t *testing.T) {
	tests := []struct {
		mode EntityMode
		motion string
		minIntensity float64
	}{
		{EntityIdle, "breathe", 0.30},
		{EntityListening, "listen", 0.45},
		{EntityThinking, "orbit", 0.60},
		{EntityWorking, "flow", 0.75},
		{EntitySuccess, "resolve", 0.70},
		{EntityError, "alert", 0.80},
	}
	for _, tt := range tests {
		got := VisualState(EntitySnapshot{Mode: tt.mode, Detail: "safe status"})
		if got.Mode != tt.mode || got.Motion != tt.motion || got.Intensity < tt.minIntensity {
			t.Fatalf("mode %s mapped to %#v", tt.mode, got)
		}
		if got.Label != "safe status" || !got.Interactive || got.PulseMS <= 0 {
			t.Fatalf("mode %s returned incomplete visual contract: %#v", tt.mode, got)
		}
	}
}

func TestVisualStateRejectsUnknownMode(t *testing.T) {
	got := VisualState(EntitySnapshot{Mode: EntityMode("unknown"), Detail: "do not expose"})
	if got.Mode != EntityError || got.Motion != "alert" || got.Label != "invalid state" {
		t.Fatalf("unexpected unknown-mode fallback: %#v", got)
	}
}
