# Deterministic metrics RFC

Status: design-only. This RFC defines a vocabulary and adoption contract for
turning repeated human or agent reasoning into deterministic predicates. It
does not add a verifier, workflow, merge rule, or new blocking check.

## 1. DecisionMetric contract

Every metric is a record, not a score. Its observation is bound to one
immutable snapshot:

```text
DecisionMetric {
  id: stable identifier
  question: deterministic question being answered
  authority_inputs: named authoritative inputs and their digests
  normalization: canonicalization procedure
  value: normalized value
  unit: value's unit or finite domain
  predicate: exact accept/reject expression
  decision: PASS | FAIL_CLOSED | NOT_APPLICABLE | UNKNOWN
  evaluation_state: EVALUATED | DEFERRED | NOT_RUN | ERROR
  applicability: catalog-derived APPLICABLE | NOT_APPLICABLE
  scope: feature | global | integration | promotion | retirement
  failure_domain: FEATURE_LANE | DEPENDENCY_LOCAL | REPOSITORY_INTEGRITY | INTEGRATION_BARRIER | PROMOTION_ONLY | RETIREMENT_ONLY | null
  failure_domain_rule: immutable catalog rule that selects failure_domain | null for PASS or catalog-proven NOT_APPLICABLE
  failure_reason: NONE | PREDICATE_FALSE | REQUIRED_INPUT_MISSING | INPUT_AMBIGUOUS | SNAPSHOT_MISMATCH | APPLICABILITY_UNPROVEN | CATALOG_MISMATCH | EVALUATOR_ERROR | CATALOG_NOT_APPLICABLE | EXCEPTION_WINDOW_ACTIVE | NOT_RUN
  dependency_edges: sorted metric/lane IDs that this result blocks or depends on
  monotonicity: APP | SNAP | EQ | COND
  ci_layer: L0 | L1 | L2 | L3 | L4 | L5
  evidence_retention: R0 | R1 | R2 | R3
  owner: registered authority
  failure_code: deterministic MetricID-specific code for failure_reason | null when failure_reason=NONE
  retry: owner-scoped retry trigger and next action
  evidence_refs: immutable artifact/ledger references
  observed_at: RFC3339 observation time
  expires_at: RFC3339 expiry or null when the catalog permits no expiry
}
```

`evaluation_state` is separate from `decision`: `DEFERRED` and `NOT_RUN` are
not decisions and never mean `PASS`; `ERROR` means the evaluator itself could
not establish a result. Applicability is not inferred from a prose question,
PR body, actor choice, or missing data. It is derived only from an immutable
catalog policy row containing `metric_id`, scope, CI layer, event, path
predicates, and semantic predicates, all bound to the catalog digest.

Every result has exactly one closed `failure_reason` value. `PASS` has
`failure_reason=NONE`, no `failure_code`, and no failure domain. A valid
catalog-proven `NOT_APPLICABLE` has `failure_reason=CATALOG_NOT_APPLICABLE`, a
deterministic code, and no failure domain; it is not an error. All actual
failure or unknown outcomes carry one concrete `failure_domain` selected by
the catalog's `failure_domain_rule`. Failure codes are constructed
deterministically as `<MetricID>#<failure_reason-kebab-case>`, so
`PREDICATE_FALSE` cannot share a reason or code with missing or ambiguous
input. `EVALUATOR_ERROR` is reserved for an evaluator that cannot establish
the other reasons; it is not a generic replacement for an unresolved authority
input.

The valid decision/state/reason combinations are closed:

| decision | evaluation_state | failure_reason/code | required condition |
| --- | --- | --- | --- |
| `PASS` | `EVALUATED` | `NONE` / absent | Applicable catalog row and exact predicate are true. |
| `FAIL_CLOSED` | `EVALUATED` | `PREDICATE_FALSE` / exact code | Applicable predicate is false. |
| `FAIL_CLOSED` | One of `EVALUATED`, `DEFERRED`, `NOT_RUN`, `ERROR` | Any non-`NONE`, non-`CATALOG_NOT_APPLICABLE` reason / exact code | Catalog-declared blocking adapter projects the raw non-PASS result without rewriting its state, reason, code, or domain. |
| `NOT_APPLICABLE` | `EVALUATED` | `CATALOG_NOT_APPLICABLE` / exact code | Immutable catalog proof establishes `NOT_APPLICABLE`. |
| `UNKNOWN` | `EVALUATED` | missing, ambiguous, mismatch, applicability-unproven, or evaluator reason / exact code | Required input or exact result cannot be established. |
| `UNKNOWN` | `DEFERRED` | `EXCEPTION_WINDOW_ACTIVE` / exact code | The union of cataloged exception intervals is active; waived is an annotation, never `PASS`. |
| `UNKNOWN` | `NOT_RUN` | `NOT_RUN` / exact code | Evaluation did not run; it never implies success. |
| `UNKNOWN` | `ERROR` | `EVALUATOR_ERROR` / exact code | The evaluator itself failed before establishing another reason. |

