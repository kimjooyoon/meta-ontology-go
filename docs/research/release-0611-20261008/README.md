# Gooo 0.6.11 candidate regression

The frozen [plan](PLAN.md) and implementation are commit
`60b1d7cb274f44531aa2501b00a29aed94b3958f`. A clean candidate built with Go
1.27.1 reports 0.6.11-dev and SDK v0.2.26-experimental. These local observations
precede public release and installation; later evidence-only changes do not
alter the tested implementation.

## Native platform observation

The macOS arm64 platform witness completed PASS/EXACT. Two native builds and
two archives matched. Existing division (8 inputs), retry (12), filenames (12),
their replays, source graph export and three input-only filename cases passed.

The new three-package diagnostic was constructed and replayed with explicit
metadata and with names/imports read from Gooo. Each of those four runs matched
the same four input rows and eight named outputs, including the large-integer
invalid-count branch and Korean text. Generated code was identical across all
four runs; runtime and replay inference counts were zero. These remain four
distinct input rows, repeated with two manifest forms and two execution modes.
Raw values, independent expected values and the candidate's platform receipt
are inside the observation archive.

## Actual use of the existing model

The public standard-library consumer and graph model remain pinned by PLAN.md.
Their sources, ten evaluation inputs and model weights were unchanged.

| Operation | Expected outputs | Construction attempts | New inference calls |
| --- | --- | --- | --- |
| Explicit metadata, fixed order | 10/10 | 8 | 0 |
| Source metadata, fixed order | 10/10 | 8 | 0 |
| Source metadata, own model | 10/10 | 8 | 1 |
| Saved model construction replay | 10/10 | prior 8 retained | 0 |

The entire normalized package graph, actual input/output traces and generated
program matched across manifest forms. Original manifest digests stayed distinct.
The first model candidate met 2/5 construction examples and 9/15 fields. Both
construction orderings needed all eight candidates; this observation shows no
reduction in completion attempts.

One local prediction took 110,791 ns. Model setup took 0.610458 ms and resident
tensors occupied 2,096 bytes. These are one-run observations; host CPU utilization,
incremental process memory and comparative latency were not measured.

Input-only replay returned three actual records without assigning a correctness
score: `(true, 13, "한글 예제")`, `(false, 1, "untitled")` and
`(false, 0, "doc")`, in available/bytes/title order. It made no new prediction.
The records retain the original construction history separately.

These are existing-program release regressions. The observations do not measure
new training, unseen-program transfer or arbitrary-intent completion.

## Retained evidence

- `build.json`: exact clean compiler identity.
- `summary.json`: exact-number comparison output with original attempt histories.
- `targeted-tests.log`: command, metadata and package-result regressions.
- `release-tests.log`: full release and CLI witness package race tests.
- `observations.tar.gz`: both workspace copies, original execution/resolution
  records, native platform observations, preparation/comparison Go code and logs.
- `SHA256SUMS`: digests of this retained set.

No model weights, compiler binaries or platform binary archives are duplicated
here. Extract the observation archive to a new directory and run
`go run /path/to/extracted/verify.go /path/to/extracted` to compare the retained
library results. The archived platform receipt preserves the excluded binary
archive identity. Four-platform CI and publication are subsequent steps.
