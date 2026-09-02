# Gooo meta-policy compilation

This experiment treats `policy.gooo` as the semantic authority. The producer
parses and lowers that raw source, then emits a standalone Go judge whose
reduction rows are generated from the source metadata. The independent consumer
parses the raw Gooo source again and reconstructs the same rows without
importing the producer package or generated judge.

The fixed policy denominator is eight source-declared obligations and the
canonical conformance denominator is three cases, evaluated by source,
generated judge, and independent consumer (`3/3/3`):

| Case | Expected decision | Meaning |
| --- | --- | --- |
| `pass-semantic-equivalence` | `PASS` | all digest coordinates agree |
| `refuted-valid-source-drift` | `FAIL_CLOSED` | a well-formed contradictory digest is REFUTED first |
| `unknown-malformed-source-digest` | `UNKNOWN` | malformed evidence preserves six UNKNOWN fields |

UNKNOWN preserves `stage`, `step`, `reason`, `unknown_class`,
`next_operation`, and `blocked_by`. An unrecognized upper decision is never
accepted as a fixed point; it reduces to
`FAIL_CLOSED / FEEDBACK_COVERAGE_DECISION_UNKNOWN`.

Synthetic case evidence and current runner-temp evidence are separate receipt
sections. Repository net-write is an exact start/end file snapshot comparison;
the six generated files are required to stay in runner temp.

