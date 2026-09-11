# Public partial test reuse

This is the canonical v15 example for semantic impact-based test reuse. It is
intentionally small but real: `CreateReceipt` is the independently testable
orders partition and `SnapshotInventory` is the independently testable
inventory partition. Both use `SharedContract`, so the canonical metadata
declares exactly two dependency edges:

```
shared-contract -> orders
shared-contract -> inventory
```

The `computes` metadata on the `.gooo` activities is the policy table. It binds
source roots, generated symbols, test units, dependency edges, the fixed six
acceptance cases, and the v14 orchestration-to-evidence journey. The Go
implementation lowers this source, derives the semantic subgraphs and actual
edges, computes impact closure, and validates immutable partition receipts.
It does not select behavior from filenames or maintain a parallel case switch.

The v15 acceptance denominator is fixed at six cases: two `CLOSED`, two
`UNKNOWN`, and two `REFUTED`. The source is registered as one append-only
language package member; the repository language corpus therefore advances
from 58 to 59 `.gooo` sources and from 1011 to 1022 physical `.gooo` lines.

Selective reuse is explicit opt-in and fail-closed. A receipt is reusable only
when its canonical source and semantic subgraph, generated partition artifact,
compiler/released identity, Go 1.27 toolchain, exact test contract, dependency
graph, successful result, and v14 orchestration provenance all match.

The same canonical source also carries the v16 counterexample-driven semantic
resolution repair table. A proven hidden dependency first falls back to the
full project test contract, then produces a deterministic `shared-contract ->
inventory` proposal from the preserved v15 REFUTED record. Only an explicit
human authorization permits a caller-owned immutable graph overlay; the
overlay replay proves complete affected closure and retains selectivity for
an independent unchanged partition. Ambiguous or unsupported proof evidence
is `UNKNOWN`, while tampered counterexamples and unauthorized overlays are
`REFUTED` with REFUTED dominance.
