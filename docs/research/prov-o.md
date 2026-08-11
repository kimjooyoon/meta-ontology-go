# PROV-O mapping for the `.gooo` Semantic IR

Status: research proposal. This document changes no compiler, ontology, verifier, or
storage code. It defines a compatibility profile that can be implemented later without
making PROV-O the authority for business meaning.

## 1. Scope and design position

The `.gooo` DSL is authoritative for business intent. The Semantic IR is authoritative
for the normalized, technology-independent meaning of that intent. PROV-O is the
vocabulary and constraint family for provenance: what was used, generated, derived,
associated, or attributed. It is not a replacement for the business ontology.

The mapping therefore has three layers:

1. **Declared facts** come from `.gooo` declarations and explicit assertions.
2. **Deterministic facts** are compiler deductions whose inputs and rule are fully
   specified, such as one `used` fact per declared input port.
3. **Candidate facts** are observations that need a domain assertion or policy before
   they can affect the authoritative graph, such as an ambiguous Go call or a guessed
   derivation.

Candidate facts MUST NOT be silently exported as authoritative PROV facts. A candidate
may be shown to an agent, stored as review evidence, or promoted by an explicit DSL
assertion and an independent policy check.

This profile uses the following current repository contracts as its boundary:

- `docs/spec.md`: a declared activity has entity inputs and outputs; the compiler
  derives `used` and `wasGeneratedBy` facts.
- `docs/architecture.md`: the semantic package owns identities, PROV vocabulary,
  graph facts, and normalization; provenance and evidence are downstream projections.
- `AGENTS.md`: stable semantic IDs are identity, provenance facts are append-only, and
  ontology/verifier policy is protected.

## 2. Proposed IR shape

The IR need not copy RDF's storage shape, but it must retain enough information to
round-trip to a PROV representation without losing direction, identity, or evidence.
The conceptual shape is:

```text
Node       = { id, kind, types, label, aliases, attributes }
Fact       = { subject, predicate, object, status, source, bundle }
Qualified  = { relationID, predicate, subject, object, attributes, status, source }
Bundle     = { id, facts, qualified, attributes, contentDigest }
```

`Node.kind` is one primary role (`Entity`, `Activity`, or `Agent`) for the core
profile; `types` retains explicit PROV/application types and deliberate multi-typing.
A `prov:Bundle` is an `Entity` with bundle metadata, not a fourth competing node kind.
`Qualified` represents a relation occurrence; the Go/RDF projection materializes it as
an instance of `prov:Usage`, `prov:Generation`, `prov:Derivation`, `prov:Association`,
or `prov:Attribution`.

`status` should preserve at least `deterministic` and `candidate`. It is useful to
retain `declared` as source metadata even if declared facts are normalized into the
deterministic set. `source` must identify the authoritative view and source span or
symbol evidence; a status without evidence is not sufficient for merge or scope
verification.

The canonical IR should use a set interpretation for unqualified binary facts and a
relation identity for qualified occurrences. It should not make RDF blank-node identity
the source of Semantic IR identity.

## 3. Core class and relation mapping

The PROV-O starting point has three classes. Their direction and meaning are not
interchangeable with ordinary application nouns.

| PROV-O term | Semantic IR mapping | Direction / invariant |
| --- | --- | --- |
| `prov:Entity` | Entity node with a stable semantic ID | A physical, digital, conceptual, or other thing with fixed aspects. |
| `prov:Activity` | Activity node with a stable semantic ID | An occurrence that acts on entities over time; do not treat it as an Entity. |
| `prov:Agent` | Agent node with a stable semantic ID | Something bearing responsibility for an Activity, Entity, or Agent. |
| `prov:used` | `Fact(Activity, used, Entity)` | Activity is the subject; Entity is the object. |
| `prov:wasGeneratedBy` | `Fact(Entity, wasGeneratedBy, Activity)` | Generated Entity is the subject; generating Activity is the object. |
| `prov:wasDerivedFrom` | `Fact(Entity, wasDerivedFrom, Entity)` | Derived/output Entity is the subject; source/input Entity is the object. |
| `prov:wasAssociatedWith` | `Fact(Activity, wasAssociatedWith, Agent)` | Activity is the subject; responsible Agent is the object. |
| `prov:wasAttributedTo` | `Fact(Entity, wasAttributedTo, Agent)` | Entity is the subject; responsible Agent is the object. |
| `prov:Bundle` | Bundle record and an `Entity` node | A named set of provenance descriptions that can itself have provenance. |

