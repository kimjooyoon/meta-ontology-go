# Counterexample-first meta compilation

This is a bounded, read-only vertical experiment over one real Gooo
transformation: `syntax.ParseFile` followed by `bidir.Lower`. The source
declares `CanonicalEntityID(... ) computes "identity:v1"` and a connected
four-stage meta graph:

```text
CanonicalEntityID
  -[CompilationClaim]-> DiscoverMinimalCounterexample
  -[MinimalCounterexample]-> BindResolutionEvidence
  -[ResolutionEvidence]-> PromoteOnlyAfterResolution
```

The compiler reconstructs those `used`/`wasGeneratedBy` edges from lowered
facts. Missing or disconnected activities authorize no counterexample,
resolution, or promotion decision. The meta-operation intervention removes
the binding activity and must change per-case receipts and claim transitions;
the comment-only intervention must preserve them.

The corpus contains raw source, five unique claim propositions, and five
distinct observation predicates. It contains no failing/minimal/accepted or
expected-decision fields. The producer executes each supplied source, compares
observed lowering output with the predicate, discovers violations, and applies
a deterministic finite neighborhood shrinker. The claim is only
`FINITE_NEIGHBORHOOD_IRREDUCIBLE`: every candidate in the declared immediate
neighborhood was executed and did not preserve the violation. This is not a
global or cost minimum.

For the resolved case, the resolution input names an operation only. The
producer generates the repair source by transforming the minimum counterexample
itself, records before/after source digests and a repair delta digest, then
reruns the same claim and predicate. Only that observed pass can append
`REFUTED -> DISCHARGED` and promote.

The independent judge repeats ParseFile, Lower, fact-graph inspection,
predicate evaluation, shrinking, and repair from raw source/corpus inputs. It
does not import the producer package or a shared conclusion table. Receipts
preserve parse/lower observations, semantic digests, graph authorization,
producer/consumer/meta-operation/proof choice, per-claim transition evidence,
and exact unknown coordinates. A missing source is
`INPUT_OBSERVATION/source-acquisition/SOURCE_NOT_PROVIDED` with decision
`UNKNOWN`, resolution `LOWER_RESOLUTION`, and claim state `OPEN`.

This does not claim general compiler correctness, global or cost minimality,
theorem proving, unbounded corpus coverage, or repository mutation authority.
