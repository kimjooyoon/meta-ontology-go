# DSL↔Go bidirectional transformation

This note defines the research contract for synchronizing `.gooo`, semantic IR,
and Go. It is deliberately a design and conformance note: it does not change
`internal/bidir`, the generator, or the CLI.

## Decision summary

The system should be treated as a partial, resourceful delta lens rather than as
two inverse serializers:

- The `.gooo` declarations are authoritative for business intent.
- The normalized semantic IR is the alignment space. Stable semantic IDs, not
  display names, align nodes across views.
- Go is a partial view. Generated regions are owned by the projection; handwritten
  slots, comments, source order, and implementation-only code are complement
  information that must be retained.
- Go analysis produces a typed fact delta. It must pass provenance, registry,
  locality, and conflict policy before changing the IR.
- Candidate facts are uncertainty, not semantic truth. Promotion requires an
  explicit authoritative assertion or review activity.
- Equality for laws is semantic equality, not byte equality. A second equality
  relation is still needed for retained source complement and generated-region
  locality.

The proposed pipeline is:

```text
S (.gooo) --Get--> M (semantic IR) --Project--> G (Go + slots)
  ^                  ^                  |
  |                  |                  +--Lift--> A (observed Go facts)
  |                  +--Reconcile(M,A,policy)--> M'
  +--PutDSL(S,M')--------------------------------+
```

`Get`, `PutDSL`, `Project`, `Lift`, and `Reconcile` are intentionally separate
operations. Calling every reverse operation `Put` hides the fact that a Go edit
first becomes an observation and only becomes a model update after policy accepts
it.

## Model and equivalence

Let:

- `S` be a parsed DSL source view, including source order and source spans;
- `M` be a normalized semantic model containing nodes, relations, and evidence;
- `G` be a Go source view containing generated regions and handwritten slots;
- `A` be a Go analyzer observation, with deterministic facts, candidates, and
  implementation details;
- `Δ` be an explicit add/remove delta; absence is not deletion;
- `≈sem` mean equality of the fields declared semantic by the IR contract;
- `≈src` mean equality of source complement outside the owned region being
  updated.

`≈sem` must be specified once and reused by the generic model, semantic IR,
generator adapter, cache key, and CI. At minimum it includes stable node ID,
node kind, namespace/identity domain, relation predicate, relation endpoints, and
semantic relation attributes. Package identity, candidate evidence, and IR
version must either be included or explicitly assigned to a separate evidence or
projection equivalence; silently dropping them is not a law.

Presentation fields such as display name, alias spelling, source span, and
formatting may be ignored by `≈sem`, but they are not disposable. They are part of
the complement needed to make a useful `PutDSL` and to preserve source locality.

### Directional contracts

The DSL direction is:

```text
GetDSL : S -> M
PutDSL : S × M -> S
```

The Go direction is deliberately two-stage:

```text
Project : M × G? -> G
Lift    : G -> A
Reconcile : M × A × Policy -> (M, Δ, Decision)
```

`Project` may use an existing `G` to recover handwritten slot bodies. `Lift` must
not treat all Go calls as semantic; only registered semantic identities cross the
boundary. A call to `strings.TrimSpace` remains an implementation detail. A
registered `fraud://activity/check` reference can become a deterministic or
candidate fact depending on resolution.

## Laws

These are the laws to implement and test. They are stronger and more precise than
checking that one generated file parses.

### Get-Put (source retention)

For a valid source `s`:

```text
PutDSL(s, GetDSL(s)) ≈src s
```

At the semantic boundary this also requires:

```text
GetDSL(PutDSL(s, GetDSL(s))) ≈sem GetDSL(s)
```

The source form need not be byte-identical if canonical formatting is an allowed
policy. It must retain stable IDs, explicit relation attributes, ordered activity
ports, source order when order is authoritative, and all non-owned complement
data. For the Go projection, the analogous no-op law is:

```text
Reconcile(M, Lift(Project(M, g))) = (M, empty-Δ, accept)
```

where generated-region and source-map metadata are ignored only where the
projection is allowed to normalize them. A generated Go file must not create a
new deterministic fact merely because it was generated.

### Put-Get (accepted update visibility)

For an accepted, representable model update `m'`:

```text
GetDSL(PutDSL(s, m')) ≈sem m'
```

The Go form includes the reconciliation step:

```text
m1 = Reconcile(M, Lift(g'), policy).model
GetDSL(source-after-accepted-update) ≈sem m1
```

This law is conditional. A candidate, an unknown identity, an unsupported
relation, or a conflict must produce a non-accepting decision and must not be
reported as a successful Put-Get result.

### Semantic round-trip

For every supported DSL fixture:

```text
M0 = GetDSL(s)
g0 = Project(M0, nil)
A0 = Lift(g0)
M1 = Reconcile(M0, A0, strict-policy).model
M1 ≈sem M0
```

For an accepted Go edit `g'`, the stronger loop is:

```text
M0 = GetDSL(s)
A  = Lift(g')
M1 = Reconcile(M0, A, policy).model
g1 = Project(M1, g')
GetDSL(EncodeDSL(PutDSL(s, M1))) ≈sem M1
```

Textual equality is not expected: formatting, generated block order, and source
spans can change under an explicit policy. The semantic graph and the declared
complement must not change accidentally.

### Idempotence and normalization

Repeated operations must converge:

```text
Project(M, Project(M, g)) = Project(M, g)       (bytes, after policy normalization)
Reconcile(M, Lift(Project(M, g)))               = M
Normalize(Normalize(x))                         = Normalize(x)
Diff(x, x)                                       = empty-Δ
```

Fact normalization must not make evidence disappear. If two observations have
the same semantic edge but different source spans, reasons, or attestations,
their edge may be one semantic fact, but both evidence records must remain
addressable in the provenance layer.

### Locality

Let `closure(Δ)` be the policy-defined semantic scope of a delta. For an edit
whose semantic delta is `Δ`:

- only generated regions whose stable IDs are in `closure(Δ)` may change;
- unrelated generated regions must remain byte-identical;
- unrelated handwritten slot bodies and marker-outside text must remain
  byte-identical;
