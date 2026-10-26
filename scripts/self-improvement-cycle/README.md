# Self-Improvement Cycle Envelope

This v3 metaprogram binds one exact-head Gooo self-improvement cycle without
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

The source-owned [lifecycle policy](../../internal/meta/cyclecontext/policy.gooo)
classifies observations through an ordinary generated Go library. `success` and
`failure` can produce `BOUND` when every artifact, identity, link and metric check
passes. `cancelled` produces `INTERRUPTED` with reason `SOURCE_CI_CANCELLED` under
those same obligations. A missing or invalid binding remains `OPEN`, including
when its source CI was cancelled. Other source outcomes remain unsupported and
`OPEN`; strings are not trimmed or normalized.

`-check` succeeds when a valid `BOUND` or `INTERRUPTED` observation is published.
This says the observation was constructed, not that source CI succeeded. The
original conclusion, source head/run, digests and `promotion_authorized=false`
remain in the envelope. The existing feedback consumer continues to require both
`BOUND` and source CI `success` before selecting another meta operation.

The compiled Gooo contract also declares an isolated read-only language
observation input. This envelope does not claim that the generation plan has
consumed it; the separate observation workflow binds the real experiment
receipt while keeping candidate generation and every authority bit at zero.

## v3 root metric witnesses

The envelope content-addresses the project-root exception policy together with eleven
canonical witnesses: eight Go/Gooo and folder/file observations, two
ROOT_TOPOLOGY_EXEMPT topology decisions, and one ROOT_README_EXEMPT decision.
FOUNDATION selects schema and exception axioms, COHERENCE binds roots, observations, witnesses, and semantic digests, and
REGRESSION binds canonical replay. The project root still does not require a
README.md.
