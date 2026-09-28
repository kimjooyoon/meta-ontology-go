# Capability discovery to domain completeness

This bridge keeps capability discovery useful without turning it into a truth score.

## Evidence flow

1. A gooo declaration is inspected by the capability discovery layer.
2. The discovery result is emitted as a jev capability receipt with source, declaration, query, evidence, catalog, and toolchain identities.
3. A domain completeness receipt records the observation as one bounded domain observation.
4. The completeness validator checks identity continuity, stage localization, and metric arithmetic. It does not infer that a high coverage value means the domain is complete.

## Required observation binding

Each observation must retain:

- the exact declaration digest and source digest;
- the capability receipt digest and catalog digest;
- the toolchain identity;
- the capability state: AVAILABLE, DEFERRED, or UNKNOWN;
- the first missing stage when the state is not AVAILABLE;
- the actionable next operation, when one exists;
- explicit non-executing and non-authorizing flags.

An observation without these bindings is not silently repaired. It is rejected or recorded as UNKNOWN with its missing stage.

## Metrics that remain interpretable

The domain receipt may report:

- declaration binding rate;
- evidence-boundary rate;
- first-missing-stage localization rate;
- replay continuity rate;
- actionable-next-operation rate;
- investment allocation by missing stage and domain.

These are diagnostic measurements, not a single correctness score. A value can improve because more declarations were observed, and later decrease when the domain boundary expands. That change is expected and must remain visible.

## Investment decision

An investment decision should name:

- the domain boundary and intended users;
- the missing stage or capability;
- the evidence required before calling the stage available;
- the expected implementation cost;
- the rollback or review condition;
- the next observation that can falsify the decision.

Capability discovery therefore helps choose the next useful language feature, while the domain contract prevents suggestions from being mistaken for execution, authorization, or semantic completeness.
