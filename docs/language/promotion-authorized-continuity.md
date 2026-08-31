# Promotion-authorized continuity

## Purpose

This receipt answers one bounded question: after a failed-closed observation is
recovered at an exact zero-effect fixed point, does the next successful merged
generation return to the ordinary `AUTHORIZED` path?

It does not increase the language-readiness numerator. The fixed contract stays
at `12/24` until a separate language or toolchain obligation is implemented.
This is a conformance proof, not a quantified improvement claim.

## Executable use case

1. A merged transformation run emits a v1-compatible proposal receipt.
2. The next merged CI run completes successfully for one exact head SHA.
3. The transformation workflow emits an `AUTHORIZED / EXACT` guard receipt.
4. Its recovery join emits `PASS / EXACT / PROMOTION_AUTHORIZED`.
5. A continuity job in that same workflow consumes both exact-run artifacts.
6. The evaluator, independent of both producers, runs twice and must emit
   byte-identical receipts.

The observer writes only to its Actions artifact directory. It grants no
repository mutation authority and performs no source-tree mutation.
It adds zero workflow files: composing one job into the existing transformation
workflow keeps workflow entry cardinality unchanged, while repository projection
remains the authority for `storage.direct-entry = 0`.

## Fixed indicator registry

| Class | Metric | Target |
| --- | --- | ---: |
| OUTCOME | `gooo.metric.language.promotion-continuity-readiness-bps.v1` | 10000 |
| DRIVER | `gooo.metric.language.promotion-continuity-authorized-guards.v1` | 1 |
| DRIVER | `gooo.metric.language.promotion-continuity-authorized-routes.v1` | 1 |
| GUARDRAIL | `gooo.metric.language.promotion-continuity-unresolved.guardrail.v1` | 0 |
| GUARDRAIL | `gooo.metric.language.promotion-continuity-effects.guardrail.v1` | 0 |
| GUARDRAIL | `gooo.metric.language.promotion-continuity-writes.guardrail.v1` | 0 |
| GUARDRAIL | `gooo.metric.language.promotion-continuity-authority.guardrail.v1` | 0 |
| GUARDRAIL | `gooo.metric.language.promotion-continuity-source-mutations.guardrail.v1` | 0 |
| OUTCOME | `gooo.metric.language.promotion-continuity-terminal-preserved.v1` | 1 when the mixed terminal is applicable |

The evaluator's continuity summary has eight coordinates (`summary.total=8`).
Separately, all nine indicator definitions are emitted by
`promotioncontinuity.Evaluate` and consumed by `self-improvement-cycle`. The ninth
indicator is bound to the released `PreserveNonPromotingTerminal` activity and is
applicable only to the exact mixed terminal; it never grants promotion. Mixed
terminal coordinates and proofs use terminal-neutral names, while the two
authorization indicators remain unsatisfied. Unknown schema,
subject, decision, resolution, effect, write, or authority evidence produces
`FAIL_CLOSED / LOWER_RESOLUTION`; default zero values cannot authorize a route.

## Munchhausen choices

| Choice | Exact role |
| --- | --- |
| FOUNDATION | Bind versioned guard and recovery receipts to one 40-hex head SHA. |
| COHERENCE | Require `AUTHORIZED` and `PROMOTION_AUTHORIZED` to describe the same generation. |
| REGRESSION | Reject effects, source mutation, repository writes, or mutation authority. |

## Interpretation boundary

A passing receipt proves only this repository's versioned CI transition. It is
not a claim that the mechanism is unique among programming languages, and it is
not evidence of commercial demand or production reliability.
