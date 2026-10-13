# Language syntax round-trip corpus

`corpus.json` is the versioned denominator for `LANGUAGE-SYNTAX-ROUNDTRIP`.
Each observation records its own repository source-file and physical-line counts.
The registry contains 84 single-file cases: 81 valid source units and three
fail-closed fixtures, split into 82 language capabilities and two governance
observations. `batch-planner` and `batch-handoff` are executable ordered-list
examples checked under the explicit EntityFields V5 profile. Their source,
canonical formatting, semantic round trip, GetPut and PutGet enter the same
registered denominator. Earlier 82-case observations keep their own counts.

CI rejects an unregistered `.gooo` file, a missing registered file, an unknown
registry field, a replay mismatch, any repository write, or mutation authority.
Every complete single-file case records its physical Gooo line count and proves
AST shape, canonical bytes, semantic hash, GetPut, and PutGet. The invalid cases
must emit their registered diagnostic and fail closed.

The five members of `examples/billing-package` and
`examples/symbolic-invocation-schema` are not complete programs when read
alone. Registry `v2` routes them as two package units to
`languagepackageexecution.Evaluate`, `PACKAGE_SOURCE_FILES`, and
`PACKAGE_EXECUTIONS`; the syntax witness therefore does not falsely lower either
member at single-file resolution or count either package twice.

The witness only reads source. It does not format or rewrite repository files.
