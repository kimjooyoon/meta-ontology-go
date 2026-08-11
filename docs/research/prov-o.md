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

## 12. Comparative evidence experiment: Go-hosted versus `.gooo`-hosted

This experiment compares two ways of hosting the same compiler/CI provenance contract.
It is a fixture and measurement design, not an implementation claim. The comparison
must keep the runner, Go toolchain, source payload, policy revision, and evidence
capture budget equal; otherwise a richer capture path could be mistaken for a better
semantic model.

### 12.1 Define the two lanes

“Hosted” names the authoritative evidence boundary, not the programming language used
to implement the compiler.

| Lane | Authoritative source view | Compiler/CI Activities | Typical Agents |
| --- | --- | --- | --- |
| **Go-hosted** | Go source, `go.mod`, workflow configuration, and command receipts | `go test`, `go vet`, `go build`, workflow gate, artifact publication | developer, Go toolchain, CI runner, review gate |
| **`.gooo`-hosted** | `.gooo` declarations, normalized Semantic IR, policy, and explicit handwritten slots | parse/lower, normalize, generate, Go-symbol lift, semantic reconcile, verification, Go build, artifact publication | DSL author, `.gooo` compiler, semantic gate, CI runner, review gate |

The Go-hosted lane is not a strawman: it may emit structured PROV evidence and signed
receipts. The question is whether the evidence contract can preserve semantic identity,
relation direction, dependency freshness, and authority boundaries without the `.gooo`
IR. Conversely, the `.gooo` lane receives no automatic trust benefit merely because its
records use PROV names.

Implementation status is an explicit control: this repository's initial stage is
Go-hosted, while `.gooo`-hosted compiler/CI is a self-hosting target. Until the target
pipeline exists, its fixture verdict is `not-run`, not pass or fail. The tables below
specify expected observations and acceptance criteria; they are not evidence that the
future lane has already succeeded. This follow-up remains documentation-only.

### 12.2 Common scenario and evidence contract

Both lanes process the same small billing change: an input source is compiled, tests
run, a verification result is produced, and an artifact is published. The common
logical IDs are intentionally explicit and stable:

```text
exp://prov-compare/entity/source
exp://prov-compare/entity/policy
exp://prov-compare/entity/ast
exp://prov-compare/entity/ir
exp://prov-compare/entity/generated-go
exp://prov-compare/entity/test-report
exp://prov-compare/entity/build-artifact
exp://prov-compare/entity/verification

exp://prov-compare/activity/compile
exp://prov-compare/activity/parse
exp://prov-compare/activity/lower
exp://prov-compare/activity/generate
exp://prov-compare/activity/test
exp://prov-compare/activity/verify
exp://prov-compare/activity/build
exp://prov-compare/activity/publish

exp://prov-compare/agent/author
exp://prov-compare/agent/compiler
exp://prov-compare/agent/ci-runner
exp://prov-compare/agent/gate
```

These are comparison role slots, not a claim that both lanes contain the same literal
Entity or Activity. The Go-hosted lane may have no IR or generated-Go Entity; the
`.gooo`-hosted lane is expected to have both. A missing slot is measured as missing
evidence only when the lane's declared contract requires it. Any cross-lane equivalence
must be an explicit mapping from the lane-local ID to a comparison role.

The `.gooo` lane uses parse, lower, and generate Activity IDs; a future lift Activity
may be added when Go analysis is part of the run. The Go lane may collapse equivalent
implementation steps into a compiler receipt, but it must declare the collapse rather
than pretending the steps were observed. Both lanes must export
the following minimum evidence fields for each artifact or relation occurrence:

| Field | Purpose |
| --- | --- |
| `stableID` | Semantic identity; never a display label or path alone. |
| `kind` | Entity, Activity, or Agent primary role plus explicit types. |
| `digest` | Content or canonical-record hash for the observed object. |
| `inputs` | Ordered only for display; semantically a set of referenced input IDs and digests. |
| `toolchainID` | Compiler/runner identity and version or digest. |
| `policyID` | Verification policy/ontology revision used by the Activity. |
| `relationID` | Stable identity for a qualified Usage, Generation, Derivation, Association, or Attribution. |
| `bundleID` | Named evidence context and normalized content digest. |
| `status` | Declared, deterministic, or candidate. |
| `attestation` | Optional signature/proof binding Agent, Activity, inputs, outputs, and policy. |

The Entity/Activity/Agent evidence comparison is:

