# Reuse a Gooo decision inside another body

Gooo can call a fixed, pure activity declared in the same source. The
[diagnostic example](main.gooo.fixture) gives a small rule a name:

```text
activity IsPartial(Integer, Integer) -> Boolean computes
    `return input1 > 0 && input0 >= 0 && input0 < input1`
```

`Diagnose` uses `IsPartial(input0, input1)` inside its condition. Its three
record-field alternatives still belong to the Gooo source. The optional own
model ranks those alternatives; the helper's meaning remains defined by its
typed body. The [assembly policy](policy.gooo.fixture) also reuses a Gooo
`CanContinue` activity to decide whether more candidates can be considered.

## Run one entry and its helpers

From this revision with Go 1.27.1:

```sh
go build -o /tmp/gooo-calls ./cmd/gooo
/tmp/gooo-calls body-compose \
  --source examples/pure-activity-calls/main.gooo.fixture \
  --entry Diagnose \
  --cases examples/assembly-policy/cases.json \
  --assembly-policy examples/pure-activity-calls/policy.gooo.fixture \
  --policy-activity Explain --out /tmp/gooo-calls-program
```

The output directory must be new. Use `--go-bin /path/to/go1.27.1` if needed.
Add `--model /path/to/shared-qat/model.json` for the existing
[own compact model](https://huggingface.co/asketeddy/gooo-record-shared-field-tiny-v1/tree/5800946afb35506d66357ee3ea6f956f506795b9/models/qat_ternary).
Without that option, selection follows deterministic ordering.

`--entry Diagnose` executes that activity and its explicit `bind` producers.
`IsPartial` receives its arguments from the calling expression. It requires no
separate external input. The entry is stored in the composition plan. Replay
uses that saved entry and the saved construction:

```sh
/tmp/gooo-calls body-compose \
  --source examples/pure-activity-calls/main.gooo.fixture \
  --cases examples/assembly-policy/cases.json \
  --composition /tmp/gooo-calls-program/composition.json
```

Checkpoint and continuation use the [existing saved-construction
workflow](../assembly-policy/README.md#continue-a-saved-partial-construction).
Supply `--entry` when first constructing the program. Replay and continuation
take the entry from the retained plan. Both reconstruct the callee identities
and bodies from the supplied source without new model inference.

## Meaning and measurement

Arguments are evaluated in source order and checked against the callee's declared
input types. Calls work in conditions, local initializers, assignments, record
fields and returns. Boolean `&&`/`||` keep their short-circuit behavior. Each call
has its own local bindings; scalar and supported record inputs retain value
semantics. The callee must return one value on every path.

`generation.report.call_closure` records transitive callee IDs and program
digests, call edges with ordinals within each caller, and a conservative call
count bound. The route-equivalence digest includes the typed callee bodies.
With calls present, source/lowered semantic-unit counts cover the root and each
distinct helper once. Finite case and field percentages retain their own
denominators. A call-count bound includes both branches syntactically; it is an
upper bound, not a measured invocation count or a claim about elapsed time.

The closure accepts up to 32 activities including the root, 16 nested calls and
4,096 calls per root invocation. The compiler rejects recursive cycles, unknown
or indirect callees, external calls, and callees that still have an `assembling`
contract. Current bodies use 1..16 scalar or supported record inputs and one
result. This profile reuses fixed activities from one source; package-qualified
body calls and model construction inside a callee need additional language work.
Existing source-IR fill/search profiles retain their own expression constraints.

## 한국어: 판단에 이름을 붙여 재사용하기

`IsPartial`은 “일부만 완료됐는가”를 계산하는 작은 Gooo 활동입니다. 다른
본문의 조건식에서 호출할 수 있고, 조립 정책도 `CanContinue`라는 판단을
재사용합니다. 자주 쓰는 공구에 이름을 붙여 작업대에 놓는 방식입니다.

이름뿐 아니라 실제 본문, 입력·출력 타입, 안정 ID와 호출 관계를 함께 기록합니다.
모델은 선언된 코드 후보의 순서를 제안하고, 재사용하는 판단의 의미는 Gooo에
남습니다. 실행 진입점을 지정하면 보조 활동에 따로 입력을 제공할 필요가 없습니다.
다음 단계는 이 호출 관계를 패키지 경계까지 연결하고, 실제 도구에서 반복되는
판단을 라이브러리로 옮기는 것입니다.
