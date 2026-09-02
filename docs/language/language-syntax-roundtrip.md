# Language syntax round-trip evidence

## Decision

A syntax claim has no readiness value without executable meta-code. The
`languagesyntax.Evaluate` producer calls the existing parser, formatter, semantic
lowerer, and bidirectional lens over a versioned complete corpus. Its external
receipt is consumed by `self-improvement-cycle`; a catalog row alone earns zero
readiness credit.

## Fixed denominator

The `v2` registry contains exactly 44 cases: 41 valid sources and three invalid
fixtures. The current registry observation contains 44 cases, while the
independent repository observation contains 48 `.gooo` files and 794 physical
Gooo lines. Each observed file carries its individual line count and source
digest. Of the 44 cases, 43 are
`LANGUAGE_CAPABILITY` and one (`live-governance-snapshot`) is a separate
`GOVERNANCE_OBSERVATION` case. EntityFields is a language capability case; its
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
language registry transition is observed as overall `43/43` to `44/44`, with
language capability `42/42` to `43/43` and governance observation `1/1` to
`1/1`; this is a denominator observation, not an improvement score. The
EntityFields operation has no exact before/after pair, so improvement remains
`UNKNOWN`.