The raw metric result is the source of truth. A catalog-declared blocking
adapter may project any raw `UNKNOWN` or `ERROR` result to `FAIL_CLOSED` for
enforcement, retaining its original evaluation state, failure reason, code, and
domain; it may not project `NONE` or an unproved `CATALOG_NOT_APPLICABLE`.
Other decision/state/reason combinations, including `PASS+ERROR`,
`PASS+PREDICATE_FALSE`, `FAIL_CLOSED+NONE`, and `NOT_APPLICABLE` without
catalog proof, are invalid and rejected.

For every L4 or L5 observation, the packet must contain repository, event,
event_ref, base ref and SHA, head ref and SHA, workflow SHA, run and attempt,
toolchain, policy/catalog/scope digests, `evidence_refs`, `observed_at`, and
`expires_at`. It must also contain each required artifact's role, ID, digest,
and expiry. A catalog-proven `artifact_requirement=none` is represented as an
explicit empty role set, never by omitting the field. Metrics whose catalog
predicates touch branch protection or topology must include the live protection
and/or topology digest; a catalog proof is required to mark either input
not-applicable. There is no “when inputs exist” escape: missing mandatory data
is `UNKNOWN`, and for a declared blocking metric `NOT_APPLICABLE` without
catalog proof, `DEFERRED`, `NOT_RUN`, `UNKNOWN`, or evaluator `ERROR` all map to
`FAIL_CLOSED`.

`APP` means append-only evidence can only add or retain facts, `SNAP` is a
current snapshot value, `EQ` is an exact deterministic equality, and `COND` is
monotonic only under its stated event condition.

The layers are ordered cheap-first: L0 shape and identity, L1 semantic/BX/
projection laws, L2 query/cache/diagnostics/LSP, L3 race/fuzz/performance, L4
exact PR tuple/ownership/evidence, and L5 topology/post-merge/promotion/
retirement. R0 is a log or summary, R1 a commit-bound metric manifest, R2 a
run/head/artifact-bound immutable bundle, and R3 a release or retirement
ledger with recovery evidence.

## 2. Adoption state machine

```text
UNOBSERVED -> OBSERVED -> EXACT_BOUND -> SHADOW -> BLOCKING
```

- `UNOBSERVED`: no reproducible observation exists; no claim is made.
- `OBSERVED`: a value was measured, but its authority inputs are incomplete.
- `EXACT_BOUND`: inputs, normalization, and snapshot binding are complete.
- `SHADOW`: the predicate runs in CI and records PASS/FAIL without changing a
  merge or promotion decision.
- `BLOCKING`: the catalog explicitly declares the metric required for its
  scope and layer, and its exact failure is allowed to gate that scope.

The transition to `BLOCKING` requires an exact-bound fixture, retained shadow
evidence, a stable failure code, and a documented owner/retry path. `UNKNOWN`
blocks only after that declaration; before then it remains an adoption gap.
`DEFERRED` and `NOT_RUN` are never `PASS`, regardless of a successful job.

### Applicability and evaluation

The catalog is the sole authority for applicability. Its immutable row evaluates
the metric's scope, layer, event, changed-path predicate, and semantic-scope
predicate against the same snapshot used by the metric. An evaluator may report
`NOT_APPLICABLE` only with the matching catalog row and predicate evidence.
Retrospective interpretation of a missing check, a human instruction, or a
successful unrelated job cannot make a metric inapplicable.

## 3. Determinization debt

The catalog should measure the remaining human judgment as meta-metrics:

| Meta-metric | Deterministic observation |
| --- | --- |
| Repeated instruction occurrence | Count identical normalized DecisionIDs across dated instructions, issues, runbooks, or handoffs, preserving source locations. |
| Predicate coverage | Required decision questions with a named predicate, authority inputs, fixture, and retained observation divided by declared questions. |
| Observability | Required authority inputs actually observed, normalized, and digest-bound divided by required inputs. |
| Replay determinism | Repeated execution over the same pinned snapshot yields identical normalized values, decisions, failure codes, and canonical evidence payload. |
| Explanation completeness | Failure packets containing question, value, predicate, code, owner, evidence, retry, and next poll. |
| Human-authority exclusion | CI decisions whose inputs include review, approver identity, or agent preference; CI-only paths require zero such dependencies. |
| Catalog drift | Current catalog/policy digest differs from the digest bound to the observation or declared gate. |

