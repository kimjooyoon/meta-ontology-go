# `.gooo` language sketch

The surface language is intentionally narrow. A file currently contains a
package, a namespace, entity declarations, and activity declarations:

```gooo
package billing
namespace billing

entity Order id "billing://entity/order"
entity PaymentMethod id "billing://entity/payment-method"
entity Payment id "billing://entity/payment"

activity PayOrder(Order, PaymentMethod) -> Payment
```

The checked-in [billing example](../examples/billing/main.gooo) is the canonical
small input. It is a conformance fixture, not a promise that all future syntax
will remain source-compatible.

## Public EntityFields V1 contract

This section is normative for the profile-bound EntityFields V1 contract. The
ordinary/default parser remains `DEFERRED`; explicit V1 routes used by `gooo
check`, `gooo generate`, and the atomic witness are profile-bound and observed
by the generated [support view](entity-fields-support.md). The view is a
bounded fixture-and-evidence claim, not a claim that every parser or LSP server
supports EntityFields.

### Surface grammar

The V1 surface adds an optional block to an `entity` declaration. Whitespace
and comments are insignificant between terminals. A block has no implicit
separator: each `field` keyword starts one complete field declaration.

```ebnf
entity-declaration = "entity", identifier, "id", string-literal,
                     [ fields-block ] ;
fields-block      = "fields", "{", { field-declaration }, "}" ;
field-declaration = "field", identifier, "id", string-literal,
                    "type", type-reference, presence, cardinality ;
presence          = "required" | "optional" ;
cardinality       = "one" | "many" ;
type-reference    = unqualified-type | qualified-type | stable-type-id ;
unqualified-type  = identifier ;
qualified-type    = string-literal ;
stable-type-id    = string-literal ;
```

The decoded value of a `qualified-type` string is exactly
`namespace:name`: one valid namespace, one colon, and one valid type name.
The decoded value of a `stable-type-id` string is an absolute identity accepted
by the semantic identity parser, including the live `urn:` and `scheme://`
forms. The lexical categories overlap intentionally; the value is classified
as a stable ID first, then as a qualified lookup, and otherwise is invalid for
V1. A quoted unqualified type is not a second spelling; use the identifier form.

Minimal contract example (illustrative of the explicit profile-bound routes;
the ordinary/default parser remains deferred):

```gooo
package billing
namespace billing

entity Order id "billing://entity/order" fields {
    field OrderNumber id "billing://field/order-number"
        type string required one
}

activity PayOrder(Order) -> Order
```

### Identity, ownership, and ordering

- Field identity is exactly `Field.ID`. It is mandatory in V1, must be a valid
  absolute identity, and must never be derived from the display name, source
  spelling, resolved type, presence, cardinality, declaration order, or path.
  Name, source spelling, resolved type, presence, and cardinality are state or
  presentation, never identity.
- `Parent` records the exact stable ID of the enclosing entity. Parent
  ownership is immutable in V1: a field ID cannot move to another entity or to
  a non-Entity declaration. Moving a field is an explicit source delete plus
  source add, and therefore creates a new field identity.
- The stable-ID collision domain is the complete normalized model: Entity,
  Activity, Agent, and Field IDs share one globally unique keyspace. A
  duplicate ID of any kind, including a Field ID equal to a declaration ID,
  is rejected before publication with `GOOO-EF-V1-ID-COLLISION`. A field ID
  reused under another parent is the same collision. Different field IDs may
  have the same name under different parents; names do not merge fields.
- A name or type spelling change preserves `Field.ID`. A resolved type,
  presence, or cardinality change updates semantic state while preserving
  `Field.ID`. Only an explicit source delete/add changes identity.
- A field name is presentation metadata. V1 accepts one identifier and does
  not define aliases.
- Field declarations retain source order in the syntax carrier, BX model,
  semantic IR, and every later projection. Normalization may normalize
  presentation text, but it must not sort or drop fields.

### Type references and registry resolution

V1 has exactly three source spellings:

