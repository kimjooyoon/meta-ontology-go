# Changed-surface receipt eligibility

This stage consumes three exact capsules from merged commit
`25ee2f076d67eafceec4eac3ff4d85a90abb597f`: official language assurance,
the changed-surface receipt meta report, and its fixed six-case suite.

| Case | Decision | Resolution | Effect |
| --- | --- | --- | --- |
| Three exact merged capsules | `ELIGIBLE` | `EXACT` | `NO_EFFECT` |
| Capsule unavailable | `FAIL_CLOSED` | `UNKNOWN` | `BLOCK` |
| Capsule digest mismatch | `FAIL_CLOSED` | `INVARIANT_ONLY` | `BLOCK` |

The exact case projects `8/12 -> 9/12` and `6666 -> 7500 bps`. Official
language assurance remains `8/12`; repository writes and applied promotions
remain zero. A later activation must consume the merged eligibility receipt.
