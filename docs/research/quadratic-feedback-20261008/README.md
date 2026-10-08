# Gooo changes the grammar for a repeated hole value

Observed with clean compiler `942d3e642499d07ffad947eef9380e28c1f3c291`,
Go 1.27.1, Darwin arm64. [Summary](summary.json) · [full run](refinement.json) ·
[saved replay](replay.json) · [hashes](sha256.json) ·
[source and command](../../../examples/quadratic-feedback/README.md).

The activity keeps negative inputs unchanged and adds the square of a constructed
value to other inputs. The source contains an `if`, two local declarations, a
return expression, and a declared alternative grammar. The existing Gooo policy
observes the failed construction and chooses that alternative.

| Stage | Grammar | Feedback outputs | Gooo decision |
| --- | --- | --- | --- |
| First round | `integer-hole-residual/v1` | 1/3 | Advance to the declared quadratic setting |
| Second round | `integer-hole-quadratic/v1` | 3/3 | Stop and retain the program |
| Post-selection evaluation | Selected program | 3/3 evaluation outputs | No further source decisions |
| Saved replay | Retained source/composition | 3/3 evaluation outputs | Zero new model calls |

The selected inner value is `-3`; its square is nine. The grammar also proposes
`3`. The receipt preserves the three probe outputs and both integer roots.
The final three input tuples are distinct from recorded construction/probe inputs
under the native input-separation check. They are evaluated after source selection.

The two probe grammars share the original body and expectations. The Gooo policy
changes only the active grammar and its declared bounds. The negative branch
already satisfies its expectation in the first round. The squared adjustment
accounts for the two initially unmatched outputs.

## Checks and limits

Race checks passed in assemblyspec, syntax, bidir, bodycodegen, bodyrefinement
and the CLI package. Focused native CLI and semantic round-trip checks passed.
The exact integer root calculation was checked over 350 coefficient/root
combinations and explicit large-number, non-integral and insensitive cases.
Those are algebra checks, separate from this one language execution experiment.

Additional body tests cover input-dependent fitted values, local variables,
negative curvature, inactive branches, omitted candidates, unavailable probes,
cancellation, held-out input isolation and altered receipts. A square required
to equal five and a cubic example remain partial after whole-body scoring.
The three-point fit supplies proposals within a bounded grammar; finite execution
results determine the reported outcome.

This run uses deterministic construction and Gooo decisions. No model training
or downloads occurred. The raw record embeds the source, policy, cases, candidate
observations, generated code and native results. Reproduction creates a fresh
output directory; this publication keeps the aggregate records to limit disk use.

## 한국어

빈칸에 들어갈 값을 제곱해서 사용하는 예제입니다. 처음 방법은 맞는 값을 찾지 못해
세 실행 사례 중 하나만 맞혔습니다. Gooo가 소스에 적어둔 다른 문법으로 바꾸자
나머지 두 사례도 맞혔고, 마지막에 따로 실행한 세 입력도 통과했습니다.

새 문법은 빈칸에 -1, 0, 1을 넣어 본 결과로 후보를 만듭니다. 실제 본문을 실행해
후보를 검사하며, 정수 답을 찾지 못한 경우는 그 상태를 남깁니다. 이번에는 작은
모델의 도움 없이 이 조립 경로를 실행했습니다.
