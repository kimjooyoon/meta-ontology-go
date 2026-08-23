# Language diagnostic provenance

## Decision

The language diagnostic obligation is satisfied only by an exact CI report
over the versioned 18-case registry. Registering the concept without that
report leaves `LANGUAGE-DIAGNOSTIC-PROVENANCE` not satisfied.

The metaprogram reuses existing boundaries rather than introducing a second
error system:

1. `go/token` retains physical and line-directed logical positions.
2. `go/scanner` and `go/types` provide ordered parse and typed diagnostics.
3. `internal/formatter` supplies stable diagnostic codes and spans.
4. `internal/generator` maps generated ranges to semantic IDs.
5. `internal/lsp` receives the normalized source range.

## Fixed denominator

The registry is `gooo/language-diagnostic-provenance-cases/v1`, version
`2026-08-23`. It contains 3 syntax cases, 3 type cases, 4 semantic source-map
cases, and 8 guardrail cases. The denominator cannot change without a schema
or version change.

Exact success is:

| Quantity | Exact target |
| --- | ---: |
| Cases | 18/18 |
| Positive traces | 10 |
| Guardrail rejections | 8 |
| Physical and logical positions | 10 and 10 |
| Semantic ID bindings | 4 |
| LSP projections and canonical replays | 10 and 10 |
| Ordered diagnostic sets | 6 |
| Line-directive remaps | 1 |
| Type hardness classifications | 3 |
| Provenance steps | 50 |

## Indicators

The report has exactly 18 indicators: 3 outcome, 8 driver, and 7 guardrail.
Every indicator names its producer, consumer, proof choice, and meta
operation. Unknown, missing, ambiguous, invalid, or effectful observations
must remain zero.

## Munchhausen choices

`FOUNDATION` binds the fixed registry, Go 1.27 position/type APIs, concept,
code, metrics, and use cases. `COHERENCE` replays the
physical-to-logical-to-semantic-to-LSP chain. `REGRESSION` proves all eight
unknown boundaries fail closed.

## Positive concepts observed

Dual physical/logical coordinates preserve generated-code debugging without
discarding the source author view. Semantic source-map reversal makes a
diagnostic actionable by stable identity rather than only line text.
Canonical trace replay makes diagnostic changes comparable across CI runs.
These are observed combinations in this repository, not novelty claims.

The AST-reification approach is informed by
[gomacro](https://github.com/cosmos72/gomacro), but this package adds no
gomacro dependency or runtime authority. Position semantics follow
[`go/token`](https://pkg.go.dev/go/token), diagnostic ordering follows
[`go/scanner`](https://pkg.go.dev/go/scanner), type error classification
follows [`go/types`](https://pkg.go.dev/go/types), and generated-source
remapping follows [Go line directives](https://go.dev/wiki/Comments#line-directives).