The canonical relation names should remain the PROV-O local names. Inverse names such
as `usedBy`, `generatedBy`, or `attributedEntity` may be query projections, but MUST NOT
be stored as competing canonical facts. This prevents a graph query view from becoming
a second authority with a different direction.

### 3.1 Lowering the current `.gooo` example

For:

```gooo
entity Order id "billing://entity/order"
entity PaymentMethod id "billing://entity/payment-method"
entity Payment id "billing://entity/payment"

activity PayOrder(Order, PaymentMethod) -> Payment
```

the normalized design-time graph is at least:

```text
Entity   billing://entity/order
Entity   billing://entity/payment-method
Entity   billing://entity/payment
Activity billing://activity/pay-order

used(billing://activity/pay-order, billing://entity/order)
used(billing://activity/pay-order, billing://entity/payment-method)
wasGeneratedBy(billing://entity/payment, billing://activity/pay-order)
```

The activity ID shown above is a proposed stable ID for the example; a compiler MUST
use an explicit ID when available and a documented, deterministic namespace rule when
the surface declaration omits one. A display name such as `PayOrder` is never enough
to identify the Activity.

`wasDerivedFrom(payment, order)` and
`wasDerivedFrom(payment, payment-method)` are not business facts merely because an
Activity has inputs and an output. They may be materialized as deterministic
**application/profile inference** when the `.gooo` profile declares that every output
depends on every input. This is a project rule, not a claim that PROV-O's base
vocabulary alone proves the dependency. Otherwise the facts remain absent until an
explicit dependency assertion is present. This distinction matters because PROV
derivation means that one Entity had an influence on another; it is not just a function
signature edge. PROV constraints also state that a derivation is not transitive by
default.

No `wasAssociatedWith` or `wasAttributedTo` fact is inferred from the example because
it declares no Agent. The compiler must not invent a human, repository, or runtime
process as an Agent.

## 4. Qualified relation mapping

PROV-O's qualification pattern restates an unqualified relation through an intermediate
influence instance. The host resource points to the qualified relation, and the
qualified relation points back to the same influencing resource. Additional attributes
such as role, time, plan, generation, or usage belong to the relation occurrence.

The IR should keep the base edge and its qualifier linked by `relationID`; the exporter
should emit the base edge whenever the qualified relation asserts that edge. An
unqualified edge without details remains valid and must not be padded with a fabricated
qualified node.

| Base edge | Qualified class and host property | Required counterpart | Useful attributes |
| --- | --- | --- | --- |
| `Activity used Entity` | `Activity --prov:qualifiedUsage--> prov:Usage` | `Usage --prov:entity--> Entity` | `prov:hadRole`, `prov:atTime`, source span |
| `Entity wasGeneratedBy Activity` | `Entity --prov:qualifiedGeneration--> prov:Generation` | `Generation --prov:activity--> Activity` | `prov:hadRole`, `prov:atTime` |
| `Entity wasDerivedFrom Entity` | `Entity --prov:qualifiedDerivation--> prov:Derivation` | `Derivation --prov:entity--> source Entity` | `prov:activity`, `prov:hadGeneration`, `prov:hadUsage` |
| `Activity wasAssociatedWith Agent` | `Activity --prov:qualifiedAssociation--> prov:Association` | `Association --prov:agent--> Agent` | `prov:hadRole`, `prov:hadPlan` |
| `Entity wasAttributedTo Agent` | `Entity --prov:qualifiedAttribution--> prov:Attribution` | `Attribution --prov:agent--> Agent` | `prov:hadRole` |

