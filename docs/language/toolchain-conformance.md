# Toolchain conformance ledger

Toolchain conformance is a meta operation over nine existing CI receipts. It
does not add another evaluator or mutate source. It joins the exact schemas,
commit identity, fixed denominators, indicators, proofs, and effect boundaries
already emitted by the language and toolchain.

The fixed `gooo/toolchain-conformance-corpus/v1` denominator contains:

- 9 tool surfaces
- 154 executable source cases
- 151 source indicators
- 27 source proofs
- 13 in-memory drift mutations

The conformance evaluator emits 3 outcome, 10 driver, and 15 guardrail
indicators. Every indicator names
`close-toolchain-conformance-ledger` as its meta operation. Unknown decisions,
lower resolution, schema or SHA drift, incomplete evidence, writes, and
mutation authority fail closed.

## Munchhausen choices

- `FOUNDATION` fixes the versioned surface registry and concept bindings.
- `COHERENCE` requires all nine exact-head receipts to agree.
- `REGRESSION` requires all thirteen bounded mutations to be rejected.

## Structural reference

[gomacro](https://github.com/cosmos72/gomacro) separates interactive, script,
library, and staged macro boundaries. Gooo uses that separation only as a hint
to treat tool surfaces independently before joining them. Unlike gomacro
macros, this ledger inherits no ambient file or network authority. No gomacro
dependency, interpreter claim, or novelty claim is introduced.
