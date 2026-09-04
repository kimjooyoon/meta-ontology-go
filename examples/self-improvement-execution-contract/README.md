# Candidate execution contract

This v1 contract is a declaration boundary for the v24 non-executing candidate
authorization bridge. It projects the candidate binding, exact observation
evidence, and v24 authorization request into a typed, bounded pre-execution
contract.

The contract never executes a candidate, grants execution, writes the
repository, runs tests, compares runtime outputs, or adopts a candidate. A
`CLOSED/DECLARED` result means only that the twelve pre-execution fields are
exact and bound to the fixed operation registry. Execution remains blocked by a
separate execution grant, with `max_executions=1` and
`repository_writes_allowed=false`.

Missing evidence or an unsupported operation mapping remains `UNKNOWN` with the
six causal fields and a `missing_fields` frontier. Contradictions are
`REFUTED`, and contradiction takes precedence over UNKNOWN.
