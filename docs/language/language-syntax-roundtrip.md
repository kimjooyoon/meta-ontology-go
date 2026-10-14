# Language syntax round-trip evidence

## Decision

A syntax claim has no readiness value without executable meta-code. The
`languagesyntax.Evaluate` producer calls the existing parser, formatter, semantic
lowerer, and bidirectional lens over a versioned complete corpus. Its external
receipt is consumed by `self-improvement-cycle`; a catalog row alone earns zero
readiness credit.

## Fixed denominator

The `v2` registry declares 89 cases: 86 valid sources and three invalid
fixtures. Of these, 87 are `LANGUAGE_CAPABILITY` and two
(`self-improvement-ci-continuation` and `domain-completeness-profile`) are
`GOVERNANCE_OBSERVATION` cases. The five native origin readers are registered
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
| COHERENCE | replay-ast-bytes-semantics-and-lens-laws | all 86 valid cases satisfy all five preservation laws |
| REGRESSION | reject-invalid-syntax-with-zero-effects | all three diagnostics reject and writes and authority remain zero |

## CI authority

GitHub Actions runs the syntax producer for the submitted revision, retains its
first and replay artifacts, and checks source registration, preservation laws
and the downstream scope partition. An unregistered example makes that
revision incomplete. Its own CI results must satisfy the declared checks.
