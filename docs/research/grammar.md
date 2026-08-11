# `.gooo` grammar research

Status: proposal only. This note reviews the current `.gooo` example and the
syntax implementation on `agent/syntax` (lexer/parser/AST/diagnostics). It does
not change `internal/syntax` or claim that the proposed extensions are already
implemented.

## Recommendation

Keep the current surface language as the compatibility baseline:

```gooo
package billing
namespace billing

entity Order id "billing://entity/order"
activity PayOrder(Order, PaymentMethod) -> Payment
```

The next syntax revision should add only two things:

1. An optional, explicit source-version header: `version "gooo/v1"`.
2. An optional activity identity clause, in the same position as the entity
   identity clause: `activity PayOrder id "billing://activity/pay-order" (...) -> ...`.

The version header is optional for the v1 grammar, so existing files continue
to parse as v1. An explicit activity ID is the only rename-safe identity. A
legacy activity without an ID may be lowered with a deterministic inferred ID
as a migration bridge, but the compiler should mark that identity as inferred
and recommend adding an explicit ID.

Keep syntax small and fixed-position. Do not add a general attribute bag,
implicit relation declarations, or a second expression language until the
semantic authority and round-trip behavior of each feature are specified.

## Current contract

The current parser accepts the following shape. Whitespace, newlines, `//`,
`#`, and `/* ... */` comments are not significant, so declarations may also
appear on one line.

```ebnf
file           = package-decl , namespace-decl , { declaration } , EOF ;
package-decl  = "package" , identifier ;
namespace-decl = "namespace" , identifier ;
declaration   = entity-decl | activity-decl ;
entity-decl   = "entity" , identifier , "id" , string ;
activity-decl = "activity" , identifier , "(" , [ parameter-list ] , ")" ,
                "->" , identifier ;
parameter-list = identifier , { "," , identifier } ;
```

Observed lexical rules are deliberately narrow:

- An identifier starts with `_` or a Unicode letter and continues with those
  characters or Unicode digits. The words `package`, `namespace`, `entity`,
  `id`, and `activity` are reserved in identifier position.
- Strings use double quotes. The lexer decodes `\"`, `\\`, `\n`, `\r`, `\t`,
  and `\uXXXX` while retaining the exact source spelling in the token.
- `->`, parentheses, and commas are the only non-string punctuation currently
  needed by the grammar.
- Token spans are half-open. Offsets count UTF-8 bytes; lines and columns are
  one-based, and columns count Unicode code points.

The AST keeps source spans for the file, headers, declarations, names, IDs, and
activity references. The parser returns a partial AST together with ordered
diagnostics rather than failing on the first error. Lexer diagnostics precede
parser diagnostics, and the implementation guarantees a deterministic result
for repeated parses.

This baseline has one important semantic limitation: entities carry an
explicit ID, but activities do not. Therefore an old activity declaration can
be parsed, but its identity cannot be rename-safe unless the lowering layer has
already established a durable mapping. The absence of an activity ID should
not be mistaken for a stable identity.

## Proposed versioned grammar

The additive grammar is:

```ebnf
file            = package-decl , [ version-decl ] , namespace-decl ,
                  { declaration } , EOF ;
version-decl    = "version" , string ;
declaration     = entity-decl | activity-decl ;
entity-decl     = "entity" , identifier , id-clause ;
activity-decl   = "activity" , identifier , [ id-clause ] , "(" ,
                  [ parameter-list ] , ")" , "->" , identifier ;
id-clause       = "id" , string ;
parameter-list  = identifier , { "," , identifier } ;
```

For the first revision, the only accepted version string is exactly
`"gooo/v1"`. A missing version means `"gooo/v1"`. The version declaration is
allowed once, immediately after `package`, and must precede `namespace`.
Unknown versions must produce an explicit unsupported-version diagnostic; a
parser must not silently guess the nearest grammar.

This gives the new parser both of the following forms without changing the
meaning of the original form:

```gooo
package billing
namespace billing
entity Order id "billing://entity/order"
activity PayOrder(Order) -> Payment
```

```gooo
package billing
version "gooo/v1"
namespace billing
entity Order id "billing://entity/order"
entity Payment id "billing://entity/payment"
activity PayOrder id "billing://activity/pay-order" (Order) -> Payment
```