- unrelated DSL declarations, explicit relations, and their source spans must
  remain unchanged;
- analyzer output must contain no deterministic edge outside the registered
  semantic references present in the edited source.

One-hop locality is a useful minimum for the graph, but it is not automatically
the generator locality contract. A changed Activity signature may legitimately
touch its generated ports and source map, while an implementation-only edit may
touch no generated region at all. The two closures must be tested separately.

### Provenance and no false promotion

Every accepted deterministic fact needs source-backed evidence or a separately
authorized trusted adapter. A candidate remains a candidate until an explicit
promotion records who/what decided it and why. Reconciliation is transactional:
if any fact in a batch conflicts, no partial semantic update is committed.

The following implication is a safety property:

```text
ambiguous(A) or missing-source(A) or unknown-ID(A)
  => decision != accept-deterministic
```

## BX research implications

### Why ordinary lenses are insufficient

Classical lenses assume a source/view relationship where `put` receives the old
source so that information discarded by `get` can be retained. That is exactly
the situation here: Go does not contain all DSL declarations, and generated code
does not contain all IR relations. The old DSL, marker regions, slot bodies, and
evidence are required as complement data.

This is also why stable IDs are necessary but not sufficient. IDs solve alignment
when a declaration is renamed, inserted, deleted, or reordered; they do not by
themselves decide whether an edited Go parameter is a rename, a deletion, or a
new port. The adapter must retain an alignment map and an explicit policy for
ordered boundaries.

### Delta lenses and change propagation

The Go analyzer should be viewed as delta discovery, not as a second authority.
The useful split is:

1. align source symbols and stable semantic IDs;
2. discover a typed add/remove delta with source evidence;
3. classify it as deterministic, candidate, implementation-only, or conflict;
4. propagate only an accepted delta into the IR;
5. project the affected closure.

This prevents a full re-analysis from being mistaken for an instruction to delete
every fact that was not observed in one partial Go file. Deletion must be explicit
and authoritative.

### Partial information and uncertainty

Go is incomplete in at least four ways:

- implementation-only calls have no semantic registration;
- a registered symbol can have multiple semantic identities;
- a generated region can be absent or stale;
- a DSL feature can be richer than the current Go projection.

The safe result is a partial delta plus an uncertainty set, not a guessed model.
For ambiguity, retain all candidate identities and a deterministic reason. For an
unsupported but explicit DSL relation, fail at the adapter boundary with a
representability error rather than silently dropping it.

### Conflict resolution

Conflict resolution should be explicit and deterministic:

1. align both updates by stable identity and relation key;
2. compute the overlap of their semantic closures;
3. auto-merge only equal, idempotent updates with compatible kinds and attributes;
4. preserve the union of provenance evidence for a merged fact;
5. emit a conflict containing both deltas and source spans for non-mergeable
   updates;
6. require an authority-bearing resolution activity to choose a winner or amend
   the model.

Disjoint closures should commute. Two updates for the same stable ID with
different kinds, namespaces, port order, or semantic attributes must not be
resolved by map iteration order or last-writer-wins unless that policy is explicit
and evidenced.

## Review of PR #7

