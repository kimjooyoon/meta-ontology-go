# Language deterministic query

This example is a serialized query-plan receipt, not an instruction to mutate
the repository. CI reifies the language concept catalog as a detached PROV
graph and evaluates the fixed plan twice plus one insertion-permuted replay.

Expected result:

- fixed cases: `32/32`
- binding cases: `28/28`
- fail-closed laws: `4/4`
- canonical replay checks: `56/56`
- permutation replay checks: `28/28`
- candidate promotions, unknown acceptances, graph mutations, effects: `0`
