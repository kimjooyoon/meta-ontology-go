# External conformance assurance activation

This use case consumes three immutable capsules: the merged `11/12` language
assurance report, the PR #474 eligibility report, and the GitHub merge relation.
The evaluator exposes exactly one operating meta-operation only when all three
are byte-exact and semantically exact.

- assurance predecessor: `2104e3fcede21572951c266cd7b61fd38f386aef`;
- merged eligibility head: `0fa760bdcb9c90e386d85978bff1c6552e8b503a`;
- metric: `gooo.metric.ecosystem.external-conformance.v1`;
- operation: `verify-external-conformance`;
- assurance transition: `11/12 = 9166 bps` to `12/12 = 10000 bps`;
- activation denominator: `8/8`;
- activation indicators: `10 = 2 OUTCOME + 4 DRIVER + 4 GUARDRAIL`;
- Munchausen partition: `FOUNDATION 4 + COHERENCE 2 + REGRESSION 4`;
- eligibility indicators: `18/18`;
- parent whole-project result preserved: `6/8`, known failures `2`;
- selected executable capabilities preserved: `10/10`, executions `4`;
- applied transitions: `1`;
- evaluator repository writes: `0`.

Only explicit `ELIGIBLE_SHADOW / EXACT` activates. `UNKNOWN`, `FIXED_POINT`,
and unrecognized top-level decisions all become `FAIL_CLOSED / UNKNOWN / BLOCK`.
Byte drift becomes `FAIL_CLOSED / INVARIANT_ONLY / BLOCK`.

`12/12` is the fixed language-assurance kernel denominator, not overall
language completion and not whole-project gomacro compatibility. The latter
remains the independently reported `6/8` result.
