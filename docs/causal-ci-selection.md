# Causal CI selection meta-program

This experiment tests a stronger proposition than changed-file pattern
matching:

```text
PR changed-file observation
  -> Gooo claim
  -> Gooo surface
  -> registered check
  -> proof choice and plan receipt
```

The policy is the actual
[`main.gooo`](../examples/causal-ci-selection/main.gooo) source. Its typed
activity contracts and semantic value programs define `changed-file-to-claim`,
`claim-to-surface`, `surface-to-check`, and prior claim state operations. The
producer reconstructs that graph only after canonical parse, canonical format,
lowering, and semantic digest. The raw observation is generated from the PR's
actual `git diff`; it has no final known/decision/choice/reason fields.
Its predecessor state comes from the raw
[`prior-claims.json`](../examples/causal-ci-selection/prior-claims.json)
ledger observation and is joined to each observed path by CI.

## Decision boundary

The receipt separates conformance from subject resolution:

- `conformance.decision=PASS` means the registered Gooo policy was parsed,
  lowered, and reconstructed. It does not mean any check ran.
- A subject with a complete claim-mediated path is `SELECTED` and receives a
  selective plan with `CLAIM_IMPACT_REASON`.
- A subject not bound to the source authority is `UNKNOWN`; the plan descends
  to the fixed six-check suite and carries the exact
  `CAUSAL_SELECTION/observe-subject/SOURCE_NOT_BOUND_TO_POLICY` cause.
- A contradictory policy path is `FAIL_CLOSED`; each subject has no plan and
  the conformance coordinate is retained.

This is explicitly a `PLAN_ONLY` artifact. The workflow validates the plan and
uploads it; it does not execute the selected checks or authorize a merge.

## Claim ledger

CI observes prior `OPEN` claims. The producer appends a digest-linked
transition, preserving stage, step, reason, evidence digest, and provenance:

- complete observed path: `OPEN -> DISCHARGED`;
- unresolved path: `OPEN -> OPEN`, resolution `DESCEND_TO_FULL_SUITE`;
- explicit semantic contradiction: `OPEN -> REFUTED`, resolution `NO_PLAN`.

Evidence digests are SHA-256 values over canonical observed values, not
placeholders. The transition chain is append-only and independently replayed.

## Interventions and falsification

CI emits four source reconstructions from one raw observation:

1. base policy selects `go-test` for the source subject;
2. semantic intervention changes the policy target to `go-vet`, changing the
   semantic digest, plan digest, and subject proof choice;
3. nonsemantic comment/layout intervention changes the raw digest while
   preserving parsed/semantic/plan digests;
4. contradiction intervention declares two selective targets for one surface,
   producing `FAIL_CLOSED` and `REFUTED` claim transitions.

The proposition is falsified if a filename-only rule selects a check without a
reconstructed semantic path, a raw observation can declare a conclusion, an
unknown subject omits the full six-check descent, a contradiction produces a
plan, a comment changes the semantic/plan digest, or the independent consumer
can verify a producer receipt without reparsing and relowering the raw source.

## Build-graph research boundary

The experiment adopts explicit path reasoning from Bazel's first-party query
model (`deps`, `rdeps`, `somepath`, `allpaths`) and records an explanation for
the path. It rejects Bazel's build target/configuration as CI semantic
authority. See the [Bazel Query Guide](https://bazel.build/query/guide) and
[Bazel Query Reference](https://bazel.build/reference/query).

It adopts Nix's first-party transitive closure intuition for conservative full
descent when a path is unknown. It rejects store reachability and cache reuse
as evidence that a CI proof was justified. See the [Nix Derivation
Reference](https://nix.dev/manual/nix/2.22/language/derivations.html) and
[Nix Closure Glossary](https://nix.dev/manual/nix/2.35/glossary.html).
