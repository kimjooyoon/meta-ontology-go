# Provenance-bound change proposal

A reverse-observation receipt can justify a *proposal* for a language change, but it cannot apply that change by itself.

The proposal must preserve source, generated-artifact, reverse-observation, target-scope, rationale, and review evidence digests. Missing or mismatched receipt stages remain `UNKNOWN` or `DEFERRED`; a proposal is only `PROPOSED` when an explicit application boundary exists.

Applying the proposal remains a separate, reviewable operation. This makes self-improvement possible without turning observation evidence into implicit authorization or mutation.
