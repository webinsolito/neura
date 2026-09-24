package main

import "fmt"

// EntityRenderFrame is the concrete, dependency-free render description for NEURA's
// central entity. A desktop/web shell can consume these values without duplicating
// state semantics or coupling the agent loop to a graphics framework.
type EntityRenderFrame struct {
	Mode         EntityMode `json:"mode"`
	Label        string     `json:"label"`
	Motion       string     `json:"motion"`
	Scale        float64    `json:"scale"`
	RotateXDeg   float64    `json:"rotate_x_deg"`
	RotateYDeg   float64    `json:"rotate_y_deg"`
	Glow         float64    `json:"glow"`
	CoreOpacity  float64    `json:"core_opacity"`
	HaloOpacity  float64    `json:"halo_opacity"`
	ParticleRate int        `json:"particle_rate"`
	TransitionMS int        `json:"transition_ms"`
	Interactive  bool       `json:"interactive"`
}

// RenderFrame converts the stable visual contract into bounded renderer values.
// phase is normalized to [0,1) so callers can animate deterministically without
// allocating timers inside the core. reducedMotion disables rotation/particles.
func RenderFrame(state EntityVisualState, phase float64, reducedMotion bool) EntityRenderFrame {
	if phase < 0 || phase >= 1 {
		phase = phase - float64(int(phase))
		if phase < 0 {
			phase += 1
		}
	}

	intensity := clamp01(state.Intensity)
	frame := EntityRenderFrame{
		Mode:         state.Mode,
		Label:        state.Label,
		Motion:       state.Motion,
		Scale:        1 + 0.025*intensity,
		Glow:         0.24 + 0.58*intensity,
		CoreOpacity:  0.72 + 0.24*intensity,
		HaloOpacity:  0.18 + 0.42*intensity,
		ParticleRate: int(4 + 20*intensity),
		TransitionMS: maxInt(180, state.PulseMS/3),
		Interactive:  state.Interactive,
	}

	// A small bounded parallax/orbit gives depth while avoiding continuous heavy
	// 3D work. Renderers may map these values to CSS transforms or a real 3D scene.
	frame.RotateYDeg = (phase*2 - 1) * 5 * intensity
	frame.RotateXDeg = (0.5 - phase) * 3 * intensity

	switch state.Mode {
	case EntitySuccess:
		frame.Scale += 0.025
		frame.ParticleRate += 4
	case EntityError:
		frame.Scale += 0.012
		frame.TransitionMS = 180
	case EntityIdle:
		frame.ParticleRate = minInt(frame.ParticleRate, 8)
	}

	if reducedMotion {
		frame.Motion = "still"
		frame.RotateXDeg = 0
		frame.RotateYDeg = 0
		frame.ParticleRate = 0
		frame.TransitionMS = maxInt(frame.TransitionMS, 300)
	}
	return frame
}

func clamp01(v float64) float64 {
	if v < 0 { return 0 }
	if v > 1 { return 1 }
	return v
}

func minInt(a, b int) int { if a < b { return a }; return b }
func maxInt(a, b int) int { if a > b { return a }; return b }

func (f EntityRenderFrame) String() string {
	return fmt.Sprintf("%s/%s scale=%.3f glow=%.3f", f.Mode, f.Motion, f.Scale, f.Glow)
}
