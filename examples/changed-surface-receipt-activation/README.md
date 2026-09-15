# Changed-surface receipt activation

This stage consumes the exact merged assurance and eligibility capsules from
`1b9acfaac0ff2d1a2353de6c5019d515ab542eb3`. It is the only stage that can add
`totalize-changed-surface-receipts` to the official operating operation set.

| Case | Decision | Resolution | Transition |
| --- | --- | --- | --- |
| Exact predecessor | `APPLIED` | `EXACT` | `1` |
| Evidence unavailable | `FAIL_CLOSED` | `UNKNOWN` | `0` |
| Capsule digest mismatch | `FAIL_CLOSED` | `INVARIANT_ONLY` | `0` |

The exact transition changes official assurance from `8/12 = 6666bps` to
`9/12 = 7500bps`. Every case reports zero repository writes. Unknown or
invariant-only evidence cannot activate the metric.
