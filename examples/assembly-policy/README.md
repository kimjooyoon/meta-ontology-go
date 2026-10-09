# Let a Gooo tool control the next construction attempt

The existing [Gooo explainer](../assembly-explainer/main.gooo.fixture) can now run
between record-construction attempts. After a candidate is type-checked and
scored, its Gooo conditions decide whether to try the next permitted candidate
or finish with an observed candidate. The compiler still owns the declared
alternatives, attempt budget, type checks and finite measurements.

## Construct and execute

Build this revision with Go 1.27.2, then run from the repository root:

```sh
go build -o /tmp/gooo-assembly-policy ./cmd/gooo
/tmp/gooo-assembly-policy body-compose \
  --source examples/package-diagnostic-replay/diagnostics.gooo.fixture \
  --cases examples/assembly-policy/cases.json \
  --assembly-policy examples/assembly-explainer/main.gooo.fixture \
  --policy-activity Explain --out /tmp/gooo-policy-construction
```

The output directory must be new. Add `--go-bin /path/to/go1.27.2` when the Go
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

See [continuing a saved construction](#continue-a-saved-partial-construction)
to apply these operations to an existing attempt history.

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

## Continue a saved partial construction

Use the source-owned checkpoint policy to finish after the first scored candidate:

```sh
/tmp/gooo-assembly-policy body-compose \
  --source examples/package-diagnostic-replay/diagnostics.gooo.fixture \
  --cases examples/assembly-policy/cases.json \
  --assembly-policy examples/assembly-policy/checkpoint.gooo.fixture \
  --policy-activity Checkpoint --out /tmp/gooo-checkpoint

/tmp/gooo-assembly-policy body-compose \
  --source examples/package-diagnostic-replay/diagnostics.gooo.fixture \
  --cases examples/assembly-policy/cases.json \
  --resume-composition /tmp/gooo-checkpoint/composition.json \
  --assembly-policy examples/assembly-explainer/main.gooo.fixture \
  --policy-activity Explain --out /tmp/gooo-continued
```

The first command may include the optional own-model path. The second reads the
saved ranking and uses no model file or inference. Gooo evaluates the new policy
against the retained prefix before it considers another candidate. The original
alternatives, finite cases and total attempt budget remain fixed. Exhausting that
budget leaves the same partial result; changing the source starts a new experiment.

The continuation verifies prior attempts and policy decisions, then records
`control_history`, the new policy's `entry` decision, and retained/added attempt
counts. Reconstruction evaluates historical candidates again; those evaluations
are validation work, separate from newly considered candidates. Original model
calls and prediction timing stay historical, while `continuation.new_model_calls`
reports zero for the new operation. A subsequent `--composition` replay checks
the complete saved history and runs the updated native program.

Each composition carries a reference to its parent's canonical digest. Keep the
parent file for comparison: this reference alone cannot establish its origin.
History is bounded to 16 prior control stages per record activity. Continuation
currently accepts graphs whose assembling activities all use record choices;
ordinary bound activities are regenerated from source. Other assembly profiles
return an explicit unsupported result. This is an explicit CLI continuation;
the compiler does not schedule follow-up invocations itself.

[Called record helpers](../dependent-continuation/README.md) also resume in
dependency order. A changed helper causes retained caller masks to be rechecked
before the new policy reads their scores. `rechecked_attempts` distinguishes this
work from additional candidates; source-bound history reconstructs the original
observations. The declaration's budget still counts distinct attempted masks.

### 한국어: 남은 조립을 이어가기

`Checkpoint`는 첫 후보의 관측을 저장하고, `Explain`은 그 기록을 읽어 다음
후보를 시도할지 판단합니다. 작업대에 부품과 조립 기록을 남겨 두는 방식입니다.
다시 시작할 때 모델을 불러오지 않아도 기존 후보 순서로 작업을 계속할 수 있습니다.
앞서 쓴 예산과 실패한 후보도 그대로 계산합니다. 선언한 범위에 답이 없는 경우에는
그 부족한 부분을 결과에 남기며, 새 부품을 추가하는 일은 별도의 소스 변경입니다.

### Recorded continuation

The [continuation study](../../docs/research/assembly-resume-20261008/summary.json)
uses clean compiler `f23bc3c29df5e7869549347f6d5952917385bee5` and the unchanged
own QAT model. Its first proposal matched 3/5 construction cases and 13/15 fields.
A separate process continued without a model path, retained the first attempt,
added one candidate and reached 5/5 cases, 15/15 fields and 4/4 separate native
expectations. Deterministic continuation retained one attempt and added three to
reach the same generated program. Saved continued-program replay matched 4/4
with zero new model calls.

The unsolved variant retained 3/5 cases, 13/15 fields and 2/4 native expectations
after eight cumulative attempts. Another continuation retained all eight and
added zero. The same total budget applied across process boundaries.

The initial prediction took 18,250 ns; setup took 0.594 ms and resident tensors
occupied 2,096 bytes. Model-free continuation took 0.40 s wall, 0.16 s user and
0.11 s system time, with 86,818,816 bytes maximum RSS for the measured command.
This includes historical validation, Go compilation and native execution. Host
CPU utilization and the model's CPU increment were not sampled. These single
sequential runs establish a reproducible example; broader performance and model
training exposure remain open measurements.