The language version describes the source grammar and parser contract. It must
not be overloaded to version the PROV vocabulary, business ontology, or Go
generator. Those are separate compatibility surfaces and need their own
versioned metadata at the IR or projection boundary.

### Extension rules

Future syntax should follow these rules:

- Add new top-level forms behind a declared language version or a documented
  backward-compatible minor capability. Do not reinterpret an existing token
  sequence.
- Keep new clauses in fixed positions and make their authority explicit:
  authoritative source, deterministic derived fact, or candidate assertion.
- Preserve declaration boundaries at `entity`, `activity`, and future
  declaration keywords so an editor can recover after an incomplete edit.
- Require a round-trip rule before adding syntax: the accepted source must
  lower deterministically, and a projection/lift must not invent a new fact
  without source-backed evidence.
- Prefer quoted values for opaque identifiers and metadata. Do not introduce
  punctuation solely to make the grammar look more conventional; every new
  delimiter increases lexer, recovery, and compatibility surface.

Explicit relations, constraints, capabilities, assertions, handwritten slots,
and projection options remain future forms. They should not be encoded as
unrecognized trailing tokens on v1 declarations.

## Stable semantic identity

An ID is an identity, not a display name. The quoted spelling is source syntax;
the decoded string is the semantic value. The following rules keep identity
independent from presentation:

- Every durable entity must have an explicit `id` clause. A durable activity
  should have one as well; the optional activity clause exists only to read old
  v1 files.
- An explicit ID must be non-empty, absolute, URI-like, and free of whitespace
  or control characters after string decoding. URI parsing and policy checks
  belong to semantic validation, not to the lexer.
- IDs are compared as decoded, exact strings. Do not lowercase, Unicode-fold,
  normalize path segments, or derive a new ID from a renamed display name.
- A project-wide duplicate ID is an error. An ID changing its declaration kind
  is also an error unless a migration explicitly models the old ID as
  deprecated and the new declaration as a replacement.
- Namespace and display name are lookup and presentation data. They do not
  become identity by string coincidence. For example,
  `billing::Payment` and `settlement::Payment` remain distinct even when their
  display names match.
- The recommended convention is a kind-bearing URI-like path such as
  `billing://entity/order` or `billing://activity/pay-order`. This convention
  is a lint/policy aid, not a parser rule; externally owned absolute IDs must
  remain possible.
- A rename that keeps the ID is an in-place semantic rename. A changed ID is a
  remove-plus-add migration and must produce a provenance/evidence record for
  downstream references.

For a legacy activity without an ID, the lowering layer may temporarily infer
`<namespace>://activity/<lower-kebab(display-name)>` when the mapping is
unambiguous. This is compatibility behavior, not a stable-identity promise:
the lowering result should carry `inferred=true` and emit a migration warning.
Non-ASCII or colliding names must require an explicit ID rather than guessing.

The ID clause is intentionally placed immediately after the declaration name:

```gooo
activity AuthorizePayment id "billing://activity/pay-order" (Order) -> Payment
```

Here the display name changed but the semantic identity did not. The generated
Go symbol, documentation label, and aliases may change; references keyed by
`billing://activity/pay-order` must continue to resolve to the same node.

## Recovery and diagnostics

The current diagnostic API is a good compatibility foundation: machine-readable
codes, severity, source spans, and a partial AST. The following behavior should
be treated as the intended contract for the versioned parser.

### Recovery behavior

1. Lexing always makes progress. An illegal character becomes an illegal token
   plus a diagnostic; later source is still tokenized.
2. Header failures recover at the next header or declaration boundary. A
   missing `package` must not cause the same token to be reported as both a
   missing package and a missing namespace without attempting synchronization.
3. Entity failures recover at the next `entity`, `activity`, or EOF boundary.
4. Activity parameter failures recover at `,`, `)`, `->`, the next declaration
   keyword, or EOF. A missing comma should not discard a later valid activity.
5. An unsupported version blocks semantic lowering, but the parser may retain
   a best-effort syntax tree for editor features. It must not lower unknown
   constructs under v1 rules.
6. Every recovery path must advance or return. Recovery must be deterministic,
   bounded by the source, and panic-free for malformed UTF-8, unterminated
   strings, and incomplete edits.

