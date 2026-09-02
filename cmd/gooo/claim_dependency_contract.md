# Claim dependency causality

`gooo claim dependencies <file.gooo> --json` reconstructs structural claim
dependencies from Gooo activity dataflow and activity value programs.

A recoverable foundation uses `claim.observe:recoverable|<label>`. A dependent
activity uses `claim.edge:<kind>|<label>`, where kind is `requires`, `supports`,
`contradicts`, or `failure-entailment`. Each input entity of that activity binds
one edge from the unique activity that produces the entity. Source order fixes
node and edge ordinals.

The candidate `gooo.primitive.claim-dependency-causality.v1` is present in the
Gooo example and every JSON receipt. Every indicator carries the exact Gooo
activities that define its value. FOUNDATION covers recoverable roots,
COHERENCE covers typed structural edges, and REGRESSION covers missing inputs,
cycles, and repository writes.

A missing input producer returns `INCOMPLETE` with `UNKNOWN`, stage, step,
reason, unknown class, next operation, and blocked-by entities. Unsupported edge
kinds, ambiguous producers, and cycles return `FAIL_CLOSED` with `REFUTED`.

Version 1 claims structural reconstruction only. It does not claim proposition
truth, state propagation, external evidence, semantic promotion, repository
mutation, or automatic merge authority.
