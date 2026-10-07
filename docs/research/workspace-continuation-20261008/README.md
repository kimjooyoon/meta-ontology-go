# Saved package continuation: October 8 observation

Clean compiler `e411aafeaf674e61d6998197f6d80b94d155d0f1`, Go 1.27.1,
Darwin arm64. [Build](build.json), [raw summary](summary.json), and
[commands](../../../examples/called-body-construction/README.md#continue-a-saved-package-checkpoint).
This is source after the published 0.6.6-dev binary; build the pinned revision.

The diagnostic app imports its constructing helper, which imports a rule
package. A separate two-package Gooo policy stops after its first candidate.
`package resume` then consumes that receipt and an explicit continuing policy.

| Route | Checkpoint native cases | Continued native cases | Continuation work | Replay |
| --- | --- | --- | --- | --- |
| Deterministic | 2/4 | 4/4 | Retain 1, add 3 | 4/4, zero new predictions |
| Own compact model | 2/4 | 4/4 | Retain 1, add 1 | 4/4, zero new predictions |

The helper's construction cases progress from 3/5 to 5/5, and fields from
13/15 to 15/15. Four distinct native root inputs are disjoint from the helper's
declared construction cases. Both routes produce the same final generated
program. Historical policy packages and model ranking survive continuation and
replay; continuation itself makes zero new predictions. Input disjointness from
these cases does not establish independence from model training data.

The unchanged [own model](https://huggingface.co/asketeddy/gooo-record-shared-field-tiny-v1/tree/5800946afb35506d66357ee3ea6f956f506795b9/models/qat_ternary)
ranked mask 7 first, then retained search selected mask 3. Prediction took
16,083 ns, setup 0.775917 ms, with 2,096 bytes of decoded tensors. The model
checkpoint command took 0.52 s wall, 0.17 s user and 0.10 s system CPU with
86,999,040 bytes maximum RSS. Continuation and replay took 0.37 and 0.34 s.
Those are whole commands including compilation and execution. CPU times give
a rounded 52% average of one core for the checkpoint, including child work.
Host utilization and the model's isolated CPU increment were not sampled.
Each row is one sequential local observation with uncontrolled caches.

## Development checks and remaining boundaries

The CLI regression first failed because `package resume` did not exist. It now
passes an imported-helper checkpoint through continuation and replay, removes
the old policy files, and checks that a configured local provider receives zero
requests. Related package tests and a focused race run pass locally.

Regression coverage includes nested dependent helpers, repeated rounds, an
initial built-in policy, source-body fills retained before record construction,
fresh input-only replay, cancellation, and changed or missing policy histories.
The scalar body-fill parser now accepts coexisting record declarations through
the existing body-profile parser; the scalar fill signature remains Integer.

Only saved record-choice construction is continued. Earlier fills are replayed
unchanged; source fill/search within the composition requires further work.
Policy history is bounded to 16 stages and candidate budgets remain cumulative.
Parent digests refer to earlier artifacts; complete artifact retention is the
caller’s responsibility. Calls during nested candidate scoring still have
unobserved input history. Broader correctness, external adoption, and a general
speed advantage remain unmeasured.

## 한국어

이번에는 파일 하나에서 하던 조립 이어가기를 여러 Gooo 패키지에 연결했습니다.
설명 도구가 만든 중간 결과를 저장한 뒤, 다른 Gooo 정책으로 남은 선택지를
확인합니다. 앞서 만든 부품과 당시 정책의 원본도 함께 보관합니다.
모델을 다시 부르지 않고 앱 사례 2/4에서 4/4로 이어갔으며, 재실행도 같은
네 사례를 충족했습니다. 이 수치는 위에 공개한 작은 진단 예제의 관측입니다.