Diagnostics should keep stable code strings while allowing human messages to
improve. Tests and LSP clients should rely on `Code`, `Severity`, and `Span`,
not message text. The current codes cover unexpected characters, unterminated
comments/strings, invalid escapes, missing grammar elements, and unexpected
declarations. The versioned extension should add, at minimum:

```text
parse.unsupported-version
parse.duplicate-version
semantic.invalid-id
semantic.duplicate-id
semantic.implicit-id
semantic.unknown-reference
```

`semantic.implicit-id` should be a warning for legacy activity inference;
`parse.unsupported-version`, invalid IDs, duplicate IDs, and unresolved
references should be errors before IR lowering. Keeping syntax and semantic
diagnostics distinct prevents a URI policy change from silently changing the
lexer grammar.

Diagnostic ordering should remain deterministic: source order within a phase,
lexer phase before parser phase, and semantic diagnostics after a successful
syntax phase. A future related-span or fix-it field may be added, but the
primary span and code must remain stable for existing clients.

## Backward-compatibility fixtures

These are documentation fixtures for the parser/IR conformance suite. They are
shown here rather than added as source files so this research change remains
limited to `docs/research/grammar.md`.

### Fixture A: current v1 billing source

Input (accepted today, with no version header):

```gooo
package billing
namespace billing

entity Order id "billing://entity/order"
entity PaymentMethod id "billing://entity/payment-method"
entity Payment id "billing://entity/payment"

activity PayOrder(Order, PaymentMethod) -> Payment
```

Expected compatibility behavior for a versioned parser:

- Parse as implicit `gooo/v1` with no syntax diagnostics.
- Preserve three entity declarations and one activity declaration, including
  source spans and the exact decoded entity IDs.
- Derive `PayOrder used Order`, `PayOrder used PaymentMethod`, and
  `Payment wasGeneratedBy PayOrder` as deterministic facts only after name
  resolution succeeds.
- Treat the activity identity as inferred (recommended warning
  `semantic.implicit-id`) until the source adds an explicit activity ID.
- Do not infer any relation from the display-name coincidence between
  `billing` and the `billing://` URI scheme.

### Fixture B: explicit v1 spelling

```gooo
package billing
version "gooo/v1"
namespace billing

entity Order id "billing://entity/order"
entity Payment id "billing://entity/payment"
activity PayOrder id "billing://activity/pay-order" (Order) -> Payment
```

Expected result: the same v1 declaration and derived-fact shape as Fixture A,
except the activity identity is explicit and therefore rename-safe. A baseline
parser that does not know the optional version/id clauses should reject this as
a newer source form; that is expected. Backward compatibility means that a
newer parser accepts Fixture A without reinterpretation, not that an old binary
can parse every future source form.

### Fixture C: rename with stable identity

```gooo
package billing
version "gooo/v1"
namespace billing

entity Order id "billing://entity/order"
entity Payment id "billing://entity/payment"
activity AuthorizePayment id "billing://activity/pay-order" (Order) -> Payment
```

The semantic ID and graph node must match Fixture B's activity exactly. Only
the display name and source spans change. A generator may rename a Go symbol or
documentation heading, but it must retain the identity in generated markers,
source maps, and provenance facts.

### Fixture D: recovery must preserve later declarations

```gooo
package billing
version "gooo/v1"
namespace billing

activity PayOrder(Order Payment) -> Payment
entity Payment id "billing://entity/payment"
activity Audit() -> Payment
```

Expected diagnostics include one `parse.expected-comma` at the second activity
parameter boundary. Recovery must make progress, retain the later `entity` and
`Audit` declarations in the partial AST, and never create a semantic fact for
the malformed `PayOrder` signature. This fixture is intentionally about
recovery, not accepted IR.

## Scope for implementation

This note recommends a future syntax change but makes no implementation change
in this branch. When the syntax agent implements it, the focused checks should
cover:

- old files with no version header;
- explicit `version "gooo/v1"` and activity IDs;
- duplicate/unknown versions and invalid/duplicate IDs;
- Unicode names, escaped strings, CRLF, comments, and malformed UTF-8;
- partial AST preservation and deterministic diagnostic order;
- rename round-trip with an unchanged semantic ID.

