# Semantic delta receipt

This independent philosophy experiment treats a source change as three
different objects:

1. `textual_delta`: the raw byte change between two `.gooo` files;
2. `structural_delta`: the change in normalized nodes and directed facts; and
3. `semantic_claim_delta`: the change in stable claims, including explicit
   claim transitions.

It is a Gooo meta-operation, not a second line-oriented diff. The operation is
`separate-text-structural-semantic-deltas`. `semanticdeltareceipt.ProduceFiles`
is the producer; `semanticdeltareceiptconsumer.AdjudicateFiles` is a separate
consumer and independent judge. Neither path writes the repository.

## Research and adopted principles

The design is grounded in three primary sources:

- Microsoft Research's [Abstract Semantic Diffing of Evolving Concurrent
  Programs](https://www.microsoft.com/en-us/research/publication/abstract-semantic-diffing-evolving-concurrent-programs/)
  treats semantic diffing as comparison of program abstractions and behavioral
  differences, not just changed lines.
- LLVM's [Program Analysis for Compiler
  Validation](https://llvm.org/pubs/2008-11-PASTE-CompilerValidation.html)
  describes translation validation as a validation pass after each compiler
  run, proving the produced target is a correct translation of that source.
- The CompCert [semantic-preservation
  theorem](https://compcert.org/man/manual001.html) states the desired boundary
  over observable source and target meaning, while allowing compilation to
  fail rather than inventing a result.

The experiment adopts these rules:

| Principle | Implementation |
| --- | --- |
| Syntax and semantics are different observations | byte digests never decide semantic class |
| Validation is per change pair | each receipt is rebuilt and adjudicated from its two raw sources |
| Meaning is anchored to stable identities | nodes and claims use immutable IDs, not display order |
| Approximation is not exact proof | unsupported syntax becomes `INDETERMINATE / FAIL_CLOSED / LOWER_RESOLUTION` |
| Evidence is read-only | CI derives tracked-plus-untracked workspace writes and combines that empty set with reported source/output paths |

The raw decision is recorded independently as `RAW_CHANGED` or
`RAW_FIXED_POINT`; the semantic decision is separately recorded as
`SEMANTIC_PRESERVED`, `SEMANTIC_CHANGED`, or `SEMANTIC_UNKNOWN`. An unknown
subject is `FAIL_CLOSED / LOWER_RESOLUTION` at stage `bind-subject`, step
`resolve-subject`, with reason `SEMANTIC_DELTA_SUBJECT_UNKNOWN`. A parse or
lowering failure is likewise `FAIL_CLOSED / LOWER_RESOLUTION` at stage
`project-source`, step `parse-lower`, with reason
`SEMANTIC_TRANSLATION_VALIDATION_UNAVAILABLE`; there is no separate
`resolution=UNKNOWN` state.

The experiment rejects line-count equality as a semantic proof, raw text equality
as equivalence, and a semantic-diff approximation promoted to `EXACT`. It also
does not claim whole-program behavioral equivalence or compiler correctness.

## Fixed denominator and cases

The denominator is fixed at three cases and is never widened by the evaluator:

| Case | Text | Structure | Claims | Result |
| --- | --- | --- | --- | --- |
| `equivalent` | changed | empty | empty | `SEMANTIC_PRESERVED / FIXED_POINT / EXACT` |
| `semantic-change` | changed | one node plus one relation replacement | one changed claim | `SEMANTIC_CHANGED / DELTA_OBSERVED / EXACT` |
| `indeterminate` | changed | unavailable | unavailable | `INDETERMINATE / FAIL_CLOSED / LOWER_RESOLUTION` |

The suite reports `3/3 = 10000` basis points. Its fixed counts are textual
changes `3`, structural observations `3`, semantic preservation `1`, semantic
change `1`, indeterminate `1`, and repository writes `0`.

## Decision rule

For a pair that both bounded projections can parse:

```text
textual_delta.changed && structural_delta == empty && semantic_claim_delta == empty
  => SEMANTIC_PRESERVED / FIXED_POINT / EXACT

textual_delta.changed && (structural_delta != empty || semantic_claim_delta != empty)
  => SEMANTIC_CHANGED / DELTA_OBSERVED / EXACT
```

If either projection cannot parse, the result is `INDETERMINATE` and
`FAIL_CLOSED / LOWER_RESOLUTION` at the exact unavailable coordinates above. A receipt is accepted only if the independent
adjudicator recomputes the text, structure, claims, transitions, digests, and
classification from raw bytes. A digest or field mismatch is
`INVARIANT_ONLY / FAIL_CLOSED`.

## Proof choices and claim transitions

Every indicator carries its `producer`, `consumer`, `meta_operation`, `stage`,
`step`, `reason`, and `proof_choice`:

- `FOUNDATION` binds raw bytes, stable identities, and the canonical graph;
- `COHERENCE` checks that the graph, claim delta, transition, and independent
  verdict agree; and
- `REGRESSION` checks the no-write boundary.

A transition records the claim ID, claim kind, status before and after, object
before and after, stage, step, reason, and (for preservation rows) the
preserved object claim ID. Object propositions and cross-version preservation
propositions are separate persistent rows. Persistent statuses are `OPEN`,
`DISCHARGED`, and `REFUTED`:

- bounded semantic equivalence is `OPEN -> DISCHARGED` when preserved,
  `OPEN -> REFUTED` when changed, and `OPEN -> OPEN` when indeterminate;
- every before-object preservation row is discharged when the same proposition
  remains and refuted when it changes or disappears;
- after-only object propositions are source observations and are discharged by
  the canonical observation; they are never refuted merely because they are
  new.

Object claim IDs are digests of their propositions, preservation IDs are
digests of the preservation proposition, and the receipt retains both old rows
and their transitions. The semantic-change case therefore refutes the old
payment preservation proposition and discharges the new reversal observation;
it does not claim that the reversal proposition is false.

The conformance suite's `FIXED_POINT` decision means only that the fixed
three-case contract was reproduced. `subject_semantic_equivalence` is recorded
separately as `NOT_ASSERTED`.

## Falsifiability

The result is falsifiable. Reordering declarations or adding comments should
keep the first case semantically preserved. Changing a stable ID, output entity,
or supported relation must produce a non-empty structural or claim delta.
Putting an unsupported declaration into the source must remain indeterminate.
Mutating one field of a receipt without resealing it must be rejected by the
independent judge. These tests are intentionally narrower than behavioral
equivalence: the bounded projector does not cover runtime values, macros,
external dependencies, or arbitrary future Gooo syntax.
