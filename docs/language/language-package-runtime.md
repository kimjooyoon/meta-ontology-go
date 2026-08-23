# Language package runtime

## Decision

`LANGUAGE-PACKAGE-RUNTIME` is satisfied only when CI executes the versioned
18-case corpus at the exact PR head. A catalog entry without that report is
not sufficient.

The runtime normalizes a manifest, orders a package import DAG, parses and
lowers every `.gooo` source through the existing syntax and semantic IR, and
resolves one activity contract. It produces an immutable runtime image and
invocation plan. It does not execute handwritten Go function bodies, load Go
plugins, access the network, or write the repository.

## Fixed denominator

| Indicator | Exact target |
| --- | ---: |
| Corpus | 18/18 |
| Positive paths | 10 |
| Guardrail rejections | 8 |
| Compiled packages | 40 |
| Lowered sources | 50 |
| Resolved imports | 40 |
| Package initializations | 40 |
| Entry bindings | 10 |
| Semantic bindings | 50 |
| Canonical replays | 10 |
| Order-invariant replays | 3 |

The report contains exactly 18 indicators: 3 outcome, 8 driver, and 7
guardrail. `FOUNDATION` binds the corpus and semantic pipeline,
`COHERENCE` binds package order to the entry contract, and `REGRESSION`
requires all invalid boundaries and effects to remain zero.

## Structural reference

[gomacro](https://github.com/cosmos72/gomacro) demonstrates treating Go AST
as data evaluated in stages and separates package import from evaluation.
This implementation adopts those boundary ideas without adding a gomacro
dependency or claiming interpreter capability. That distinction avoids the
platform-dependent runtime import and plugin authority described by gomacro.

## Observed positive concepts

A package can be compared by its canonical semantic image rather than input
ordering. A multi-source entry activity becomes inspectable before any
effect is authorized. Invalid package graphs become named evidence rather
than incidental loader errors. These are repository observations, not
novelty claims.
