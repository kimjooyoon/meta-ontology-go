# Laya local evaluation baseline — 2026-09-29

This is a local smoke measurement of the Gooo `decide` path against a warm,
loopback Laya service. It records latency and resource use; it is not an
accuracy benchmark.

## Setup

- Host: Apple M4, arm64, 10 logical CPUs
- Runtime: Go 1.27.0, Python 3.11.15, `laya[serve]==0.3.21`
- Device: CPU; Laya server bound to `127.0.0.1:8787`
- Loaded checkpoints: English and multilingual; Korean request routed to
  multilingual
- Checkpoint revision: `55cf4c4ebb4ebe31b2550e8bdf3bd21b99753851`
- Request: `examples/laya-decision/request.json`
- Measurement: one warm-up call followed by 30 sequential `gooo decide --json`
  processes. Each Laya round trip includes the decision request and the
  `/health` revision lookup. A separate 30-call deterministic-fallback batch
  used the same CLI request.

## Results

| Path | Calls | p50 | p95 | Mean | Min–max |
| --- | ---: | ---: | ---: | ---: | ---: |
| Gooo CLI → Laya → revision lookup | 30 | 61.59 ms | 63.24 ms | 62.06 ms | 60.43–72.23 ms |
| Gooo deterministic fallback | 30 | 4.55 ms | 5.22 ms | 4.63 ms | 4.09–5.66 ms |

The separately measured warm-up call took 256.53 ms and is excluded from the
table. It is not a cold model-load measurement; the server and checkpoints were
already running. The warm Laya path was about 13.5 times the fallback p50 in
this run, adding about 57 ms at p50.

| Resource observation | Result |
| --- | ---: |
| Laya process CPU while idle, 2-second baseline | 0.13% of one core |
| Laya process CPU during the 30 Laya calls | 105.21% of one core on average |
| Increase over that idle baseline | 105.08 percentage points |
| Approximate share of this 10-core host during the batch | 10.5% |
| Laya process RSS after loading | 565.3 MiB |

CPU percentage is process CPU time divided by elapsed wall time, normalized to
one core. Values above 100% mean the process used more than one core on average.
The host-wide CPU was not used as the primary signal because unrelated system
activity varied during the short run.

All 30 repeated inputs selected `implement`, with identical returned
`confidence=0.4368` and `answer_confidence=0.7682`. This is one repeated input,
not 30 independent examples, so it says nothing about classification accuracy
or confidence calibration.

## Code-generation implications

Gooo currently lowers `.gooo` source into semantic IR and emits Go through its
deterministic generator. Laya can fit before that boundary as a bounded router:
choose among existing generation recipes or classify an intent into a small
known set, then let the semantic parser, generator, and conformance checks do
the construction and acceptance. Keep the selected recipe, probabilities,
checkpoint revision, source digest, generated digest, and deterministic check
outcomes in the observation record.

Laya should not directly author Go or become a source-of-truth path. Its model
card describes it as a typed-decision model that does not generate text. It
also reports base-checkpoint accuracy below the majority baseline on its
typed-decisions benchmark; the higher score belongs to a separately fine-tuned
checkpoint. Confidence therefore needs calibration against Gooo's own recorded
outcomes. The current `gooo decide` uses a valid Laya choice even at low
confidence; it falls back for missing, unavailable, or malformed provider
results. A future code-generation router should add an uncertainty gate only
after calibration, with the explicit deterministic choice as its fallback.
See the [Laya model card](https://huggingface.co/convaiinnovations/laya)
for its capabilities and benchmark limits.

For a fully automatic improvement loop, first run the classifier in shadow
mode over real Gooo tasks. Measure recipe-selection accuracy against outcomes
that can be derived from parser success, generated-artifact replay, semantic
conformance, and regression checks. Promote a model or routing rule only when
those deterministic measures improve; do not use Laya's self-reported
confidence as the improvement score.
