package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

type EntityPresentation struct {
	Snapshot EntitySnapshot   `json:"snapshot"`
	Visual   EntityVisualState `json:"visual"`
	Frame    EntityRenderFrame `json:"frame"`
}

func currentEntityPresentation(reducedMotion bool) EntityPresentation {
	snapshot := DefaultEntityState.Snapshot()
	visual := VisualState(snapshot)
	return EntityPresentation{
		Snapshot: snapshot,
		Visual: visual,
		Frame: RenderFrame(visual, 0.5, reducedMotion),
	}
}

func writeEntityEvent(w http.ResponseWriter, p EntityPresentation) error {
	b, err := json.Marshal(p)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "event: entity\ndata: %s\n\n", b)
	return err
}

const entityUIHTML = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>NEURA</title>
<style>
:root{color-scheme:dark;--bg:#05070c;--panel:rgba(12,18,31,.72);--line:rgba(135,174,255,.16);--text:#eef4ff;--muted:#8d9bb3;--accent:#83a8ff;--glow:.42;--scale:1.01;--transition:500ms}
*{box-sizing:border-box}body{margin:0;min-height:100vh;background:
radial-gradient(circle at 50% 35%,rgba(37,74,145,.20),transparent 33rem),
linear-gradient(180deg,#070a12 0%,var(--bg) 70%);font-family:Inter,ui-sans-serif,system-ui,-apple-system,Segoe UI,sans-serif;color:var(--text);overflow:hidden}
.shell{min-height:100vh;display:grid;grid-template-rows:auto 1fr auto;padding:28px}
.top{display:flex;justify-content:space-between;align-items:center;gap:16px}.brand{letter-spacing:.34em;font-size:12px;font-weight:700}.badge{border:1px solid var(--line);background:var(--panel);backdrop-filter:blur(18px);padding:8px 12px;border-radius:999px;color:var(--muted);font-size:12px}
.stage{display:grid;place-items:center;perspective:1100px;position:relative}
.entity{width:min(54vw,520px);aspect-ratio:1;position:relative;transform-style:preserve-3d;transform:scale(var(--scale));transition:transform var(--transition) ease,filter var(--transition) ease;filter:drop-shadow(0 0 calc(70px * var(--glow)) rgba(92,135,255,.34))}
.entity[data-motion="orbit"]{animation:orbit 8s ease-in-out infinite}.entity[data-motion="flow"]{animation:flow 4.5s ease-in-out infinite}.entity[data-motion="resolve"]{animation:resolve 1.4s ease-out 1}.entity[data-motion="alert"]{animation:alert .9s ease-in-out infinite}
.halo,.ring,.core,.lobe{position:absolute;inset:50%;transform-style:preserve-3d;border-radius:50%}
.halo{width:90%;height:90%;margin:-45%;background:radial-gradient(circle,rgba(121,160,255,.10),transparent 62%);box-shadow:0 0 120px rgba(79,124,255,.16)}
.ring{width:78%;height:78%;margin:-39%;border:1px solid rgba(135,174,255,.22);box-shadow:inset 0 0 40px rgba(91,137,255,.07);transform:rotateX(66deg) rotateZ(18deg)}
.ring.r2{width:64%;height:64%;margin:-32%;transform:rotateY(67deg) rotateZ(-24deg);opacity:.65}
.core{width:54%;height:54%;margin:-27%;background:
radial-gradient(circle at 36% 28%,rgba(255,255,255,.34),transparent 12%),
radial-gradient(circle at 50% 50%,rgba(91,132,255,.40),rgba(24,36,79,.64) 46%,rgba(7,11,22,.95) 74%);
border:1px solid rgba(162,190,255,.30);box-shadow:inset 0 0 80px rgba(108,149,255,.18),0 0 80px rgba(79,123,255,.18)}
.lobe{width:28%;height:40%;margin:-20% -14%;background:linear-gradient(145deg,rgba(145,174,255,.22),rgba(37,55,104,.10));border:1px solid rgba(143,176,255,.18);backdrop-filter:blur(12px)}
.l1{transform:translate3d(-55px,-20px,64px) rotateY(-22deg)}.l2{transform:translate3d(55px,-20px,64px) rotateY(22deg)}.l3{transform:translate3d(-38px,55px,38px) rotateX(18deg)}.l4{transform:translate3d(38px,55px,38px) rotateX(18deg)}
.status{position:absolute;bottom:7%;text-align:center;max-width:min(80vw,620px)}.mode{font-size:clamp(26px,4vw,54px);font-weight:650;letter-spacing:-.03em;text-transform:capitalize}.detail{min-height:1.5em;margin-top:8px;color:var(--muted);font-size:14px;letter-spacing:.02em}
.bottom{display:flex;justify-content:center;color:var(--muted);font-size:12px}.live{display:flex;align-items:center;gap:8px}.dot{width:7px;height:7px;border-radius:50%;background:#6ae0a5;box-shadow:0 0 14px rgba(106,224,165,.8)}
@keyframes orbit{0%,100%{transform:scale(var(--scale)) rotateX(-2deg) rotateY(-5deg)}50%{transform:scale(var(--scale)) rotateX(2deg) rotateY(5deg)}}@keyframes flow{0%,100%{transform:scale(var(--scale)) translateY(0)}50%{transform:scale(calc(var(--scale) + .018)) translateY(-8px)}}@keyframes resolve{0%{transform:scale(.97)}55%{transform:scale(calc(var(--scale) + .045))}100%{transform:scale(var(--scale))}}@keyframes alert{0%,100%{transform:scale(var(--scale)) translateX(0)}50%{transform:scale(var(--scale)) translateX(3px)}}
@media (max-width:700px){.shell{padding:20px}.entity{width:min(82vw,460px)}.top{align-items:flex-start}.badge{max-width:52vw;text-align:right}}
@media (prefers-reduced-motion:reduce){.entity{animation:none!important;transition:none!important}.ring{transform:none}.ring.r2{transform:none}}
</style>
</head>
<body>
<div class="shell">
<header class="top"><div class="brand">NEURA</div><div class="badge" id="health">LOCAL CORE</div></header>
<main class="stage">
  <div class="entity" id="entity" data-motion="breathe" aria-label="NEURA digital entity">
    <div class="halo"></div><div class="ring"></div><div class="ring r2"></div><div class="core"></div>
    <div class="lobe l1"></div><div class="lobe l2"></div><div class="lobe l3"></div><div class="lobe l4"></div>
  </div>
  <div class="status"><div class="mode" id="mode">Idle</div><div class="detail" id="detail">Ready</div></div>
</main>
<footer class="bottom"><div class="live"><span class="dot"></span><span id="connection">connected locally</span></div></footer>
</div>
<script>
const entity=document.getElementById('entity'),mode=document.getElementById('mode'),detail=document.getElementById('detail'),connection=document.getElementById('connection');
const reduced=matchMedia('(prefers-reduced-motion: reduce)').matches;
function apply(p){
 const v=p.visual,f=p.frame;
 entity.dataset.motion=reduced?'still':v.motion;
 entity.style.setProperty('--scale',String(f.scale||1));
 entity.style.setProperty('--glow',String(f.glow||.35));
 entity.style.setProperty('--transition',(f.transition_ms||420)+'ms');
 mode.textContent=v.mode||'idle';
 detail.textContent=v.label||'';
 document.body.dataset.mode=v.mode||'idle';
}
fetch('/entity?reduced_motion='+(reduced?'1':'0')).then(r=>r.json()).then(apply).catch(()=>{});
const events=new EventSource('/entity/events?reduced_motion='+(reduced?'1':'0'));
events.addEventListener('entity',e=>{try{apply(JSON.parse(e.data));connection.textContent='connected locally'}catch(_){}});
events.onerror=()=>{connection.textContent='reconnecting…'};
</script>
</body>
</html>`

func (s *Server) entityUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline'; script-src 'unsafe-inline'; connect-src 'self'; img-src 'self' data:; base-uri 'none'; frame-ancestors 'none'")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(entityUIHTML))
}

func (s *Server) entitySnapshot(w http.ResponseWriter, r *http.Request) {
	reduced := r.URL.Query().Get("reduced_motion") == "1"
	writeJSON(w, http.StatusOK, currentEntityPresentation(reduced))
}

func (s *Server) entityEvents(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "streaming unsupported"})
		return
	}
	reduced := r.URL.Query().Get("reduced_motion") == "1"
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.Header().Set("Cache-Control", "no-cache, no-store")
	ch, cancel := DefaultEntityState.Subscribe()
	defer cancel()

	for {
		select {
		case <-r.Context().Done():
			return
		case snap, ok := <-ch:
			if !ok {
				return
			}
			visual := VisualState(snap)
			p := EntityPresentation{Snapshot: snap, Visual: visual, Frame: RenderFrame(visual, 0.5, reduced)}
			if err := writeEntityEvent(w, p); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

func entityPhase(raw string) float64 {
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil { return 0.5 }
	return v
}
