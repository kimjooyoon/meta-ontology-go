# Gooo meta-policy compilation

This is a read-only philosophical experiment: a Gooo policy is treated as a
semantic source, not as text to be copied into another DSL. The compiler lowers
the policy through the repository's semantic IR, validates only typed safety
invariants, compiles eight source-declared rules and a source-declared
reduction, and emits a standalone Go decision kernel.

The checked-in cases are intentionally small and fixed:

| Case | Validator expectation | What is demonstrated |
| --- | --- | --- |
| `pass-semantic-equivalence` | `PASS` | producer, consumer, and independent source digests agree |
| `fail-closed-source-drift` | `FAIL_CLOSED` | a source mismatch cannot be treated as equivalent |
| `unknown-missing-consumer` | `UNKNOWN` | missing evidence is not promoted to a denial or pass |

The producer records the source digest, semantic IR digest, and fixed `8`
obligations. The consumer independently parses raw policy/cases and binds its
source digest and generated-judge digest without importing producer-only code.
The source interpreter, generated judge, and independent reconstruction each
derive the decision from source-compiled reduction rows. The receipt requires
all three validator expectations, source/generated/independent coordinate
equivalence, and a `48`-event append-only claim ledger (`3 cases × 8 claims × 2
transitions`). Synthetic fixtures are explicitly `SYNTHETIC_FIXTURE`; the
consumer's runner-temp observation is separately `CURRENT_EVIDENCE`.

This experiment does not claim a general policy language, cryptographic
attestation, deployment freshness, or equivalence outside the modeled digest
and decision semantics.
