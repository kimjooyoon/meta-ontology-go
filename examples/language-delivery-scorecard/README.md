# Language delivery scorecard

This directory fixes the `gooo-v0.2-observable-delivery` denominator at 36 obligations.

- USER owns 12 outcomes.
- TOOL_AUTHOR owns 12 drivers.
- GOVERNOR owns 12 guardrails.
- Reader views are cumulative: USER is contained by TOOL_AUTHOR, which is contained by GOVERNOR.
- `NOT_IMPLEMENTED` is a known zero, while malformed or unknown evidence is `UNKNOWN` and fails closed.
- The internal 24-obligation readiness receipt remains visible only as `INTERNAL_SELF_IMPROVEMENT_CONTRACT`.
- A separately uploaded source-execution artifact backs three user outcomes; five fixed gaps remain.

The JSON contract is duplicated by canonical Go meta-code. Changing one without the other is contract drift, not progress.