PR [#7, “Add bidirectional compiler contract”](https://github.com/kimjooyoon/meta-ontology-go/pull/7)
is a useful foundation: it adds stable-ID alignment, deterministic normalization,
explicit add/remove fact deltas, candidate separation, transactional reconciliation,
and a one-hop locality helper. Its tests cover the happy path for generic
Get-Put, Put-Get, fact layers, presentation-insensitive equivalence, and simple
locality. The following review items are the gaps that must be closed before the
laws can be treated as end-to-end guarantees.

### P1: Activity port order is lost by Put

declarationFromNode reconstructs Activity inputs and outputs from relations and
sorts them by semantic ID. The generic SemanticEquivalent check treats those
relations as a set, but the generator contract says port order is part of a Go
function boundary. Consequently a test can pass while the generated signature
changes.

Failure example:

~~~gooo
activity PayOrder(PaymentMethod, Order) -> Payment
~~~

If billing://entity/order sorts before
billing://entity/payment-method, a Get -> Put -> Get test still sees the same
two prov:used edges, but projection can change from:

~~~go
func PayOrder(paymentMethod PaymentMethod, order Order) Payment
~~~

to:

~~~go
func PayOrder(order Order, paymentMethod PaymentMethod) Payment
~~~

The conformance policy must either declare port order non-semantic everywhere or
preserve and test it as an ordered contract. Given the current generator, it must
be preserved.

### P1: Fact normalization can discard provenance evidence

FactSet.Normalized deduplicates by layer, subject, predicate, and object. Two
observations of the same edge with different source spans or reasons therefore
collapse to one record. The reconciliation fast path also returns when an
existing relation is semantically equal, without merging the new evidence.

Failure example:

~~~text
payment.go:10  PayOrder --invokes--> AuditPayment
audit.go:22     PayOrder --invokes--> AuditPayment
~~~

The semantic edge can be one edge, but both evidence records are needed for
provenance, review, and freshness. A property must assert evidence-set union,
not merely edge-set equality.

### P1: Unknown endpoints are inferred instead of being registry conflicts

ensureEndpoint can create a node for a source-backed deterministic fact when a
kind is supplied or inferred. A typo or stale registry entry can therefore extend
the model instead of producing an unknown-identity conflict.

Failure example:

~~~text
billing://activity/pay-order --gooo:invokes-->
fraud://activity/chekc
~~~

Under a strict semantic namespace policy, chekc must be rejected unless the same
transaction explicitly registers that identity. Auto-creation may be a trusted
import mode, but it must not be the default CI mode and must be recorded as a
declaration delta.

### P1: Generic and semantic-IR vocabularies are not yet closed under round-trip

The generic model admits gooo:invokes, gooo:represents, and
prov:specializationOf, while LowerDocument currently maps only used,
wasGeneratedBy, and wasDerivedFrom into the semantic IR. It also rejects
declaration and relation attributes that the generic model can carry.

Failure example:

~~~text
Get(Document with invokes or relation attributes)
  -> Put(Document, Model)
  -> LowerDocument(updated Document)
  -> representability error or lost attributes
~~~

The supported vocabulary and attribute projection must be declared per adapter.
Unsupported data must fail loudly or remain in an explicit extension/evidence
channel; it must not be silently omitted from a round-trip claim.

### P2: Equivalence scope is under-specified across layers

The generic SemanticEquivalent ignores model package/namespace and candidate
evidence, while the semantic IR canonical form includes version, package,
namespace, and fact status. This can make a generic law pass while the IR law
fails. The project needs named relations such as EquivalentGraph,
EquivalentEvidence, and EquivalentSourceComplement, or one documented
equivalence that includes all required fields.

### P2: Source validity is weaker than source-backed provenance

SourceSpan.Valid accepts a span with only a non-zero offset and no file. That is
useful for an internal adapter but insufficient as CI evidence. Strict policy
should require a source identity plus a valid range, or explicitly distinguish
trusted synthetic evidence from source evidence.

### P2: Name collisions and duplicate-name alignment need one policy

The generic Get path rejects duplicate declaration names, while the semantic-IR
document lowerer builds a name map without the same explicit collision check. A
property suite should feed duplicate names, aliases, and cross-namespace same-name
declarations through every adapter and require either one deterministic resolution
or a conflict.

## Review of generator and CLI boundaries

The generator and CLI branches make the projection boundary concrete, but they do
not yet establish the bidirectional laws by themselves.

### Generator checks are syntax-level, not build-level

The generator parses and formats output, but it does not type-check the complete
file. Entity and Activity Go names are validated in separate maps, so a cross-kind
collision can produce parser-valid Go with duplicate top-level declarations.
Likewise, a field or port with an unknown GoType can be formatted successfully
and fail only when a real package is compiled.

Failure examples:

~~~text
entity GoName = PayOrder
activity GoName = PayOrder
~~~

and:

~~~text
Port{GoType: PaymentTypo}
~~~

The generator conformance gate needs a go/types or package-build check for
generated output, plus a negative test that expects these cases to fail before
the CLI overwrites the output.

### CLI generate bypasses the reverse path

gooo generate currently lowers the DSL and calls the generator directly. The
gooo analyze command emits JSON facts, but there is no CLI path that loads the
DSL model, maps analyzer facts to the bidir fact vocabulary, calls Reconcile,
writes an accepted DSL/IR update, and regenerates the affected Go closure.

Failure example:

~~~text
1. gooo analyze semantic.go  # reports an added invokes fact
2. gooo generate main.gooo --out generated
3. generated output still reflects only main.gooo
~~~

This is not a failed unit test of internal/bidir; it is an unconnected product
path. An end-to-end Put-Get test must prove that an accepted Go delta is visible
to the next check, query, and generate invocation. The same test must prove that
a candidate or conflict leaves the old output untouched.

### CLI adapter shape and write safety

The CLI projection adapter reads the descriptive Parameters/Result fields of the
syntax AST. The parser currently populates those and the compact Inputs/Output
fields, but other AST producers may populate only the compact form; the generic
syntax adapter already handles both forms. A conformance fixture should exercise
both representations and require identical generator IR.

The CLI also reads the previous output and then writes the final file directly.
For a failed write or interrupted process, this can truncate an otherwise valid
generated file. Generate to a temporary file in the same directory, validate it,
then atomically rename it. Atomic replacement is part of locality/safety even
though it is not a semantic lens law.

## Partial information and conflict matrix

| Observation | Default classification | Model effect | Required evidence/decision |
| --- | --- | --- | --- |
| Registered, unique semantic call with valid span | deterministic | add/remove explicit fact | accept and append evidence |
| Registered symbol with multiple identities | candidate | no deterministic edge | retain all options and reason |
| Unregistered helper or standard-library call | implementation detail | no graph change | retain source span outside semantic graph |
| Unknown semantic ID in strict registry mode | conflict | no graph change | reject atomically |
| Missing/invalid source span | conflict | no graph change | trusted adapter or explicit review required |
| Same edge, compatible update, new evidence | merge | one edge, evidence union | deterministic evidence merge |
| Same ID with different kind/namespace | conflict | no partial change | authority-bearing resolution |
| Explicit relation absent from a partial view | no deletion | preserve existing edge | only explicit authoritative remove deletes |
| Unsupported relation/attribute in an adapter | representability conflict | no output claim | extension channel or loud failure |

The key negative rule is that a partial Go observation is not a complete desired
state. Lift must return an explicit delta; Reconcile must not interpret absence
from A as removal from M.

## Property and conformance test specification

The following is a proposed test contract, not an implementation in this
research branch. It can be implemented with the standard library's
testing/quick or a deterministic table-driven generator. Every property should
run with both hand-built fixtures and generated small graphs.

### Fixture generators

Generate valid and invalid cases for:

- stable URI-like IDs, renamed display names, aliases, and same-name nodes in
  different namespaces;
- ordered Activity inputs and outputs, zero/multiple outputs, and explicit
  relation attributes;
- generated regions with stable IDs, unrelated regions, marker-outside text,
  and handwritten slot bodies;
- deterministic facts, duplicate observations with distinct evidence, candidate
  options, implementation-only calls, and explicit removals;
- malformed markers, unknown endpoints, duplicate IDs, conflicting kinds, stale
  packages, and unsupported predicates;
- generator names/types/imports that are parser-valid but build-invalid.

The generator must shrink failures to the smallest graph that still changes a
stable ID, relation, port order, evidence record, or unrelated region.

### Core law properties

| ID | Property | Assertion |
| --- | --- | --- |
| L1 | DSL Get-Put | PutDSL(s, GetDSL(s)) preserves source complement and semantic model; no new fact |
| L2 | DSL Put-Get | Every accepted representable m' satisfies GetDSL(PutDSL(s,m')) ≈sem m' |
| L3 | Go Get-Put | Project followed by Lift and strict Reconcile yields an empty delta |
| L4 | Go Put-Get | Accepted analyzer delta is visible after reconcile, encode, check, and regenerate |
| L5 | Semantic round-trip | DSL → IR → Go → Lift → Reconcile returns an equivalent graph and evidence state |
| L6 | Projection idempotence | Repeating generation with identical IR and previous output produces identical bytes |
| L7 | Normalization idempotence | Normalizing models, facts, and deltas twice equals normalizing once |
| L8 | Stable identity | Renaming a node with unchanged ID changes presentation only, not semantic hash or links |
| L9 | Ordered boundary | Reordering Activity ports is preserved as a semantic change or rejected; never silently sorted |
| L10 | Evidence monotonicity | Merging equal edges retains the union of source spans, reasons, and attestations |
| L11 | Candidate safety | Candidate facts never appear as deterministic facts without explicit promotion |
| L12 | Transactionality | Any conflict leaves model, source, evidence, and generated output unchanged |
| L13 | Explicit deletion | Only an authoritative remove delta deletes a fact; partial absence does not |
| L14 | Disjoint commutativity | Deltas with disjoint semantic closures commute and produce the same normalized result |
| L15 | Conflict determinism | Overlapping incompatible deltas always yield the same ordered conflict record |

