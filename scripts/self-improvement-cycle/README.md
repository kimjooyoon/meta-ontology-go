# Self-Improvement Cycle Envelope

This v2 metaprogram binds one exact-head Gooo self-improvement cycle without
changing the schemas of its existing artifacts.

It consumes the canonical source metrics, generation plan, execution
manifest, receipt report, artifact provenance, and compiled Gooo semantic
contract. Raw file hashes and each artifact's semantic digest form an ordered
artifact-set digest.

The source-metrics reference carries both the logical workspace root and the
content-addressed storage root. Its semantic digest binds folder/file counts,
Go/Gooo files and lines, and the catalog indicators that produced those values.

Project-root topology remains an explicit `NOT_APPLICABLE` exception. A root
`README.md` is not required, while root counts remain observable.

The CI workflow context is bound separately so dev and main executions remain
distinguishable while their artifact sets can still be compared.

The envelope validates schema versions, exact head and base identities,
plan-to-execution-to-receipt-to-provenance links, the shared indicator ledger,
the Gooo contract indicators, executor coverage, and deterministic replay.

The result never authorizes promotion.

## v3 root metric witnesses

The envelope content-addresses the project-root exception policy together with ten
canonical witnesses: eight Go/Gooo and folder/file observations plus two
ROOT_TOPOLOGY_EXEMPT topology decisions. FOUNDATION selects schema and exception
axioms, COHERENCE binds roots, observations, witnesses, and semantic digests, and
REGRESSION binds canonical replay. The project root still does not require a
README.md.
