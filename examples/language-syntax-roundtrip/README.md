# Language syntax round-trip corpus

`corpus.json` is the versioned denominator for `LANGUAGE-SYNTAX-ROUNDTRIP`.
It registers all 13 repository `.gooo` files and two invalid fixtures. The
registered valid corpus contains exactly 174 physical Gooo lines.

CI rejects an unregistered `.gooo` file, a missing registered file, an unknown
registry field, a replay mismatch, any repository write, or mutation authority.
Every valid case records its physical Gooo line count and proves AST shape,
canonical bytes, semantic hash, GetPut, and PutGet. The invalid cases must emit
their registered diagnostic and fail closed.

The witness only reads source. It does not format or rewrite repository files.
