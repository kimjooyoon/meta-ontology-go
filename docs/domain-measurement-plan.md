# Domain measurement plan

domaincapability.MeasureDomain defines a bounded measurement contract for deciding what to build next.

The plan deliberately does not produce a language-completeness score. It binds four independent evidence identities:

- the declared domain scope
- the source being measured
- the capability surface
- the observed use-case evidence

It keeps observed_use_cases, required_use_cases, and the derived remaining count as explicit facts. The result is BOUND when the evidence and policy are bound, even when more in-scope use cases remain. Its decision then records whether to add use cases, defer investment because the budget is exceeded, or hold the declared scope for review.

A larger observed count does not authorize expanding the declared scope. Scope expansion requires a new scope digest and a separate review boundary.

The plan is review-only. It does not expand the scope, authorize execution, invoke providers, or mutate the ontology. The .gooo contract in examples/domain-measurement-plan/plan.gooo is the declaration-level counterpart.
