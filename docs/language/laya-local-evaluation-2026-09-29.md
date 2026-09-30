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

The stable `gooo generate` path lowers `.gooo` source into semantic IR in
[`generateWithDeadlineCore`](../../cmd/gooo/generate_pipeline_part04.go) and
emits Go through the [deterministic generator](../../internal/generator/generator_part01.go).
The experimental `gooo body-codegen` path now uses Laya as a bounded router
between equivalent lowering shapes for a small pure activity-body profile.
The initial alternatives preserve two branch returns, rewrite them as a
guard-return plus fallthrough, or create an explicit result local at the
control-flow join. This is a bounded IR-structure choice; it does not ask Laya
to write code. The selected route, probabilities, checkpoint
revision, source digest, generated digest, structural completeness, typecheck,
and deterministic replay are recorded in its JSON report. A missing or
unavailable endpoint selects the exact `preserve` route.

Laya should not directly author Go or become a source-of-truth path. Its model
card describes it as a typed-decision model that does not generate text. It
also reports base-checkpoint accuracy below the majority baseline on its
typed-decisions benchmark; the higher score belongs to a separately fine-tuned
checkpoint. Confidence therefore needs calibration against Gooo's own recorded
outcomes. The current `gooo decide` uses a valid Laya choice even at low
confidence; it falls back for missing, unavailable, or malformed provider
results. The body-codegen experiment currently uses valid route choices
without a confidence threshold; all three routes are defined by fixed lowering
rules, and its report does not treat confidence as correctness evidence.
Expand the route set only after CI and recorded codegen outcomes show a useful
distinction.
See the [Laya model card](https://huggingface.co/convaiinnovations/laya)
for its capabilities and benchmark limits.

The `body-codegen` JSON report includes route-decision latency in milliseconds.
It covers the local decision request and model-revision lookup; whole-command
duration also includes Go process startup and lowering.

## Body-codegen integration smoke — 2026-09-30

This follow-up exercised the real local Laya server through
`gooo body-codegen` after the checkpoint finished loading. It used the same
single pure conditional fixture 30 times, with three eligible lowering routes.
The local runtime was Python 3.11.15 with `laya[serve]==0.3.21`, PyTorch 2.14.0,
and Transformers 5.17.0; model inference used CPU.
The table reports the Gooo receipt's decision latency, which includes Laya's
choice request and the local model-revision lookup.

| Measurement | Result |
| --- | ---: |
| Calls | 30 |
| Route-decision p50 | 108.48 ms |
| Route-decision p95 | 123.10 ms |
| Mean / min / max | 112.84 / 105.36 / 219.90 ms |
| End-to-end batch wall time | 4.10 s |
| Selected route | `preserve`, 30 of 30 calls |
| One returned confidence / answer confidence | 0.4649 / 0.8211 |
| Laya process CPU-time increase during batch | 5.72 s |
| Laya CPU share over the 4.10-second batch | 139.5% of one core |
| Laya RSS before / after batch | 82.1 / 625.0 MiB |

The CPU percentage is normalized to one core; on this 10-logical-CPU Apple M4,
139.5% of one core is about 14.0% of total CPU capacity during this short
batch. The `ps` snapshot was 0.1% at idle before the batch and 161.2% just
after it. The local `.venv-laya` occupied 758 MiB and the model cache occupied
3.5 MiB at measurement time; 32 GiB remained available on the volume.

The first cold request exceeded the three-second planner budget while Laya
loaded the checkpoint. Gooo returned a typechecked `preserve` result with
`PROVIDER_UNAVAILABLE`; the full `go run` command took 5.01 seconds including
Go startup. The largest Laya RSS observed during loading was about 983 MiB.
Once the model was available, the receipt returned the Laya model revision
`55cf4c4ebb4ebe31b2550e8bdf3bd21b99753851` and generated code with 10/10
semantic units covered. Repeating one input 30 times measures latency and
resource use, not route accuracy. The model consistently preferred preserving
the source shape, so this run does not show that the alternate joins improve
generated code.

For a fully automatic improvement loop, first run the classifier in shadow
mode over real Gooo tasks. Measure recipe-selection accuracy against outcomes
that can be derived from parser success, generated-artifact replay, semantic
conformance, and regression checks. Promote a model or routing rule only when
those deterministic measures improve; do not use Laya's self-reported
confidence as the improvement score.

## Typed IR body-fill and test-score gate — 2026-09-30

The measurements below describe the initial body-fill implementation at
`2fc19ea5b094f424f550e2b0a9ebe6477759d1a2` (merged to `dev` as
`fa33c7532a8c0bebfc5fdb594fbfd4ba4b2d2e2a`). A subsequent differential review
found evaluator errors for shadowed Boolean names and large constant
arithmetic, plus expression-grouping and CLI error-propagation defects.
The simple clamp fixture below does not contain those counterexamples.
These historical timing measurements are not a benchmark of the repaired
type-information-based evaluator.

This experiment asks Laya to choose a typed expression for one Gooo IR hole in
a body that declares a local value, branches on `input < 0`, assigns the hole
in the negative branch, and returns the local. Gooo first typechecks three
closed candidates and scores them against nine declared `int64` inputs,
including the minimum and maximum values. The local evaluator reports exact
pass counts for this suite; it does not execute the generated Go binary or
prove behavior for every `int64` value.

The live run used the CPU English checkpoint at revision
`55cf4c4ebb4ebe31b2550e8bdf3bd21b99753851`, Laya 0.3.21, Python 3.11.15,
PyTorch 2.14.0, Transformers 5.17.0, and `LAYA_THREADS=4` on an Apple M4 with
10 logical CPUs. The loopback server was already warm. The request sent the IR
skeleton, candidate expressions and pass counts, plus the test count and
digest; individual test inputs remained in Gooo's local receipt.

| Measurement | Result |
| --- | ---: |
| Sequential calls | 10 |
| Laya decision p50 / p95 | 364.46 / 388.95 ms |
| Laya decision min / max | 360.41 / 388.95 ms |
| Whole CLI p50 | 370.34 ms |
| Batch wall time | 3.77 s |
| Laya process CPU-time increase | 8.50 s |
| Average Laya CPU | 225.6% of one core, about 22.6% of this 10-core host |
| Laya process peak RSS | 1,613.9 MiB |

Laya proposed `negate` in all 10 calls. That candidate passed 5/9 cases
(55.56%); `zero` passed 9/9 (100%). The receipt records Laya's proposal and a
44.44 percentage-point selection regret. Gooo's deterministic score gate
replaced the proposal with `zero` in all 10 runs; the emitted body passed 9/9,
then passed Go typechecking and deterministic replay. Repetition of one fixture
measures latency and repeatability, not general code-generation accuracy.

A paired context check used the same body skeleton and candidates, with five
warm calls per form. Sending all nine test cases produced a 394.26 ms median
and Laya selected `negate` in 5/5. Sending candidate score summaries, test
count, and digest produced a 381.78 ms median and Laya selected `zero` in 5/5.
That small sample does not establish a reliable speedup; the more useful
finding is that the decision changed when Laya saw the compact test evidence.
Gooo still scores every individual case locally and gates the emitted choice,
so the provider cannot erase or overrule test outcomes.

With no provider configured, ten body-fill CLI runs on the final local-variable
fixture had a 0.478 ms median reported generation time and a 5.86 ms median
process wall time. The 8-second body-fill provider decision limit is bounded and
cancellation-tested; it does not bound parsing, candidate scoring, or final
emission. It accommodates the measured warm path while leaving
deterministic fallback for missing or slow providers. Lazy model startup can
exceed that budget, so preload the intended checkpoint before latency-sensitive
use. The temporary
Python environment measured 943 MiB and lived outside the repository; the
checkpoint stayed in the existing Hugging Face cache and no model weights were
added to Git.

This experiment supports Laya as a bounded proposal source for codegen. It does
not support trusting its selection without deterministic test evidence: the
full-input call repeatedly chose a candidate with a 44.44-point observed gap.
The effective path is Gooo-owned IR, test-scored candidates, synchronous Laya
proposal, deterministic best-score arbitration, then emission and replay.
The score context here is local TDD evidence prepared before Laya is called,
not a GitHub Actions result. The exact generated revision's CI run can validate
this outcome and feed a subsequent iteration.
