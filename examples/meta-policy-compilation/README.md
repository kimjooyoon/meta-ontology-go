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
| `unknown-malformed-source-digest` | `UNKNOWN` | malformed evidence is lower-resolution, not a valid contradiction |

The producer records the raw source digest, source-derived semantic contract
digest, and fixed `8` obligations. The consumer independently parses raw
policy/cases and binds raw source, artifact source, generated-judge, and
independent-reconstruction digests without importing producer-only code. The
source interpreter, generated judge, and independent reconstruction each
derive the decision from source-compiled reduction rows. The receipt keeps
eight distinct predicates—source bound, artifact bound, generated execution,
independent replay, proof selection, ledger chain, decision reduction, and
lineage seal—so a final PASS is not copied to every claim. It has a `64`-event
append-only claim ledger (`4 cases × 8 claims × 2 transitions`). Synthetic
fixtures are explicitly `SYNTHETIC_FIXTURE`; the consumer's runner-temp
artifact observation is separately `CURRENT_EVIDENCE` and is the only current
subject evidence.

This experiment does not claim a general policy language, cryptographic
attestation, deployment freshness, or equivalence outside the modeled digest
and decision semantics. The write observation is deliberately named
`repository_net_change_observed`: it proves unchanged repository status before
and after the run, not the number of filesystem writes.
