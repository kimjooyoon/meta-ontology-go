# Gooo-declared completeness receipt

## Typed-path selection and finite TDD

`body-codegen --path-plan` uses the same generated receipt structure, with the
`gooo/body-codegen-typed-path-v1` profile. Its plan identity binds the original
source, complete typed-path document (including Korean/English intent and test
expectations), search controls, and the final selected source. Two requests that
emit identical Go can therefore retain different plan identities.

Five additional axes record source binding, final finite-suite accuracy,
provider accounting, candidate attempts, and completed candidate scores. The
first three participate in the scoped decision. A selected body can have 100%
lowering coverage and only 2/3 functional matches. An observed 0/3 is `PROGRESS`;
an unobserved final score is `UNKNOWN`, even if search candidates were evaluated.

`scope.typed_path` binds the actual path observation, model metadata/weights,
local prediction count (including feedback), external calls, and model context.
`scope.emission_decision_provider` describes final lowering; the nested path
provider describes selection. Local predictions never become Laya calls or an
implicit network request. A source rejection before the prediction entry point
records known zero calls; a started search with a missing result stays unknown.

The final suite is interpreted by the bounded Go AST evaluator and cross-checked
against the typed interpreter. It is a selection suite, not an independent
holdout or generated-package runtime execution. Execution, reverse observation,
real workflow coverage and full-domain semantics remain unresolved until a
separate producer supplies source-bound observations. No aggregate percentage
replaces these axes, and no human action is needed to continue the bounded search.

## Declared structure

The compiler's existing `gooo/metaprogramming-completeness-receipt/v2` wire
structure is now owned by
[`receipt.gooo`](../internal/completeness/receipt.gooo). That declaration generates
the actual Go types used by body code generation and its JSON Schema. Generated
Go fields retain the corresponding stable Gooo field ID in a `gooo` tag; the
schema retains declaration offsets and the exact source digest.

The `gooo/receipt-structure-projection/v1` profile supports declared text, four
measurement states, nonnegative counters, numbers, opaque scope records, status
counts, required lists and nullable single values. Structural references retain
their declared entity identity. This is an explicit receipt projection profile;
EntityFields V1's general semantic profile still supports only its existing
string/required-one shape. Unknown types, ID/name collisions, unsupported
cardinality, activity/binding execution and oversized declarations are rejected.
This compiler profile has no inference, network, execution or approval step.

## Generate and inspect

```sh
go generate ./internal/completeness
go run ./cmd/gooo receipt-schema --format go internal/completeness/receipt.gooo
go run ./cmd/gooo receipt-schema --format json internal/completeness/receipt.gooo
```

The CLI writes its result to standard output. The build tool writes only the two
explicitly requested generated files. CI tests regenerate both artifacts and
compare every byte. Reflection of the compiled Go fields independently checks
field order, type, JSON key and stable ID against the source declaration.

The complete language corpus registers this declaration as a required projection
unit with the exact profile, root and both artifact paths. Its evidence replays
the AST, canonical formatting and typed structural identity, then regenerates and
compares both checked-in artifacts byte for byte. Missing artifacts remain
UNKNOWN; changed source or stale outputs fail the overall corpus result. The
projection has its own fixed denominator and evidence digest, bound to the source
inventory and revision. The existing 81 syntax cases and their 78 general
Get-Put/Put-Get checks remain separate; this profile does not count as an extra
general semantic or BX proof. Removing or redirecting the projection registration
fails corpus validation.

The `schema` wire value and existing JSON keys remain v2-compatible. All receipt
producers use the generated shared types through aliases, including body-fill,
body-search and source/parse failures. `scope.receipt_declaration` binds the
embedded declaration digest, generated declaration digest and projection profile.
A separate `receipt_schema_binding` core axis fails closed when the generated
structure is stale. Matching this schema proves structural consistency only.

## Independent consumption

`internal/completeness.Decode` rejects unknown fields, duplicate JSON keys,
trailing input, invalid UTF-8, inputs over 1 MiB and JSON deeper than 64 levels.
Scope JSON numbers retain their original precision. `Validate` checks the exact
declaration binding, observed denominators, four-state accounting, unique core
axis references and the ordered unresolved claims. The first unresolved cause is
preserved even when a subsequent axis fails. Aggregated scores are rejected;
UNKNOWN is never converted into a score or a successful observation.

The declaration covers the existing declaration, generation, source binding,
reverse observation, use-case, boundary and provenance measurements without
inventing observations. Schema conformance cannot turn a fixture count into
semantic evidence or convert an external authority failure into product coverage.

## Remaining work for issue #1023

The shared declaration/generated structure requirement now has executable
evidence. Natural-language capability discovery and separate runtime/reverse
observation producers still need to use this common receipt and bind independent
inputs, source revisions, tool identities and observation artifacts. Comparable
before/after domain deltas, regression accounting and real workflow coverage
remain separate obligations. This change does not close #1023 or establish
universal language completeness, production utility or permission authority.