Qualified relations are not merely annotations on the subject. For example, the
`prov:agent` field of an Association points to the influencing Agent; it does not point
back to the Activity. Likewise, the `prov:activity` field of a Generation points to the
generating Activity; it does not reverse the `wasGeneratedBy` edge.

### 4.1 Relation-specific rules

#### Usage

`used(a, e)` means that Activity `a` began using Entity `e`. A qualified Usage can say
that `e` was an `input-order`, `payment-method`, `configuration`, or other role. This
is the recommended way to distinguish ports in a `.gooo` Activity without creating
application-specific predicates for every port shape.

#### Generation

`wasGeneratedBy(e, a)` means that Entity `e` was generated by Activity `a`. A qualified
Generation can retain output role and generation time. One logical Entity ID MUST NOT
be reused for unrelated runtime versions merely because their display names match;
versioned or specialized Entities need distinct IDs.

#### Derivation

`wasDerivedFrom(e2, e1)` is Entity-to-Entity and points from newer/derived to older/source.
When the activity, generation, and usage occurrences are known, a qualified Derivation
may link them with `prov:activity`, `prov:hadGeneration`, and `prov:hadUsage`. If any
part is unknown, the IR must represent that absence explicitly rather than guessing.
In particular, an explicit qualified derivation with generation and usage references
must satisfy the PROV ordering constraints available to the verifier.

#### Association

`wasAssociatedWith(a, ag)` assigns responsibility for an Activity to an Agent. A
qualified Association adds role and optional Plan. It does not prove that the Agent
executed every implementation detail, and it does not imply attribution of every
Entity generated by the Activity unless the relevant generation/attribution policy is
also satisfied.

#### Attribution

`wasAttributedTo(e, ag)` ascribes an Entity to an Agent. A qualified Attribution adds
the Agent's role. In PROV, attribution is compatible with an unspecified generating
Activity that was associated with the Agent; it is therefore useful when the exact
Activity is unknown. In `.gooo`, attribution should be emitted only from an explicit
assertion or from a documented, evidence-backed inference. Association alone is not a
license to attribute every output.

## 5. Design-time activities versus runtime provenance

PROV-DM defines an Activity as something that occurs, happens, or unfolds over time.
The current `.gooo` `activity` declaration is normally a design-time semantic
definition: it describes an allowed business operation and its ports. Treating that
declaration as a completed runtime occurrence would make a generated graph claim more
than the source proves.

The implementation should choose and label one of these profiles:

- **Design profile:** use the PROV vocabulary for topology, mark the graph with a
  project-specific type such as `gooo:ActivityDefinition`, and do not attach runtime
  timestamps or execution claims.
- **Runtime profile:** create a fresh Activity occurrence ID per execution, version
  concrete Entity instances, and link the occurrence to the declared plan or activity
  definition through an application-specific relation and/or `prov:hadPlan` in a
  qualified Association.

The two profiles may share stable definitions but MUST NOT share an occurrence ID. A
Go symbol reference proves a registered semantic dependency, not that a runtime call
actually happened. Runtime evidence is required before promoting an execution claim.

## 6. Stable identifiers and normalization

### 6.1 Node identity

1. Every Entity, Activity, Agent, and Bundle crossing the IR boundary MUST have a
   non-empty absolute URI/IRI-like stable identifier. The current examples use
   `billing://...`; the profile may accept that scheme while it remains the documented
   project convention.
2. The identifier is opaque identity. Normalize only the representation required by
   the chosen URI/IRI parser; do not case-fold path segments, decode arbitrary percent
   escapes, or rewrite a scheme-specific ID unless that scheme's rules are part of the
   identity contract.
