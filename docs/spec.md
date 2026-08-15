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

This section is normative for the upcoming EntityFields V1 implementation
closure. It is not a current support claim: the live parser still rejects
`EntityFields` source and the implementation Gate must land before this syntax
can be parsed, lowered, formatted, generated, or exposed by LSP. Until then,
the public parser must continue to reject it without constructing a partial
field list.

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

Minimal contract example (illustrative only; it is not runnable on the live
parser until the implementation Gate lands):

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

- A field ID is mandatory in V1. It is the field's stable semantic identity;
  it must be a valid absolute identity and must never be derived from the
  display name, entity name, declaration order, or source path.
- The parent is implicit only from the enclosing entity block and is recorded
  as that entity's exact stable ID. Fields are valid only on `Entity` nodes;
  an activity, agent, top-level field, or parent move is invalid.
- A field name is presentation metadata. V1 accepts one identifier and does
  not define aliases. Renaming the name is not an identity change when the
  field ID is preserved.
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
All four combinations are valid V1 semantic values, and all four are part of
the field's semantic identity. An unknown value is a deterministic error.

The live generator model currently carries only `Field.GoType` and renders it
verbatim; it does not carry field presence or cardinality. Therefore V1 makes
no pointer, slice, nullability, or other generated-Go shape promise. Until a
separate generator contract defines and tests that mapping, generated Go for
EntityFields is deferred and a generator must fail closed rather than choose
or silently omit a shape.

### Deterministic validation and malformed input

Validation is transactional and returns no authoritative partial model on any
failure:

| Condition | Required result |
| --- | --- |
| Duplicate field ID in one entity or across entities | Reject deterministically |
| Duplicate name or alias collision within entity | Reject; aliases are not a V1 surface feature |
| Same field name under different entities | Allowed; parent identity separates the fields |
| Unknown or ambiguous type reference | Reject before semantic publication |
| Wrong parent or field on a non-Entity | Reject before semantic publication |
| Missing field ID, name, type, presence, cardinality, or required span | Reject before semantic publication |
| Malformed or unterminated `fields` block | Emit ordered diagnostics and publish no fields from that entity |

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
  registry, and returns no partial IR on error.
- BX `Get`/`Put` must preserve field order, IDs, parent immutability, type
  presentation, and all field subspans. `Put` must validate before writing;
  missing provenance, semantic-only additions, non-source origins, conflicts,
  or invalid fields return the original document unchanged and must not promote
  candidate or derived observations.
- The generator may project a field only after it has a complete deterministic
  presence/cardinality-to-Go contract. Its generated region must use stable
  markers, preserve handwritten slots, and reject stale or incomplete field
  input rather than omit a field.
- A field-aware source map must bind each field ID to its source span and
  generated range in declaration order. The current generator maps only entity,
  activity, and slot regions, so this obligation is deferred with generation.
- LSP must carry field diagnostics and exact ranges without inventing semantic
  identities. Field symbols, references, and source-map capabilities may be
  advertised only after their implementation and conformance evidence land;
  the current adapter exposes entity/activity symbols only.

`.gooo` declarations remain authoritative for business intent and explicit
contracts; stable IDs remain authoritative identity; generated Go remains a
derived structural projection; handwritten slots remain irreducible logic; and
candidate or derived views remain non-authoritative. EntityFields do not add
PROV relations, query nodes, aliases, or capabilities by implication.

### Explicitly deferred or unsupported

V1 does not support field aliases, alias-based identity or lookup, query-node
semantics, field-specific capability declarations, implicit field IDs, implicit
parent inference, pressure selectors, or generated-Go presence/cardinality
mapping. These are deferred until a separately implemented and verified
contract exists. Parser recognition, a latent AST carrier, or semantic
constructors alone do not change the current support status.

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
