# Counterexample-first meta compilation

## Claim and vertical boundary

This experiment tests one causal ordering claim over a real Gooo language
transformation:

```text
raw .gooo
  -> syntax.ParseFile
  -> bidir.Lower
  -> lowered nodes/facts and semantic digest
  -> observed predicate comparison
  -> finite-neighborhood counterexample search
  -> repair transform from that counterexample
  -> same-claim repair rerun
  -> compile decision and append-only claim transition
```

The source's meta code is executable authority. The producer and independent
consumer reconstruct the required activities and their lowered data edges:
`CanonicalEntityID -> DiscoverMinimalCounterexample -> BindResolutionEvidence
-> PromoteOnlyAfterResolution`. Removing an activity or changing an edge
produces an unauthorized graph and changes receipt/decision/transition
observations; a comment-only change preserves the semantic digest and evidence.

## References and limits

- [QuickCheck `shrink` API documentation](https://hackage.haskell.org/package/QuickCheck/docs/Test-QuickCheck.html)
  describes immediate, type-specific shrink candidates. A shrinker must retain
  the relevant invariant; a malformed or incomplete candidate is not proof.
- [QuickChick's Software Foundations chapter](https://deepspec.github.io/sf/qc-current/QC.html)
  describes greedy shrinking until every immediate shrink succeeds. This
  experiment names the resulting claim precisely
  `FINITE_NEIGHBORHOOD_IRREDUCIBLE`; it does not call it global or cost
  minimality.
- [Isabelle's official Nitpick manual](https://www.isabelle.in.tum.de/website-Isabelle2021/dist/Isabelle2021/doc/nitpick.pdf)
  documents bounded scopes and counterexample search limits. No observed model
  or unavailable input therefore becomes a safety proof.

The shrink relation here is intentionally finite and explicit: remove
`noise=1`, then replace `drift=1` with the rule's canonical ID. Every immediate
candidate is executed. The numerator is the number of final-neighborhood
candidates that no longer violate the predicate; the denominator is the
number executed. The report and receipt use only the finite-neighborhood name
and retain global/cost minimality in `not_claimed`.

## Claims, resolution, and UNKNOWN

The v3 contract fixes five unique propositions and five distinct observation
predicates rather than cloning one identity predicate across five labels.
Every transition contains the claim ID, proposition digest, predicate ID,
predicate evidence digest, and actual observation evidence digest.

The resolved case provides only `canonicalize-entity-id` as a resolution
operation. The compiler derives the repair source from the minimum
counterexample, records original/repair source digests and a repair delta
digest, and reruns the same claim and predicate. Only a passing rerun appends
`REFUTED -> DISCHARGED`; an unresolved violation remains `REFUTED` with
`LOWER_RESOLUTION`.

An absent source is not represented by a generic unknown. It is recorded at
`INPUT_OBSERVATION/source-acquisition/SOURCE_NOT_PROVIDED`, with decision
`UNKNOWN`, resolution `LOWER_RESOLUTION`, and an `OPEN` claim transition.

## Evidence and falsifiability

The independent judge reconstructs from raw `.gooo`, raw corpus, and actual
ParseFile/Lower observations. It does not import the producer package or a
shared expected-decision table. A tampered receipt must be rejected; a policy
or graph intervention must alter per-case evidence; a comment-only source must
preserve semantic digest and all per-case decision/transition evidence; an
unresolved counterexample must never promote; and missing capability evidence
must leave mutation authority `UNKNOWN` rather than infer denial from an
unchanged status.

CI writes artifacts under the runner's temporary directory, outside the
repository. Repository effects are bound to before/after CI status digests;
the read-only indicator additionally requires explicit workflow permission
evidence (`contents: read`) and `mutation_authority: DENIED`.

The fixed denominator is v3: five cases, five unique claims, five unique
predicates, twelve indicators, fifteen transitions, one unknown coordinate,
and five corpus inputs. It does not claim general compiler correctness,
theorem proving, global/cost minimality, unbounded corpus coverage, or
repository mutation authority.
