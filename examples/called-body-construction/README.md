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

From this revision with Go 1.27.1:

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
using its actual finite observations. Use `--go /path/to/go1.27.1` when required.

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

The runtime traces graph activities. It does not yet retain every helper-call
argument tuple. Consequently, input separation reports
`UNKNOWN / CALLED_ASSEMBLY_INPUTS_NOT_OBSERVED`, even when graph outputs pass.
New root inputs may map to previously observed helper inputs. Construction case
coverage and native output checks retain their own denominators.

Each helper is selected against its own declared cases. A later caller failure
does not automatically reopen an earlier helper choice. The native regression
suite includes an ambiguous helper that satisfies 1/1 local cases while its caller
matches 0/1 expectations. Retaining those two results exposes where stronger
obligations or a future search across dependent choices is needed.

Saved continuation currently rejects compositions with called-body preparations:
changing a helper may affect later caller constructions and requires a new
construction. Ordinary saved replay is supported. The next extension should
define which dependent constructions can be retained when a helper changes.

## 한국어: 필요한 부품부터 만들어 연결하기

앱은 진단 함수를 호출하고, 진단 함수는 자신의 선택지와 사례로 본문을 조립합니다.
컴파일러는 필요한 부품부터 순서대로 만든 뒤 호출하는 쪽을 연결합니다. 같은
함수를 여러 곳에서 불러도 부품을 매번 다시 조립하지 않습니다.

각 부품에서 무엇을 시도했고 몇 사례를 만족했는지 기록이 남습니다. 최종 앱의
실행 결과와 부품의 조립 점수는 따로 읽습니다. 아직 보조 함수에 전달된 모든
인자를 기록하지 않으므로, 새로운 입력에서도 잘 되는지에 대한 지표는 미확인으로
남깁니다. 실제로 관측한 범위부터 조금씩 넓히는 메타프로그래밍 경로입니다.

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
Input-only replay returned three observations with 0/0 expectations. All called
construction runs retain unknown input separation because helper arguments are
not traced. The complete deterministic, model and replay routes generated the
same program.

Prediction took 16,584 ns; setup took 0.733292 ms with 2,096 bytes of tensors.
The whole model command took 0.35 s wall, 0.17 s user and 0.11 s system time,
with 86,818,816 bytes maximum RSS. Rounded CPU/wall totals correspond to 80% of
one core including child work; host utilization and the model's separate CPU
increment were not sampled. These are single sequential local observations
after tests, with uncontrolled caches. Training exposure and external adoption
remain unmeasured.
