# Semantic resolution descent lattice

This is a bounded Gooo-only experiment. `UNKNOWN` is an observation result,
not a scalar: it carries `stage`, `step`, and `reason`, while the transition
records the exact `EXACT -> INVARIANT_ONLY` descent and its
`LOWER_RESOLUTION` operation.

The checked-in source is [main.gooo](main.gooo). Its four
`resolution-lattice.case;...` value programs carry `required`, `observed`,
`reason`, `repository_writes`, `mutation_authority`, `claim_id`, and
`claim_state`. The producer parses those activity value programs from the
Gooo AST; no case fixture is hard-coded in Go. The checked-in result receipt
is [receipt.json](receipt.json). The independent adjudicator parses the
activity declarations and value programs independently, reconstructs the
cases and claims, and then compares its generated result:

```sh
go run ./cmd/semantic-resolution-lattice-judge \
  -source examples/semantic-resolution-lattice/main.gooo \
  -receipt examples/semantic-resolution-lattice/receipt.json \
  -check
```

The same command can require the two source counterexamples:

```sh
go run ./cmd/semantic-resolution-lattice-judge \
  -source examples/semantic-resolution-lattice/main.gooo \
  -receipt examples/semantic-resolution-lattice/receipt.json \
  -check -counterexamples
```

The semantic counterexample changes only `observed=2` to `observed=3` for the
partial case: `UNKNOWN` becomes `PASS`, `invariant_only` becomes
`exact_operation`, and the claim transition becomes `OPEN -> DISCHARGED`.
The non-semantic counterexample adds a comment: the source digest changes but
the semantic digest, decision, and claim transition remain unchanged.

The four-case denominator is fixed. The receipt records one `PASS`, two
`FAIL_CLOSED`, and one structured `UNKNOWN`. Repository writes are `0` and
mutation authority is `false`; emitting the receipt is an evidence artifact,
not authority to mutate the source tree.

The fixed case denominator is `4`; the fixed counterfactual denominator is
`2`. Metrics expose decision influence `1/2`, claim-transition influence
`1/2`, and combined semantic influence `1/2`, each with a producer, consumer,
meta-operation, and FOUNDATION/COHERENCE/REGRESSION proof level.

## Research boundary

- [Cousot & Cousot, Abstract interpretation (POPL 1977)](https://www.di.ens.fr/~cousot/COUSOTpapers/POPL77.shtml): adopt finite ordered abstraction and monotone coarsening; reject claiming a general static analyzer or fixpoint engine here.
- [W3C PROV-O](https://www.w3.org/TR/prov-o/): adopt explicit Entity/Activity producer-consumer relations; reject treating provenance alone as proof of truth.
- [Open Policy Agent partial evaluation](https://www.openpolicyagent.org/docs/filtering/partial-evaluation): adopt an explicit unknown-input boundary and residual condition; reject undefined-to-false laundering and preserve Gooo's fail-closed decision boundary.
