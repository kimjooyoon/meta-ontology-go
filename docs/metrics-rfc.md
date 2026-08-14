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
  scope: feature | global | promotion
  failure_domain: FEATURE_LANE | DEPENDENCY_LOCAL | REPOSITORY_INTEGRITY | PROMOTION_ONLY
  dependency_edges: sorted metric/lane IDs that this result blocks or depends on
  monotonicity: APP | SNAP | EQ | COND
  ci_layer: L0 | L1 | L2 | L3 | L4 | L5
  evidence_retention: R0 | R1 | R2 | R3
  owner: registered authority
  failure_code: stable code when not PASS
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

Raw file and function limits of 300 and 75 lines remain resource-safety caps.
They are not evidence that the language is DAMP or DRY and are reported
separately from semantic reuse.

For a semantic unit `u`, define:

```text
SemanticUnitID(u) = stable semantic identity from the authoritative ID registry
L(u) = canonical logical weight of u after AST/IR normalization
UseSiteID = hash(containing_semantic_id,
                 resolved_target_semantic_id,
                 canonical_ast_path,
                 implementation_slot_id)
AuthoritativeUseSet(u) = reachable, resolved, authoritative UseSiteIDs for u
R(u) = distinct containing semantic IDs represented by AuthoritativeUseSet(u)
```

`AuthoritativeUseSet` is built only from reachable resolved semantic references.
It excludes candidate facts, aliases as identities, unresolved symbols,
reflection strings, unreachable code, and test-only references unless the
immutable metric policy explicitly includes tests. A generated call-site is
mapped to its source origin exactly once; generated expansion is not a new use.

The component closure is rooted at `SemanticUnitID(u)` over a catalog-pinned
edge set:

```text
E_catalog = containment ∪ canonical_calls ∪ owned_implementation_slots
Closure(u) = reachable nodes from SemanticUnitID(u) through E_catalog
L(u) = count(normalized AST operation tokens in Closure(u), deduplicated by origin)
```

Inverse, derived, and candidate edges are excluded from `E_catalog`. The
canonical implementation is the unique authority node referenced by all use
sites in `AuthoritativeUseSet(u)`. If the root, edge catalog, origin mapping,
resolution, or reachability cannot be determined, the result is `UNKNOWN` and
never `PASS`.

The profile is:

| Classification | Predicate | Gate meaning |
| --- | --- | --- |
| `ORPHAN` | `|R(u)| = 0` | Descriptive only; no reuse claim. |
| `DAMP` | `|R(u)| = 1` and the component closure weight is `<= 300` | One-use intent may remain local. |
| `DRY` | `|R(u)| >= 2` and one stable canonical implementation has local weight `<= 75` | Shared intent must have one authoritative implementation. |
| `UNKNOWN` | `R(u)` or `L(u)` cannot be determined exactly | Never PASS; adoption state decides whether it gates. |

`L(u)` counts normalized semantic operations, not physical lines. Comments and
whitespace, generated duplicates, aliases, candidate facts, unreachable code,
and fake or unresolved references do not create consumers. A generated region
counts once through its source-map origin. Splitting one semantic closure across
files or helpers does not reduce its closure weight; otherwise the metric is
being gamed. A DRY result also requires a stable identity and an authoritative
consumer set, not merely similar text.

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

Every result and failure packet carries `failure_domain` and sorted
`dependency_edges`:

| Domain | Propagation rule |
| --- | --- |
| `FEATURE_LANE` | A feature failure stops only its own lane. |
| `DEPENDENCY_LOCAL` | A dependency failure stops its transitive dependents; disjoint lanes continue. |
| `REPOSITORY_INTEGRITY` | Global stop is allowed only for protection or policy tamper, unknown ownership/scope, corrupted evidence, unsafe topology, or cross-scope semantic conflict. |
| `PROMOTION_ONLY` | Promotion remains an independent gate and is never inferred from feature green. |

Dependency edges must identify the exact upstream metric or lane and the
transitive closure used for propagation. A missing or ambiguous edge is a
repository-integrity observation failure, not permission to stop unrelated
feature work or to promote a green feature.

## 7. Metric family catalog

The first catalog is by family; individual metrics must still satisfy the
`DecisionMetric` contract before adoption.

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
