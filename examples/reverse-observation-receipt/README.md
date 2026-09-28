# Reverse-observation receipt

This example defines a durable provenance receipt for a generated language artifact.

The receipt binds the source, `.gooo` declaration, intermediate representation, generated artifact, and reverse observation. `evidence_prefix_digest` identifies the immutable prefix of evidence already observed. If a stage is missing or the reverse observation disagrees with the generated artifact, the result remains `DEFERRED`, `UNKNOWN`, or `MISMATCH`; it is never relabeled as success.

The receipt records origin and observation identity. It does not execute generated code, authorize a provider, or claim that the language is complete.
