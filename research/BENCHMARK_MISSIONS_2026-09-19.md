# Candidate benchmark missions — 2026-09-19

These are evidence missions, not promotions. Stable must remain untouched.

## ZS-20260919-01 — BitNet 270M embedding lane
Goal: determine whether microsoft/BitNet embedding-270M materially improves NEURA's local memory retrieval efficiency on Windows.

Candidate only. Measure against the current NEURA embedding/retrieval baseline on the same corpus and hardware:
1. cold start and warm start;
2. peak working set / RAM;
3. model + runtime disk footprint;
4. embeddings/sec and p50/p95 latency;
5. retrieval quality on the existing memory regression corpus (same top-k and scoring protocol);
6. offline behavior after model is present locally;
7. installation/build complexity and number of redistributable runtime files.

Acceptance: no retrieval-quality regression, >=20% RAM or latency improvement, EUR 0 recurring cost, license/redistribution documented, Windows-native run PASS, rollback demonstrated. Otherwise REJECT/KEEP-AS-OPTION.

## ZS-20260919-02 — sherpa-onnx voice lane
Goal: evaluate an offline Windows voice front-end without weakening Warden/capability enforcement.

Candidate only. Test VAD + denoise + STT first; TTS and wake word remain separate missions. Measure:
1. first-audio latency and end-of-utterance latency;
2. real-time factor on CPU;
3. peak RAM and disk footprint;
4. word error rate on a fixed small Italian/English noisy corpus;
5. cancellation/interruption behavior;
6. offline operation;
7. process crash containment;
8. proof that every privileged post-STT action still passes through NEURA's existing capability lease / Warden path.

Acceptance: real-time on target Windows hardware, no safety-path bypass, no recurring cost, deterministic fallback/disable switch, rollback PASS. Otherwise do not replace the current voice path.

## ZS-20260919-03 — Source CI failure triage
Current GitHub Source CI is not green: formatting passed but `go vet ./...` failed in the observed run. Do not claim the imported C9.38 source is verified until the exact vet error is captured and fixed or classified. This mission outranks feature integration because reliability comes first.
