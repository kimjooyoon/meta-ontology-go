# Language syntax round-trip evidence

## Decision

A syntax claim has no readiness value without executable meta-code. The
`languagesyntax.Evaluate` producer calls the existing parser, formatter, semantic
lowerer, and bidirectional lens over a versioned complete corpus. Its external
receipt is consumed by `self-improvement-cycle`; a catalog row alone earns zero
readiness credit.

## Fixed denominator

The `v2` registry contains exactly 43 cases: 40 valid sources and three invalid
fixtures. CI also walks the repository and requires the observed `.gooo` path
set to equal the registered valid paths plus five package members. The current
fixed corpus observes 46 `.gooo` files and 766 physical Gooo lines; each file
carries its individual line count and source digest. Of the 43 cases, 42 are
`LANGUAGE_CAPABILITY` and one (`live-governance-snapshot`) is the separate
`GOVERNANCE_OBSERVATION` case. These are fixed observation denominators, not a
quality score.

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
| COHERENCE | replay-ast-bytes-semantics-and-lens-laws | all 40 valid cases satisfy all five preservation laws |
| REGRESSION | reject-invalid-syntax-with-zero-effects | both diagnostics reject and writes and authority remain zero |

## CI authority

GitHub Actions produces the receipt twice outside the repository, compares the
bytes, consumes it again, and only then passes its digest into language
readiness. Under the unchanged 24-obligation registry, a proven receipt makes
the mechanically checkable candidate transition `13/24 = 5416 BPS` to
`14/24 = 5833 BPS`, a delta of `+417 BPS`. This is not reported as achieved
until the PR transition artifact proves it.