3. Names, aliases, labels, Go identifiers, and source spans are indexes or evidence;
   none can replace the stable ID. Renaming `PayOrder` while retaining
   `billing://activity/pay-order` is an identity-preserving change.
4. Namespace expansion happens once at the parser/lowering boundary. A bare reference
   is invalid outside a declared namespace, and two namespaces with the same local name
   MUST remain distinct unless an explicit relation connects them.
5. A single ID with incompatible primary roles is a validation error by default.
   Explicit multi-typing may be supported only as a deliberate profile feature and must
   be recorded in `types`; it must not arise from name matching. PROV-O permits some
   resources to be both an Agent and an Entity, while `prov:Bundle` is always modeled as
   an Entity subtype in this profile.

### 6.2 Qualified relation identity

The base triple is not enough to identify two Usage or Association occurrences with
different roles or times. Each qualified occurrence needs a stable `relationID`.

- Prefer an explicit relation ID when a DSL author needs to refer to the occurrence.
- Otherwise derive an ID from a versioned, tagged canonical tuple containing the host
  ID, PROV predicate, counterpart ID, and all identity-bearing qualifiers. Use length-
  delimited or otherwise unambiguous encoding before hashing.
- Include an explicit occurrence key when the same tuple can happen more than once.
- Do not derive identity from source order, line number, display name, or a transient
  Go symbol. Source spans belong in evidence and may change during formatting.
- Decide whether bundle scope is part of relation identity. A relation that represents
  the same global occurrence should remain equal across bundles; a claim-local relation
  should include its claim scope. The choice must be versioned and tested.

Qualified relation blank nodes imported from RDF MUST be assigned an import-scoped
identity and marked non-portable or candidate until an explicit stable ID is available.
Otherwise a round trip can silently attach evidence to a different relation occurrence.

### 6.3 Canonical ordering and equivalence

Normalization should be deterministic and semantic rather than textual:

1. Expand IDs and namespaces to their canonical IR strings.
2. Validate node kinds and relation directions before sorting.
3. Sort nodes by `(kind, id)` and unqualified facts by
   `(subject, predicate, object)`.
4. Sort qualified relations by `(host, predicate, relationID)` and attributes by
   `(attribute name, typed value)`; repeated `prov:role` values are a set unless the
   application explicitly gives them sequence semantics.
5. Preserve typed time/value information and canonicalize only its serialization.
6. Deduplicate identical unqualified triples. Do not collapse distinct qualified
   relation IDs merely because their base triples are equal.
7. Compare DSL/IR/Go round trips by the normalized semantic graph plus status/evidence
   policy, not by labels, source spans, or generated text.

## 7. Bundle and attribution policy

A Bundle is a named set of provenance descriptions and is itself an Entity. It is not
the same thing as a `prov:Collection`; `prov:hadMember` should not be used as a shortcut
for “this fact belongs to this bundle.” The IR should represent bundle membership as
container topology and export it as a named graph, PROV-N bundle, or another format
that preserves the set boundary.

Recommended Bundle fields:

- stable Bundle ID;
- canonical set of included facts and qualified relation records;
- content digest over the normalized set;
- creation/build metadata and source authority;
- optional provenance of the Bundle itself (`wasGeneratedBy`, `wasDerivedFrom`,
  `wasAttributedTo`, and `generatedAtTime`).

Changing a description changes the named set and therefore changes the content digest.
The same fact may be present in several bundles without becoming several different
semantic facts; the bundle is provenance context, not a replacement for fact identity.

Attribution needs a stricter policy than association:

- explicit `.gooo` attribution is deterministic and source-backed;
- attribution inferred from a known generation plus known association can be emitted only
  under an explicit application rule. It is not a reverse PROV inference, so the IR must
  retain the application rule and evidence;
- attribution inferred from a display name, commit author, or nearby Go code is a
  candidate until an authority or policy binds that Agent ID;
- a missing attribution is not evidence that no Agent was responsible (PROV is not a
  closed-world business ontology).

