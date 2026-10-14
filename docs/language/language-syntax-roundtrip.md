# Language syntax round-trip evidence

## Decision

A syntax claim has no readiness value without executable meta-code. The
`languagesyntax.Evaluate` producer calls the existing parser, formatter, semantic
lowerer, and bidirectional lens over a versioned complete corpus. Its external
receipt is consumed by `self-improvement-cycle`; a catalog row alone earns zero
readiness credit.

## Fixed denominator

The `v2` registry declares 101 cases: 98 valid sources and three invalid
fixtures. Of these, 99 are `LANGUAGE_CAPABILITY` and two
(`self-improvement-ci-continuation` and `domain-completeness-profile`) are
`GOVERNANCE_OBSERVATION` cases. The ten native origin readers are registered
as language capability cases under the exact V5 field profile.

Case totals come from the [corpus](../../examples/language-syntax-roundtrip/corpus.json)
and [native model](../../internal/meta/languagereadiness/languagesyntax/model.go).
The repository file inventory is a separate population: package members, meta
sources and structure projections also have explicit registrations. Each
observed file carries its own source digest and line count.

The [registration observation](../observations/native-origin-registration/README.md)
records the first four reader additions and denominator versions 50 through 53.
The [local-variable observation](../observations/native-local-origins/README.md)
adds the fifth reader through the same registration operation, producing version 54.
The [local assembly observation](../observations/local-fill-origins/README.md)
adds the sixth reader at version 55 and bounds generated selection helpers to
16 capability entries each.
The [derived-origin example](../../examples/derived-origin/README.md) and its Gooo
coverage reducer are registered at versions 56 and 57.
The [original-file reader](../../examples/native-origin-reader/local-files/README.md)
and [workspace local-origin example](../../examples/workspace-local-origins/README.md)
are registered at versions 58 and 59. Their source snapshots and generated
members are retained in the [file-origin observation](../observations/workspace-local-file-origins/README.md).
The [typed source and Gooo ancestry examples](../../examples/typed-value-origins/README.md)
are registered at versions 60 and 61. Their requests, input snapshots and native
registration outputs are retained in the [typed-origin observation](../observations/typed-value-origins/README.md).
The [native record-field example and reader](../../examples/native-record-fields/README.md)
are registered at versions 62 and 63. Their exact requests and generated members
are retained in the [native-field observation](../observations/native-record-fields/README.md).
The [Gooo field ancestry reader](../../examples/native-field-lineage/README.md)
is registered at version 64 through the same nine-role operation. Its corpus
entry uses ordered collections; the [new observation](../observations/native-field-lineage/README.md)
retains the request, exact generated members and bounded execution results.
The [field/local bridge source and Gooo reader](../../examples/local-field-bridge/README.md)
are registered at versions 65 and 66. The [bridge observation](../observations/local-field-bridge/README.md)
retains the exact source-bound proposals, including the replaced unpublished reader
proposal and the new source operation that preserves known roots under partial data.
These counts identify what a submitted revision must check; CI supplies its
actual results.

## Indicators

The receipt contains 19 indicators under one schema and denominator.

| Class | Count | Exact contents |
| --- | ---: | --- |
| outcome | 1 | readiness basis points |
| driver | 11 | executed cases, valid files, invalid fixtures, AST replay, byte replay, semantic replay, GetPut, PutGet, diagnostic rejection, projection units, source-bound projections |
| guardrail | 7 | unregistered Gooo, missing registered source, unresolved evidence, repository writes, mutation authority, registry drift, unresolved projections |

All seven guardrails have target `0`. Unknown registry shape or unavailable source
does not become a fixed point; it produces `FAIL_CLOSED` at `LOWER_RESOLUTION`.

## Munchhausen choice

| Choice | Meta-operation | Passing condition |
| --- | --- | --- |
| FOUNDATION | bind-versioned-complete-gooo-corpus | exact registry, exact commit, complete path set, bound concept artifact |
| COHERENCE | replay-ast-bytes-semantics-and-lens-laws | all 98 valid cases satisfy all five preservation laws |
| REGRESSION | reject-invalid-syntax-with-zero-effects | all three diagnostics reject and writes and authority remain zero |

## CI authority

GitHub Actions runs the syntax producer for the submitted revision, retains its
first and replay artifacts, and checks source registration, preservation laws
and the downstream scope partition. An unregistered example makes that
revision incomplete. Its own CI results must satisfy the declared checks.
