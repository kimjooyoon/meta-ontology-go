# Causal CI selection meta-program

## Proposition

CI selection is a proof-choice program, not a filename heuristic. The causal
route is:

```text
changed file -> semantic claim -> impact path -> check -> proof choice
```

The receipt explains the route used for every selected check. If an edge is
unknown, the program records its exact stage/step/reason and descends to the
whole fixed six-check suite. A malformed or unregistered graph is rejected and
never silently widened into an apparently valid selective plan.

The program is read-only. `producer` describes the observation boundary,
`consumer` describes GitHub Actions and its independent verifier,
`meta_operation` is `causal-ci-select`, and `proof_choice` is data describing
why a check was selected. None of these fields grants repository or merge
authority.

## What was adopted and rejected

The experiment borrows a narrow causal idea from build-graph systems while
keeping CI policy separate from build execution:

- Bazel's official query documentation treats declared rules as a dependency
  graph and provides `deps`, `rdeps`, `somepath`, and `allpaths` to trace why a
  target depends on another. The experiment adopts explicit graph paths and
  records a path explanation. It rejects Bazel's configuration/build-target
  model as a CI authority: a build graph is not proof that every repository
  policy obligation was evaluated. See the [Bazel query guide](https://bazel.build/query/guide)
  and [Bazel query reference](https://bazel.build/reference/query).
- Nix's official reference describes derivations as specifications whose inputs
  induce build-time dependencies, and defines a closure as the transitive set
  of reachable store paths. The experiment adopts transitive closure as the
  intuition for whole-suite descent when a causal path is not known. It rejects
  store reachability and cache reuse as a semantic answer to “which CI proof is
  justified.” See the [Nix derivation reference](https://nix.dev/manual/nix/2.22/language/derivations.html)
  and [Nix closure glossary](https://nix.dev/manual/nix/2.35/glossary.html).

The resulting boundary is intentionally weaker than a build system: it
selects or widens a proof plan and emits evidence, but does not run a target,
reuse a cache, write the repository, or promote a branch.

## Fixed observables and falsification

The corpus fixes three cases, six checks, and six binary indicators. A passing
receipt must show one `SELECTED`, one `FULL_FALLBACK`, and one `REJECTED` case;
the fallback must select all 6 checks with an explicit unknown cause; the
rejection must select none; and all transition digests must form one chain.

The proposition is falsified if a filename-only change can select a check with
no claim-mediated path, an unknown path does not select all 6 checks, an
unregistered reference produces a plan, a receipt can be replayed with a
different input, or the transition chain is rewritten rather than extended.
These are observable counterexamples, not qualitative review judgments.
