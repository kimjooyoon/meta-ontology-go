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
`stage/step/reason` coordinate. The receipt contains separate baseline and
replay input/operation/output/digest observations. Each fixed case is an
actual `computes` value in the Gooo source: an `int64` input, an `add:n`
candidate, an expected value, an invariant relation, a replay recipe/capability,
and an effect policy. The producer and judge parse and execute the candidate
and recipe independently; a replay label alone is never evidence. The judge
also recomputes the approved artifact bytes from the source and checks their
`RUNNER_TEMP` path, size, and digest.

## State and effect model

Each claim starts `OPEN` and records one explicit transition to `DISCHARGED`,
`REFUTED`, or `OPEN`. `DISCHARGED` means the required evidence is present and
bound. `REFUTED` carries a counterexample. `OPEN` means the evidence is absent;
it never becomes authorization by optimism or by a green producer result.

The four-case denominator is deliberately mixed: one preserved translation,
one semantic violation, one missing replay recipe, and one approved artifact.
The last case creates actual bytes under `RUNNER_TEMP` and records one
`APPROVED_ARTIFACT_RECORDED` effect, while every case keeps
`repository_writes=0` and `mutation_authority=false`. `AUTHORIZED` is scoped
to the semantic transformation receipt or temporary artifact emission; it does
not grant permission to edit the repository or promote a branch.

The current exact receipt is `4/4` cases and `10000` basis points, with `2`
authorized cases, `1` refuted case, and `1` open case. Across the `16` claim
values, `13` are discharged, `2` refuted, and `1` open. These are conformance
counts, not a claim that the violation or evidence gap is a successful
transformation.

## Research basis and limits

The design follows the per-translation viewpoint of [George C. Necula,
“Translation Validation for an Optimizing Compiler,” PLDI 2000](https://dl.acm.org/doi/10.1145/349299.349314):
check the concrete output of each transformation and use an explicit
simulation-style witness where possible. It also follows [Nik Sultana and
Simon Thompson, “Mechanical Verification of Refactorings”](https://kar.kent.ac.uk/23959/),
which formalizes refactorings in Isabelle/HOL and makes behavior preservation
the correctness condition.

The experiment is smaller and weaker than either line of work. Program
equivalence is undecidable; finite evidence cannot prove arbitrary programs or
all effects. A validator can be incomplete or report false alarms when it
cannot explain a transformation, and validation has runtime cost. This model
therefore makes only a synthetic, bounded claim over four explicitly declared
integer fixtures; it does not authorize arbitrary transformations or claim a
verified refactoring engine, complete semantic equivalence, toolchain
correctness, repository mutation, or promotion authority. The negative cases
make the claim falsifiable: changing the `.gooo` expected value or candidate
recipe changes the independently recomputed result to `REFUTED`, while an
unavailable recipe leaves it `OPEN`.

## Intervention separation

The intervention witness is separate from the four-case authority score. It
publishes three cases with three non-aggregated fixed denominators, each `1`:

* The semantic-expected slice changes `expected=3` to `expected=4`. Its parsed
  projection, receipt, and independent decision change from `AUTHORIZED` to
  `REFUTED` with reason `SEMANTIC_POSTCONDITION_REFUTED`.
* The semantic-operation slice changes the candidate and replay recipe from
  `add:1` to `add:2`, changing the candidate output, postcondition, and
  decision.
* The non-semantic slice appends only whitespace and a comment. Its raw
  `SourceDigest` and receipt digest change, while the parsed/lowered fixture
  projection, replay output and semantic digest, decision, resolution, reason,
  claims, and effect remain equal.

All three slices require zero repository writes. Their claims are bound to
persistent `OPEN -> DISCHARGED` transitions with exact coordinates: the
semantic-expected step/reason are
`compare-semantic-expected-projection-and-decision` /
`SEMANTIC_EXPECTED_VALUE_AND_DECISION_CHANGED`; the semantic-operation pair
is `compare-semantic-operation-projection-and-decision` /
`SEMANTIC_OPERATION_AND_DECISION_CHANGED`; and the non-semantic pair is
`compare-nonsemantic-projection-and-decision` /
`NONSEMANTIC_PROJECTION_AND_DECISION_PRESERVED`.

The top-level intervention decision and resolution are derived from those
adjudicated cases and both source variants. An observed contradiction is
`FAIL_CLOSED` with `REFUTED`; an unobservable obligation is `FAIL_CLOSED` with
`OPEN` at lower resolution. The separate `interventionconsumer` package and
its witness command use their own wire structs and functions, parse the raw
`.gooo`, lower it canonically, execute each candidate and replay recipe, and
compare receipts, decisions, coordinates, transitions, and effects without
importing producer or calling `Build`. The separately named
`DeterministicReplay` uses a second `Build` only for repeatability. CI also
changes one case outcome together with its claim, denominator, top decision,
and resealed digest, then proves that the source-bound consumer rejects the
coherent-looking artifact.

CI exposes fixed evidence counts: producer imports `0/0`, reconstructed cases
`3/3`, actual semantic replays `2/2`, artifact observation `1/1`, and coherent
tamper rejection `1/1`.
