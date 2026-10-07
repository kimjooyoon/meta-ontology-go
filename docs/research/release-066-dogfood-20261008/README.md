# Gooo 0.6.6 development candidate: construct, call and explain

This local observation uses clean compiler
`da09721eaf4cf835a6c620b5794fe3b95d3eb490`, Go 1.27.1 and the unchanged
[own compact model](https://huggingface.co/asketeddy/gooo-record-shared-field-tiny-v1/tree/5800946afb35506d66357ee3ea6f956f506795b9/models/qat_ternary).
It was recorded before publication of the release. The [summary](summary.json)
binds the raw receipts, resource observations, generated starter and source files.

## What ran

| Route | Candidate attempts | Construction cases | Native expectations | New model calls | Whole command wall |
| --- | ---: | --- | --- | ---: | ---: |
| New diagnostic starter, deterministic | 4 | 5/5 | 8/8 | 0 | 0.53 s |
| Same starter, own model | 2 | 5/5 | 8/8 | 1 | 0.34 s |
| Starter replay on three unscored inputs | 2 retained | 5/5 retained | 0/0 | 0 | 0.29 s |
| Called helper, deterministic | 4 | 5/5 | 4/4 | 0 | 0.61 s |
| Same helper and policy, own model | 2 | 5/5 | 4/4 | 1 | 0.34 s |
| Prior compiler's saved construction replay | 2 retained | 5/5 retained | 4/4 | 0 | 0.28 s |
| Own model with early-stop policy | 1 | 3/5 | 2/4 | 1 | 0.51 s |
| Gooo explanation of that partial record | — | — | 0/0 | 0 | 0.90 s |

Each deterministic/model pair generated the same executable source and outputs.
The historical replay also reconstructed the same called-helper program. Four
root inputs were disjoint from the recorded construction inputs in each scored
program. Input-only starter replay observed one overlapping and two disjoint
inputs; it carries no accuracy score.

The early-stop result preserves a helper with 3/5 cases and 13/15 fields matched,
then observes 2/4 caller expectations matched. The Gooo explainer returns
`PROGRESS / CONTINUE_CANDIDATES`. This invocation only explains the record.
Saved continuation with called-body preparations still requires implementation;
a caller currently starts a fresh construction to revisit those choices.

## Follow the path

Build the compiler at the revision above. Set `GOOO` to its executable,
`GO` to a matching Go 1.27.1 executable, and `MODEL` to the linked model metadata.
From the repository root, use a fresh output directory:

```sh
"$GOOO" init --template diagnostic --module example.org/release066/diagnostic my-diagnostic
"$GOOO" package execute --json --go "$GO" --assembly-model "$MODEL" \
  --cases my-diagnostic/cases.json my-diagnostic/gooo.workspace.json > starter-model.json
"$GOOO" package replay --json --go "$GO" --receipt starter-model.json \
  --inputs my-diagnostic/inputs.json my-diagnostic/gooo.workspace.json
"$GOOO" package execute --json --go "$GO" --assembly-model "$MODEL" \
  --assembly-policy-workspace examples/package-assembly-policy/checkpoint.workspace.json \
  --cases examples/called-body-construction/cases.json \
  examples/called-body-construction/gooo.workspace.json > called-checkpoint.json
"$GOOO" package execute --json --go "$GO" --construction-receipt called-checkpoint.json \
  examples/assembly-explainer/gooo.workspace.json
```

Omit `--assembly-model` for deterministic ordering. Replace the checkpoint
policy with `examples/package-assembly-policy/gooo.workspace.json` to let this
task continue to the complete finite candidate. The [called-body guide](../../../examples/called-body-construction/README.md)
explains saved replay, helper arguments and the remaining dependency boundary.

## Resource and measurement scope

Starter prediction took 15,875 ns and setup 0.4165 ms. Called-helper prediction
took 76,084 ns and setup 0.168666 ms. Both loaded 2,096 bytes of tensors. Whole
model commands peaked at 86,179,840 and 87,326,720 bytes RSS respectively.
Rounded user-plus-system CPU time divided by wall time gives 73.53% and 79.41%
of one core, including child work. Host CPU was not sampled; the model's separate
CPU increment is unmeasured.

These are eight sequential local workload invocations after focused tests, with
uncontrolled caches and host work. This run's timing difference is a local
observation. The prior paired study was slightly slower with model ranking.
Neither establishes a general speed advantage. Model weights were unchanged.
Source cases were authored by the maintainer; model-training exposure, nested
candidate-scoring argument histories and external reproduction remain unknown.
The command caller started each operation; no candidate edits or interactive
approvals occurred inside them.

## 한국어: 만든 도구로 덜 완성된 도구를 설명하기

진단 도구의 조건과 선택지를 Gooo로 선언하고, 작은 모델로 시도 순서를 정했습니다.
같은 선언을 모델 없이도 조립했으며, 두 경로는 같은 결과를 만들었습니다.
부품의 본문을 만든 뒤 앱이 그 부품을 호출하는 경로도 확인했습니다.

조립을 중간에 멈추면 부품의 사례는 3/5, 앱의 기대 출력은 2/4가 맞습니다.
Gooo로 만든 설명 도구가 이 기록을 읽고 다음 후보 탐색을 제안합니다.
공방에서 부품 검사표와 완성품 작동 기록을 함께 보며 다음 작업을 정하는 과정과
비슷합니다. 보조 함수를 바꿀 때 필요한 호출자 재조립과 이어하기가 다음 언어 과제입니다.