Occurrences are counted after normalization, not by string search alone.
Recurrence `>= 2` creates a `DecisionID` candidate for cataloging; it is not an
automatic blocker. Meta-metrics describe determinization debt until separately
adopted through the state machine.

## 4. DAMP/DRY semantic profile

Raw LOC caps of 300 lines per file and 75 lines per function remain
resource-safety limits. They are not evidence that the language is DAMP or DRY
and are reported separately from semantic reuse and its semantic weights.

For a semantic unit `u`, define:

```text
SemanticUnitID(u) = stable semantic identity from the authoritative ID registry
UseSiteID = hash(containing_semantic_id,
                 resolved_target_semantic_id,
                 canonical_ast_path,
                 implementation_slot_id)
ExecutionRootSet(snapshot) = sort(unique(
  immutable_metric_catalog.execution_root_ids(snapshot)
  ∪ canonical_adapter_registry.adapter_root_ids(snapshot)))
G_exec = (V_exec, E_exec)
V_exec = immutable adapter-root IDs ∪ authoritative SemanticUnitIDs
E_exec = (adapter_root_id -> bound SemanticUnitID)
         ∪ (caller SemanticUnitID -> resolved callee SemanticUnitID
            for canonical authoritative calls)
Reach_exec = Reach(ExecutionRootSet(snapshot), G_exec)
AuthoritativeUseSet(u) = {UseSiteID(call) | call is canonical and authoritative,
  containing_semantic_id(call) ∈ Reach_exec,
  target_semantic_id(call) = SemanticUnitID(u)}
R(u) = distinct containing semantic IDs represented by AuthoritativeUseSet(u)
ComponentWeight(u) = count(normalized AST operation tokens in the catalog-pinned
                            component closure, deduplicated by source origin)
LocalImplementationClosure(u) = canonical implementation of u plus its
                                eligible private intra-unit helpers
LocalWeight(u) = count(normalized AST operation tokens in
                       LocalImplementationClosure(u), deduplicated by origin)
```

`ExecutionRootSet` is not inferred from exported names, file paths, changed
files, tests, a `main` function, or a scan of the call graph. Every root must
be an immutable metric-catalog row or a canonical adapter-registry binding,
with a stable ID; the set is sorted after duplicate removal and its
catalog/registry digest is recorded in the observation. An explicitly empty
root set is valid; a missing root declaration or an ambiguous root binding is
`UNKNOWN`.

`G_exec` contains exactly the two directed edge forms shown above: an
adapter-root ID points to its bound semantic ID, and a caller semantic ID
points to its resolved callee semantic ID for a canonical authoritative call.
`implementation_slot_id` remains part of `UseSiteID` and its source binding; it
is not a `G_exec` vertex and no implicit semantic-unit-to-slot edge exists. A
UseSite is reachable if and only if its `containing_semantic_id` is in
`Reach_exec`; it contributes to `AuthoritativeUseSet(u)` only when its target
is exactly `SemanticUnitID(u)`. Incoming edges may be selected for use-site
collection but are never traversed in reverse for reachability.

Containment, ownership, inverse or derived relations, candidate facts, aliases,
unresolved symbols, reflection or string names, unreachable code, and test-only
edges unless the immutable metric policy includes tests are excluded from
`E_exec`. Generated call-sites map to one source-origin edge and do not add
generated expansion edges. A missing or ambiguous adapter, call, owner,
source-origin, stable-ID, or reachability resolution is `UNKNOWN`, never an
empty set.

Candidate facts are not references; aliases are not identities. The containing
semantic ID, target semantic ID, source binding, and call edge must each
resolve uniquely. Thus an unresolved, ambiguous, or unreachable reference
cannot create a consumer or a `PASS` result.

The component closure is rooted at `SemanticUnitID(u)` over a catalog-pinned
edge set used only for component weight:

```text
E_catalog = containment ∪ canonical_calls ∪ owned_implementation_slots
Closure(u) = reachable nodes from SemanticUnitID(u) through E_catalog
ComponentWeight(u) = count(normalized AST operation tokens in Closure(u),
                            deduplicated by origin)
```

Inverse, derived, and candidate edges are excluded from `E_catalog`. The
canonical implementation is the unique authoritative implementation slot for
`u`; it is not selected by the shortest file, function, or text match.

