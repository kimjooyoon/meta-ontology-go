# Language syntax round-trip corpus

`corpus.json` is the versioned denominator for `LANGUAGE-SYNTAX-ROUNDTRIP`.
Each observation records its own repository source-file and physical-line counts.
The current registry contains 137 single-file cases: 134 valid source units
and three fail-closed fixtures, split into 135 language capabilities and two
governance observations. Each case has a stable source identity and contributes
to the fixed syntax replay denominator.

CI rejects an unregistered `.gooo` file, a missing registered file, an unknown
registry field, a replay mismatch, any repository write, or mutation authority.
Every complete single-file case records its physical Gooo line count and proves
AST shape, canonical bytes, semantic hash, GetPut, and PutGet. The invalid cases
must emit their registered diagnostic and fail closed.

The `collection-filter` program is a complete source and enters this fixed
denominator. Its separate package execution cases check its runtime behavior.

The registry also has four package units with seven source members. Their
members are tracked as package sources because they are not all complete
programs when read alone. Package execution supplies the corresponding
checks; the syntax witness does not count those members as single-file cases.

The witness only reads source. It does not format or rewrite repository files.
