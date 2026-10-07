# Build a diagnostic tool from Gooo library functions

Three packages share a small diagnostic tool:

- `tools/rules` defines `IsPartial` and its local helper `BelowTotal`.
- `tools/diagnostics` calls `rules.IsPartial(input0, input1)` inside `Diagnose`.
  Its three record-field choices and five construction cases remain in Gooo.
- `app/explain` receives the diagnostic through an explicit `bind` and renders
  a message. Cases name the two executed graph activities. Called helpers
  receive arguments from the calling expressions.

This builds on the [diagnostic replay example](../package-diagnostic-replay/README.md)
and the [fixed pure activity profile](../pure-activity-calls/README.md).

The [package assembly policy](../package-assembly-policy/README.md) uses another
Gooo workspace to decide whether these construction attempts should continue.
Its own decision helper is imported from a reusable package.

## Construct, execute and keep the result

From this compiler revision with Go 1.27.1:

```sh
go build -o /tmp/gooo-package-calls ./cmd/gooo
/tmp/gooo-package-calls package execute --json \
  --cases examples/package-body-calls/cases.json \
  examples/package-body-calls/gooo.workspace.json > package-calls.json
```

Add `--assembly-model /path/to/shared-qat/model.json` for the optional
[own compact model](https://huggingface.co/asketeddy/gooo-record-shared-field-tiny-v1/tree/5800946afb35506d66357ee3ea6f956f506795b9/models/qat_ternary).
Without it, candidate ordering is deterministic. Use `--go /path/to/go1.27.1`
when the Go binary on PATH differs.

Replay the saved construction on fresh inputs without supplying a model:

```sh
/tmp/gooo-package-calls package replay --receipt package-calls.json \
  --inputs examples/package-body-calls/inputs.json \
  examples/package-body-calls/gooo.workspace.json
```

For repeated finite checks, replace `--inputs` with `--cases .../cases.json`.
The source workspace and saved receipt can move together; source paths are
relative to the workspace manifest. A changed helper source requires a new
construction record. Replay reconstructs the call closure and performs fresh
native executions.

## What a call means

`rules.IsPartial(...)` resolves through the import alias in the calling source
file. An unqualified name refers to the local package, including its other source
files. The compiler includes reachable fixed helpers, checks their argument and
result types, and lowers their calls to unique executable names. Local variables
and string contents keep their meaning. Record constructors and signatures retain
the package's entity identities, including when two packages use the same type
name.

The call uses explicit argument values; the binding delivers a graph activity's
result into another activity's port. Only the entry and its explicit bind
producers require case inputs and have independent graph results. Helpers can
appear in conditions, local computations, returns, record fields and declared
record-field alternatives.

`result.program.pure_calls` records each participating activity's original package,
name, namespace-derived semantic ID, source body digest and lowered name/ID.
Call sites identify the source surface and byte range of the callee expression.
These offsets refer to the decoded `computes`, baseline or candidate expression.
Package source digests identify the containing files. Sites include calls in
declared alternatives; a selected body's `call_closure` records the helpers it
actually includes.

The compiler checks the same pure-call bounds as standalone generation. Cycles,
indirect calls, unknown imports, shadowed call names and still-assembling callees
are rejected. Imported helpers have fixed Gooo bodies. Source-IR fill/search keep
their existing narrower expression profiles. Finite case results and projection
unit preservation remain separate measurements.

## 한국어: 작은 판단을 도구 상자에 나누어 담기

`tools/rules`는 “일부만 충족됐는가”라는 판단을 보관합니다. 진단 도구는 그
판단을 호출하고, 앱은 진단 결과를 문장으로 보여줍니다. 각 패키지는 자신의
역할을 Gooo로 표현하며, 모델은 선언된 조립 후보의 순서를 제안합니다.

함수가 어느 파일에서 왔고 실행할 때 어떤 이름이 됐는지 함께 기록합니다.
다른 도구에서도 같은 판단을 가져다 쓸 수 있고, 저장된 결과를 다시 실행할 때
원본 함수가 바뀌었는지도 확인할 수 있습니다.

## Recorded own-model use

The [raw study](../../docs/research/package-body-calls-20261008/summary.json)
uses clean compiler `a0af6fd8bb01da787e893a888dd54f82aaf58844` and the unchanged
own compact model. The three packages contain two independently executed graph
activities and two pure helper calls.

| Operation | Construction cases | Fields | Attempts | Native expectations | New inference calls |
| --- | --- | --- | --- | --- | --- |
| Deterministic | 5/5 | 15/15 | 4 | 8/8 | 0 |
| Own compact model | 5/5 | 15/15 | 2 | 8/8 | 1 |
| Saved replay | 5/5 retained | 15/15 retained | 2 retained | 8/8 | 0 |
| Input-only replay | 5/5 retained | 15/15 retained | 2 retained | 0/0, no expectations | 0 |

The model's first candidate matched 3/5 construction cases; the second matched
5/5. Deterministic and model-guided construction produced the same program.
The eight native expectations cover two outputs for each of four unique input
rows. Those rows are disjoint from the recorded construction inputs; model
training exposure remains unknown. Input-only replay reports `OBSERVED` and
preserves the historical construction measurements without inventing a current
accuracy percentage.

Prediction took 19,292 ns, setup 0.779 ms and resident tensors 2,096 bytes. The
whole model command took 0.32 s wall, 0.16 s user and 0.10 s system time, with
86,425,600 bytes maximum RSS. Native compilation and execution are included.
Local tests had finished before these sequential observations; host activity and
caches were uncontrolled. Host CPU utilization and the model's CPU increment were
not sampled. This small task establishes the recorded behavior and reuse path;
broader performance and external adoption remain open measurements.
