# Conflict-free registry projection

This experiment keeps concept registration local and makes the shared view a
deterministic projection. The bounded slice has three real local inputs:

- `examples/language-syntax-roundtrip/concept.manifest.json`
- `examples/language-semantic-model/concept.manifest.json`
- `examples/toolchain-conformance/concept.manifest.json`

The producer is `scripts/conflict-free-registry-projection`. It discovers local
manifests, sorts them by stable ID, validates the manifest boundary, reads the
referenced corpus, registry, and documentation files, and emits the generated
catalog, corpus, registry, denominator, README, manifest digests, and full
projection. `generate` is the only write mode; `check` is read-only and fails
closed for a missing, unexpected, or stale generated projection.

The independent consumer in
`scripts/conflict-free-registry-projection-consumer` parses raw manifests and
reconstructs `projection.json` without importing the producer package. CI
compares its bytes with the generated projection.

## Fixed baseline

For this slice, one touchpoint is one unique shared path whose current content
contains a selected concept, corpus, registry, README, documentation, or fixed
denominator token. The current `origin/dev` baseline is exactly 12/12 observed
manual shared registration touchpoints. The three local manifests provide
3/3 concept-local touchpoints. The projected path removes the need for manual
global edits: 12/12 becomes 0/12, and the conflict surface becomes 10000 to 0
basis points. The baseline list is executable in the producer, so a missing or
drifted baseline path fails closed.

The existing global catalog and readiness registries remain unchanged. This is
intentional: full migration is outside the slice, and root topology and the
root README remain exceptions.

## Meaning gates

`prove` executes the following evidence contract without writing the source
repository:

- two projections are byte-for-byte equal;
- reversing manifest discovery order preserves every output byte;
- a semantic manifest change changes catalog, denominator, documentation, and
  digest surfaces while corpus and registry projections stay unchanged;
- a comment-only manifest change changes raw manifest digests while the
  semantic projection stays byte-identical;
- a new concept fixture changes only generated outputs and leaves existing
  concept inputs untouched;
- duplicate stable IDs, missing manifests, and stale generated output fail
  closed while preserving `stage`, `step`, and `reason`;
- `FOUNDATION`, `COHERENCE`, and `REGRESSION` are selected from the local
  manifests and each strategy must discharge its own checks;
- every claim has a distinct proposition and target digest with an explicit
  `OPEN` to `DISCHARGED` or `REFUTED` transition chain.

Repository net-state observations are reported separately from generated output
path, digest, and byte metadata. Mutation authority is reported as `UNKNOWN`
unless an authorized observer establishes it; a clean net-state observation is
not treated as proof of mutation authority.
