# Gooo meta-policy compilation

This is a read-only philosophical experiment: a Gooo policy is treated as a
semantic source, not as text to be copied into another DSL. The compiler lowers
the policy through the repository's semantic IR, binds eight declared
`computes` metadata records to a fixed ontology, and emits a standalone Go
decision kernel.

The checked-in cases are intentionally small and fixed:

| Case | Expected decision | What is demonstrated |
| --- | --- | --- |
| `pass-semantic-equivalence` | `PASS` | producer, consumer, and independent source digests agree |
| `fail-closed-source-drift` | `FAIL_CLOSED` | a source mismatch cannot be treated as equivalent |
| `unknown-missing-consumer` | `UNKNOWN` | missing evidence is not promoted to a denial or pass |

The producer records the source digest, semantic IR digest, and fixed `8`
obligations. The consumer re-reads the compiled artifact and binds its source
digest and generated-judge digest. The generated judge is executed as a
standalone Go program; an independent verifier re-derives the decision from the
case and policy. The receipt requires all three expected decisions, equivalent
generated/independent coordinates, and a `48`-event append-only claim ledger
(`3 cases × 8 claims × 2 transitions`).

This experiment does not claim a general policy language, cryptographic
attestation, deployment freshness, or equivalence outside the modeled digest
and decision semantics.
