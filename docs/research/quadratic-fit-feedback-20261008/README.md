# One shared expression across ten inputs

Clean compiler `14717427b49ddd7562503e25ada41245e0577262`, Go 1.27.1,
Darwin arm64. [Summary](summary.json) · [native run](refinement.json) ·
[replay](replay.json) · [hashes](sha256.json) ·
[source and command](../../../examples/quadratic-fit-feedback/README.md).

Increasing the construction cases to ten exposed a candidate-order limitation.
Each input needs a different inner value before it is squared. The v1 grammar
fills its retained prefix with individual constants and omits a shared affine
expression. A regression first reproduced 1/10 construction cases and 0/2
holdouts despite the correct expression existing later in the grammar.

The new v2 grammar keeps the same candidate universe. It first lists root-derived
constants and affine expressions compatible with all available integral probe
roots, then the remaining v1 expressions in their original order. Compatibility
comes from construction inputs only. The original typed body still scores every
attempted candidate.

## Native source-policy run

| Observation | v1 round | v2 round |
| --- | --- | --- |
| Enumerated expressions | 108 | 108 |
| Retained candidate cap | 16 | 16 |
| Per-round attempt allowance | 8 | 8 |
| Actually attempted | 8 | 1 |
| Feedback outputs matched | 1/10 | 10/10 |
| Candidate-space coverage | 16/108 (14.81%) | 16/108 (14.81%) |
| Whole grammar retained | No | No |
| Selected hole expression | `-3` | `input * -2 + -1` |
| Gooo decision | Advance to declared `fit` | Stop and retain |

Each round observes 30 probe evaluations: -1, 0 and 1 for each of ten inputs.
The selected factor's square supplies the required adjustment. After selection,
three separate inputs match 3/3; all three are disjoint under the native
input-separation check. Saved replay again matches 3/3 with zero model calls.

Functional case coverage and candidate-space coverage remain distinct. The
runtime satisfies the supplied expectations while the grammar receipt continues
to report truncation. These results describe one finite program and its cases.
They do not establish broader natural-language accuracy or whole-domain behavior.

## Compatibility and checks

The published v1 grammar retains its original order. A native CLI regression
loads the earlier published quadratic record and replays its 3/3 result with
zero model calls under the new compiler code.

Checks also verify deterministic order, unchanged candidate membership, absent
and contradictory root observations, exact overflow handling, held-out input
isolation and a cubic counterexample. The cubic example's fitted roots are
compatible with the probe observations but fail full-body scoring and remain
partial. Race checks passed in the six affected compiler packages; both grammar
examples preserve their meaning through the source/semantic round trip.

The native record contains source, Gooo policy, input cases, each round's chosen
grammar, candidate observations and compiled execution. No model training or
downloads occurred. The clean compiler identity is recorded in the summary.

## 한국어

사례가 열 개가 되자 후보 목록의 앞부분이 개별 숫자로 채워졌습니다. 여러 입력에서
함께 쓸 수 있는 식은 뒤로 밀려 빠졌고, 열 개 중 하나만 맞았습니다.

새 버전은 관측한 조건들을 함께 만족할 수 있는 식부터 후보에 담습니다. Gooo 정책이
그 버전으로 넘어가자 같은 후보 한도에서 10/10을 맞혔습니다. 마지막에 따로 실행한
세 입력도 통과했습니다. 후보 전체 108개 중 16개를 담았다는 사실은 그대로 남깁니다.
기능이 어느 사례까지 맞는지와 탐색을 얼마나 했는지를 함께 보는 실험입니다.
