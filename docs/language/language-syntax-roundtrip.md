# Language syntax round-trip evidence

## Decision

A syntax claim has no readiness value without executable meta-code. The
`languagesyntax.Evaluate` producer calls the existing parser, formatter, semantic
lowerer, and bidirectional lens over a versioned complete corpus. Its external
receipt is consumed by `self-improvement-cycle`; a catalog row alone earns zero
readiness credit.

## Fixed denominator

At corpus blob `b5181791baa2f347f44bedbc99597a72c43a311e`, the `v2`
registry declares 137 cases: 134 valid sources and three invalid fixtures.
The valid sources comprise 132 `LANGUAGE_CAPABILITY` cases and two
`GOVERNANCE_OBSERVATION` cases (`self-improvement-ci-continuation` and
`domain-completeness-profile`). All three invalid fixtures are language
capability cases. The twelve native origin readers are registered as language
capability cases under the exact V5 field profile. New registrations change
these totals; the linked corpus is authoritative for the current checkout.

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
The [original field-file reader](../../examples/field-source-map/README.md) is
registered at version 67. Its [source-location observation](../observations/field-source-locations/README.md)
retains the nine generated roles and source snapshot. The consumer distinguishes
field target/value locations and assembly fragments in Gooo.
The [scoped-counter origin coverage reader](../../examples/scoped-origin-coverage/README.md)
is registered at version 68 through the same nine-role operation. It expresses
row mapping and page scope in Gooo with scoped `for let` counters and unit updates.
The [length-aware summary](../../examples/length-aware-summary/README.md) is
registered at version 69. It joins length-derived values with Integer contexts
and retains the source of newly accepted length locals in native observations.
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
| COHERENCE | replay-ast-bytes-semantics-and-lens-laws | all valid cases in the pinned corpus satisfy all five preservation laws |
| REGRESSION | reject-invalid-syntax-with-zero-effects | all three diagnostics reject and writes and authority remain zero |

## CI authority

GitHub Actions runs the syntax producer for the submitted revision, retains its
first and replay artifacts, and checks source registration, preservation laws
and the downstream scope partition. An unregistered example makes that
revision incomplete. Its own CI results must satisfy the declared checks.

Version72 registers the two explicit-candidate and two derived-candidate source/reader
programs. The syntax capability target is108; governance cases remain2 and invalid
fixtures remain3. Earlier version65..71 denominator bytes and digests are preserved.

Version 73 registers the [ordinary variables/condition program](../../examples/everyday-program/README.md)
through the existing source-bound registration operation. Its syntax target is 109.
All 110 preceding cases and their classifications remain, with two governance
observations and three invalid-source fixtures. The v72 bytes retain target 108;
the other five boundaries and 13 links remain unchanged.
