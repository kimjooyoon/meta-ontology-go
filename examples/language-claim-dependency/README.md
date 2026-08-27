# Claim dependency causality experiment

This bounded experiment distinguishes a direct `UNKNOWN` observation from a
dependent claim that is only `DEPENDENCY_BLOCKED`. It also demonstrates that a
refuted upstream claim does not refute ordinary `SUPPORTS` or `REQUIRES`
dependents. Only an explicitly satisfied `CONTRADICTS` or
`FAILURE_ENTAILMENT` edge can propagate `REFUTED`.

The producer and consumer both start with raw `.gooo` bytes, run
`syntax.ParseFile -> bidir.Lower`, and reconstruct the graph from canonical IR.
The six activity claims are joined by eight semantic relations formed from
`prov:wasGeneratedBy + prov:used`; the downstream activity's semantic value
program declares the typed edge. `Root` and `Derived` therefore have a real
semantic relation through `RootState`, rather than a case-name-only graph.

| source | observation predicate | direct / dependent state | edge result |
| --- | --- | --- | --- |
| `unknown.gooo` | `UNKNOWN` | 1 `OPEN` / 5 `DEPENDENCY_BLOCKED` | 8 blocking |
| `refuted.gooo` | `EXPLICIT_CONTRADICTION` | 1 `REFUTED` / 3 open / 2 dependency-refuted | 5 blocking, 2 refuting |
| `main.gooo` after `unknown.gooo` | `EVIDENCE_ACCEPTED` | 1 direct / 5 dependency `DISCHARGED` | 8 recovery edges |

The fixed denominator is six claims, eight typed edges, and twelve initial
transitions. Recovery is not a new ledger: it preserves the twelve unknown
transitions and appends six recovery transitions, verifies the prior receipt
digest, prior transition head, prior claim states, and observation digest.

`edge-intervention.gooo` changes a semantic edge value from `CONTRADICTS` to
`SUPPORTS`; the CI intervention artifact compares its state propagation with
`refuted.gooo`. The comment-only difference between `unknown.gooo` and
`main.gooo` must preserve the canonical IR/graph digest and decision.
