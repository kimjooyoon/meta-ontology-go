# Laya local observation (2026-10-06)

This note records one local Gooo body-generation experiment using the public
`convaiinnovations/laya` English checkpoint. It is a measurement of this fixture
and machine, not a benchmark claim.

## Setup

- Machine: Apple M4, macOS, `darwin/arm64`.
- Go: `go1.27.1`.
- Python server: isolated Python 3.11 virtual environment, Laya `0.3.28`.
- Model: `english`, revision `7b928d828b7b0e022f929d9bd2e44165aa270148`.
- Device: PyTorch MPS; Laya reported zero CPU fallbacks.
- Server bound to `127.0.0.1`; only the English checkpoint was loaded.
- Input: `source-ir-fill-probe-choice.gooo.fixture`, with two candidates tied at
  100% training accuracy and a generated probe that distinguishes them.

## Observations

| Run | Provider | Decision time | Total generation time | Selected candidate | Holdout |
| --- | --- | ---: | ---: | --- | ---: |
| First request after startup | Laya | 703.11 ms | 705.29 ms | `strict_positive` | 100% (1/1) |
| Warm request sample 1 | Laya | 243.90 ms | 244.95 ms | `strict_positive` | 100% (1/1) |
| Warm request sample 2 | Laya | 171.56 ms | 172.29 ms | `strict_positive` | 100% (1/1) |
| Warm request sample 3 | Laya | 171.20 ms | 171.87 ms | `strict_positive` | 100% (1/1) |
| Warm request sample 4 | Laya | 171.99 ms | 172.65 ms | `strict_positive` | 100% (1/1) |
| Warm request sample 5 | Laya | 171.36 ms | 172.08 ms | `strict_positive` | 100% (1/1) |
| One disconnected fallback sample | deterministic | 0.028 ms | 0.607 ms | `strict_positive` | 100% (1/1) |

The Laya receipt had confidence `0.0222` and probabilities `0.5875` vs `0.4125`.
Laya logged a warning that some checkpoint temperature values were invalid or
outside its supported range and that affected confidence should be treated as
uncalibrated. The selected candidate matched the deterministic selection, so
this run shows integration working, but does not show a correctness gain from
Laya. The held-out value `5` also does not distinguish these two candidates;
the `0` probe does. The holdout result therefore must not be read as evidence
that the choice between the candidates is correct.

Immediately after requests, the server process reported 0.0–0.1% CPU and about
286–314 MiB RSS across two point-in-time samples. This is not peak CPU or total
model memory: inference ran on MPS and the sample missed active inference. No
claim about peak utilization can be made from these observations.

## What this suggests for Gooo

The useful role for this model is a bounded proposal step after Gooo has built a
typed IR plan, enumerated legal candidates, scored the declared training suite,
and exposed candidate behavior profiles. Gooo should remain responsible for
typechecking, emitting code, replay checks, and reporting finite-suite and
holdout results. The model should not invent unrestricted source or be treated
as a correctness oracle.

For everyday use, deterministic mode is the fast path when no provider is
configured. Laya is an optional chooser for ambiguous candidate sets, where its
latency is material and the receipt can reveal whether it improved on the
baseline. The next useful experiments are cases with holdouts that distinguish
the candidates, Korean and mixed-language intent, and repeated measurements of
cold-start time, warm latency, peak RSS, and process CPU. Keep those outcomes
separate from candidate coverage and accuracy; each answers a different
question.
