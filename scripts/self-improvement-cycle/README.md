# Self-Improvement Cycle Envelope

This metaprogram binds one exact-head Gooo self-improvement cycle without
changing the schemas of its existing artifacts.

It consumes the canonical source metrics, generation plan, execution
manifest, receipt report, artifact provenance, and compiled Gooo semantic
contract. Raw file hashes and each artifact's semantic digest form an ordered
artifact-set digest.

The CI workflow context is bound separately so dev and main executions remain
distinguishable while their artifact sets can still be compared.

The envelope validates schema versions, exact head and base identities,
plan-to-execution-to-receipt-to-provenance links, the shared indicator ledger,
the Gooo contract indicators, executor coverage, and deterministic replay.

The result never authorizes promotion.
