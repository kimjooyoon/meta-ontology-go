# Claim dependency causality experiment

This is a bounded Gooo meta experiment about responsibility for an unresolved
claim. It is intentionally separate from the operation-catalog dependency
classifier. The producer reads one `.gooo` source and emits a deterministic
receipt; the independent judge re-derives the graph, transition states, frontier
edges, and decision from that receipt.

The contract has a fixed denominator of six claims and eight directed edges.
The last claim has a direct shortcut from `producer-bound`, so the judge must
preserve the minimum cause path rather than return every reachable ancestor.
All claims carry `producer`, `consumer`, `meta_operation`, `proof_choice`, and a
`stage/step/reason` coordinate.

The three source cases are:

| source | case | exact result |
| --- | --- | --- |
| `unknown.gooo` | `direct-unknown` | 1 direct `OPEN`, 5 `DEPENDENCY_BLOCKED`, 8 blocking edges, depth 2 |
| `refuted.gooo` | `refuted` | 1 direct `REFUTED`, 5 `DEPENDENCY_REFUTED`, 8 refuting edges |
| `main.gooo` | `recovered` | 1 direct `DISCHARGED`, 5 `DEPENDENCY_RECOVERED`, 5 minimum recovery edges |

Every receipt has twelve transitions: six `CLAIM_REGISTERED` events followed by
six outcome events. Across the three receipts the outcome vocabulary exercises
`OPEN`, `DISCHARGED`, and `REFUTED`. The experiment is read-only and reports zero
repository writes and zero semantic-promotion authority.

The producer does not claim that the `.gooo` operation is semantically correct.
The source marker only selects a controlled observation case. The judge proves
receipt consistency and state propagation, not compiler correctness, causal
inference about the world, runtime behavior, or a general-purpose dependency
engine.
