# Optional own Gooo three-choice path model

An explicit `--path-model` may supply the separate
`gooo/tiny-three-choice-path-model/v1` schema from SDK v0.2.13-experimental.
The compiler binds the original Gooo body to the declared typed fallback before
opening or predicting with a model. It projects three complete source-v3 inputs
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

The model has 18,656 parameters, a fixed 3,200-byte caller workspace and 74,624
FP32 weight bytes. Five-trit files occupy 3,854 bytes; decoding uses 18,624 int8
matrix bytes, 128 FP32 bias bytes and eight separate scale bytes. These figures
describe model arrays/storage, rather than process RAM or packed arithmetic.
Metadata is capped at 64 KiB and three-choice weights at 128 KiB. Existing
operation, independent path and two-choice model contracts retain their ABI.

This integration verifies source binding, all eight finite candidates, explicit
feedback accounting, declines and independently compiled generated Go against
authored arithmetic. Controlled test weights exercise the ABI, rather than
trained quality. No default model is selected. Fresh three-choice training and
the larger preregistered study remain separate phases:
[frozen protocol](https://github.com/kimjooyoon/gooo-neural-decision-experiments/blob/d74a8a455ceed5949fcbad482375405b4704dc9a/docs/own-three-choice-completeness-preregistration-20261002.md).