The semantic layer should then decide whether inferred legacy IDs are acceptable
for a project policy. The generator and bidirectional layers must use the ID,
not the display name, as their join key.

## Experiment matrix

The following experiments turn the proposal into small conformance fixtures.
They are intentionally embedded in this research note so the syntax agent can
lift them into focused tests without waiting for another PR to merge.

Each fixture has three observable results:

1. `Parse`: ordered diagnostics identified by stable code and primary span;
2. `AST`: declaration names and source spans retained after recovery; and
3. `IR`: a canonical set of IDs and facts, compared after sorting by
   `(subject, predicate, object)`.

`Parse` errors prevent `IR` lowering. A partial AST is still useful to an editor,
but it must not be treated as an accepted semantic model. For successful
fixtures, semantic equivalence ignores display names, declaration order where
the IR is set-like, and source spans; it compares stable IDs, declaration kinds,
resolved references, and derived facts.

### R1: recovery at an activity parameter boundary

Input:

```gooo
package billing
namespace billing

activity Broken(Order Payment) -> Payment
entity Payment id "billing://entity/payment"
activity Healthy() -> Payment
```

Expected result:

- `Parse` emits exactly one `parse.expected-comma` diagnostic at the `Payment`
  parameter on line 4, column 23 (span `4:23-4:30`), in the parser phase.
- `AST` retains the later `entity Payment` and `activity Healthy` declarations;
  the malformed `Broken` declaration may remain as a partial node but is
  marked unusable for lowering.
- `IR` is not produced for `Broken`. If the implementation lowers valid
  declarations from a partial file, it may lower `Healthy` only after explicit
  policy approval; the default conformance mode rejects the file as a whole.
- Repeating the parse produces byte-for-byte equal diagnostics and an equivalent
  partial AST. The parser must not loop or report a second comma/arrow error for
  the same missing separator.

This fixture tests recovery without depending on an implementation-specific
error message. A companion test may delete `entity Payment` and assert that an
unresolved-reference diagnostic is semantic and occurs only after syntax is
valid.

### E1: grammar evolution and version skew

Legacy input, produced by the current example shape:

```gooo
package billing
namespace billing
entity Order id "billing://entity/order"
entity Payment id "billing://entity/payment"
activity PayOrder(Order) -> Payment
```

Versioned input, using the proposed additive clauses:

```gooo
package billing
version "gooo/v1"
namespace billing
entity Order id "billing://entity/order"
entity Payment id "billing://entity/payment"
activity PayOrder id "billing://activity/pay-order" (Order) -> Payment
```

Expected compatibility matrix:

| Producer source | Consumer parser | Expected diagnostics | IR result |
| --- | --- | --- | --- |
| Legacy input | version-aware parser | none; missing version is implicit `gooo/v1` | accepted v1 model; activity ID is inferred |
| Versioned input | version-aware parser | none | accepted v1 model; activity ID is explicit |
| Versioned input | baseline parser | rejection at the version/extended clause boundaries | no trusted IR; partial AST only |
| `version "gooo/v9"` | v1 parser | one `parse.unsupported-version` at the version literal; no feature guessing | no IR |

For the unknown-version case, use this minimal input:

```gooo
package billing
version "gooo/v9"
namespace billing
entity Order id "billing://entity/order"
```

The diagnostic must point to the quoted version literal on line 2, and parsing
must not silently fall back to v1. The version-aware parser may continue to
produce a syntax tree for editor services, but `IR` lowering is blocked. This
is the safety boundary for a newer producer and older compiler: rejecting an
unknown grammar is preferable to accepting a different meaning.

### S1: stable-ID rename

Source A:

```gooo
package billing
version "gooo/v1"
namespace billing
entity Order id "billing://entity/order"
entity Payment id "billing://entity/payment"
activity PayOrder id "billing://activity/pay-order" (Order) -> Payment
```

Source B:

```gooo
package billing
version "gooo/v1"
namespace billing
entity Order id "billing://entity/order"
entity Payment id "billing://entity/payment"
activity AuthorizePayment id "billing://activity/pay-order" (Order) -> Payment
```

Expected result:

