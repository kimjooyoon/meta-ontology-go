# Language syntax round-trip corpus

`corpus.json` is the versioned denominator for `LANGUAGE-SYNTAX-ROUNDTRIP`.
It observes all 20 repository `.gooo` files and 275 physical Gooo lines. Its
single-file semantic denominator is exactly 20 cases: 17 valid source units and
three fail-closed fixtures.

CI rejects an unregistered `.gooo` file, a missing registered file, an unknown
registry field, a replay mismatch, any repository write, or mutation authority.
Every complete single-file case records its physical Gooo line count and proves
AST shape, canonical bytes, semantic hash, GetPut, and PutGet. The invalid cases
must emit their registered diagnostic and fail closed.

The two `examples/billing-package` members are not complete programs when read
alone. Registry `v2` routes them as one package unit to
`languagepackageexecution.Evaluate`, `PACKAGE_SOURCE_FILES`, and
`PACKAGE_EXECUTIONS`; the syntax witness therefore does not falsely lower either
member at single-file resolution or count the package twice.

The witness only reads source. It does not format or rewrite repository files.
