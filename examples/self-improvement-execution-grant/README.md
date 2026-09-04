# Self-improvement execution grant v26

This policy is the bounded grant layer between the v24 candidate authorization and the v25 pre-execution contract. It can declare a separate, one-use execution grant only when the v24 request and resolution, v25 contract, source artifact, scope, safety envelope, and decision evidence are exact.

The grant is not execution. An `ALLOW` resolves to `CLOSED/GRANTED_UNCONSUMED` with `remaining_uses=1`, `execution_count=0`, and `one_use_enforced=false`. The next bounded executor must verify the receipt, verify that it is unconsumed, and consume it exactly once. This v26 layer neither consumes the grant nor runs the candidate, produces output, compares results, or adopts changes.

Automatic `workflow_run` requests contain no explicit grant decision and therefore resolve to `UNKNOWN/LOWER_RESOLUTION` with the six causal fields and `blocked_by=explicit_execution_grant_decision`. A `workflow_dispatch` decision binds the GitHub event actor, repository, run ID, run attempt, event, and explicit decision as actor evidence; it is not represented as a cryptographic signature or authenticated identity.