| Evidence role | Go-hosted observation | `.gooo`-hosted observation | Comparison invariant |
| --- | --- | --- | --- |
| Source Entity | Go source revision, `go.mod`, workflow file, and content digests | `.gooo` source, policy, handwritten slot, and content digests | Same canonical source digest contract; paths are labels only. |
| Semantic Entity | Usually absent or inferred from symbol/package names | Normalized IR node with stable ID, kind, and source span | Missing semantic nodes are measured as missing evidence, not guessed. |
| Generated Entity | Compiler output or binary tied to command receipt | Generated Go, source map, and binary tied to projection Activities | Output digest and generator inputs must bind identically. |
| Compiler Activity | One or more `go` command receipts | Named parse/lower/normalize/generate/lift Activities | Collapsed Go steps must declare their observation granularity. |
| CI Activity | Workflow job, test, vet, build, and publish steps | Semantic verification, scope, freshness, build, and publish steps | Each pass is bound to the same input, policy, and toolchain tuple. |
| Human/automation Agent | Commit author, Go toolchain, runner, reviewer | DSL author, `.gooo` compiler, runner, semantic gate, reviewer | Names/aliases do not establish identity; attestation rules are shared. |
| Evidence Entity | Logs, receipts, test report, artifact metadata | IR snapshot, semantic delta, generated source map, evidence Bundle | Both lanes use append-only records and content-addressed freshness. |

The minimum semantic obligations are represented by these lane-specific paths:

```text
Go-hosted:
used(compile, source)
used(test, source)
wasGeneratedBy(test-report, test)
used(verify, test-report)
used(verify, policy)
wasGeneratedBy(verification, verify)
used(build, source)
used(build, verification)
wasGeneratedBy(build-artifact, build)
wasAssociatedWith(compile, compiler)
wasAssociatedWith(test, ci-runner)
wasAssociatedWith(verify, gate)
wasAssociatedWith(build, ci-runner)

.gooo-hosted:
used(parse, source)
wasGeneratedBy(ast, parse)
used(lower, ast)
wasGeneratedBy(ir, lower)
used(generate, ir)
wasGeneratedBy(generated-go, generate)
used(test, generated-go)
wasGeneratedBy(test-report, test)
used(verify, test-report)
used(verify, policy)
wasGeneratedBy(verification, verify)
used(build, generated-go)
used(build, verification)
wasGeneratedBy(build-artifact, build)
wasAssociatedWith(parse, compiler)
wasAssociatedWith(lower, compiler)
wasAssociatedWith(generate, compiler)
wasAssociatedWith(test, ci-runner)
wasAssociatedWith(verify, gate)
wasAssociatedWith(build, ci-runner)
```

The `.gooo` lane therefore tests the design chain `source.gooo -> IR -> generated-go`,
while the Go lane records the directly observed Go-source-to-build chain. The comparison
normalizes both paths into the same required build and verification obligations without
inventing a Go IR node. `wasDerivedFrom` is added only where the lane's evidence
contract asserts an actual dependency, not merely because two files appear in one
directory. `wasAttributedTo(build-artifact, gate)` is an optional final edge and is
valid only when an explicit attribution policy says the gate is responsible for the
published artifact; association with the gate alone is insufficient.

### 12.3 Hypotheses

Each hypothesis is falsifiable. A result is not a win for one lane if the other lane was
given a weaker evidence schema or a different verification policy.

| ID | Hypothesis | Measurement | Falsifier |
| --- | --- | --- | --- |
| H1 | A typed Semantic IR makes relation-direction errors easier to reject than path/log-only evidence. | Direction-mutant rejection rate and diagnostic precision. | Go-hosted lane matches or exceeds `.gooo` lane under the same typed contract. |
| H2 | Stable semantic IDs reduce false identity churn under label rename and declaration reorder. | ID continuity and unrelated-node churn after equivalent edits. | Both lanes preserve the same IDs, or `.gooo` churn is not lower. |
| H3 | Explicit dependency digests and projection closure reduce stale evidence acceptance. | Freshness-mutant rejection rate and time-to-diagnosis. | Equal rejection rates and diagnosis quality, or `.gooo` misses stale closure. |
| H4 | PROV vocabulary alone does not prevent provenance spoofing; authenticated binding is the differentiator. | Forged/replayed claim rejection with and without attestations. | A lane accepts an unsigned or digest-mismatched trusted-agent claim, or either lane claims authenticity from PROV triples alone. |
| H5 | `.gooo` produces a more inspectable semantic path, at the cost of more evidence records and possibly more build time. | Required-edge coverage, path length, record count, bytes, and wall time. | No coverage improvement, or overhead exceeds the pre-declared budget without a locality benefit. |

### 12.4 Counterexample fixture catalogue

Every fixture has a clean case and one mutation. The clean case MUST be accepted by
both lanes with semantically equivalent normalized facts. Mutants MUST produce a stable
failure class; a generic “build failed” is insufficient evidence.