| Source spelling | Resolution | Failure rule |
| --- | --- | --- |
| `string` | Exact unqualified registry-name lookup | Zero matches is unknown; more than one is ambiguous |
| `"gooo:string"` | Exact `namespace:name` registry lookup | Namespace and name must each match exactly |
| `"urn:gooo:type:string"` | Exact stable-ID registry lookup | The ID must be registered |

The same rules apply to other valid names, namespaces, and stable identities.
The live default registry maps `gooo:string` to the stable ID
`urn:gooo:type:string`. An unqualified lookup is a global exact-name lookup;
it does not search the enclosing namespace, nearest namespace, aliases, or a
display-name fallback. An explicit stable ID is authoritative. The original
lookup or stable-ID spelling and its exact source span remain provenance and
must not be reconstructed from a current registry display name.

### Presence and cardinality

`required` lowers to semantic `Required`; `optional` lowers to semantic
`Optional`. `one` lowers to semantic `One`; `many` lowers to semantic `Many`.
These values are semantic state, not part of field identity. All four
combinations are valid semantic input, but the immutable V1 Go projection
profile intentionally supports only one combination.

#### Immutable Go projection profile

The named authority is profile ID `gooo.entityfields.go-projection.v1`,
profile version `1`. Its canonical profile bytes are UTF-8 with LF line
endings and no trailing newline; the profile digest is SHA-256 of those exact
bytes: `7e93032618d1250cd4ff480eb7b5d6832f79bfc6921e6b9eea104151db965ec0`.
Downstream code must bind the profile ID, version, and digest together. An
unbound profile or digest mismatch fails closed with
`GOOO-EF-V1-UNBOUND-PROFILE` or `GOOO-EF-V1-PROFILE-DIGEST-MISMATCH`.

```text
profile_id=gooo.entityfields.go-projection.v1
profile_version=1
type.urn:gooo:type:string=string
shape.required.one=string
shape.required.many=UNSUPPORTED:GOOO-EF-V1-UNSUPPORTED-SHAPE
shape.optional.one=UNSUPPORTED:GOOO-EF-V1-UNSUPPORTED-SHAPE
shape.optional.many=UNSUPPORTED:GOOO-EF-V1-UNSUPPORTED-SHAPE
```

The closed mapping is:

| Resolved `TypeRef.ID` | Presence | Cardinality | Go field type |
| --- | --- | --- | --- |
| `urn:gooo:type:string` | `required` | `one` | `string` |
| `urn:gooo:type:string` | `required` | `many` | unsupported |
| `urn:gooo:type:string` | `optional` | `one` | unsupported |
| `urn:gooo:type:string` | `optional` | `many` | unsupported |

The profile binds the stable semantic type ID, never a display name. An
unknown, ambiguous, or unprofiled type fails with
`GOOO-EF-V1-UNKNOWN-TYPE`, `GOOO-EF-V1-AMBIGUOUS-TYPE`, or
`GOOO-EF-V1-UNSUPPORTED-TYPE`, respectively. The supported `string` field has
the ordinary Go zero value `""`; `required` is semantic schema state and does
not promise a non-empty runtime value or add runtime validation. Unsupported
shapes produce no generated field, generated file, source-map entry, or
evidence record. The profile defines no tags, aliases, custom codecs,
nullability convention, or slice/pointer convention.

For this profile, the generated Go identifier is exactly `Field.Name`; no
case conversion, sanitization, or name inference is permitted. The name must
be a valid Go identifier and not a Go keyword. Duplicate resulting Go names
within one entity fail with `GOOO-EF-V1-GO-NAME-COLLISION`; names in different
entities are independent. Field declaration order is the generated struct
field order. The generated structural DTO remains derived and
non-authoritative; the profile-scoped support view does not claim that it
carries presence, cardinality, resolved type IDs, or field-level source-map
identity until those bindings are observed explicitly.

### Deterministic validation and malformed input

Validation is transactional and returns no authoritative partial model on any
failure:

