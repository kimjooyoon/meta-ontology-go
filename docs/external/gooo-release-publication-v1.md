# Gooo experimental release publication v1

This contract publishes the first Gooo CLI prerelease without treating a pull
request artifact as a release. The only allowed identity is:

```text
tag      = v0.2.0-dev
version  = 0.2.0-dev
status   = development
release  = prerelease
latest   = false
```

Stable `v0.1.0` is outside this contract.

## Meta authority

The source `examples/gooo-release-publication/main.gooo` declares exactly eight
activities. CI must accept that source with `gooo check`, bind its digest to the
exact candidate SHA, and carry the digest into the release manifest and final
receipt.

| Cell | Proof choice | Required evidence |
|---|---|---|
| `VALIDATED_MERGE_EVIDENCE` | FOUNDATION | successful push-to-dev readiness run at the exact SHA |
| `EXPERIMENTAL_RELEASE_IDENTITY` | FOUNDATION | exact tag, version, status, and schema |
| `RELEASE_PAYLOAD` | COHERENCE | six payload files assembled from validated evidence |
| `PAYLOAD_CHECKSUMS` | REGRESSION | seven non-checksum release files verify byte-exact |
| `ANNOTATED_TAG` | FOUNDATION | annotated tag resolves to the validated commit |
| `DRAFT_PRERELEASE` | COHERENCE | draft prerelease is bound to the tag |
| `DRAFT_ASSET_SET` | REGRESSION | draft contains the exact eight asset names |
| `PUBLISHED_PRERELEASE` | COHERENCE | draft becomes a non-latest published prerelease |

Proof choices are fixed by the contract and cannot be selected after observing
which route is easier.

## Fixed assets

The release asset denominator is exactly 8:

1. `gooo-darwin-amd64.tar.gz`
2. `gooo-linux-amd64.tar.gz`
3. `gooo-windows-amd64.zip`
4. `release-eligibility.json`
5. `release-manifest.json`
6. `release-report.json`
7. `version.json`
8. `SHA256SUMS`

`SHA256SUMS` contains exactly 7 entries, one for every other release asset. The
manifest contains the six payload digests, the readiness run ID, exact source
SHA, report and concept digests, publication meta-source digest, version, and
tag. It does not contain its own digest.

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
an annotated tag, stages a draft prerelease with all eight assets, verifies the
tag target and exact asset set, and only then publishes it. Existing tags or
releases fail closed and are never overwritten.

A failure after tag creation can leave an annotated tag or draft release. The
workflow does not silently delete or overwrite either state. Recovery requires
an explicit inspection and separate decision.

The published release does not claim SLSA compliance, signed provenance,
performance guarantees, or stable language compatibility.
