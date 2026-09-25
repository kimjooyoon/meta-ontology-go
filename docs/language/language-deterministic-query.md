# Language deterministic query

## Purpose

A query is language behavior only when its plan, authority boundary, result,
and replay receipt are data. This capability reuses `internal/query` without
changing that engine. It projects this repository's own concept catalog into a
detached PROV graph and queries the links between one concept, six code
bindings, eighteen metric bindings, and three executable use cases.

## Fixed denominator

The versioned `gooo/language-deterministic-query-plans/v1` registry contains
exactly 32 cases:

- concept binding queries: 1
- code binding queries: 6
- metric binding queries: 18
- use-case binding queries: 3
- fail-closed laws: 4

No percentage is inferred. Full conformance is `32/32 = 10000 BPS`.

## Meta-programming connection

The query plan is reified code. CI observes the language concept artifact,
projects its bindings, normalizes a bounded request, executes it twice, rebuilds
the graph in reverse insertion order, replays the request, and seals effects.
The result therefore measures a meta operation over the program's own declared
structure rather than a disconnected repository statistic.

[gomacro](https://github.com/cosmos72/gomacro) is a useful structural reference
for treating syntax and programs as data. This project adopts the staged shape,
not ambient evaluator authority: filesystem and network access are not query
capabilities, candidates are never promoted, and unknown layers or endpoints
fail closed.

## Indicators

The report has 18 metrics with a fixed partition:

- outcome: 1
- drivers: 8
- guardrails: 9

Drivers measure plan execution and deterministic replay. Guardrails measure
registry drift, unresolved evidence, candidate promotion, unknown acceptance,
graph mutation, effects, repository writes, and mutation authority.

## Munchhausen routes

- `FOUNDATION` binds the versioned plan registry and exact concept bindings.
- `COHERENCE` requires identical canonical request/result and permutation receipts.
- `REGRESSION` rejects candidates, unknowns, mutations, effects, and writes.

This is an uncommon combination, not a novelty claim.

## Compatibility law discovered during implementation

A concept-specific witness cannot require the global catalog to remain frozen
at the count observed when that concept was introduced. The semantic-model
binding therefore records the exact current count while checking its versioned
historical floor and the continued satisfaction of its own obligation. The
global transition ledger remains the only authority for predecessor-to-current
improvement.
