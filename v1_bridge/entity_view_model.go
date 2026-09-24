package main

// EntityVisualState is the renderer-facing contract for NEURA's central entity.
// It intentionally exposes presentation hints only: no tool payloads, paths, or logs.
type EntityVisualState struct {
	Mode       EntityMode `json:"mode"`
	Label      string     `json:"label"`
	Motion     string     `json:"motion"`
	Intensity  float64    `json:"intensity"`
	PulseMS    int        `json:"pulse_ms"`
	Interactive bool      `json:"interactive"`
}

// VisualState converts core execution state into a stable, renderer-agnostic model.
// Keeping this mapping in Go lets desktop/web renderers share the same semantics
// without coupling the agent loop to any graphics framework.
func VisualState(snapshot EntitySnapshot) EntityVisualState {
	state := EntityVisualState{Mode: snapshot.Mode, Label: snapshot.Detail, Motion: "breathe", Intensity: 0.32, PulseMS: 2600, Interactive: true}
	switch snapshot.Mode {
	case EntityListening:
		state.Motion = "listen"
		state.Intensity = 0.48
		state.PulseMS = 1800
	case EntityThinking:
		state.Motion = "orbit"
		state.Intensity = 0.62
		state.PulseMS = 1200
	case EntityWorking:
		state.Motion = "flow"
		state.Intensity = 0.78
		state.PulseMS = 900
	case EntitySuccess:
		state.Motion = "resolve"
		state.Intensity = 0.72
		state.PulseMS = 1500
	case EntityError:
		state.Motion = "alert"
		state.Intensity = 0.86
		state.PulseMS = 700
	case EntityIdle:
		// defaults are deliberately calm to minimize idle GPU/CPU work.
	default:
		state.Mode = EntityError
		state.Motion = "alert"
		state.Intensity = 0.86
		state.PulseMS = 700
		state.Label = "invalid state"
	}
	return state
}