### Locality and generator properties

For each generated Go fixture, snapshot bytes for every marker-outside segment,
generated region, and slot body before and after the edit:

| ID | Property | Assertion |
| --- | --- | --- |
| G1 | Region locality | Only regions in the declared semantic closure change |
| G2 | Slot retention | Editing IR or an unrelated region never changes an existing slot body |
| G3 | Deleted region safety | Removing an ID removes only its owned region and its source-map entry |
| G4 | New region stability | Adding a node does not reorder or rewrite existing regions |
| G5 | Marker integrity | Nested, duplicate, mismatched, or unterminated markers are rejected atomically |
| G6 | Build validity | Generated source parses, type-checks, and compiles for valid IR; invalid Go names/types fail before write |
| G7 | Source map alignment | Every region and slot maps to one stable semantic ID and a valid generated range |
| G8 | CLI loop | analyze → reconcile → generate → check/query observes the same accepted delta |
| G9 | CLI rejection safety | Candidate, conflict, or failed validation does not replace the previous output |
| G10 | Adapter parity | Syntax Inputs/Output and Parameters/Result representations produce the same generator IR |

### Conformance fixtures with expected failures

These named fixtures make failures reviewable and should be part of the CI
evidence:

1. ordered_ports_preserved: two inputs whose ID order differs from source order;
   fails if the generated function signature changes on a no-op Put.
2. same_edge_two_evidence_spans: one relation observed in two files; fails if
   either evidence record disappears.
3. unknown_registered_endpoint: a typoed semantic ID; fails if it is silently
   auto-created in strict mode.
4. candidate_not_promoted: ambiguous imported symbol; fails if it enters the
   deterministic graph.
5. partial_observation_does_not_delete: analyzer sees only one of two existing
   edges; fails if the unseen edge is removed.
6. generated_name_collision: entity and Activity share a Go name; fails if
   parser-only validation lets build-invalid output through.
7. invalid_port_type: a parser-valid but unknown Go type; fails if the CLI writes
   output that cannot type-check.
8. unrelated_region_byte_stable: edit one Activity; fails if another region or
   slot changes.
9. unsupported_relation_is_loud: gooo:invokes or relation attributes at an
   adapter without support; fails if data is dropped.
10. cli_accepted_delta_visible: accepted analyze output must change the next
    generated projection; a JSON report alone is insufficient.
11. cli_rejected_delta_atomic: rejected analyzer output leaves the previous
    generated file and its hash unchanged.
12. rename_with_stable_id: display rename retains all references and slot bodies;
    an ID change is a semantic deletion plus addition, not a rename.

### Evidence required from a conformance run

Each run should publish:

- canonical semantic hashes before/after each direction;
- normalized accepted, removed, candidate, and conflicting deltas;
- evidence-set hashes and source spans for accepted facts;
- generated-region hashes and the computed semantic/locality closures;
- CLI exit codes and output hashes for accepted and rejected updates;
- the exact fixture seed when a property test fails.

The gate should fail on stale evidence, not just on a failing assertion. A passing
test with no source span, no delta record, or no generated-region comparison is
not sufficient evidence for a bidirectional claim.

## Follow-up experiment plan: separated hashes and counterexamples

This section is a research follow-up to the current PR gate. It is intentionally
not a CI policy change. The experiment should make semantic drift and evidence
drift observable as different failures, then use minimized counterexamples to
improve the conformance suite.

### Two hashes, not one

Every fixture should report at least two hashes for every durable state:

~~~text
Hsem(x)  = SHA256(Csem(N(x)))
Hevid(x) = SHA256(Cevid(Nevid(x)))
~~~

N is deterministic normalization. Csem contains stable node IDs, node kinds,
identity namespaces, semantic relation keys, semantic relation attributes, and
ordered Activity contracts when port order is declared semantic. It excludes
display names, formatting, source spans, generated bytes, and provenance records.

Cevid contains the evidence ledger: accepted fact evidence, candidate options,
implementation-only observations, conflict records, source URI/range, reasons,
promotion decisions, and evidence identifiers. It is sorted by an evidence
identity that is distinct from the semantic edge key. It must not deduplicate two
records merely because they support the same edge.

The hashes answer different questions:

- Hsem asks whether the meaning and accepted contract changed.
- Hevid asks whether the support, observation, uncertainty, or decision record
  changed.

An implementation-only edit can therefore preserve Hsem while changing Hevid. A
display rename with unchanged identity can preserve Hsem while changing Hevid if
its source span or observation record changes. This is intentional; a single
stable hash would hide that distinction.

Generated-region bytes need a third, separate measurement for locality:

~~~text
Hgen(region-id) = SHA256(bytes(owned region with region-id))
~~~

Hgen is not semantic truth or evidence. It is used to prove that only the
declared generated closure changed. The experiment must never use Hgen as a
substitute for Hsem.

### Trace format

Each run should emit a compact trace that can be attached to CI evidence:

~~~text
seed
base: Hsem Hevid
left: delta Hsem Hevid closure
right: delta Hsem Hevid closure
decision: accept | candidate | reject | conflict
result: Hsem Hevid
regions: {region-id: before-Hgen -> after-Hgen}
~~~

For DSL → Go → DSL, record:

~~~text
S0 --GetDSL--> M0 --Project--> G0 --Lift/Reconcile--> M1
  --PutDSL--> S1 --GetDSL--> M2
