# Evidence freshness experiment

This example uses the real `main.gooo` source as a subject and records a
six-part justification tuple: `subject`, `material`, `recipe`, `environment`,
`runner`, and `verifier`. The independent decider compares that tuple with a
current context and selects the earliest changed stage in a fixed order.

The contract has ten deterministic cases: one fresh claim, one stale case for
each coupling axis, one expired temporal boundary, and two unknown cases. A
fresh claim transitions to `CLAIM_PRESERVED`; stale or unknown evidence is never
preserved as current.

Independence is reported with two different semantics: the receipt and report
carry `forbidden_dependency_count` as a guardrail count (`0`, never a ratio),
while the performed and passed independence check is the positive fixed-
denominator contract `independence_contract = 1/1`. The CI summary and
validator recompute these same fields.
