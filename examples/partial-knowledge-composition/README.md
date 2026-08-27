# Partial-knowledge composition

This read-only experiment treats the five `computes` strings in
[`main.gooo`](main.gooo) as the source of truth. Each string is a structured
upstream observation receipt containing `required`, `observed`,
`observed_available`, `dependency_claim_id`, `prior_state`, and invariant
evidence. No conclusion state is supplied by a JSON fixture.

The producer and independent verifier both execute:

```text
syntax.ParseFile -> bidir.Lower -> observation receipt reconstruction
```

The calculus derives `EXACT`, `DIRECT_UNKNOWN`, `DEPENDENCY_BLOCKED`,
`INVARIANT_ONLY`, or `MIXED_UNRESOLVED` from those fields. The outcomes are:

| Case | Resolution | Decision | Claim transition |
|---|---|---|---|
| exact-pair | `EXACT` | `PASS` | `OPEN -> DISCHARGED` |
| direct-unknown | `LOWER_RESOLUTION` | `UNKNOWN` | `OPEN -> OPEN` |
| dependency-blocked | `LOWER_RESOLUTION` | `UNKNOWN` | `OPEN -> OPEN` |
| invariant-preservation | `INVARIANT_ONLY` | `HOLD` | `OPEN -> OPEN` |
| mixed-unknown-and-blocked | `LOWER_RESOLUTION` | `UNKNOWN` | `OPEN -> OPEN` |

The receipt-level decision is `CALCULUS_PROVEN`; it is separate from the
subject resolution and does not promote non-exact cases. The fixed denominator
is 5, with 1 exact case, 4 open claims, 10 indicators, 0 repository writes,
and `promotion_authorized=false`.

GitHub Actions also runs a semantic intervention that makes the direct
observation available and must change its result and claim transition. A
comment-only source intervention must preserve the semantic IR and semantic
projection while changing only the source digest. Its human-readable report
and JSON evidence are uploaded under the exact head SHA.