| Condition | Required result |
| --- | --- |
| Duplicate or cross-kind stable ID | Reject with `GOOO-EF-V1-ID-COLLISION` |
| Duplicate name or alias collision within entity | Reject; aliases are not a V1 surface feature |
| Same field name under different entities | Allowed; parent identity separates the fields |
| Unknown type reference | Reject with `GOOO-EF-V1-UNKNOWN-TYPE` |
| Ambiguous type reference | Reject with `GOOO-EF-V1-AMBIGUOUS-TYPE` |
| Wrong parent or field on a non-Entity | Reject with `GOOO-EF-V1-WRONG-PARENT` |
| Missing field ID, name, type, presence, cardinality, or required span | Reject with `GOOO-EF-V1-INCOMPLETE-FIELD` |
| Unsupported profile type or shape | Reject with the profile's stable fail-closed diagnostic |
| Generated identifier collision | Reject with `GOOO-EF-V1-GO-NAME-COLLISION` |
| Malformed or unterminated `fields` block | Emit ordered source diagnostics and publish no fields from that entity |

The field, ID, name, type, presence, and cardinality spans are half-open,
source-backed spans. A malformed block must not produce a field that later
appears only in CLI output, generated Go, or LSP. Repeated parsing of the same
source yields the same diagnostics and no partially populated field AST.

### Projection and bidirectional obligations

The implementation Gate must preserve the existing authority boundaries:

- The parser owns the source grammar and exact spans. It must preserve every
  declared field or fail closed; it may not infer IDs or silently recover by
  dropping a field.
- Lowering maps each field to `semantic.Field` in declaration order, assigns
  the enclosing entity ID as `Parent`, resolves the type through the explicit
  registry, enforces the model-wide ID collision domain, and returns no
  partial IR on error. A type lookup is resolved once to `TypeRef.ID`; no
  nearest, spelling, or display-name fallback is allowed.
- BX `Get`/`Put` must preserve field order, IDs, parent immutability, type
  presentation, and all field subspans. `Put` must validate before writing;
  missing provenance, semantic-only additions, non-source origins, conflicts,
  or invalid fields return the original document unchanged and must not promote
  candidate or derived observations.
- The generator must bind the exact profile ID, version, and digest before
  projection. It must preserve declaration order, use stable generated-region
  markers, preserve handwritten slots, and reject stale or incomplete field
  input rather than omit a field. A failed projection returns no generated
  artifacts.
- Each generated field source-map/evidence record must bind `Field.ID`,
  `Parent`, resolved `TypeRef.ID`, presence, cardinality, declaration ordinal,
  source span, name span, generated byte region, and the profile ID, version,
  and digest. A full field-level source-map claim remains an implementation
  Gate boundary; the current support view claims only the observed structural
  projection. Generated Go remains derived and non-authoritative; handwritten
  slots cannot override structural fields.
- A field document symbol is standard `SymbolKind.Field = 8`, with
  `Name = Field.Name`, `Range = Field.Span`, and
  `SelectionRange = Field.NameSpan`, nested beneath its enclosing entity.
  Standard `Detail` may carry the stable field ID; no custom JSON-RPC wire
  member may be added. Definition and reference identity is
  `TypeRefUse.ResolvedID`. Emit a `Location` only when a source-backed target
  location for that ID exists in the same snapshot; otherwise emit no link.
  Spelling, display-name, and nearest-name fallback are forbidden. Rename and
  name edits preserve `Field.ID`; type, presence, and cardinality edits change
  state only. Invalid or partial fields produce neither trusted symbols nor
  links; diagnostics remain source-backed.

`.gooo` declarations remain authoritative for business intent and explicit
contracts; stable IDs remain authoritative identity; generated Go remains a
derived structural projection; handwritten slots remain irreducible logic; and
candidate or derived views remain non-authoritative. EntityFields do not add
PROV relations, query nodes, aliases, or capabilities by implication.

### Conformance partitions for the implementation Gate

The support boundary is atomic. The parser, formatter, lowerer, BX, CLI
generator, source map, and LSP must be proven together; parser-only or
synthetic-only LSP support must not be advertised. Any layer unable to preserve
the field contract must fail closed before write or publication. The minimum
independent Gate partitions are:

- Positive: one minimal `required one` string field; a name rename preserving
  its ID; a type-state update preserving its ID; and stable declaration-order
  projection with the bound profile digest.
- Negative semantic: duplicate field ID; Field ID colliding with Entity,
  Activity, or Agent ID; wrong parent; unknown, ambiguous, or unprofiled type;
  and malformed or unterminated field block.