#### Relation direction fixtures

| Fixture | Mutation | Expected result |
| --- | --- | --- |
| `direction-used-reversed` | Replace `used(test, generated-go)` with `used(generated-go, test)`. | Reject: subject must be Activity and object Entity. |
| `direction-generated-reversed` | Replace `wasGeneratedBy(test-report, test)` with `wasGeneratedBy(test, test-report)`. | Reject: subject must be Entity and object Activity. |
| `direction-derived-reversed` | Replace `wasDerivedFrom(build-artifact, generated-go)` with the inverse. | Reject or classify as a different claim; never normalize silently. |
| `direction-associated-reversed` | Replace `wasAssociatedWith(verify, gate)` with `wasAssociatedWith(gate, verify)`. | Reject: Agent cannot be the Activity subject. |
| `direction-qualified-counterpart` | Keep the host `qualifiedGeneration` but set `Generation.activity` to another Activity. | Reject: qualified and unqualified edges disagree. |
| `direction-inverse-view` | Add `generatedBy(build, artifact)` as a canonical fact instead of a query view. | Reject duplicate authority or normalize to the canonical direction with evidence. |

The fixture output must identify the exact subject, predicate, object, qualified relation
ID, and expected domain/range violation. It must not rely on a human reading a graph
visualization.

#### Stable identity fixtures

| Fixture | Mutation | Expected result |
| --- | --- | --- |
| `identity-label-rename` | Rename `PayOrder` to `AuthorizePayment`, retain the Activity ID. | Accept; same node and relation IDs; only label evidence changes. |
| `identity-declaration-reorder` | Reorder declarations and input ports without changing semantic IDs/roles. | Accept; normalized graph and relation IDs unchanged. |
| `identity-namespace-collision` | Add `fraud://entity/Payment` beside `billing://entity/Payment`. | Accept as distinct IDs; reject any implicit merge. |
| `identity-source-span-shift` | Reformat the source so every span moves. | Accept; spans may change, identity and semantic delta must not. |
| `identity-qualified-duplicate` | Create two same-base Usage occurrences with different roles. | Preserve two relation IDs; do not deduplicate by base triple. |
| `identity-unauthorized-rekey` | Change a stable ID but keep the display name and content. | Reject as an identity migration unless an explicit migration record exists. |

For the Go-hosted lane, the baseline fixture uses a path/Go-symbol label in addition to
the explicit stable ID. A result that passes only because a path happened not to change
does not count as stable identity.

#### Freshness fixtures

Freshness is content- and dependency-based. Timestamps are diagnostic metadata, not the
sole freshness proof. A record is fresh only if its source, transitive input, toolchain,
policy, output, and bundle digests match the observed state.

| Fixture | Mutation | Expected result |
| --- | --- | --- |
| `freshness-source-replay` | Change source bytes while replaying the old test/build evidence. | Reject: source digest mismatch. |
| `freshness-ir-replay` | Change `.gooo` or Go semantic input while retaining the old IR digest. | Reject: IR input closure mismatch. |
| `freshness-generated-replay` | Replace generated Go but retain the previous generation receipt. | Reject: generated-output digest mismatch. |
| `freshness-policy-drift` | Change ontology/verifier policy revision but reuse an old pass. | Reject: policy digest mismatch. |
| `freshness-toolchain-drift` | Change Go compiler/toolchain digest without rebuilding. | Reject or mark stale according to policy; never report fresh. |
| `freshness-bundle-edit` | Remove one assertion from a named Bundle but retain its content digest. | Reject: Bundle digest mismatch. |
| `freshness-cache-key-alias` | Reuse a cache result whose key has the same label but a different canonical input tuple. | Reject: cache key is not semantic identity. |

The `.gooo` lane is expected to expose more projection edges (`DSL -> IR -> generated
Go -> verification`) and therefore more freshness checkpoints. The Go lane may pass if
it records an equivalent digest closure; the experiment must measure the contract, not
reward record count.

#### Provenance spoofing fixtures

PROV statements describe responsibility; they do not authenticate who wrote the
statement. These fixtures distinguish structural validity from trust. An unsigned
claim may be structurally well-typed but MUST be reported as `untrusted`, not as a
verified pass.

