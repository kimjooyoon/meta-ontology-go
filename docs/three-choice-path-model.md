# Optional own Gooo three-choice path model

## Choosing a compatible model — 2026-10-03

Use the [compact V3 bundle](https://huggingface.co/asketeddy/gooo-shared-judgment-tiny-v1/tree/985999a89caba6a31cc7147f66ba29a5ce76a1d9/research/compact-runtime-20261003)
for the current compiler and SDK v0.2.14-experimental. Its generated-program
measurements are linked below. The same Hugging Face repository also publishes
new [full-input research models](https://huggingface.co/asketeddy/gooo-shared-judgment-tiny-v1/tree/main/research/full-input-initial-20261003).
Their V4 whole-text fragment feature contract runs in the research runtime and
[SDK v0.2.15-experimental](https://github.com/kimjooyoon/gooo-decision-runtime/releases/tag/v0.2.15-experimental).
Explicit arithmetic now matches the frozen arm64/Linux observations, including
every full ranking. The SDK replay passed on both platforms at revision `59c8d34`:
18,432 complete inputs and 36,864 predictions per platform.
[Linux SDK replay](https://github.com/kimjooyoon/gooo-decision-runtime/actions/runs/37070241916).
The compiler still depends on v0.2.14; use the pinned V3 files above for its
current `--path-model` contract. Native V4 generation and execution are the next
stage in the [registered protocol](https://github.com/kimjooyoon/gooo-neural-decision-experiments/blob/daecfea3583614e006c960de263448a1645a9190/docs/full-input-sdk-native-protocol-20261003.md).
The [language direction guide](language-direction.ko.md) explains the progression
and the remaining language work.

## Current compiler contract

An explicit `--path-model` may supply the separate
`gooo/tiny-three-choice-path-model/v1` schema, or the compact shared schema below,
from SDK v0.2.14-experimental. Direct codegen binds the original Gooo body to the
declared typed fallback before opening a model. Retained generators load during
construction; every request binds its source before predicting. The compiler projects three complete source-v3 inputs
in declared order and ranks all eight absolute masks in one initial prediction.
It never substitutes eight marginal-label or four-mask predictions.

    gooo body-codegen --json --path-plan plan.json --path-model three/model.json --path-step-attempts 1 --path-feedback-rounds 7 --activity ChoosePath source.gooo

Actual finite failures can re-rank remaining masks with one explicit prediction;
the final sole mask needs none. Seeded initial sampling is reproducible and keeps
the seed outside model text. Tests decide finite acceptance, and CI feedback is
caller context rather than edit authority. No weight updates occur during codegen.

Disconnected execution remains deterministic. Unsupported arity and complete
source/intent overflow record all original declared input text and its hash,
skip seed/feedback and continue deterministically with zero predictions.
`complete_declared_inputs` identifies caller text; successful prediction receipts
separately identify actual source-projected canonical model bytes. Later feedback
overflow preserves complete attempted input and each part's hash, consumes one
bounded round and continues using unchanged scores. Text is never shortened.

`NewTypedPathGenerator` and `gooo-body-worker` load this immutable model once.
Each request owns its source, plan, workspace, session, progress and feedback.
Concurrent requests share only immutable tensors. Tests remove the on-disk model
files after construction and confirm eight independent requests still complete.

The expanded model has 18,656 parameters, a fixed 3,200-byte caller workspace and 74,624
FP32 weight bytes. Five-trit files occupy 3,854 bytes; decoding uses 18,624 int8
matrix bytes, 128 FP32 bias bytes and eight separate scale bytes. These figures
describe model arrays/storage, rather than process RAM or packed arithmetic.
Metadata is capped at 64 KiB and three-choice weights at 128 KiB. Existing
operation, independent path and two-choice model contracts retain their ABI.

## Compact shared judge

SDK v0.2.14 also accepts `gooo/tiny-shared-three-choice-path-model/v1` through the
same `--path-model` flag and retained worker. It stores one 8x256 input matrix,
eight hidden biases and a 2x8 output matrix reused across the three decisions.
`hidden_dim: 8` names the stored judge; the full feature/workspace/mask dimensions
remain 768/24/8. FP32 weights/resident tensors occupy 8,288 bytes. Ternary files
occupy 446 bytes and decode to 2,096 tensor bytes plus eight scale bytes.
Compact weights are capped at 16 KiB; metadata is capped at 64 KiB.

The loader requires the explicit schema and a strict three-tensor layout.
Initial, feedback and retention receipts record the compact artifact's actual
schema and hashes. Seeds bind those hashes: repeated requests with one artifact
are reproducible, while the same seed across expanded/compact artifacts need
not choose the same mask. All unsupported-input and cancellation behavior above
also applies to compact models.

[Published own compact models](https://huggingface.co/asketeddy/gooo-shared-judgment-tiny-v1/tree/985999a89caba6a31cc7147f66ba29a5ce76a1d9/research/compact-runtime-20261003)
include exact 10,739-state parity in all three variants. Those local kernel
measurements exclude native codegen. Compiler regression tests use controlled
weights to exercise all eight candidates, actual feedback, retained concurrent
requests, full-input declines and independently compiled arithmetic for FP32,
PTQ and QAT. They are not a new trained-model accuracy or native speed claim.

The integration's controlled weights cover source binding, all eight candidates,
feedback accounting, input declines and compiled arithmetic. Training and actual
model measurements have since completed as separate studies:
[fresh shared-model quality](https://github.com/kimjooyoon/gooo-neural-decision-experiments/blob/main/docs/shared-three-judgment-results-20261003.md)
and [compact native generation](https://github.com/kimjooyoon/gooo-neural-decision-experiments/blob/main/docs/compact-shared-native-results-20261003.md).
The compact study retains 96 generations, 298 actual predictions, 192 compiled
runs and 48 matching unseeded representation pairs. Its 2,304 supplied finite
expectations all pass. Bilingual first-choice disagreement and family regressions
remain recorded in the quality study. Callers explicitly select the model.