- Negative projection: `optional one`, `required many`, and `optional many`;
  generated Go identifier collision; unbound profile or digest mismatch; and
  any unsupported input producing an artifact or partial source map.
- Negative LSP: a missing source-backed target for `TypeRefUse.ResolvedID`, an
  invalid or partial field, and a cross-snapshot target. Each produces no link
  or trusted symbol and retains source-backed diagnostics.
- Negative CLI/BX: a failed generation or `Put` leaves the original source and
  generated outputs unchanged, with no partial write, evidence, or authority.

### Explicitly deferred or unsupported

V1 does not support field aliases, alias-based identity or lookup, query-node
semantics, field-specific capability declarations, implicit field IDs, implicit
parent inference, pressure selectors, tags, custom codecs, or any Go shape
outside the immutable profile above. Unsupported shapes are fail-closed, not
deferred inference. Other features are deferred until a separately implemented
and verified contract exists. Parser recognition, a latent AST carrier, or
semantic constructors alone do not change the current support status.

## Current implementation boundary

This language sketch defines the `.gooo` source and semantic contracts only. It
does not define self-hosting, a self-hosted verifier, or a promotion authority.
The checked-in files under `examples/bootstrap/` and
[bootstrap-evidence.md](bootstrap-evidence.md) are non-promoting evidence-shape
fixtures; `deferred`, `not-run`, and candidate results are not success or policy.

This contract has no separate research dependency. The current branch and
promotion rules are documented separately in [governance.md](governance.md) and
do not change the meaning of `.gooo` declarations.

## Identity and namespaces

An identity is an absolute, stable, URI-like string. It is the semantic identity
of a node. A namespace is a separate scope used for lookup and validation; it is
not inferred from the URI. Names and aliases are presentation or lookup metadata,
not identity. Renaming a declaration is therefore safe only when its semantic ID
is preserved and all references remain valid.

The lowerer derives an activity ID from the namespace and activity name when the
surface form has no explicit activity ID. Entity IDs may be written explicitly;
the lowerer has a deterministic namespace/kind/name fallback for omitted entity
IDs. These defaults are convenience behavior, not a substitute for stable IDs in
long-lived business models.

## PROV-inspired core

The compact vocabulary is intentionally limited:

- nodes: `prov:Entity`, `prov:Activity`, and `prov:Agent`;
- relations: `prov:used`, `prov:wasGeneratedBy`, `prov:wasDerivedFrom`, and
  `prov:wasAssociatedWith`;
- an activity input derives `Activity used Entity`;
- an activity result derives `Entity wasGeneratedBy Activity`.

The activity signature is the only implicit relation source in the initial DSL.
The compiler does not infer domain facts such as delegation, authorization, or
validation from ordinary names or helper calls. Such facts require an explicit
assertion or a separately governed adapter.

Facts have separate statuses:

- deterministic facts are unambiguous, source-backed semantic facts;
- candidate facts are plausible observations retained for review but never
  promoted automatically;
- syntactic observations describe what a parser or analyzer saw and do not change
  semantic state.

## Bidirectional contract

The parser-neutral BX boundary exposes two directions:

1. `Get` lowers a DSL document to canonical semantic IR.
2. `Put` writes an accepted semantic model back to a representable DSL document.

Go adapters use the same boundary by emitting an explicit fact delta. Reconcile
requires provenance for strict semantic updates, is transactional on conflict,
and requires explicit removals rather than interpreting an incomplete analysis as
deletion.

The laws are specified in [docs/governance.md](governance.md): Get-Put, Put-Get,
semantic round-trip, locality, and provenance. Textual equality is not required;
semantic equivalence, stable IDs, and generated-region integrity are required.

## Evolution boundary

Future syntax may add explicit relations, constraints, capabilities, assertions,
handwritten slots, or projection options. Each addition must state whether it is
authoritative, derived, candidate evidence, or protected policy. A new syntax
feature is not supported merely because a parser can recognize it; it needs a
semantic contract, runnable conformance evidence, and—if it participates in
self-hosting—the seed/candidate comparison required by the bootstrap plan.
