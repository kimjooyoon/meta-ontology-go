# Gooo-declared completeness receipt

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
