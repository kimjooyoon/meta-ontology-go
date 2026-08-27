# Counterexample-first meta compilation

## Claim and vertical boundary

This experiment tests a causal ordering claim over one real Gooo language
transformation:

```text
raw .gooo source
    -> syntax.ParseFile
    -> bidir.Lower
    -> semantic IR digest and lowered node observations
    -> predicate comparison
    -> deterministic shrinking of a discovered violation
    -> resolution rerun on the same minimal source
    -> compile decision
```

The source itself declares the candidate policy with
`CanonicalEntityID(CompilationClaim) -> CompilationClaim computes "identity:v1"`.
The predicate is not a decision label: it compares each observed lowered Entity
ID with `gooo://counterexample-first/entity/<kebab-case-name>`. A mismatch is an
observed violation. Parse diagnostics or lower errors remain UNKNOWN evidence.

The producer and consumer each run the parser and lowerer. The consumer
reconstructs observations from raw source and corpus inputs without importing
the producer package and without importing a shared expected-outcome table.
The JSON contract fixes the operation, predicate, corpus shape, and
denominators, but contains no decision, resolution, failure, minimality, or
acceptance assertion.

## What the references contribute

- [QuickCheck `shrink` API documentation](https://hackage.haskell.org/package/QuickCheck/docs/Test-QuickCheck.html)
  defines immediate shrink candidates and recommends type-specific shrinkers.
  Its relevant limitation is that an invariant-bearing value needs a shrinker
  that preserves its invariants; a malformed candidate is not automatically a
  useful proof.
- [QuickChick's Software Foundations chapter](https://deepspec.github.io/sf/qc-current/QC.html)
  describes greedy shrinking from a known failing value until every immediate
  shrink succeeds. This establishes local minimality under a strictly
  decreasing shrink order, not global minimality over all possible inputs.
- [Isabelle's official Nitpick manual](https://www.isabelle.in.tum.de/website-Isabelle2021/dist/Isabelle2021/doc/nitpick.pdf)
  presents bounded model finding and documents scopes, timeouts, and potential
  counterexamples. Therefore this experiment preserves UNKNOWN
  stage/step/reason and lowers resolution instead of treating lack of an
  observed model as proof of safety.

## Claim transitions and falsifiability

Each receipt retains three append-only transition records. An observed
predicate violation appends `OPEN -> REFUTED`; an unresolved violation remains
REFUTED with `LOWER_RESOLUTION`; an observed passing rerun appends
`REFUTED -> DISCHARGED` and only then permits `PASS`. An unobserved input
preserves `OPEN`, `UNKNOWN/UNKNOWN/UNKNOWN`, and `LOWER_RESOLUTION`.

The experiment is falsifiable in concrete ways: a producer receipt mutation
must be rejected by the independent judge; a changed candidate rule must
change the semantic digest, minimal counterexample, or transition evidence; a
comment-only change must not change semantic digest or decision evidence; an
unresolved counterexample must never be promoted; and any tracked repository
write or producer dependency must fail the read-only/independence indicators.

The custom shrinker is deliberately narrow. It removes the two fixed identity
noise suffixes in order and executes every immediate candidate at the final
minimal input. It does not establish global minimality, completeness of the
fixed corpus, or correctness of the general Gooo compiler.

## Fixed denominator and evidence

The contract pins five corpus cases, twelve indicators, and fifteen claim
transitions under `counterexample-first-denominator/v2`. The CI report exposes
corpus executions, discovered counterexamples, shrink-candidate executions,
minimality numerator/denominator, resolution reruns, source reconstruction
numerator/denominator, producer-import numerator/denominator, and repository
write effects. The producer is replayed byte-for-byte and the independent judge
replays the same raw inputs before accepting the report.
