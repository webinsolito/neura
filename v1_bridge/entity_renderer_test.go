package main

import "testing"

func TestRenderFrameBoundsAndState(t *testing.T) {
	states := []EntityMode{EntityIdle, EntityListening, EntityThinking, EntityWorking, EntitySuccess, EntityError}
	for _, mode := range states {
		visual := VisualState(EntitySnapshot{Mode: mode, Detail: "safe"})
		frame := RenderFrame(visual, 0.75, false)
		if frame.Mode != mode { t.Fatalf("mode %s became %s", mode, frame.Mode) }
		if frame.Scale < 1 || frame.Scale > 1.06 { t.Fatalf("unsafe scale for %s: %f", mode, frame.Scale) }
		if frame.Glow < 0 || frame.Glow > 1 { t.Fatalf("unsafe glow for %s: %f", mode, frame.Glow) }
		if frame.CoreOpacity < 0 || frame.CoreOpacity > 1 || frame.HaloOpacity < 0 || frame.HaloOpacity > 1 {
			t.Fatalf("opacity out of bounds for %s", mode)
		}
		if frame.RotateXDeg < -5 || frame.RotateXDeg > 5 || frame.RotateYDeg < -5 || frame.RotateYDeg > 5 {
			t.Fatalf("rotation out of bounds for %s", mode)
		}
		if frame.ParticleRate < 0 || frame.ParticleRate > 30 { t.Fatalf("particle rate out of bounds: %d", frame.ParticleRate) }
	}
}

func TestRenderFrameReducedMotion(t *testing.T) {
	visual := VisualState(EntitySnapshot{Mode: EntityWorking, Detail: "working"})
	frame := RenderFrame(visual, 0.4, true)
	if frame.Motion != "still" || frame.RotateXDeg != 0 || frame.RotateYDeg != 0 || frame.ParticleRate != 0 {
		t.Fatalf("reduced motion must disable movement: %+v", frame)
	}
}

func TestRenderFrameClampsIntensityAndPhase(t *testing.T) {
	visual := EntityVisualState{Mode: EntityWorking, Motion: "flow", Intensity: 99, PulseMS: 1, Interactive: true}
	frame := RenderFrame(visual, -2.25, false)
	if frame.Glow > 1 || frame.CoreOpacity > 1 || frame.HaloOpacity > 1 { t.Fatalf("intensity not clamped: %+v", frame) }
	if frame.TransitionMS < 180 { t.Fatalf("transition too fast: %d", frame.TransitionMS) }
}
