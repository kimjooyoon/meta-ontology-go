# Called-body continuation: October 8 observation

Clean compiler `395fb538fae6c55edb173e660b22daef8544f01a`, Go 1.27.1,
Darwin arm64. [Build identity](build.json), [summary and raw identities](summary.json),
and [runnable example](../../../examples/dependent-continuation/README.md).
The binary identifies as 0.6.6-dev; this source adds functionality after the
published v0.6.6-dev release. Pin the commit to reproduce this observation.

## Observed results

| Program / ordering | Saved native cases | Continued native cases | Search work during continuation |
| --- | --- | --- | --- |
| Nested `Seed → Wrap → Main` | 0/2 | 2/2 | Seed: retain 1, add 1. Wrap: retain 1, recheck 1 under changed Seed, add 1 |
| Called diagnostic / deterministic | 2/4 | 4/4 | Retain 1, add 3 |
| Called diagnostic / own QAT model | 2/4 | 4/4 | Retain 1, add 1 |

Each continued program also replayed successfully in a new process. Every
continuation and replay made zero new predictions. Both diagnostic paths
generated the same program. Their helper construction scores rose from 3/5
to 5/5, independently of the four native application expectations. The nested
program ended with 2/2 construction cases for each helper. Its input-separation
report still identifies missing observations for calls during candidate scoring.

The unchanged [own model](https://huggingface.co/asketeddy/gooo-record-shared-field-tiny-v1/tree/5800946afb35506d66357ee3ea6f956f506795b9/models/qat_ternary)
predicted mask 7 first. Prediction took **11,375 ns**; setup took **0.716167 ms**;
resident decoded tensors were **2,096 bytes**. Continued search selected mask 3.
This is a retained ranking observation, with no fine-tuning during the experiment.

Whole-command times for deterministic checkpoint/continuation/replay were
0.54/0.33/0.30 seconds; model-guided times were 0.53/0.35/0.31 seconds.
The model-guided checkpoint recorded 0.16 seconds user CPU and 0.11 seconds
system CPU, with 85,950,464 bytes maximum RSS. These figures include Go
compilation and native execution. They do not isolate the model's utilization
increase or establish a speed advantage. Every row is one sequential local run
with uncontrolled caches. Raw `.time` files preserve all reported counters.

## Reproduce the own-model pair

Use the example commands, replacing `main.gooo.fixture` with
`diagnostic.gooo.fixture` and `cases.json` with `diagnostic-cases.json` in the
same directory. Use new output directories. For the model condition, add
`--model /path/to/shared-qat/model.json` only to the checkpoint command.
The deterministic condition omits it. Continuation and replay use the saved
composition and require no model path.

The nine raw JSON outputs preserve the selected source, generated program,
policies, construction histories and native receipts. The summary binds their
SHA-256 digests. These are maintainer observations; independent reproduction,
external use and broad task-level advantage need further evidence.

## Development consequence

A reusable language tool needs to survive changes to the small functions it
uses. This experiment implements that next step for record-choice construction.
Historical stages keep their original dependency sources, while current scores
are calculated with current helpers. Regression tests additionally exercise an
already complete caller becoming incomplete, a dependency used only by an
alternative, unchanged unrelated helpers, exhausted budgets, repeated rounds,
origin-bound model ranking, and a helper used as both a bound producer and call.

The next integration work is package-workspace continuation and input histories
for nested candidate evaluation. Those are concrete gaps from constructing and
reusing Gooo tools; they provide the next language tasks and future training
observations.
