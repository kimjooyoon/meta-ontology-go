# Reflective query sandbox

This experiment asks a Gooo program about its own structure, claim catalog,
and metric catalog through the existing semantic-IR query projection. The
projection is a detached read view. Its result is sealed as a receipt, but no
query is allowed to add a semantic fact, edit `.gooo`, write the repository, or
grant mutation authority.

The source contains five entities and four activities. The producer runs five
fixed attempts:

| Attempt | Result | Meaning |
|---|---|---|
| `reflect.structure` | `PASS / EXACT` | the source structure is queryable |
| `reflect.claims` | `PASS / EXACT` | the source claim catalog is queryable |
| `reflect.metrics` | `PASS / EXACT` | the source metric catalog is queryable |
| `mutation.attempt` | `DENIED / INVARIANT_ONLY` | a mutation-shaped request is not a query |
| `unknown.target` | `UNKNOWN / LOWER_RESOLUTION` | an absent stable ID is preserved as unknown |

The fixed denominator is 12 indicators: `OUTCOME 4`, `DRIVER 4`, and
`GUARDRAIL 4`. Proof choices are fixed at `FOUNDATION 4`, `COHERENCE 4`, and
`REGRESSION 4`. Every indicator records its producer, consumer,
meta-operation, proof choice, and `stage/step/reason` coordinate. The receipt
contains 24 append-only transitions: 12 `UNRECORDED -> OPEN` registrations and
12 `OPEN -> DISCHARGED` observations. An unknown target discharges the
preservation claim; it never becomes a successful query.

## Research decisions

The design records these primary sources:

- Go's official [`reflect` package documentation](https://pkg.go.dev/reflect)
  distinguishes a runtime type/value view from mutation. It says `CanSet`
  reports whether a value can be changed and that setters panic when it is
  false. Adopted: stable query identity, detached observations, and an
  explicit mutation boundary. Rejected: `MakeFunc`, `Set`, `SetMapIndex`, or
  any arbitrary runtime value mutation; this sandbox is not a general Go
  reflection API.
- The official [OCaml 5.2 effect-handler
  manual](https://ocaml.org/manual/5.2/effects.html) models effects as named
  operations handled by an enclosing boundary, and documents that unhandled
  effects fail at the point of use. Adopted: name the meta-operation and keep
  unhandled/unknown cases visible. Rejected: continuations, scheduling, and
  runtime effect execution; the Gooo operation only observes.
- The official [Koka language book](https://koka-lang.github.io/koka/doc/book.html)
  treats effect typing and handlers as core language concepts. Adopted:
  effect-shaped receipts and an explicit zero-authority effect summary.
  Rejected: effect polymorphism and an inferred effect algebra; this PR proves
  only the fixed read-only contract.

## Running in CI

The workflow uses Go `1.27.0`, checks the `.gooo` source with `gooo check`,
replays the producer twice, transfers the exact producer artifact, and replays
the independent consumer twice. It expects:

- `9` semantic nodes and `8` deterministic relation facts;
- `3` exact safe queries, `1` denied mutation attempt, and `1` unknown target;
- `12/12` indicators, `24` transition events, `0` repository writes, and
  `mutation_authority=false`;
- `0` promotion credit.

This does not claim general reflection equivalence, complete source coverage,
hostile-process isolation, runtime memory/performance bounds, or permission to
change the compiler or repository. It is falsified if any query changes the
semantic or graph digest, if a mutation request is accepted, if an unknown
target is upgraded to `PASS`, if a transition is removed/reordered/rewritten,
or if the fixed effect summary changes.