- Both parses have no errors.
- `IR(A)` and `IR(B)` contain the same declaration IDs and kinds:
  `billing://entity/order` (`Entity`), `billing://entity/payment` (`Entity`),
  and `billing://activity/pay-order` (`Activity`).
- Their canonical fact sets are equal:
  `billing://activity/pay-order uses billing://entity/order` and
  `billing://entity/payment wasGeneratedBy billing://activity/pay-order`.
- `semanticEquivalent(A, B) = true`, while textual equality is false.
- The semantic delta is a display-name rename only. It must contain no
  ID-keyed remove/add pair, no changed provenance edge, and no unrelated
  locality expansion. Go symbol names, documentation labels, and source spans
  may change.

This fixture catches the most dangerous identity regression: treating the
activity's display name as the graph join key. A rename with a changed ID is a
different fixture and must instead be reported as a migration.

### F1: formatter idempotence

Input with intentionally irregular whitespace:

```gooo
package   billing
namespace billing
entity Order id "billing://entity/order"
entity PaymentMethod id "billing://entity/payment-method"
entity Payment id "billing://entity/payment"
activity PayOrder( Order,PaymentMethod )->Payment
```

Proposed canonical output:

```gooo
package billing
namespace billing

entity Order id "billing://entity/order"
entity PaymentMethod id "billing://entity/payment-method"
entity Payment id "billing://entity/payment"
activity PayOrder(Order, PaymentMethod) -> Payment
```

Expected result:

- `format(input)` equals the canonical output byte-for-byte.
- `format(canonical output)` equals the same canonical output, establishing
  `F(F(source)) = F(source)`.
- Parsing the input and canonical output yields no diagnostics and equivalent
  IR, including exact decoded IDs and declaration order.
- Formatting never rewrites ID contents, changes declaration kind, or uses a
  display-name-derived ID as a replacement for an explicit ID.

The current AST does not retain comment trivia, so this fixture deliberately
contains no comments. Comment-preserving formatting needs a separate contract:
either trivia must be retained losslessly, or the formatter must document that
comments are not round-tripped. That choice must not be made implicitly while
claiming formatter idempotence.

### Experiment acceptance table

| Fixture | Primary invariant | Failure meaning |
| --- | --- | --- |
| R1 | Recovery is bounded and preserves later declarations | Parser may cascade, loop, or lower malformed facts |
| E1 | Newer parsers accept legacy input; unknown versions never guess | Version skew can silently change semantics |
| S1 | Stable IDs survive display-name changes | Identity is coupled to presentation |
| F1 | Formatter reaches a canonical fixed point without semantic drift | Repeated formatting creates noisy or meaning-changing diffs |

These fixtures should be run at the syntax boundary first, then again after
lowering and generation. The second pass must compare semantic IDs and facts,
not generated text alone; generated text is allowed to change when a display
name changes, provided locality and stable-ID source maps remain correct.

## Self-hosting evidence boundary

Self-hosting is a roadmap, not a current success claim. The present stage is
Go-hosted: Go owns lexing, parsing, diagnostics, lowering, and any formatter
implementation, while `.gooo` is the authored source view. A future gooo-hosted
stage may describe those compiler components in `.gooo`, but that stage is not
implemented by this repository state.

The stages should share one comparable contract rather than a narrative claim:

| Host stage | Authoritative implementation | Required evidence | Status in this note |
| --- | --- | --- | --- |
| Go-hosted bootstrap | Go lexer/parser/lowering/formatter | R1/E1/S1/F1 diagnostics, canonical IR, and formatter fixed-point results | Current target; syntax implementation exists on a separate lane |
| Transitional dual host | Go implementation plus a `.gooo` declaration of the same contract | Same fixtures produce equivalent diagnostics and IR; differences are explicit deltas | Future experiment; not implemented |
| gooo-hosted compiler | `.gooo` compiler declarations bootstrap the compiler projection | Independent host runs agree on IDs, facts, diagnostics, formatter output, and provenance of the build | Future goal; no success evidence yet |

For each future host, record `host`, source hash, grammar version, fixture ID,
diagnostic list, canonical IR hash, formatter output hash, and provenance
evidence. A missing host executable or missing comparison result is `blocked`,
not `passed`. This keeps self-hosting progress falsifiable and prevents the
future projection from being reported as implemented merely because its
contract is documented.