~~~

The primary assertion is Hsem(M2) = Hsem(M0) for a no-op projection and
Hsem(M2) = expected for an accepted edit. Evidence assertions compare Hevid
separately, and locality assertions compare only the affected Hgen entries.

For Go → IR, record the shorter path independently:

~~~text
G' --Lift--> A --Reconcile--> M'
~~~

An accepted deterministic fact must change Hsem(M') exactly as its delta says
and must add the corresponding evidence to Hevid(M'). A candidate or rejected
fact must leave Hsem unchanged while still producing an auditable evidence or
decision record.

### Counterexample property matrix

The matrix is the minimum experiment set. An equals sign means equality to the
expected baseline, not byte equality. Union means the evidence hash must reflect
both records even when the semantic edge is one edge.

| ID | Flow and mutation | Expected decision | Hsem expectation | Hevid expectation | Counterexample caught |
| --- | --- | --- | --- | --- | --- |
| X1 | DSL no-op → Go → DSL | accept | equals baseline | equals baseline | Generated code invents a fact |
| X2 | Stable-ID display rename | accept | equals baseline | equals or span-aware change | Name used as identity |
| X3 | Registered source-backed semantic call | accept | changes by exact delta | changes with source evidence | Accepted edge lacks evidence |
| X4 | Implementation-only helper edit | no semantic delta | equals baseline | local change in observation ledger | Helper becomes a graph edge |
| X5 | Ambiguous registered symbol | candidate | equals baseline | changes with all options | Candidate is promoted |
| X6 | Deterministic fact without source | reject | equals baseline | changes with rejection record | Unattributed fact accepted |
| X7 | Unknown endpoint in strict mode | conflict | equals baseline | changes with conflict record | Endpoint auto-created |
| X8 | Partial Go observation omits an old edge | no deletion | equals except explicit additions | observation changes | Absence is treated as removal |
| X9 | Edit one Activity implementation | accept | unrelated scope equals | local change only | Unrelated region rewrites |
| X10 | Reorder Activity ports | ordered change or conflict | changes when order is semantic | changes with source evidence | Put silently sorts ports |
| X11 | Same edge observed at two spans | merge | one semantic edge | union of evidence | Evidence deduplicated by edge key |
| X12 | Unsupported predicate or attribute | reject as unrepresentable | equals baseline | changes with diagnostic | Data silently dropped |
| X13 | Three-way disjoint left/right edits | merge | equals applying both deltas | evidence union | Merge depends on order |
| X14 | Three-way equal edge, distinct evidence | merge | one semantic edge | evidence union | One provenance record lost |
| X15 | Three-way same edge, differing attributes | conflict | base unchanged | changes with both sources | Last-writer-wins hides conflict |
| X16 | Three-way delete versus modify same ID | conflict | base unchanged | changes with both deltas | Partial delete or modify |
| X17 | Three-way incompatible kind or namespace | conflict | base unchanged | changes with identity conflict | Identity boundary weakened |
| X18 | Corrupt or stale generated markers | reject atomically | model equals baseline | changes with diagnostic | Old output overwritten |

The matrix must assert both sides of a split result. X5 is not correct merely
because the candidate is absent from the deterministic graph; the candidate
options and reason must be present in Hevid. X4 must not be called a failure
because its evidence changed: the intended result is semantic stability with
observation freshness.

### Three-way conflict experiment

Use a common base model B and two independently edited models L and R:

~~~text
DeltaL = Diff(B, L)
DeltaR = Diff(B, R)
Mmerge = Merge3(B, DeltaL, DeltaR)
~~~

The merge procedure is evaluated by semantic closure, not by text overlap:

1. align additions, removals, and updates by stable node or relation identity;
2. compute closure(DeltaL) and closure(DeltaR);
3. auto-merge disjoint closures and verify both application orders have the same
   Hsem;
4. merge equal semantic updates while taking the Hevid union;
5. emit a conflict for incompatible updates to one semantic key;
6. leave B unchanged when the decision is conflict or unresolved candidate.

Required three-way cases:

- left adds PayOrder invokes AuditPayment; right adds an unrelated Activity:
  accept both and verify commutativity;
- both add the same invokes edge from different source spans: one edge, evidence
  union;
- left changes a relation attribute while right removes the same relation:
  conflict, no partial update;
- left renames a stable-ID node to AuthorizePayment, right renames it to
  CapturePayment: either an explicit presentation conflict or a declared
  authority policy, never map-order selection;
- left changes the ordered input contract while right changes the implementation
  slot against the old order: conflict or explicit compatibility transform;
- left deletes a node while right adds a relation to it: conflict with both
  source spans and no dangling endpoint;
- left and right both promote different candidates for the same edge: conflict
  unless the promotion policy proves the choices equivalent.

The experiment should also test associativity where it is meaningful:

~~~text
Merge3(Merge3(B, Delta1, Delta2), Delta3)
  is semantically equivalent to
Merge3(B, Delta1, Merge3(B, Delta2, Delta3))
~~~

Associativity is expected only for disjoint or explicitly mergeable deltas. The
test must record a counterexample rather than weaken the assertion when an
overlapping conflict makes associativity undefined.

### Experimental procedure and stopping criteria

1. Generate bounded graphs with up to five nodes, six relations, three ordered
   ports per Activity, and at least two evidence records for selected edges.
2. Generate one or two edit scripts from the same base, including add, remove,
   rename, reorder, attribute change, candidate promotion, and implementation
   edits.
3. Run DSL → Go → DSL and Go → IR independently, then run three-way merge for
   every pair of compatible scripts.
4. Capture Hsem, Hevid, Hgen, decision, closure, and source spans before and
   after each operation.
5. Shrink any violation to the smallest base plus edit scripts and store the
   trace as a named conformance fixture.

The follow-up is ready to graduate from research when:

- no accepted case changes Hsem outside its declared delta;
- no rejected or candidate case changes the accepted semantic model;
- every source-backed accepted fact has a corresponding evidence record;
- equal-edge merges preserve evidence union;
- disjoint three-way merges are commutative and locality-safe;
- incompatible three-way edits produce deterministic conflicts with no partial
  state;
- DSL → Go → DSL and Go → IR report separate, explainable semantic and evidence
  hash outcomes.

## Hosting-stage comparison: Go-hosted now, gooo-hosted later

The current implementation is Go-hosted: the parser, IR adapters, reconciler,
generator, and verification gate are handwritten Go. The future vision is
gooo-hosted: declarations in the semantic language describe enough of the
compiler, evidence, and projection topology to generate or verify more of the
host itself.

This comparison is a contract, not a claim that the future stage exists:

| Contract dimension | Go-hosted initial stage | gooo-hosted future stage |
| --- | --- | --- |
| Authoritative intent | .gooo declarations plus protected Go policy | .gooo declarations plus declared ontology and policy activities |
| Execution host | Handwritten Go parser, adapters, and gate | Generated Go bootstrap followed by gooo-described compiler activities |
| Semantic proof | Go tests, conformance traces, Hsem, Hevid, and Hgen | The same hashes plus a self-hosting derivation from source declarations |
| Evidence authority | Protected verifier records evidence about Go-hosted steps | Independent verifier records evidence about gooo-hosted steps |
| Failure status | Implemented only where current packages and checks pass | Not implemented; no future self-hosting step is a passing result |
| Migration criterion | Stable contract across DSL, IR, Go, and evidence | Bootstrap output is semantically equivalent and evidence-complete across two hosts |

The future stage must not redefine success as “the compiler generated itself.”
It must show a comparable trace:

~~~text
gooo source
  -> semantic IR
  -> Go bootstrap projection
  -> independent verification
  -> gooo-hosted projection
  -> Hsem and Hevid comparison
~~~

Until that trace exists, any gooo-hosted row is a planned contract and must be
reported as unimplemented. The current Go-hosted checks remain the evidence for
the current stage and are not retroactively upgraded by the vision.

## Falsifiable hypotheses and implementation contracts

The previous matrix names the laws and counterexamples. This section makes them
experimentally falsifiable and gives each implementation layer the same small
input/output vocabulary. A result is not a pass merely because a parser,
generator, or future self-hosting stage is absent: an unavailable capability is
recorded as deferred.

### Hypotheses

Each hypothesis has one observable assertion. The fixture identifier, input
revision, and policy revision must be included in the receipt so that a failed
result can be reproduced without relying on wall-clock order.

| ID | Falsifiable hypothesis | Minimal fixture | Measurement | Pass | Fail | Deferred |
| --- | --- | --- | --- | --- | --- | --- |
| H1 | Semantic identity is independent of presentation. | F0, then rename a display name and reformat without changing stable IDs or relations. | Hsem before/after; changed semantic facts. | Hsem is equal and changed facts are zero. | Hsem changes, or a semantic fact changes. | The canonical semantic serializer is not available. |
| H2 | Evidence identity is separate from semantic identity. | F1, then add a second source span supporting the same accepted edge. | Hsem, Hevid, evidence-record count, provenance IDs. | Hsem stays equal; Hevid changes; both evidence records remain. | Hevid is collapsed into Hsem, or one record is lost. | Provenance records or evidence hashing are not implemented. |
| H3 | Get-Put is a no-op on a normalized view. | F0 loaded into an unchanged view and written back. | Semantic hash, accepted delta cardinality, generated-region hashes. | Hsem is equal, accepted delta is zero, and every generated region is equal. | A no-op changes semantics, emits a non-empty update, or rewrites an unrelated region. | The reverse adapter cannot produce a write result. |
| H4 | Put-Get exposes every accepted update and no candidate update. | F1 plus F2's unknown call: one registered call and one ambiguous call. | Accepted, candidate, rejected, and conflict counts; re-read Hsem. | Registered edge is visible after Get; ambiguous edge is still candidate; no rejected fact appears. | An accepted fact disappears, a candidate is promoted, or an unknown endpoint is invented. | The analyzer or candidate-state representation is missing. |
| H5 | Locality is bounded by the semantic dependency closure. | F5: edit PayOrder while AuditPayment is unrelated. | Changed semantic IDs and Hgen for each generated region. | Only the PayOrder closure changes; unrelated Hgen values and bytes are equal. | An unrelated region changes, or a changed dependency is omitted. | Generated-region markers or dependency closure are unavailable. |
| H6 | Partial information is monotone and absence is not deletion. | F2 first has an unknown endpoint, then receives a registry entry. | Fact state transition, accepted/candidate counts, Hsem and Hevid. | Unknown remains candidate; registry completion promotes exactly that fact; no other fact is deleted. | Missing information deletes an accepted fact, or a candidate is silently accepted. | The implementation has no explicit partial-information state. |
| H7 | Three-way merge is deterministic and conservative. | F3 with disjoint, same-value, and overlapping left/right edits. | Merge decision, conflict count, merged semantic hash, replay order. | Disjoint edits commute; equal edits deduplicate; overlap emits a conflict and no partial winner. | Output depends on replay order, or an overlap is silently chosen. | Three-way merge is not implemented in the current lane. |
| H8 | Ordered semantic ports survive both directions. | F4 with two input ports whose types differ and whose order is significant. | Port sequence, positional IDs, Hsem, generated signature. | The exact port sequence is preserved; a reorder is an explicit semantic delta. | Ports are alphabetized, matched by display name, or reordered without a delta. | Port-order metadata is not represented by the IR. |
| H9 | Adapters fail closed on unsupported information. | F6 includes an unsupported predicate and malformed generated markers. | Diagnostic code, accepted/candidate/rejected counts, mutation count. | Unsupported input has a stable diagnostic and causes no silent mutation. | It is dropped, guessed, or emitted as valid Go/IR. | The adapter contract or diagnostic taxonomy is not available. |
| H10 | Cache reuse is valid only for the complete semantic/evidence input. | F0/F1 with one semantic-only edit and one evidence-only edit. | Cache key, hit/miss, Hsem, Hevid, policy revision. | A semantic change misses; an evidence-policy change misses; an exact input hits. | A stale artifact is returned, or evidence changes are ignored when policy requires them. | Cache receipts and invalidation policy are not implemented. |
| H11 | Source-facing diagnostics remain aligned after a local edit. | F7 edits one span in PayOrder while preserving AuditPayment. | Diagnostic set hash, source ranges, semantic IDs, changed URI/ranges. | Diagnostics refer to the edited span/ID only; unrelated ranges and IDs are stable. | A diagnostic moves to an unrelated span, or the same error loses its stable ID. | LSP/source-map output is not present. |
| H12 | Go-hosted and future gooo-hosted stages have comparable evidence, not implied parity. | F0 plus a bootstrap declaration that names the hosting stage. | Host stage, receipt schema version, Hsem, Hevid, implementation status. | Go-hosted output passes only its implemented checks and records the stage. | A future stage is reported as passing without an implementation. | gooo-hosted execution is not implemented; this is the required result today. |

The pass assertions are exact for semantic and provenance relations. Performance
is a separate measurement, not a substitute for a law: record parse, analyze,
normalize, generate, and receipt-write durations as milliseconds and report
count, minimum, median, p95, and maximum over at least five repetitions when a
benchmark exists. No latency threshold is claimed until a repository baseline
and fixture-size class are agreed.

### Minimal fixture catalog

The fixtures are intentionally smaller than a billing example. They contain
stable IDs, a display name, one ordered activity signature, one unrelated
activity, and one source-backed observation. A test may serialize the same
fixture as DSL, Go, or IR, but the stable IDs and relation keys must remain
identical.

~~~text
fixture F0
  namespace billing
  entity order   id billing:Order
  entity payment id billing:Payment
  activity pay_order id billing:PayOrder
    input  payment id billing:Payment
    input  order   id billing:Order
    output payment id billing:Payment
  source: examples/billing/main.gooo
  expected: accepted=1 candidate=0 rejected=0 conflict=0

fixture F1
  F0 plus source observation:
    billing:PayOrder calls billing:PaymentRepository.Save
  second observation may use a different URI/range
  expected: accepted=1 candidate=0 rejected=0 conflict=0

fixture F2
  F0 plus:
    billing:PayOrder calls billing:UnknownRepository.Save
  expected: accepted=1 candidate=1 rejected=0 conflict=0
  completion input: registry adds billing:UnknownRepository

fixture F3
  base: F0
  left:  add relation billing:PayOrder -> billing:PaymentRepository.Save
  right: add relation billing:PayOrder -> billing:Audit.Log
  overlap: both sides change the same relation with different attributes
  expected: disjoint merge succeeds; overlap is a conflict

fixture F4
  F0 with input port order [payment, order], then [order, payment]
  expected: a reorder is visible in the ordered semantic delta

fixture F5
  F0 plus unrelated activity:
    activity audit_payment id billing:AuditPayment
      input payment id billing:Payment
  edit: only the PayOrder body or its accepted relation
  expected: AuditPayment region is unchanged

fixture F6
  F0 with an unsupported predicate and a malformed generated-region marker
  expected: stable diagnostic; no silent acceptance or destructive rewrite

fixture F7
  F0 with source spans for PayOrder and AuditPayment
  edit: one token inside PayOrder only
  expected: diagnostics and source mapping retain the unrelated activity span
~~~

The fixture catalog is a contract, not a claim that every fixture is executable
today. A conformance runner must report an unavailable fixture adapter as
deferred with its missing capability, rather than converting it to pass.

### Shared experiment receipt

Every layer should be able to emit or consume the following logical record.
Field names may be adapted to a package's native type, but their meaning and
presence are not optional once that layer claims support for the experiment.

~~~text
receipt
  schema_version: string
  fixture_id: string
  host_stage: go-hosted | gooo-hosted
  source_revision: hash
  policy_revision: hash
  input:
    ast_hash: hash or absent
    ir_hash: hash or absent
    semantic_hash: Hsem
    evidence_hash: Hevid
    generated_region_hashes: map[region_id]hash
  delta:
    added_semantic_ids: ordered set
    removed_semantic_ids: ordered set
    candidate_ids: ordered set
    rejected_ids: ordered set
    conflict_ids: ordered set
  output:
    semantic_hash: Hsem
    evidence_hash: Hevid
    generated_region_hashes: map[region_id]hash
    diagnostic_hash: hash or absent
    cache_key: hash or absent
  decision: accepted | candidate | rejected | conflict | deferred
  measurements:
    parse_ms: number or absent
    analyze_ms: number or absent
    normalize_ms: number or absent
    generate_ms: number or absent
    receipt_ms: number or absent
  status: pass | fail | deferred
  reason: stable code
~~~

The input and output hashes are deliberately separate. Hsem answers whether the
meaning changed. Hevid answers whether the supporting observations or decisions
changed. Generated-region hashes answer locality. A diagnostic hash answers
source-facing behavior. A cache key answers reuse eligibility. Combining these
values into one opaque digest would make it impossible to distinguish a semantic
regression from an evidence refresh or a permitted local rewrite.

### Measurement and decision protocol

For each fixture, run the baseline, one mutation, and the relevant negative
case. Record exact cardinalities before and after the operation:

| Measurement | Required interpretation |
| --- | --- |
| accepted/candidate/rejected/conflict | State counts after normalization; candidates never count as accepted facts. |
| added/removed semantic IDs | Set delta over stable IDs, not display names or source line numbers. |
| Hsem | Canonical semantic output; equal means semantic equivalence under the declared scope. |
| Hevid | Canonical provenance/evidence output; equality is required only when observations and policy are unchanged. |
| Hgen by region | Byte or canonical-region hash owned by a stable region ID; changed IDs must equal the locality closure. |
| diagnostic hash and ranges | Source-facing diagnostics plus stable IDs/ranges; absent means the LSP layer is not in scope. |
| cache hit/miss and key | Hit only when all declared semantic, evidence, policy, and generator inputs match. |
| durations | Milliseconds with repetition statistics; informational until a baseline exists. |

Use these decision rules:

1. Pass only when every exact assertion for the fixture holds and the receipt
   has no missing required field for the claimed layer.
2. Fail on false promotion, false deletion, evidence loss, nondeterministic
   merge output, unrelated-region change, stale cache reuse, or an unsupported
   input silently becoming valid.
3. Defer only when the missing capability is named in reason and the fixture
   still has a deterministic receipt. Deferred is not evidence that the law
   holds.
4. A mixed run is fail if any required assertion fails, even if an unrelated
   optional layer is deferred. The report must preserve both statuses.
5. Re-run a failure with the same fixture, source revision, and policy revision.
   If the receipt differs without an input change, classify the result as
   nondeterminism before investigating the individual law.

### Reusable layer contracts

These contracts define the follow-up implementation seam. They are deliberately
directional: the DSL remains authoritative for business intent, stable IDs
remain authoritative for identity, and observations never become business
facts merely because an adapter can parse them.

| Layer | Input | Output | Required invariant | Current status |
| --- | --- | --- | --- | --- |
| AST/parser | DSL or Go text, URI, version | AST, spans, syntax diagnostics, ast_hash | Preserve source ranges and stable declaration IDs where present; syntax errors are not semantic facts. | Existing parser surface must be measured; exact receipt adapter is follow-up. |
| Analyzer/LSP | AST, registry, generated markers, source map | accepted observations, candidates, diagnostics, source-to-ID links | Unknown or ambiguous symbols stay explicit candidates; diagnostics do not mutate authority. | Candidate and LSP adapters may be absent; report deferred. |
| Semantic IR | normalized declarations and accepted relations | canonical IR, Hsem, ordered semantic delta | Normalize display-only changes; preserve ordered ports and stable relation keys. | Contract proposed; implementation conformance is follow-up. |
| BX/lens | view, accepted delta, partial state, merge base | updated source/view, decision, delta, conflicts, Hevid | Get-Put no-op; Put-Get accepted visibility; no implicit deletion; three-way overlap is explicit. | Research contract; do not infer implementation from this document. |
| Codegen | IR, previous generated regions, locality closure | Go text, region hashes, source map, generator receipt | Change only owned regions and retain stable markers; implementation-only code is not semantic evidence. | Generator/CLI risks are documented; exact receipt is follow-up. |
| Provenance/evidence | observations, source ranges, candidate and conflict decisions | append-only evidence records, Hevid | Distinct observations remain distinct; promotion records explain why a fact was accepted. | Evidence model is a target contract; no passing result until records are emitted. |
| Cache | source, Hsem, Hevid, policy, generator and tool revisions | artifact, cache key, hit/miss receipt | Never reuse an artifact for a changed declared input; evidence policy participates when evidence is consumed. | Deferred until cache invalidation is implemented. |
| CI/gate | receipts, allowed scope, branch and policy metadata | pass/fail/deferred gate result | Never weaken a check; deferred capability is visible; docs-only scope cannot claim code conformance. | Existing verification can run; ownership alias registration is delegated. |
| Hosting stage | Go-hosted or gooo-hosted compiler inputs | stage-labelled artifacts and receipts | Host stage is explicit; future gooo-hosted work cannot be marked implemented in advance. | Go-hosted is current baseline; gooo-hosted is deferred. |

The minimum implementation order is AST/IR receipt serialization, BX decision
and locality reporting, provenance hashing, then codegen/source-map and cache
adapters. LSP and gooo-hosted execution consume the same receipt rather than
inventing a second equivalence relation. CI should gate only on receipts emitted
by implemented layers and should preserve deferred statuses for the rest.

### Negative-case fixtures that must remain visible

The following cases are intentionally non-passing outcomes:

| Case | Expected result | Regression if |
| --- | --- | --- |
| Unknown endpoint in F2 | candidate plus diagnostic; no accepted edge | the endpoint is guessed or omitted |
| Same semantic edge with two source spans | one semantic edge plus two evidence records | evidence is deduplicated by edge key |
| Delete versus modify in F3 | explicit conflict and unchanged base for that edge | either side wins by replay order |
| Ordered-port reorder in F4 | semantic delta or declared positional conflict | names are sorted and the reorder disappears |
| Malformed marker in F6 | diagnostic and no generated-region mutation | the generator rewrites outside ownership |
| Evidence-only refresh | Hsem equal and Hevid changed | one combined hash hides the distinction |
| Missing cache invalidation input | miss or deferred | stale artifact is a hit |
| gooo-hosted stage without executor | deferred | CI reports self-hosting success |

These negative cases are useful as mutation tests: remove one guard from a
prototype and require the corresponding fixture to fail. A mutation that still
passes indicates that the experiment does not observe the intended law and the
hypothesis or fixture must be tightened.

## Integration cautions

Before integrating this research with PR #7, generator, and CLI work:

- choose one authoritative definition of port order and encode it in every
  model, adapter, generator, and equivalence test;
- separate edge equivalence from evidence equivalence;
- make strict registry closure and unknown-ID handling an explicit policy;
- define a representability matrix for every predicate and attribute channel;
- connect the CLI analyzer to FactDelta and transactional reconciliation before
  claiming Put-Get end to end;
- type-check generated packages and replace output atomically;
- add the conformance fixtures above before changing kernel semantics or relaxing
  verification.

## References

- Foster, Barbosa, Cretin, Greenberg, and Pierce, [Matching Lenses: Alignment
  and View Update](https://www.cs.cornell.edu/~jnfoster/papers/matching-lenses.pdf).
  Alignment is the key reason stable IDs and ordered-port tests belong in the
  lens contract.
- Pacheco, Cunha, and Hu, [Delta Lenses over Inductive
  Types](https://eceasst.org/index.php/eceasst/article/view/1907). Delta-based
  synchronization separates state alignment from meaningful change propagation.
- Diskin, Xiong, and Czarnecki, [From State- to Delta-Based Bidirectional Model
  Transformations](https://www.jot.fm/contents/issue_2011_01/article6.html).
  The delta-lens framing motivates explicit add/remove deltas and composition
  laws.
- Bohannon, Foster, Pierce, Pilkiewicz, and Schmitt, [Boomerang: Resourceful
  Lenses for String Data](https://www.cis.upenn.edu/~bcpierce/papers/boomerang-tr.pdf).
  Resource/complement retention is the relevant analogy for handwritten slots
  and source text.
- [W3C PROV-O](https://www.w3.org/TR/prov-o/), for the provenance vocabulary and
  evidence boundary used by the project.