| Fixture | Attack | Expected result |
| --- | --- | --- |
| `spoof-trusted-agent` | Forge `wasAssociatedWith(build, ci-runner)` using a trusted Agent ID but no valid attestation. | Reject as untrusted; an ID collision is not identity proof. |
| `spoof-output-relabel` | Copy a real build receipt and replace the artifact digest while retaining the old Activity and Agent. | Reject: output digest is not bound to the receipt. |
| `spoof-ci-pass-replay` | Attach a previous commit's signed CI pass to a new source digest. | Reject: signature is valid for the wrong input tuple. |
| `spoof-bundle-rewrite` | Add a “verified” assertion to a Bundle without updating its digest/signature. | Reject: named-set digest and attestation mismatch. |
| `spoof-candidate-promotion` | Promote an ambiguous Go call to deterministic without a DSL assertion or policy decision. | Reject: status transition lacks authority and evidence. |
| `spoof-agent-alias` | Use a display alias or email that resembles the gate Agent ID. | Reject or keep candidate; aliases cannot establish Agent identity. |
| `spoof-generated-marker` | Add a generated-region marker to handwritten Go without a generator receipt. | Reject: text marker is not generation evidence. |

For each spoof, run two subcases: (a) no cryptographic attestation, and (b) an
attestation with a deliberately wrong input/output/policy digest. The pass criterion is
not “the parser rejects malformed RDF”; it is “the gate refuses a well-typed but
unauthorized claim.”

### 12.5 Run protocol and measures

1. Freeze the source payload, Go version, compiler build, CI image, policy revision, and
   fixture seed. Record them as Entities or Agent/toolchain metadata.
2. Run 10 clean repetitions per lane to detect nondeterminism. Run each mutation 10
   times per lane with the same mutation ID. A deterministic verifier must produce the
   same normalized verdict and failure class each time.
3. Export both lanes to the same canonical fact format. Map lane-local Activity IDs to
   a declared comparison ID only when the mapping is explicit and evidence-backed.
4. Normalize before measuring. Do not compare file order, log order, labels, timestamps,
   or generated text as semantic differences.
5. Record raw evidence separately from the derived comparison report. The report is a
   projection and must not overwrite the raw append-only evidence.

The minimum metrics are:

```text
direction_detection = rejected_direction_mutants / direction_mutants
identity_continuity  = equivalent_edits_preserving_ids / equivalent_edits
freshness_detection  = rejected_stale_mutants / stale_mutants
spoof_rejection      = rejected_unauthorized_claims / spoof_mutants
clean_acceptance     = accepted_clean_runs / clean_runs
edge_coverage        = present_required_edges / expected_required_edges
false_acceptance     = accepted_invalid_runs / invalid_runs
```

Also record median and p95 verification time, evidence record count, evidence bytes,
and the number of diagnostics needed to identify the first invalid edge. Report these
as secondary tradeoffs, not as semantic correctness scores.

### 12.6 Pass criteria

The experiment passes only if all of the following hold for both lanes, unless a lane's
declared capability explicitly excludes a fixture:

- clean acceptance is 100% and normalized evidence is equivalent;
- every relation-direction mutant is rejected with a typed direction/range failure;
- label, declaration-order, and source-span changes preserve stable IDs and do not
  widen unrelated semantic locality;
- every freshness mutation is rejected or marked stale, including old evidence replay;
- every spoofed claim is rejected as invalid or untrusted; no PROV-only claim is treated
  as authenticated identity;
- candidate facts cannot satisfy an authoritative required edge or merge gate;
- repeated runs produce the same normalized verdict, failure class, and digest chain;
- Bundle content digests change when assertions change and remain stable when only
  serialization order changes.

For a comparative claim that `.gooo` is better, it must additionally show a predeclared
improvement in at least one primary metric (direction detection, identity continuity,
freshness detection, spoof rejection, or edge coverage) without a clean false-acceptance
regression and without exceeding the agreed evidence/time budget. Otherwise the correct
conclusion is “equivalent under this contract,” “inconclusive,” or “regressed.”

### 12.7 Interpretation limits

The experiment must not conclude that PROV-O itself provides signatures, access control,
or authorship. It provides a shared vocabulary and a model for influence and
responsibility; an independent attestation layer must bind an Agent to the exact
Activity, input digests, output digests, policy, and Bundle. This follows the separation
between provenance semantics and trust policy in the repository's authority model.

The most important negative result is therefore useful: if both lanes reject structural
mutants but accept forged trusted-Agent claims without attestation, the ontology mapping
works while the trust boundary is incomplete. That is a failed spoof-resistance gate,
not evidence that the two lanes are semantically equivalent.

## References

- [W3C PROV-O: The PROV Ontology](https://www.w3.org/TR/prov-o/)
- [W3C PROV-DM: The PROV Data Model](https://www.w3.org/TR/prov-dm)
- [W3C Constraints of the PROV Data Model](https://www.w3.org/TR/prov-constraints/)
- [W3C PROV Model Primer](https://www.w3.org/TR/prov-primer/)