## 8. Validation rules for the IR and exports

The following rules are proposed for a deterministic verifier. They are validation
rules, not new business semantics.

### Identity and typing

- Every referenced ID resolves to one canonical identity.
- Every declared Entity, Activity, Agent, and Bundle has the appropriate node kind.
- `used` has Activity subject and Entity object.
- `wasGeneratedBy` has Entity subject and Activity object.
- `wasDerivedFrom` has Entity subject and Entity object.
- `wasAssociatedWith` has Activity subject and Agent object.
- `wasAttributedTo` has Entity subject and Agent object.
- A qualified host and its counterpart agree with the corresponding base edge.
- A relation ID is unique within its declared identity scope and has one qualified type.

### Qualification consistency

- A `qualifiedUsage` points to a `prov:Usage` whose `prov:entity` is the `used` object.
- A `qualifiedGeneration` points to a `prov:Generation` whose `prov:activity` is the
  `wasGeneratedBy` object.
- A `qualifiedDerivation` points to a `prov:Derivation` whose `prov:entity` is the
  `wasDerivedFrom` object; optional Activity/Generation/Usage links must be mutually
  consistent.
- A `qualifiedAssociation` points to an `prov:Association` whose `prov:agent` is the
  `wasAssociatedWith` object.
- A `qualifiedAttribution` points to an `prov:Attribution` whose `prov:agent` is the
  `wasAttributedTo` object.
- If a qualified relation is exported, its base edge is also exported or the format's
  documented projection rule makes the equivalence explicit.

### Temporal and provenance constraints

- When event times are present, generation precedes usage and invalidation, and a
  known derivation's usage precedes the derived generation.
- A declared design Activity with no runtime times must not be checked as if it were a
  runtime execution.
- A known qualified derivation cannot claim a generation or usage that is missing or
  points to another Activity/Entity.
- `wasDerivedFrom` is not treated as transitive unless a separate application rule
  explicitly requests a bounded derived view.
- Bundle digests are calculated after normalization; stale or mismatched digests are
  verification failures, not opportunities to rewrite history.

### Status and evidence

- A deterministic fact records the rule and source inputs that make it reproducible.
- A candidate fact records why it is ambiguous and the source span/symbol that produced
  it; candidates cannot widen semantic scope or authorize a merge.
- Promotion from candidate to deterministic requires an explicit DSL assertion,
  registered semantic identity, or trusted policy, plus fresh evidence.
- An inverse query result is derived output and must not be mistaken for a new
  authoritative fact.
- A verifier reports direction/type/status failures in stable order so evidence can be
  compared across runs.

## 9. Semantic PR/IR review checklist

Use this checklist for any future semantic PR that adds or changes a PROV-shaped fact.
For this research PR, all core code and ontology files are intentionally out of scope.

### Scope and authority

- [ ] The PR identifies the authoritative view changed (`.gooo`, IR, Go analysis,
      evidence, or projection).
- [ ] The change is limited to the approved semantic scope; no verifier, ontology, or
      CI policy is weakened to accept it.
- [ ] The change does not treat a derived query/index/cache as a new authority.
- [ ] Any candidate fact is visibly candidate and is not silently promoted.

### Relation direction and qualification

- [ ] `used` is `(Activity, Entity)`, not `(Entity, Activity)`.
- [ ] `wasGeneratedBy` is `(Entity, Activity)`, not `(Activity, Entity)`.
- [ ] `wasDerivedFrom` is `(derived Entity, source Entity)`.
- [ ] `wasAssociatedWith` is `(Activity, Agent)` and `wasAttributedTo` is
      `(Entity, Agent)`.
- [ ] A qualified host points to the qualified relation node, and the node's
      `entity`, `activity`, or `agent` field points to the same counterpart as the base
      relation.
- [ ] Any inverse (`usedBy`, `generatedBy`, or similar) is clearly a derived query view.
- [ ] The PR includes a normalized round-trip fixture for every changed relation.

