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
digests, and either one replay witness or an explicit missing-witness marker.
The judge recomputes the candidate, postcondition, and replay digests; it does
not merely trust status strings or a resealed authority decision.

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
therefore proves only the chosen fixed invariant and bounded cases. It does
not claim a verified refactoring engine, complete semantic equivalence,
toolchain correctness, repository mutation, or promotion authority. The
negative cases make the claim falsifiable: changing the semantic digest makes
the postcondition `REFUTED`, and removing the replay witness leaves it
`OPEN`.
