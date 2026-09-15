# Candidate leakage shadow

This use case keeps a candidate non-authoritative until an exact promotion
receipt binds the same subject, candidate digest, and meta-operation.

The frozen `gooo/candidate-leakage-denominator/v1` contains six cases:

| Outcome | Cases | Expected result |
| --- | ---: | --- |
| Candidate remains isolated | 1 | `PASS / EXACT / NO_EFFECT` |
| Exact promotion is bound | 1 | `PASS / EXACT / NO_EFFECT` |
| Positive official claim leaks | 2 | `FAIL_CLOSED / EXACT / BLOCK` |
| Evidence identity is unresolved | 2 | `FAIL_CLOSED / INVARIANT_ONLY / BLOCK` |

The evaluator implements `detect-candidate-leakage`; it never writes the
repository and never grants promotion credit. An unknown promotion decision
therefore cannot become `FIXED_POINT`, `ALLOW`, or `AUTHORIZED`.

This phase is shadow-only. It does not add the metric to the operating language
assurance registry, so the official frozen readiness remains `7/12`.