### Identity and locality

- [ ] Stable IDs, not display names, are used for all node and relation identity.
- [ ] Namespace expansion and ID normalization are deterministic and collision-checked.
- [ ] Relation IDs remain stable across formatting, declaration reordering, and source
      span movement.
- [ ] The semantic delta is local to the changed IDs and their documented relation
      closure.
- [ ] Renames preserve IDs unless an intentional identity migration is documented.

### Status, provenance, and evidence

- [ ] The fact is explicitly classified as declared, deterministic, or candidate.
- [ ] Deterministic facts list the rule and source inputs; candidates list ambiguity and
      source evidence.
- [ ] Attribution is not inferred from association alone without the applicable
      generation and policy evidence.
- [ ] Bundle membership and content digest are recomputed from normalized facts.
- [ ] Freshness, source spans, generated regions, and round-trip evidence are present
      where the repository gate requires them.

## 10. Risks and open questions

1. **Definition versus occurrence:** the largest semantic risk is mapping a static
   Activity declaration to a runtime occurrence. The design/runtime profile split above
   should be settled before timestamps, attribution, or execution evidence are added.
2. **Input does not imply influence:** a function signature establishes a port relation,
   not necessarily that every input affects every output. An explicit dependency form
   or application inference policy is needed for `wasDerivedFrom`.
3. **Entity versioning:** a stable domain name such as `Payment` may describe a type,
   logical resource, or concrete version. The IR needs a policy for specialization,
   alternate entities, and runtime snapshots.
4. **Temporal incompleteness:** the current DSL has no start/end or event times, so
   PROV ordering can only be checked when runtime evidence supplies them.
5. **Agent trust:** commit authors, Go package owners, CI workers, and human reviewers
   are not interchangeable Agents. Stable Agent identities and a trust boundary are
   required for attribution and attestation.
6. **Qualified relation multiplicity:** two identical base triples can represent
   different ports, executions, or roles. Unstable relation IDs will corrupt evidence
   locality even if the base graph looks correct.
7. **Bundle semantics:** JSON-lines, RDF named graphs, and PROV-N bundles have different
   container semantics. The export format must preserve named-set identity and digest
   rules rather than flattening everything into one fact stream.
8. **Open-world behavior:** absence of `wasAttributedTo` or `wasDerivedFrom` is unknown,
   not false. CI policies must state when they require completeness.
9. **Candidate leakage:** search, Agent context, and diagnostics benefit from candidate
   facts, but candidate leakage into scope calculation or merge evidence would defeat
   the builder/guardian/gate separation.
10. **Privacy and integrity:** provenance can expose source paths, identities, inputs,
    and security-sensitive workflow details. Evidence retention, redaction, signing,
    and access control are separate design work.

## 11. Recommended conformance fixtures

Before implementation, add fixtures (in a future semantic/conformance scope) for:

1. the `PayOrder` lowering above, including exact edge directions;
2. one qualified Usage and Generation with roles, preserving the base edges;
3. a qualified Derivation with explicit Activity, Usage, and Generation links;
4. qualified Association and Attribution with distinct Agents and roles;
5. a Bundle whose content digest changes when one assertion changes;
6. a candidate Go call that remains candidate until a registered semantic ID is added;
7. a label rename that leaves normalized node and relation IDs unchanged;
8. a relation-direction mutation that fails validation deterministically;
9. a design Activity that cannot be mistaken for a runtime occurrence;
10. a blank-node import that is rejected or marked non-portable until anchored.

## References

- [W3C PROV-O: The PROV Ontology](https://www.w3.org/TR/prov-o/)
- [W3C PROV-DM: The PROV Data Model](https://www.w3.org/TR/prov-dm)
- [W3C Constraints of the PROV Data Model](https://www.w3.org/TR/prov-constraints/)
- [W3C PROV Model Primer](https://www.w3.org/TR/prov-primer/)
