# Gooo experimental release publication v3

This contract publishes the next Gooo CLI development prerelease without
treating a pull request artifact as a release. The only allowed identity is:

```text
tag      = v0.6.11-dev
version  = 0.6.11-dev
status   = development
release  = prerelease
latest   = false
```

Previously published development tags remain immutable; this contract creates a new tag rather than replacing an existing release.

## Meta authority

The source `examples/gooo-release-publication/main.gooo` declares exactly eight
activities. CI must accept that source with `gooo check`, bind its digest to the
exact candidate SHA, and carry the digest into the release manifest and final
receipt.

| Cell | Proof choice | Required evidence |
|---|---|---|
| `VALIDATED_MERGE_EVIDENCE` | FOUNDATION | successful push-to-dev readiness run at the exact SHA |
| `EXPERIMENTAL_RELEASE_IDENTITY` | FOUNDATION | exact tag, version, status, and schema |
| `RELEASE_PAYLOAD` | COHERENCE | seven payload files assembled from validated evidence |
| `PAYLOAD_CHECKSUMS` | REGRESSION | eight non-checksum release files verify byte-exact |
| `ANNOTATED_TAG` | FOUNDATION | annotated tag resolves to the validated commit |
| `DRAFT_PRERELEASE` | COHERENCE | draft prerelease is bound to the tag |
| `DRAFT_ASSET_SET` | REGRESSION | draft contains the exact nine asset names |
| `PUBLISHED_PRERELEASE` | COHERENCE | draft becomes a non-latest published prerelease |

Proof choices are fixed by the contract and cannot be selected after observing
which route is easier.

## Fixed assets

The release asset denominator is exactly 9:

1. `gooo-darwin-amd64.tar.gz`
2. `gooo-darwin-arm64.tar.gz`
3. `gooo-linux-amd64.tar.gz`
4. `gooo-windows-amd64.zip`
5. `release-eligibility.json`
6. `release-manifest.json`
7. `release-report.json`
8. `version.json`
9. `SHA256SUMS`

`SHA256SUMS` contains exactly 8 entries, one for every other release asset. The
manifest contains the seven payload digests, the readiness run ID, exact source
SHA, report and concept digests, publication meta-source digest, version, and
tag. The Darwin arm64 asset is built and replayed on a native macOS arm64 runner. It does not contain its own digest.

## Permission separation

The workflow has three operational phases:

- `conformance`: `contents: read`; runs on pull requests and dispatches.
- `prepare`: `actions: read`, `contents: read`; validates and assembles outside the repository.
- `publish`: `actions: read`, `contents: write`; runs only on a manual dispatch after merge.

The terminal publication receipt runs with `if: always()`. A failed, missing, or
skipped prepare/publish phase becomes:

```text
status         = ACTIVE
state          = UNKNOWN
resolution     = OPERATION_CLASS
stage          = RELEASE
step           = PUBLISH_PRERELEASE
reason         = RELEASE_PUBLICATION_UPSTREAM_NOT_SUCCESS
next_operation = RESOLVE_RELEASE_PUBLICATION_FAILURE
```

UNKNOWN cannot be converted to PUBLISHED by a human explanation.

## Readiness boundary

For 0.6.11, each native platform witness also uses its candidate binary to
construct and replay the division, candidate-local retry and filename examples.
Their eight, twelve and twelve execution cases must all match, with zero new
model calls and the same selected program on replay. It also exports the v3
source graph, binds the source/context digests, requires zero predictions and
candidate execution, and consumes the graph through the public SDK. Platform
artifacts retain six raw language runs, one graph export and two filename
workspace execution/replay receipts each. The latter use three actual inputs;
readiness checks their fields while preserving the input-only unscored status.
They require zero runtime/replay predictions and the same generated program.
The three-package diagnostic also runs with explicit and source-derived metadata,
each constructed and replayed. Its four input rows produce eight named expected
outputs in each run. The witness compares every actual and reported expected
value with the independent case file, keeps the generated program identical and
requires zero new runtime/replay predictions. Four additional raw package receipts
are retained per platform. These finite
behavior checks are separate from the 26 structural release cases and the nine
published asset identities. No trained v3 quality is claimed by readiness.

The selected readiness run must have all of these properties:

- workflow name `Gooo release readiness`;
- event `push`;
- branch `dev`;
- conclusion `success`;
- head SHA equal to the dispatch input and current `dev` head;
- eligibility decision `EVIDENCE_CLOSED / EXACT`;
- next operation `PUBLISH_GOOO_EXPERIMENTAL_RELEASE`;
- 7/7 readiness cells closed;
- repository writes 0.

The workflow re-evaluates the release witness and bundle instead of trusting
only the upstream job conclusion.

## Publication and retry behavior

The tag and release must not exist when publication starts. The workflow creates
an annotated tag, stages a draft prerelease with all nine assets, verifies the
tag target and exact asset set, and only then publishes it. Existing tags or
releases fail closed and are never overwritten.

A failure after tag creation can leave an annotated tag or draft release. The
workflow does not silently delete or overwrite either state. Recovery requires
an explicit inspection and separate decision.

The published release does not claim SLSA compliance, signed provenance,
performance guarantees, or stable language compatibility.