`LocalImplementationClosure(u)` is the least fixed point beginning with that
canonical implementation. From a node already in the closure, it may traverse
only a uniquely resolved private helper edge when the helper is in the same
`SemanticUnitID(u)` and the helper has no separate `SemanticUnitID`. The helper
is eligible only when every authoritative incoming call to it is from a node
already in the closure. The root implementation is exempt from that incoming-
call condition because its authoritative consumers are precisely what
`AuthoritativeUseSet(u)` measures.

Traversal stops at a helper with a stable ID, a generated helper, or a helper
known to be shared; those helpers are separate boundaries and are not folded
into the local closure. A shared helper with no model or stable identity is
`UNKNOWN`, because its ownership and weight cannot be determined. Missing or
ambiguous helper resolution, private/intra-unit classification, incoming-call
set, or source-origin mapping is likewise `UNKNOWN`.

`LocalWeight(u)` counts normalized AST operation tokens across the entire local
closure. A token is identified by its normalized operation and canonical source
origin; repeated generated expansions or repeated views of one origin count
once. Comments, whitespace, aliases, candidate facts, and unreachable nodes do
not contribute. The closure is semantic-ID anchored and file-independent, so
splitting a helper or moving it across files cannot lower `LocalWeight`; all
eligible operations remain in the same closure and are counted as one union.

The reuse classification and its budgets are separate metrics:

```text
ReuseKind(u) = UNKNOWN if any required authority, resolution, or reachability
               input for R(u) is unresolved
             = ORPHAN if |R(u)| = 0
             = DAMP if |R(u)| = 1
             = DRY if |R(u)| >= 2
```

`design.intent-cardinality` reports only `ReuseKind(u)`. The component budget
applies only when `ReuseKind(u)=DAMP` and passes exactly when
`ComponentWeight(u) <= 300`. The local budget applies only when
`ReuseKind(u)=DRY` and passes exactly when one stable canonical implementation
exists and `LocalWeight(u) <= 75`. An `ORPHAN` budget is descriptive and
`NOT_APPLICABLE` only with catalog proof; it never produces `PASS`.

The raw resource metrics `resource.file-loc-cap` and `resource.function-loc-cap`
measure physical file LOC `<= 300` and function LOC `<= 75` respectively. They
are design-only resource checks, not semantic DAMP/DRY evidence.

The profile is:

| Classification | Predicate | Gate meaning |
| --- | --- | --- |
| `ORPHAN` | `|R(u)| = 0` | Descriptive only; no reuse claim. |
| `DAMP` | `ReuseKind(u)=DAMP` | One-use intent may remain local; evaluate the separate component budget. |
| `DRY` | `ReuseKind(u)=DRY` | Shared intent must have one authoritative implementation; evaluate the separate local budget. |
| `UNKNOWN` | `ReuseKind(u)=UNKNOWN` | Never PASS; adoption state decides whether it gates. |

`ComponentWeight(u)` and `LocalWeight(u)` count normalized semantic operations,
not physical lines. A budget result with a false predicate is
`PREDICATE_FALSE`, not `UNKNOWN`; unresolved budget inputs are
`REQUIRED_INPUT_MISSING` or `INPUT_AMBIGUOUS`. A DRY result also requires a
stable identity, a uniquely resolved canonical implementation, and an
authoritative consumer set, not merely similar text.

## 5. Inference ledger and catalog immutability

The authoritative corpus is an append-only set of cataloged source records from
checked-in normative docs and runbooks, issue #106, and gate failure packets.
Each record has `source_type`, a stable URI or ID, `observed_at`, a content
digest, and an exact excerpt locator. A chat paraphrase or an uncataloged
instruction is not authoritative evidence.

Normalization is a versioned rule set that produces a `DecisionID`. The ledger
preserves the raw source digest and the mapping evidence from source record to
normalized ID. Similar wording may propose a candidate recurrence, but
paraphrase similarity cannot establish equality without a catalog rule.

Questions and catalog rows are append-only. Retiring a row requires a recorded
successor, reason, and recovery evidence; the denominator of predicate coverage
must not shrink silently. A catalog digest is therefore part of every exact
metric and evidence tuple.

## 6. Failure domains and dependency propagation

Every non-PASS result and failure packet carries `failure_domain` and sorted
`dependency_edges`:

| Domain | Propagation rule |
| --- | --- |
| `FEATURE_LANE` | A feature failure stops only its own lane. |
| `DEPENDENCY_LOCAL` | A dependency failure stops its transitive dependents; disjoint lanes continue. |
| `REPOSITORY_INTEGRITY` | Global stop is allowed only for protection or policy tamper, unknown ownership/scope, corrupted evidence, unsafe topology, or cross-scope semantic conflict. |
| `INTEGRATION_BARRIER` | A failure stops subsequent merges to the affected protected target; disjoint authoring and CI continue. |
| `PROMOTION_ONLY` | Promotion remains an independent gate and is never inferred from feature green. |
| `RETIREMENT_ONLY` | A failure affects branch retirement only; authoring, CI, and promotion remain independent. |

