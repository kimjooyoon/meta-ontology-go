# Repair handoff dogfood

This example connects a declared Gooo semantic surface to the runtime repair
boundary without granting an autonomous repair authority.

`main.gooo` declares two typed activities:

1. `ValidateCandidate` preserves a validated repair candidate.
2. `DeferCandidate` turns that candidate into a handoff for the next operation.

The CI example performs four separate observations:

1. Semantic checking of the `.gooo` declaration.
2. Generation of the bound semantic Go artifact and runtime plan.
3. Consumption of a fixed `REFUTED` candidate fixture.
4. Verification that the handoff is `DEFERRED`, carries the next operation, and
   still has `execution_allowed=false` and `repository_writes=0`.

This is not autonomous source modification. The candidate fixture is an input,
the output directory is caller-owned, and any later execution remains an
explicit operation outside this example.
