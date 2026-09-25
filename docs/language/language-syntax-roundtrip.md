# Language syntax round-trip evidence

## Decision

A syntax claim has no readiness value without executable meta-code. The
`languagesyntax.Evaluate` producer calls the existing parser, formatter, semantic
lowerer, and bidirectional lens over a versioned complete corpus. Its external
receipt is consumed by `self-improvement-cycle`; a catalog row alone earns zero
readiness credit.

## Fixed denominator

The `v2` registry contains exactly 65 cases: 62 valid sources and three invalid
fixtures. The current registry observation contains 65 cases, while the
independent repository observation contains 78 `.gooo` files and 2048 physical
Gooo lines. Each observed file carries its individual line count and source
digest. Of the 65 cases, 63 are
`LANGUAGE_CAPABILITY` and two (`live-governance-snapshot` and
`self-improvement-ci-continuation`) are separate `GOVERNANCE_OBSERVATION`
cases. EntityFields is a language capability case; its
12 proof activities live in the separate `internal/meta/entityfields/entity-fields-meta.gooo`
meta source and are not emitted into the user/domain Go projection. These are
fixed observation denominators, not a quality score.

## Indicators

The receipt contains 16 indicators under one schema and denominator.

| Class | Count | Exact contents |
| --- | ---: | --- |
| outcome | 1 | readiness basis points |
| driver | 9 | executed cases, valid files, invalid fixtures, AST replay, byte replay, semantic replay, GetPut, PutGet, diagnostic rejection |
| guardrail | 6 | unregistered Gooo, missing registered source, unresolved evidence, repository writes, mutation authority, registry drift |

All six guardrails have target `0`. Unknown registry shape or unavailable source
does not become a fixed point; it produces `FAIL_CLOSED` at `LOWER_RESOLUTION`.

## Munchhausen choice

| Choice | Meta-operation | Passing condition |
| --- | --- | --- |
| FOUNDATION | bind-versioned-complete-gooo-corpus | exact registry, exact commit, complete path set, bound concept artifact |
| COHERENCE | replay-ast-bytes-semantics-and-lens-laws | all 41 valid cases satisfy all five preservation laws |
| REGRESSION | reject-invalid-syntax-with-zero-effects | all three diagnostics reject and writes and authority remain zero |

## CI authority

GitHub Actions produces the EntityFields receipt twice outside the repository,
compares the bytes, and consumes the exact first/replay artifacts again. The
The registry denominator transition from 62 to 65 cases adds the incident,
repair, and repair-handoff declarations as explicit capability inputs; this is
a denominator observation, not an improvement score. The
EntityFields operation has no exact before/after pair, so improvement remains
`UNKNOWN`.
