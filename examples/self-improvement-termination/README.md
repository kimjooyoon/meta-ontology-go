# Self-improvement termination witness

This experiment is a source-backed proof experiment for a repeatedly executed
self-improvement meta-operation. The bounded traces are executable
`termination-case/v2` values in [`main.gooo`](main.gooo). The producer parses
and lowers that Gooo source canonically; JSON files are generated receipts and
judge outputs, never input authority.

The source binds the producer `selfimprovementtermination.Evaluate`, consumer
`self-improvement-cycle`, meta-operation `prove-self-improvement-termination`,
proof choice `TERMINATION`, stage `META_RUN`, and every step/reason. Each state
is a content digest. Adjacent no-change stuttering is removed before looking
for a repeated state, but a final no-change observation is the only fixed-point
witness.

| source case | trace / budget | state count | repeated states / period | decision | resolution | claim transition | termination proven |
| --- | ---: | ---: | ---: | --- | --- | --- | --- |
| `fixed-point` | 1/4 | 1 | 0 / 0 | `FIXED_POINT` | `EXACT` | `OPEN -> DISCHARGED` | yes |
| `cycle-2` | 2/4 | 3 | 1 / 2 | `CYCLE` | `EXACT` | `OPEN -> REFUTED` | no |
| `in-progress` | 2/4 | 3 | 0 / 0 | `IN_PROGRESS` | `LOWER_RESOLUTION` | `OPEN -> OPEN` | no |
| `divergence-possible` | 2/2 | 3 | 0 / 0 | `DIVERGENCE_POSSIBLE` | `LOWER_RESOLUTION` | `OPEN -> OPEN` | no |
| `unknown-upstream` | 1/4 | 1 | 0 / 0 | `FAIL_CLOSED` | `LOWER_RESOLUTION` | `OPEN -> OPEN` | no |

`CYCLE` is an exact witnessed cycle, not a successful termination result.
`IN_PROGRESS` and `DIVERGENCE_POSSIBLE` preserve their bounded observation and
exact coordinates while remaining open. An unrecognized upstream decision is
localized as `FAIL_CLOSED` with reason
`FEEDBACK_COVERAGE_DECISION_UNKNOWN`; it can never become a fixed point merely
because a trace contains no change.

Each generated receipt carries independently recomputed baseline and
intervened source digest, lowered semantic digest, decision, resolution, and
claim-transition evidence. The semantic intervention rewrites the selected
`trace`, `upstream`, and, when needed, `max_steps` fields in the executable
computes value, then parses, lowers, and classifies that altered source again:

| source case | baseline | semantic intervention | comment-only intervention |
| --- | --- | --- | --- |
| `fixed-point` | `FIXED_POINT/EXACT`, `OPEN -> DISCHARGED` | `IN_PROGRESS/LOWER_RESOLUTION`, `OPEN -> OPEN` | preserved exactly |
| `cycle-2` | `CYCLE/EXACT`, `OPEN -> REFUTED` | `IN_PROGRESS/LOWER_RESOLUTION`, `OPEN -> OPEN` | preserved exactly |
| `in-progress` | `IN_PROGRESS/LOWER_RESOLUTION`, `OPEN -> OPEN` | `DIVERGENCE_POSSIBLE/LOWER_RESOLUTION`, `OPEN -> OPEN` | preserved exactly |
| `divergence-possible` | `DIVERGENCE_POSSIBLE/LOWER_RESOLUTION`, `OPEN -> OPEN` | `IN_PROGRESS/LOWER_RESOLUTION`, `OPEN -> OPEN` | preserved exactly |
| `unknown-upstream` | `FAIL_CLOSED/LOWER_RESOLUTION`, `OPEN -> OPEN` | `IN_PROGRESS/LOWER_RESOLUTION`, `OPEN -> OPEN` | preserved exactly |

The fixed conformance denominator is two, without aggregation: one semantic
trace intervention must change the subject decision or resolution and produce
a canonical `OPEN` claim transition; one comment-only intervention must change
only the raw source digest while preserving semantic digest, outcome, and claim
transitions. This is a source-causality check, not a termination score and
never means “termination proven.” If a semantic digest changes without a
subject outcome change, the result is `FAIL_CLOSED/LOWER_RESOLUTION` with
`DIGEST_ONLY_BINDING`, claim `OPEN`, and the semantic indicator is not
satisfied. The receipt also binds the source digest, lowered semantic digest,
selected computes-value digest, trace digest, producer/consumer/meta-operation/
proof choice, and read-only authority.

The independent judge owns a copied wire model and independently parses,
lowers, recomputes, and seals the expected receipt. It imports zero packages
from the producer implementation. A forged fixed-point receipt, an altered
trace, a changed source digest, or a changed claim transition is rejected.

## Formal design choices

The experiment uses these sources:

1. [Baader and Nipkow, *Term Rewriting and All That*, Chapter 5:
   Termination](https://www.cambridge.org/core/books/abs/term-rewriting-and-all-that/termination/36ACECF2B0D53A6933FA6EDAF7E219FF)
   motivates treating termination as a property of a rewrite relation and
   using a well-founded reduction order rather than inferring termination from
   a finite sample.
2. [Lean 4 Reference, Recursive Definitions](https://lean-lang.org/doc/reference/latest/Definitions/Recursive-Definitions/)
   requires a well-founded recursion argument, such as a decreasing measure,
   for recursive calls; a bounded trace here is therefore only evidence at a
   boundary unless it contains an observed fixed point.
3. [Cornell CS 4110, Fixed Points](https://www.cs.cornell.edu/courses/cs4110/2014fa/lectures/slides07.pdf)
   defines a fixed point by `F(x) = x`; this equality is adopted as the local
   no-change witness but not confused with the claim that iteration reaches it.

Adopted principles:

- Parse and lower the executable Gooo source before producing any case input;
  generated JSON has no authority.
- Accept `FIXED_POINT/EXACT` only for an explicit observed final `NO_CHANGE`
  self-transition after cycle analysis.
- Preserve a contiguous state chain, fixed step budget, rank coordinates, and
  localized stage/step/reason evidence.
- Treat a non-adjacent repeated state as an exact `CYCLE`; its fixed-point
  claim is `REFUTED`, not discharged.
- Keep bounded progress, possible divergence, and unknown upstream coverage
  as `OPEN` with `LOWER_RESOLUTION`; only the explicit fixed point discharges.
- Keep semantic trace and nonsemantic comment interventions in a fixed 2-case
  conformance corpus, separate from the subject outcome.

Rejected shortcuts:

- A run limit is not a proof of infinite divergence or termination.
- An upstream `FIXED_POINT` label is not trusted without the observed digest
  equality and no-change reason.
- A repeated state is not automatically a fixed point; a period-2 witness is
  a cycle.
- A 10/10-style aggregate or a conformance score cannot be presented as a
  termination proof.
- Importing the producer package into the verifier is not independent
  evidence, so the judge duplicates the wire model and computation.
- A fixed-point theorem for an abstract recursive functional does not show
  that this concrete self-improvement workflow reaches that point.

The source is registered in the syntax, semantic, toolchain, and vertical-slice
corpora. The workflow runs Go 1.27, produces receipts from the source, replays
them, runs the independent judge, checks the exact transition matrix, and
asserts `semantic_causal_cases=1/1` and
`nonsemantic_preservation_cases=1/1`. It rejects both a forged receipt and a
forged receipt whose digest fields have been resealed. It performs no
repository writes and does not merge or promote the claim.
