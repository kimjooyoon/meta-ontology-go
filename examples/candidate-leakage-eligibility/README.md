# Candidate leakage eligibility

This stage consumes the exact merged candidate-leakage shadow and language
assurance capsules. It can project one possible transition, but it cannot apply
that transition.

| Case | Decision | Resolution | Effect |
| --- | --- | --- | --- |
| Exact merged capsules | `ELIGIBLE` | `EXACT` | `NO_EFFECT` |
| Capsule unavailable | `FAIL_CLOSED` | `UNKNOWN` | `BLOCK` |
| Capsule digest mismatch | `FAIL_CLOSED` | `INVARIANT_ONLY` | `BLOCK` |

The exact case projects `7/12 -> 8/12` and `5833 -> 6666 bps`. The official
language assurance remains `7/12`, repository writes remain zero, and promotion
applied remains zero. A later PR must consume the merged eligibility receipt to
perform any operating transition.
