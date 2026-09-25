# Rollback integrity shadow

This example does not add a second rollback algorithm. It binds the fixed
language-assurance denominator to the existing
`rollbackfixedpoint.Build` meta evaluator.

## Fixed denominator

| Measure | Current | Shadow projection |
|---|---:|---:|
| Operating obligations | 9/12 | 10/12 |
| Coverage | 7500 bps | 8333 bps |
| Executable cases | 0/7 | 7/7 |
| Existing meta reports validated | 0/7 | 7/7 |
| Existing coordinates observed | 0 | 70 |
| Repository writes | 0 | 0 |
| Promotions applied | 0 | 0 |

The projection has `NO_EFFECT`; it cannot change the official denominator.

## Use cases

| Case | Required result |
|---|---|
| fail-closed guard plus exact fixed point | `PASS / RECOVERED_FIXED_POINT` |
| already authorized guard | `PASS / PROMOTION_AUTHORIZED` |
| unknown guard decision | `FAIL_CLOSED / LOWER_RESOLUTION` |
| transformation effect | `FAIL_CLOSED / EXACT` |
| source mutation | `FAIL_CLOSED / EXACT` |
| observer write | `FAIL_CLOSED / EXACT` |
| mutation authority | `FAIL_CLOSED / EXACT` |

## Munchhausen branches

- `FOUNDATION`: exact merged assurance and replay-valid rollback reports
- `COHERENCE`: guard, ledger, subject, and terminal path agree
- `REGRESSION`: effects, writes, source mutation, and authority stay rejected

No novelty claim is made. The useful property is that an existing operational
concept becomes measurable in the same registry that decides language progress.
