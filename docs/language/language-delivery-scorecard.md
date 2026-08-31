# Language delivery scorecard

## Decision

Gooo has two non-interchangeable denominators.

| Contract | Meaning |
|---|---|
| 24 obligations | Internal self-improvement machinery |
| 36 obligations | Observable delivery to users, tool authors, and governors |

The 24/24 receipt must never be described as universal language completeness. The delivery contract reports known gaps without turning them into unknown evidence.

## Fixed reader projections

The v5 contract owns exactly 12 obligations per reader. Views are cumulative and reduce the same fact set:

`USER subset TOOL_AUTHOR subset GOVERNOR`

Every view carries both its projection decision and the global receipt decision. A local projection therefore cannot hide a global fail-closed state.

The v5 executable receipt coordinates are `USER 11/12`, cumulative
`TOOL_AUTHOR 23/24`, and cumulative `GOVERNOR 35/36`. One obligation remains
explicitly `NOT_IMPLEMENTED`; this is still an `INCOMPLETE` global decision.

## Language-test receipt boundary

`USER-LANGUAGE-TEST` is satisfied only from the independently uploaded
`test/report.json` in the `language-source-execution-<head>` Actions artifact.
The delivery observer checks the exact head, schema, `PASS / EXACT / 12/12`
decision, two declared/executed/passed tests, two deterministic digest counts,
the assertion and missing-test counterexamples, two Go 1.27 executions, three
explicit non-claims, reader projections, and zero write effects before reading
the `passed_tests` counter.

The separately uploaded unknown-top report must become `FAIL_CLOSED /
LOWER_RESOLUTION`; it cannot mint delivery credit. The only remaining known gap
is external-dependency execution.

## Status semantics

| Status | Meaning | Resolution effect |
|---|---|---|
| SATISFIED | Exact CI receipt proves the obligation | Preserve |
| NOT_IMPLEMENTED | The fixed contract has no registered receipt producer | Exact incompleteness |
| NOT_SATISFIED | A bound receipt contradicts the obligation | Invariant only |
| UNKNOWN | Identity, schema, head, digest, or decision is unknown | Lower resolution and fail closed |

## Scope references

The denominator is a versioned Gooo product contract, not a claim that all languages require the same surface. Its scope was checked against the [Go command lifecycle](https://pkg.go.dev/cmd/go), [Go 1.27 release notes](https://go.dev/doc/go1.27), [gopls feature surface](https://go.dev/gopls/features/), and [gomacro's explicit capabilities and limitations](https://github.com/cosmos72/gomacro).

These links define scope hints only. CI artifacts, canonical contract meta-code, and exact commit binding remain the decision authority.

The scorecard runs as a downstream job in the existing Transformation effect ledger. This keeps GitHub's direct workflow-file requirement inside the repository's ten-entry physical topology limit.

## No inference

Timing and RSS remain runner-scoped observations. No historical improvement is claimed without a comparable predecessor using the same 36 IDs. Adding a capability requires an executable receipt, a meta-operation binding, and an explicit transition from `NOT_IMPLEMENTED` to evidence-backed evaluation.
