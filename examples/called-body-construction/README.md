# Construct a called Gooo body before its caller

`Main` calls `diagnostics.Diagnose(...)` directly. `Diagnose` owns three
record-field choices and five finite construction cases. The compiler constructs
that helper first, then compiles `Main` against the selected callable body.
The helper's own fixed predicates are imported from `tools/rules`.

This uses the same diagnostic logic as the
[binding example](../package-body-calls/README.md). Here, the caller supplies the
helper arguments. Only `Main` has external inputs and an independently observed
activity output. A helper can be called with different arguments more than once
without rebuilding its body.

## Run and inspect

From this revision with Go 1.27.2:

```sh
go build -o /tmp/gooo-called-body ./cmd/gooo
/tmp/gooo-called-body package execute --json \
  --cases examples/called-body-construction/cases.json \
  --assembly-policy-workspace examples/package-assembly-policy/gooo.workspace.json \
  examples/called-body-construction/gooo.workspace.json > called-body.json
```

Add `--assembly-model /path/to/shared-qat/model.json` to rank the helper's choices
with the [own compact model](https://huggingface.co/asketeddy/gooo-record-shared-field-tiny-v1/tree/5800946afb35506d66357ee3ea6f956f506795b9/models/qat_ternary).
Omitting the model selects candidates deterministically. The policy flag is also
optional; when supplied, the separate Gooo policy controls the helper's search
using its actual finite observations. Use `--go /path/to/go1.27.2` when required.

Inspect these parts of the result:

| Field | Meaning |
| --- | --- |
| `composition.plan.preparations` | Called assembling activities in dependency order |
| `composition.preparations` | Each helper's exact input source, choices, attempts, finite scores and selected checkpoint |
| `composition.steps` | Projections of independently executed graph activities |
| `composition.gooo_source` | Selected executable source, with constructed helpers exposed as fixed callable bodies |
| `runtime` | Fresh native outputs and their supplied expectations |

Selection contracts remain in the original source and preparation records. The
executable helper has a fixed body so another activity can type-check and call it.
Partial construction scores stay partial in the preparation record. Each helper
is constructed once per composition, including when it is both a bound producer
and a directly called function.

## Replay and explain

```sh
/tmp/gooo-called-body package replay --json \
  --receipt called-body.json \
  --cases examples/called-body-construction/cases.json \
  examples/called-body-construction/gooo.workspace.json

/tmp/gooo-called-body package execute --json \
  --construction-receipt called-body.json \
  examples/assembly-explainer/gooo.workspace.json
```

Replay reconstructs preparations and caller projections in order, then performs
two fresh native executions without new inference. The explainer reads each
helper's verified attempt history; a helper that is also a bound producer is
counted once. Historical model calls remain in the original preparation.

Use `--inputs .../inputs.json` in place of `--cases` for unscored observations.
Input-only output establishes actual values and replay agreement without adding
an accuracy percentage.

## Bounds and measurement scope

The call plan includes dependencies from computes bodies, baselines and declared
field alternatives. Recursive dependencies are rejected. Up to 32 called bodies
can be constructed; the existing pure-call type, nesting and invocation bounds
still apply to every emitted closure. A selected composition entry determines
the externally executed graph. Standalone `body-compose --entry <name>` uses the
same construction mechanism.

Record choices, source-body fills and source IR search use their existing
generators. Independent bodies are checked before loading a model; bodies that
depend on constructed helpers are checked after those helpers are realized and
before their own selection. The package `--assembly-model` route guides supported
choice profiles. `body-compose --fill-model` supports called source fills; the
package `--tiny-model` adapter still targets its earlier workspace fill stage.

Fresh runtime observations now retain constructed helper arguments in
`traces[].calls`. The four inputs here are disjoint from the diagnostic helper's
five declared selection inputs. Input separation counts whole root cases and
compares the actual helper arguments, including int64 values above 2^53.
Construction case coverage and native output checks retain their own denominators.
Earlier published receipts describe their original compiler's narrower scope.

When one assembling helper calls another during candidate scoring, that earlier
argument history is still unobserved. Fresh inputs in such constructions remain
`UNKNOWN / CONSTRUCTION_CALL_INPUTS_NOT_OBSERVED`. A skipped helper call also
earns no new-input credit. See the [measurement contract](../../docs/native-body-composition.md#input-separation).

Each helper is selected against its own declared cases. A later caller failure
does not automatically reopen an earlier helper choice. The native regression
suite includes an ambiguous helper that satisfies 1/1 local cases while its caller
matches 0/1 expectations. Retaining those two results exposes where stronger
obligations or a future search across dependent choices is needed.

Saved record-choice construction can continue in dependency order. When a helper
changes, dependent candidates are checked with the changed helper, while their
historical scores remain bound to their original source. Earlier workspace body
fills are replayed unchanged. Continuing a source fill or source IR search inside
the composition remains a separate extension.

## Continue a saved package checkpoint

Build from this revision, then stop after the first diagnostic candidate:

```sh
/tmp/gooo-called-body package execute --json \
  --cases examples/called-body-construction/cases.json \
  --assembly-policy-workspace examples/package-assembly-policy/checkpoint.workspace.json \
  examples/called-body-construction/gooo.workspace.json > checkpoint.json

/tmp/gooo-called-body package resume --json --receipt checkpoint.json \
  --cases examples/called-body-construction/cases.json \
  --assembly-policy-workspace examples/package-assembly-policy/gooo.workspace.json \
  examples/called-body-construction/gooo.workspace.json > continued.json

/tmp/gooo-called-body package replay --json --receipt continued.json \
  --cases examples/called-body-construction/cases.json \
  examples/called-body-construction/gooo.workspace.json
```

The four finite caller expectations progress from 2/4 to 4/4. The diagnostic's
separate construction cases progress from 3/5 to 5/5. An optional
`--assembly-model /path/to/shared-qat/model.json` belongs on the initial command.
Resume keeps that ranking, the original candidate budget, and every prior policy
workspace. It accepts no inference option and makes zero new predictions. The
new policy must be explicit; no policy or source is silently substituted.

The receipt's `continued_from_sha256` identifies the consumed envelope.
`result.assembly_policy_history` retains source packages for each earlier stage,
including a null stage for the initial built-in rule. At most 16 saved stages are
accepted. `result.continuation` counts replayed fills and new model calls;
`result.composition.continuation.activities` separates retained, rechecked and
added attempts. Each resume/replay performs two fresh native executions. Use
`--inputs` instead of `--cases` for observations without an accuracy score.

The current workspace must still match the saved source. Historical policy files
need not remain on disk because their exact packages are in the receipt. Parent
digests name earlier artifacts; keep those artifacts if a complete chain is
needed. They are not signatures or a remote archive.

## 한국어: 필요한 부품부터 만들어 연결하기

앱은 진단 함수를 호출하고, 진단 함수는 자신의 선택지와 사례로 본문을 조립합니다.
컴파일러는 필요한 부품부터 순서대로 만든 뒤 호출하는 쪽을 연결합니다. 같은
함수를 여러 곳에서 불러도 부품을 매번 다시 조립하지 않습니다.

각 부품에서 무엇을 시도했고 몇 사례를 만족했는지 기록이 남습니다. 최종 앱의
실행 결과와 부품의 조립 점수는 따로 읽습니다. 실행 중 보조 함수에 전달된 인자를
기록하므로, 새 앱 입력이 기존 부품 사례와 겹치는지도 확인할 수 있습니다.
다른 부품의 후보를 평가하는 동안 발생한 간접 호출 이력은 아직 기록되지 않아
그 경로의 입력 독립성은 미확인으로 남습니다. 실제로 관측한 범위부터 조금씩
넓히는 메타프로그래밍 경로입니다.

## Recorded own-model construction

The [raw study](../../docs/research/called-body-construction-20261008/summary.json)
pins clean compiler `01f71d405d11bf1f0157a04f1791eb4347314676` and the unchanged
own QAT model. It constructs one called diagnostic body, then projects one
independently executed caller.

| Route | Helper attempts | Helper cases | Fields | Caller expectations | New model calls |
| --- | ---: | --- | --- | --- | ---: |
| Deterministic | 4 | 5/5 | 15/15 | 4/4 | 0 |
| Own model | 2 | 5/5 | 15/15 | 4/4 | 1 |
| Saved replay | 2 retained | 5/5 retained | 15/15 retained | 4/4 | 0 |
| Own model + checkpoint policy | 1 | 3/5 | 13/15 | 2/4 | 1 |
| Saved checkpoint replay | 1 retained | 3/5 retained | 13/15 retained | 2/4 | 0 |

The Gooo explainer separately interpreted the model's two helper attempts as
`CONTINUE_CANDIDATES` and `OBSERVE_NEW_INPUTS`, with zero new model calls.
Input-only replay returned three observations with 0/0 expectations. In that
original study, input separation remained unknown because its compiler did not
trace helper arguments. The complete deterministic, model and replay routes generated the
same program.

Prediction took 16,584 ns; setup took 0.733292 ms with 2,096 bytes of tensors.
The whole model command took 0.35 s wall, 0.17 s user and 0.11 s system time,
with 86,818,816 bytes maximum RSS. Rounded CPU/wall totals correspond to 80% of
one core including child work; host utilization and the model's separate CPU
increment were not sampled. These are single sequential local observations
after tests, with uncontrolled caches. Training exposure and external adoption
remain unmeasured.

### Fresh called-input observations

The [follow-up study](../../docs/research/called-input-observation-20261008/summary.json)
uses clean compiler `d5130460356646f964d01a619f0db62f166f0ea5` and the same model,
source, choices and expectations. Each root case records one actual helper call.

| Route | Helper attempts | Caller expectations | Disjoint root cases passed | New model calls | Whole command wall time |
| --- | ---: | --- | --- | ---: | ---: |
| Deterministic | 4 | 4/4 | 4/4 | 0 | 0.34 s |
| Own model | 2 | 4/4 | 4/4 | 1 | 0.36 s |
| Saved replay | 2 retained | 4/4 | 4/4 | 0 | 0.32 s |
| Prior study receipt replay | 2 retained | 4/4 | 4/4 | 0 | 0.33 s |
| Own model + checkpoint policy | 1 | 2/4 | 2/4 | 1 | 0.54 s |

The prior receipt replays without changing its construction. Its new runtime
observation supplies the previously missing helper arguments. Input-only replay
observes one overlapping and two disjoint root inputs, while its accuracy stays
unknown because no expectations were supplied. The compiler's regression cases
also cover a new root that becomes an old helper input, a skipped call, repeated
and nested calls, conflicting duplicate expectations, record inputs and a caller
failure with `0/1` disjoint cases passed.

This model invocation spent 12,042 ns in prediction and 0.734625 ms in setup,
with 2,096 bytes of tensors. Whole-command maximum RSS was 86,720,512 bytes;
0.16 s user plus 0.11 s system over 0.36 s wall gives a rounded average of 75%
of one core, including child work. These single measurements retain uncontrolled
host activity and caches. Host CPU utilization, model-only CPU increment,
training independence and external reproduction remain unmeasured.
