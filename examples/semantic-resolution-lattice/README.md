# Semantic resolution descent lattice

This is a bounded Gooo-only experiment. `UNKNOWN` is an observation result,
not a scalar: it carries `stage`, `step`, and `reason`, while the transition
records the exact `EXACT -> INVARIANT_ONLY` descent and its
`LOWER_RESOLUTION` operation.

The checked-in source is [main.gooo](main.gooo). The checked-in result receipt
is [receipt.json](receipt.json). The independent adjudicator uses only the
receipt shape, source digest, and its own transition function:

```sh
go run ./cmd/semantic-resolution-lattice-judge \
  -source examples/semantic-resolution-lattice/main.gooo \
  -receipt examples/semantic-resolution-lattice/receipt.json \
  -check
```

The four-case denominator is fixed. The receipt records one `PASS`, two
`FAIL_CLOSED`, and one structured `UNKNOWN`. Repository writes are `0` and
mutation authority is `false`; emitting the receipt is an evidence artifact,
not authority to mutate the source tree.

## Research boundary

- [Cousot & Cousot, Abstract interpretation (POPL 1977)](https://www.di.ens.fr/~cousot/COUSOTpapers/POPL77.shtml): adopt finite ordered abstraction and monotone coarsening; reject claiming a general static analyzer or fixpoint engine here.
- [W3C PROV-O](https://www.w3.org/TR/prov-o/): adopt explicit Entity/Activity producer-consumer relations; reject treating provenance alone as proof of truth.
- [Open Policy Agent partial evaluation](https://www.openpolicyagent.org/docs/filtering/partial-evaluation): adopt an explicit unknown-input boundary and residual condition; reject undefined-to-false laundering and preserve Gooo's fail-closed decision boundary.
