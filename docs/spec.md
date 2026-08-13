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
