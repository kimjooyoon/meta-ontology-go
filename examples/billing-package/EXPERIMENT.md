# Billing operation-manifest experiment

Two Gooo source files define one symbolic user operation:
`PayOrder(Order) -> Receipt`. The compiler emits a JSON operation manifest,
not Go source and not an executable business implementation.

## Reproduce the public path

Use a `gooo` CLI built from the same checkout, available on PATH, and start
at the repository root. The existing inputs are:

`examples/billing-package/activity.gooo`:
```gooo
package billing
namespace billing

activity PayOrder(Order) -> Receipt
```

`examples/billing-package/entities.gooo`:
```gooo
package billing
namespace billing

entity Order id "urn:gooo:billing:order"
entity Receipt id "urn:gooo:billing:receipt"
```

```sh
gooo emit --kind operation-manifest --entry PayOrder examples/billing-package
```

The successful JSON has schema `gooo/operation-manifest/v1`, decision
`PASS`, resolution `EXACT`, and reason `OPERATION_MANIFEST_EMITTED`.
Its `operation.activity` is `PayOrder`; the input and output IDs are
`urn:gooo:billing:order` and `urn:gooo:billing:receipt`.
Source-definition receipts and content digests remain in the full artifact.
This resolves and projects declarations; it does not execute a payment.

An unsupported emitter is a separate user-visible rejection:
```sh
gooo emit --kind not-registered --entry PayOrder examples/billing-package
```
The expected exit code is 1, with `FAIL_CLOSED / LOWER_RESOLUTION /
EMITTER_UNKNOWN`, not an empty successful manifest.

## Existing CI loop and evidence

[The workflow](../../.github/workflows/language-example-experiment.yml) invokes
[the existing producer](../../scripts/language-example-experiment/main.sh).
It emits `first.json` and `replay.json` outside the checkout, compares their
complete bytes, records five ordered resource samples, and reduces the
artifacts against the [fixed contract](operation-manifest.contract.json) and
[independent golden](operation-manifest.golden.json).
No additional runner is needed for this path.

The experiment has 15 existing indicator coordinates. The USER 6, TOOL_AUTHOR
12, and GOVERNOR 15 views overlap; they are not 33 independent achievements.
Each row in [the evaluator](../../internal/meta/languageexampleexperiment/indicators.go)
retains its class, proof choice, meta-operation label, value, target, and exact
equality result. The [decision reducer](../../internal/meta/languageexampleexperiment/evaluate.go)
requires all 15 coordinates and the fixed count to agree.

| Indicator ID | Observed quantity or check | Fixed target |
|---|---|---:|
| `value.primary-artifact` | Primary manifest count | 1 |
| `value.artifact-digest-integrity` | Content-bound artifact digest checks | 3 |
| `value.golden-match` | Agreement with the independent domain golden | 1 |
| `value.deterministic-replay` | Complete first/replay artifact agreement | 1 |
| `compiler.source-files` | Bound definition files | 2 |
| `compiler.gooo-definition-bps` | Gooo definition ratio in basis points | 10000 |
| `compiler.emitter-registry` | Registered emitter kinds, not quality | 3 |
| `resource.samples` | Ordered runner samples | 5 |
| `resource.valid-samples` | Samples satisfying the profile contract | 5 |
| `guardrail.wall` | Wall-time limit violations | 0 |
| `guardrail.rss` | Peak-RSS limit violations | 0 |
| `guardrail.binary` | Binary-size limit violations | 0 |
| `counterexample.unknown-emitter` | Unsupported emitter rejection | 1 |
| `guardrail.effects` | Reported repository writes plus mutation-authority flag | 0 |
| `guardrail.non-claims` | Preserved experiment non-claims | 5 |

The `.gooo` declarations above are real compiler inputs. The evaluator's
`meta_operation` strings are Go-defined evidence labels; their presence alone
does not establish a released Gooo activity-to-evaluator authority binding.
Likewise, the profile's zero-write field plus the producer's repository diff
check is not a full filesystem or transient-write audit.

The [six evidence-integrity counterexamples](../../scripts/language-example-experiment/counterexamples.sh)
are additional rejection checks, not extra coordinates in the 15-item score:

| Changed input | Expected reducer reason |
|---|---|
| Unknown artifact decision | `ARTIFACT_DECISION_UNKNOWN` |
| Known failed artifact decision | `ARTIFACT_DECISION_REJECTED` |
| Content changed without resealing its digest | `ARTIFACT_DIGEST_INVALID` |
| Valid artifact from a different replay operation | `ARTIFACT_REPLAY_MISMATCH` |
| Negative peak-RSS sample | `PROFILE_SAMPLE_INVALID` |
| Independent golden changed | `ARTIFACT_GOLDEN_MISMATCH` |

All six return `FAIL_CLOSED`. Only the unknown-decision case above lowers
resolution; the five known contradictions retain `EXACT`.

## Recorded native observation, not a moving-current claim

[Native run 34521688295](https://github.com/kimjooyoon/meta-ontology-go/actions/runs/34521688295)
observed source `132fb3c8d2a391aa6a5a9ea47d13493b00182e5a`.
Its artifact ID is `10169942477`; archive SHA-256 is
`75d55f76e6298ba60c8b8c6ca2684649393ebf8527cb125d3abf54a77e8ca456`.
Both `first.json` and `replay.json` have SHA-256
`c591fca2be3ca32fe3fb6c83404e3a23908e89c32587db3aabf2eb80f47d4d00`.
Actions artifact retention is finite; this record does not promise perpetual download availability.

That run reported 15/15 coordinates, two Gooo and zero Go definition files,
one unknown-emitter rejection, and 6/6 evidence-integrity rejections.
Ordered wall samples were `[5, 5, 6, 5, 5]` ms; peak-RSS samples were
`[12964, 13036, 12856, 10900, 13160]` KiB; binary size was 14301958 bytes.
These are observations from that runner and source, not a before/after pair
or evidence of a performance improvement.

## Meaning and remaining boundary

A passing report means only `MINIMAL_VALUE_OBSERVED`. It does not claim
business correctness, value-level computation, production readiness,
performance beyond the fixed runner samples, or general-purpose code generation.
The emitter count is an extension-surface measurement, not language quality.
This path is not evidence of external user utility or whole-language completeness.

The CI rejection cases provide bounded integrity evidence for this experiment.
They do not authorize promotion into self-improvement input. Such promotion
remains a separate, unestablished authority binding rather than a consequence
of this document or a passing report.
