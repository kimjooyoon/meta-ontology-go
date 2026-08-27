# Invariant-preserving transformation authority

This experiment makes transformation permission conditional on four Gooo meta
values:

```text
precondition ∧ transformation ∧ postcondition ∧ regression_witness
```

The conjunction is an authority predicate, not a generic refactoring command.
The source contract is [the checked-in Gooo value model](../../examples/invariant-transformation/main.gooo).
Its producer emits a receipt containing the four values, their
`producer/consumer/meta_operation/proof_choice` bindings, and a
`stage/step/reason` coordinate. A separate judge derives the decision from the
receipt and the fixed denominator. The receipt also carries a replayable
evidence tuple: source digest, candidate digest, before/after semantic
digests, expected semantic digest, and either one replay witness or an
explicit missing-witness marker. Each fixed case is an actual `computes` value
in the Gooo source: an `int64` input, an `add:n` candidate, an expected value,
an invariant relation, replay availability, and effect policy. The producer
and judge each parse and execute that declaration; the judge recomputes the
candidate, postcondition, and replay digests from the source and receipt. It
does not merely trust status strings or a resealed authority decision.

## State and effect model

Each claim starts `OPEN` and records one explicit transition to `DISCHARGED`,
`REFUTED`, or `OPEN`. `DISCHARGED` means the required evidence is present and
bound. `REFUTED` carries a counterexample. `OPEN` means the evidence is absent;
it never becomes authorization by optimism or by a green producer result.

The four-case denominator is deliberately mixed: one preserved translation,
one semantic violation, one missing regression witness, and one approved
artifact. The last case records one
`APPROVED_ARTIFACT_RECORDED` effect, while every case keeps
`repository_writes=0` and `mutation_authority=false`. An approved product is a
separate data effect; it does not grant permission to edit the repository or
promote a branch.

The current exact receipt is therefore `4/4` cases and `10000` basis points,
with `2` authorized cases, `1` refuted case, and `1` open case. Across the
`16` claim values, `13` are discharged, `2` refuted, and `1` open. These are
conformance counts, not a claim that the violation or evidence gap is a
successful transformation.

## Research basis and limits

The design follows the per-translation viewpoint of [George C. Necula,
“Translation Validation for an Optimizing Compiler,” PLDI
2000](https://dl.acm.org/doi/10.1145/349299.349314): check the concrete output
of each transformation and use an explicit simulation-style witness where
possible. It also follows [Nik Sultana and Simon Thompson, “Mechanical
Verification of Refactorings”](https://kar.kent.ac.uk/23959/), which formalizes
refactorings in Isabelle/HOL and makes behavior preservation the correctness
condition.

The experiment is smaller and weaker than either line of work. Program
equivalence is undecidable; finite evidence cannot prove arbitrary programs or
all effects. A validator can be incomplete or report false alarms when it
cannot explain a transformation, and validation has runtime cost. This model
therefore makes only a synthetic, bounded claim over four explicitly declared
integer fixtures; it does not authorize arbitrary transformations or claim a
verified refactoring engine, complete semantic equivalence, toolchain
correctness, repository mutation, or promotion authority. The negative cases
make the claim falsifiable: changing the `.gooo` expected value changes the
independently recomputed result to `REFUTED`, and removing the replay witness
leaves it `OPEN`.

## Intervention separation

The experiment also publishes a separate fixed two-case intervention report;
its measurements are not aggregated into the authority-suite score. The
semantic slice has denominator `1`: changing the preserved fixture's
`expected=3` to `expected=4` changes the parsed projection, receipt, and
independent decision from `AUTHORIZED` to `REFUTED` with reason
`SEMANTIC_POSTCONDITION_REFUTED`. The non-semantic slice also has denominator
`1`: adding only whitespace and a comment changes the raw `SourceDigest` and
receipt digest, but preserves the parsed/lowered fixture projection, decision,
resolution, reason, and claim transitions. Both slices require zero repository
writes.

The two intervention claims are bound to persistent `OPEN -> DISCHARGED`
transitions. The semantic claim uses stage `INTERVENTION`, step
`compare-semantic-projection-and-decision`, and reason
`SEMANTIC_PROJECTION_AND_DECISION_CHANGED`; the non-semantic claim uses the
same stage, step `compare-nonsemantic-projection-and-decision`, and reason
`NONSEMANTIC_PROJECTION_AND_DECISION_PRESERVED`.

The report top-level decision, resolution, reason, repository writes, and
mutation authority are derived from those adjudicated cases and both source
variants. An observed contradiction is `FAIL_CLOSED` with `REFUTED`; an
unobservable obligation is `FAIL_CLOSED` with `OPEN` at lower resolution. The
report consumer verifies these relationships without calling `Build`; the
separately named `DeterministicReplay` uses a second `Build` only for
repeatability and is not presented as independent evidence. CI also tampers a
case outcome and proves that a resealed `PASS`/`DISCHARGED` artifact is
rejected.
