# Let a Gooo tool control the next construction attempt

The existing [Gooo explainer](../assembly-explainer/main.gooo.fixture) can now run
between record-construction attempts. After a candidate is type-checked and
scored, its Gooo conditions decide whether to try the next permitted candidate
or finish with an observed candidate. The compiler still owns the declared
alternatives, attempt budget, type checks and finite measurements.

## Construct and execute

Build this revision with Go 1.27.1, then run from the repository root:

```sh
go build -o /tmp/gooo-assembly-policy ./cmd/gooo
/tmp/gooo-assembly-policy body-compose \
  --source examples/package-diagnostic-replay/diagnostics.gooo.fixture \
  --cases examples/assembly-policy/cases.json \
  --assembly-policy examples/assembly-explainer/main.gooo.fixture \
  --policy-activity Explain --out /tmp/gooo-policy-construction
```

The output directory must be new. Add `--go-bin /path/to/go1.27.1` when the Go
executable on PATH differs. The default uses deterministic candidate order.
Pass `--model /path/to/shared-qat/model.json` to use the
[own compact model](https://huggingface.co/asketeddy/gooo-record-shared-field-tiny-v1/tree/5800946afb35506d66357ee3ea6f956f506795b9/models/qat_ternary)
to rank the same declared choices. It loads once per composition and predicts
once per supported assembling activity. The Gooo policy has a fixed body and
runs without model calls after each scored attempt.

The diagnostic activity has five construction cases and fifteen field
expectations. This example supplies four additional runtime expectations,
including an integer above 2^53. Keep those denominators separate. Inspect
`composition.steps[].generation.report.record_assembly.control` for the policy
source, generated-code identity and each input, operation and continuation
decision. Each decision is evaluated before the following candidate is built.

## Replay the saved construction

```sh
/tmp/gooo-assembly-policy body-compose \
  --source examples/package-diagnostic-replay/diagnostics.gooo.fixture \
  --cases examples/assembly-policy/cases.json \
  --composition /tmp/gooo-policy-construction/composition.json
```

Replay reconstructs the saved policy and decisions, checks the selected source,
then builds and runs the saved program twice. It makes zero new model calls.
The policy and model paths are generation options; the receipt carries the
policy needed for replay. Retain the original target source alongside it.

## Retain a failed construction

Run the first command with `--source examples/assembly-policy/unsolved.gooo.fixture`
and a different output directory. This source keeps the five obligations and
eight-attempt budget, but neither declared code value can produce the required
`partial` result. The run retains the unresolved cases and native mismatches.
This is a different candidate space and is a failure demonstration, separate
from the paired deterministic/model comparison of the original source.

## Meaning of the operations

| Gooo result | Construction action |
| --- | --- |
| `CONTINUE_CANDIDATES`, `EVALUATE_CANDIDATES` | Try the next unattempted candidate in the retained ordering, while the source budget remains |
| `USE_OBSERVED_CANDIDATE` | Finish with the observed candidate matching the most whole cases; break ties by matched fields, then first observation |
| `OBSERVE_NEW_INPUTS`, `EXPAND_DECLARED_CHOICES`, `DECLARE_CASES`, `RECONCILE_COUNTS` | Finish the current search and retain its best observed candidate and outstanding work |

The latter operations describe follow-up work. This command constructs within
the existing source; new input collection or grammar changes are separate
operations. An unknown operation produces an explicit error. A complete finite
candidate or exhausted source budget also ends the search. A type-rejected
combination consumes an attempt and reduces the remaining scored-case budget;
it receives no invented case score or policy input.

Policy messages and states are retained as the Gooo tool's judgments. Finite
completion is computed from the actual candidate cases and fields. A policy can
stop on a partial candidate. All attempts, including an unsuccessful initial
model proposal, remain in the construction record. The model's score describes
its ranking, with task correctness measured separately.

This adapter currently applies to source-declared record choices. It is an
explicit `body-compose` option. Other assembly profiles keep their existing
construction behavior. Unit and native checks compare the interpreted Gooo
policy outputs with its compiled Go projection and cover early stop, an
unsolved space, type rejection, policy replacement, and saved replay.

## Recorded own-model use

The [October 8 observation](../../docs/research/assembly-policy-20261008/summary.json)
retains raw outputs from clean compiler `1335874bc5d113f79572d64be23efe70f3a61d2c`.
Deterministic ordering tried four candidates. The own model proposed mask 7,
matching 3/5 cases and 13/15 fields. Gooo returned `CONTINUE_CANDIDATES`; the
second candidate, mask 3, matched 5/5 and 15/15. Both paths generated the same
program and matched four separate native expectations. Saved replay matched
4/4 with zero new model calls. The model's two policy decisions also matched
2/2 outputs when the policy was compiled and executed as a native Gooo program.

The unsolved variant exhausted eight attempts, retained 3/5 cases and 13/15
fields, and matched 2/4 native expectations. Gooo returned
`EXPAND_DECLARED_CHOICES`. The command exited zero because the observation
completed; matched/total and finite status describe the remaining work.

Prediction took 20,208 ns; model setup took 0.705583 ms with 2,096 bytes of
resident tensors. Whole-command times were 0.55 s deterministic and 0.34 s
model-guided; maximum RSS was approximately 86.0 MB and 87.2 MB respectively.
These are single sequential observations including Go compilation and native
execution with uncontrolled caches. Model-only CPU utilization and a general
speed advantage remain unmeasured. The model is unchanged and its training
exposure to these task families remains unknown.