Dependency edges must identify the exact upstream metric or lane and the
transitive closure used for propagation. A missing or ambiguous edge is a
`REPOSITORY_INTEGRITY` result only when ownership, scope, evidence, catalog,
policy, or topology is corrupted; otherwise it is `DEPENDENCY_LOCAL`. It never
permits stopping unrelated feature work or promoting a green feature.

## 7. Metric family catalog

The first catalog is by family; individual metrics must still satisfy the
`DecisionMetric` contract before adoption. The compact design-only initial
catalog is [docs/metric-catalog.md](metric-catalog.md); it names questions and
fixtures but does not activate any metric or existing check. Catalog completeness
requires every family and priority claim in this RFC to resolve to at least one
concrete `MetricID` or an explicit cataloged gap; a zero denominator is never
complete.

| Family | Questions covered |
| --- | --- |
| Identity, authority, relation | Are IDs stable and unique? Are candidate facts isolated? Are typed relations closed without implicit nodes? |
| BX, locality, no-write | Do Get-Put and Put-Get preserve semantic meaning, limit the change radius, and leave rejected input unchanged? |
| Projection and leverage | Are markers, slots, source maps, generated bytes, and semantic-to-output coverage deterministic and complete? What derived views does one authority declaration supply? |
| Obligation coverage | Does every declared semantic law have a distinct input partition, oracle, and witness? |
| Query, cache, performance | Are query results sound and bounded, cache keys/freshness/invalidation correct, and pinned performance baselines reproducible? |
| Diagnostics and LSP | Are codes, spans, ordering, UTF-16 ranges, versions, cancellation, and read-only behavior deterministic? |
| Ownership and scope | Is there exactly one registered writer, is semantic closure in scope, and are roles separated? |
| Evidence and freshness | Are event/base/head/run/attempt/jobs/artifacts/catalog and write effects exact, digest-bound, current, and append-only? |
| Topology, promotion, retirement | Are protection, target topology, post-merge health, rollback, and issue-backed branch retirement exact? |
| Documentation and contract parity | Do normative docs, the catalog, and the checked-in contract agree without claiming unsupported implementation? |

## 8. Anti-gaming substitutions

Naive counts are descriptive only. The gate uses the corresponding canonical
sets or obligations:

| Naive signal | Deterministic replacement |
| --- | --- |
| Test count | `ObligationWitnessSet` keyed by obligation, fixture partition, and oracle digest. |
| Line coverage | `SemanticObligationCoverage` over declared laws and negative paths. |
| Artifact count | `ArtifactRoleCompleteness` with one allowed current role, exact tuple, and recomputed digest. |
| Relation count | `CanonicalAuthorityEdgeSet` keyed by subject, typed relation, object, and authority layer. |
| Changed-path count | `SemanticScopeClosure` over paths, previous paths, stable IDs, and affected semantic closure. |

Duplicate names or files do not increase these sets. Inverses and generated
expansions are derived views, not new authority edges or witnesses.

## 9. Current observability matrix

This is a design baseline, not a declaration that all rows are gate-ready.

| State | Metric observations |
| --- | --- |
| Observable now | Candidate-independent authority hash; in-process diagnostic determinism; cache reuse/recompute correctness. |
| Partial | DAMP/DRY use census; explicit-ID rename identity; full typed-PROV adapter matrix; durable BX receipt; end-to-end no-write observation; projection leverage; provenance freshness SLA; exact scope manifest; CI replay determinism. |

`Observable now` means the current implementation has a direct deterministic
observation for the named property. `Partial` means the missing observation
must be named before any feature or promotion claim can rely on it.

## 10. Definition priority

Define fixtures, authority inputs, predicates, and evidence packets in this
order; implementation is a later RFC or owner-scoped change.

- **P0:** determinization-debt/replay, exact CI tuple, protection snapshot, and
  issue-backed retirement safety.
- **P1:** intent cardinality, stable identity, source-backed authority, BX and
  locality, and generated boundary preservation.
- **P2:** performance baselines, host parity, rollback rehearsal, and capability
  claims tied to runnable entry points.

This RFC does not activate new blocking checks. Each future metric must first
have a fixture, an exact observation, and retained shadow evidence, then follow
`UNOBSERVED -> OBSERVED -> EXACT_BOUND -> SHADOW` before a separately documented
decision can make it `BLOCKING`. Existing CI and protection contracts remain the
only active gates.
