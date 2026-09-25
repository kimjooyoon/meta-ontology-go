# Rollback integrity eligibility

This read-only stage consumes three exact capsules produced for merged commit
`1f27d342faf7a435ca4c534e2a816f29befe21a4`: official language assurance and
two independently replayed rollback-integrity shadow reports. Both reports
bind the same seven-case rollback meta-program, including the explicit
`FAIL_CLOSED / LOWER_RESOLUTION` unknown-decision case.

| Case | Decision | Resolution | Effect |
| --- | --- | --- | --- |
| Three exact merged capsules | `ELIGIBLE` | `EXACT` | `NO_EFFECT` |
| Capsule unavailable | `FAIL_CLOSED` | `UNKNOWN` | `BLOCK` |
| Capsule digest mismatch | `FAIL_CLOSED` | `INVARIANT_ONLY` | `BLOCK` |
| Evidence subject reused as candidate | `FAIL_CLOSED` | `UNKNOWN` | `BLOCK` |

The exact case projects `9/12 -> 10/12` and `7500 -> 8333 bps`. Official
language assurance remains `9/12`; repository writes and applied promotions
remain zero. A later activation must consume the merged eligibility receipt.
