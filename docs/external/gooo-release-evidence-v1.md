# Gooo experimental release evidence v1

This contract separates a release candidate from a published release. Passing it
does not create a tag or a GitHub Release. It proves that one exact commit has the
finite evidence required for the later `v0.1.0-dev` prerelease operation.

## Authority

The semantic authority is the existing Gooo language concept artifact for
`toolchain-cross-platform-release`. Its meta operation is
`assemble-exact-cross-platform-release`. The Go witness evaluates the fixed
release corpus and binds its report to the concept digest. The workflow only
projects that report for people; it does not redefine readiness.

The authoritative machine schemas are:

- `gooo/toolchain-release-platform-receipt/v1`
- `gooo/toolchain-cross-platform-release-corpus/v1`
- `gooo/toolchain-cross-platform-release-report/v1`
- `gooo/release-eligibility/v1`

## Fixed denominators

| Evidence | Required result |
|---|---:|
| Exact candidate head | 1/1 |
| Native platform receipts | 3/3 |
| Aggregate reports | 1/1 |
| Release cases | 20/20 |
| Meta-bound indicators | 39/39 |
| Munchausen proof choices | 3/3 |
| Zero-write observations | 1/1 |

The three platforms are Linux AMD64, macOS Intel AMD64, and Windows AMD64.
The proof denominator contains exactly one `FOUNDATION`, one `COHERENCE`, and
one `REGRESSION` aggregate proof. The 39 indicators contain 3 outcomes, 16
drivers, and 20 guardrails.

## Human projection

`gooo/release-eligibility/v1` contains exactly seven cells:

| Cell | Proof choice | Closed by |
|---|---|---|
| `EXACT_HEAD_BOUND` | FOUNDATION | checkout and report bind the requested SHA |
| `PLATFORM_RECEIPTS` | FOUNDATION | three native receipts exist and validate |
| `AGGREGATE_REPORT` | COHERENCE | one deterministic aggregate report validates |
| `RELEASE_CASES` | COHERENCE | 20 of 20 fixed cases are satisfied |
| `META_INDICATORS` | COHERENCE | 39 of 39 concept-bound indicators are satisfied |
| `MUNCHAUSEN_PROOFS` | REGRESSION | all three fixed proof branches are present |
| `REPOSITORY_ZERO_WRITE` | REGRESSION | the witness reports and CI observes zero repository writes |

The projection carries the source report digest, concept digest, exact SHA, and
meta operation. A projection without those bindings is not release evidence.

## UNKNOWN and failure

The terminal `Gooo release eligibility` job runs with `if: always()`. Matrix
failure, missing artifacts, aggregate failure, and skipped upstream jobs cannot
be interpreted as success. Before enforcing the result, the job writes a
fail-closed receipt with these coordinates:

```text
status         = ACTIVE
state          = UNKNOWN
resolution     = OPERATION_CLASS
stage          = CI
step           = AGGREGATE_RELEASE_EVIDENCE
reason         = RELEASE_READINESS_UPSTREAM_NOT_SUCCESS
next_operation = RERUN_GOOO_RELEASE_READINESS
```

When all evidence closes, the candidate-evidence claim is `DISCHARGED`. This
does not discharge the separate claim that a public release exists.

## PR, merge, and publication boundary

The workflow runs against the exact pull request head and runs again on every
push to `dev`. Pull request evidence cannot be reused for the merge commit.

After a pull request run, the next operation is
`MERGE_VALIDATED_CANDIDATE`. After an exact `dev` merge-SHA run, the next
operation is `PUBLISH_GOOO_EXPERIMENTAL_RELEASE`.

Publication remains a separate operation. It must verify the successful
merge-SHA receipt, create an annotated `v0.1.0-dev` tag at that same SHA, build
the fixed assets, and publish a GitHub prerelease. Stable `v0.1.0` remains
invalid while `gooo version --json` reports version `0.1.0-dev` and status
`development`.

The public Workgraph project consumes only a released CLI, checksums, and
versioned JSON or NDJSON. It does not import core packages or block core merges.

## Refuting counterexamples

The eligibility job must fail if any of these are observed:

- only 2 of 3 platform receipts are available;
- a platform or aggregate job is skipped;
- the report is bound to a pull request merge ref instead of the exact head;
- fewer than 20 cases or 39 indicators are satisfied;
- an unknown decision value is emitted;
- any guardrail is unsatisfied;
- replay output differs;
- repository writes are nonzero or unobserved;
- the concept digest or source report digest is absent.

