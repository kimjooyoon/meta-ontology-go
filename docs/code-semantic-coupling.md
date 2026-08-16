# Code ↔ Semantic Coupling Contract

Status: normative docs contract; the two metric rows in this document are
`DESIGN_ONLY` hypotheses. This document does not add a verifier, CI job, LSP
capability, catalog row, adoption transition, or blocking check.

This contract makes explanations reproducible across the `.gooo` source,
semantic IR, code projection, and evidence. Business intent remains in the
authoritative `.gooo` declarations. Names, file paths, aliases, and display
labels are locators or presentation; stable IDs are identity. The semantic IR
is normalized meaning, not a second business source of truth.

## 1. Coupling law

A **registered semantic code surface** is a stable registry record that binds a
code symbol or generated region to one semantic owner ID and its source-map
origin. The registry, not a changed-file heuristic, defines the affected
surface set. A registered PROV-related symbol is subject to this rule just as
any other registered semantic surface is.

For one immutable snapshot, let `ChangedSurfaces` be the registry-resolved set
of changed registered surfaces. The coupling law is:

```text
receipts(snapshot) = { r | r.surface_id ∈ ChangedSurfaces }
cardinality(receipts(snapshot)) = cardinality(ChangedSurfaces)
and every receipt is current, exact, and independently verifiable
```

There is exactly one current receipt for each changed surface. A receipt has
one `change_claim`, with the wire mapping shown below:

```text
change_claim = DELTA      => receipt_kind = SEMANTIC_DELTA
change_claim = NO_DELTA   => receipt_kind = NO_SEMANTIC_DELTA

CouplingReceipt {
  receipt_id, surface_id, semantic_owner_id
  code_symbol_id, source_map_binding_digest
  snapshot_digest, registry_digest, toolchain_digest, profile_digest
  before_ir_digest, after_ir_digest
  authority_source_before_digest, authority_source_after_digest
  change_claim: DELTA | NO_DELTA
  receipt_kind: SEMANTIC_DELTA | NO_SEMANTIC_DELTA
  semantic_delta_ref: canonical delta or null
  authoritative_source_ref: exact source revision/span or null
  origin_path_id, evidence_refs, state: CURRENT
}
```

`SEMANTIC_DELTA` is valid only when the canonical before/after IR delta is
non-empty, its accepted facts have typed source-backed origins, and the
updated authoritative semantic source is bound by digest. `NO_SEMANTIC_DELTA`
is valid only when the canonical semantic digests are exactly equal under the
same `registry_digest`, `toolchain_digest`, and `profile_digest`; it records
that an implementation-only change was checked, not that semantic evidence
was unnecessary. An implementation-only refactor therefore produces
`NO_DELTA`, not meaningless semantic churn.

The canonical semantic digest excludes labels, source spans, aliases, and
ordering that normalization permits to vary. It includes stable IDs, typed
relations, authority status, and the profile/version that defines the
normalization. A delta must bind the updated `.gooo` or other registered
authoritative semantic source; a changed IR or generated file alone cannot
become authority.

The exact state table is closed:

| Observation | Result | Meaning |
| --- | --- | --- |
| One current receipt; claim and digests satisfy its kind | `PASS` | The surface is coupled. |
| Missing, orphan, duplicate, stale, mismatched, unregistered, or unbound receipt state is established | `UNKNOWN` and `FAIL_CLOSED` | The claim cannot be accepted. |
| Required snapshot, registry, toolchain, profile, source-map, or verifier input is unavailable or ambiguous | `UNKNOWN` | Retry only after the owner restores the exact input. |
| A receipt claims `NO_DELTA` with unequal IR digests, or claims `DELTA` without a non-empty canonical delta and updated source | `UNKNOWN` and `FAIL_CLOSED` | The receipt is contradictory. |

Distinct failure conditions include `missing-receipt`, `orphan-receipt`,
`duplicate-receipt`, `stale-snapshot`, `digest-mismatch`,
`surface-unregistered`, `source-unbound`, `delta-without-source`,
`no-delta-without-equality`, `profile-mismatch`, and `ambiguous-origin`; none
may be collapsed into success or an inferred `NO_DELTA`.

If a registered semantic code surface changes, its semantic contract is
re-evaluated in the same snapshot, even when the change appears to be a
PROV-O vocabulary, adapter, source-map, or terminology-only edit. The law is
not “edit a semantic file”; it is “close the changed registry surface with
one valid receipt.”

## 2. Registry and fixture design

The docs-owned registry is a canonical fixture/schema until a separately
adopted implementation owns it. Its identity-bearing fields are:

```text
TermRecord {
  term_id: stable term ID
  semantic_owner_id: stable business/semantic owner ID
  canonical_name: label only
  definition, definition_version, definition_digest
  related_term_ids: sorted stable IDs
  inference_edge_kinds: closed document shorthand values
  logical_phase: DECLARATION | DERIVATION | PROJECTION |
                 OBSERVATION | LIFT | VERIFICATION
  execution_placement: PRECOMPILE | COMPILE | RUNTIME | VERIFICATION
  authority_layer: BUSINESS_SOURCE | SEMANTIC_IR | PROJECTION |
                    OBSERVATION | VERIFICATION
  effect: DECLARES | NORMALIZES | PROJECTS | CANDIDATE |
          ACCEPTED_UPDATE | VERIFIES | RECORDS_EVIDENCE
  rule_refs: sorted stable rule IDs
  evidence_refs: sorted immutable evidence IDs
  origin_path_ids: sorted typed path IDs
}

CodeBinding {
  code_symbol_id: stable registry ID
  semantic_owner_id: stable semantic owner ID
  package_label, file_label, source_span: locations/labels only
  source_map_id, binding_digest
  registered_surface_id
}

OriginPath {
  path_id: stable path ID
  from_id, to_id: stable registry IDs
  edge_kind: document shorthand for the six typed edge names below
  logical_phase: DECLARATION | DERIVATION | PROJECTION |
                 OBSERVATION | LIFT | VERIFICATION
  execution_placement: proposed category or MISSING/UNKNOWN
  rule_ref, input_digest, output_digest, evidence_ref
}
```

`definition_digest` covers the canonical definition, version, and typed fields;
`binding_digest` covers the symbol/source-map binding. Neither is derived from
a display name or path alone. Every ID must resolve once; an unregistered kind,
missing/ambiguous endpoint, rule, or evidence makes a path incomplete.

The smallest fixture envelope is:

```text
FixtureEnvelope {
  fixture_id, source_digest, registry_digest, toolchain_digest, profile_digest
  semantic_before_digest, semantic_after_digest
  term_records[], code_bindings[], origin_paths[], coupling_receipts[]
  verifier_oracle_digest, expected_result, expected_failure_code
}
```

The positive fixture starts from the current billing grammar:

```gooo
package billing
namespace billing

entity Order id "billing://entity/order"
entity PaymentMethod id "billing://entity/payment-method"
entity Payment id "billing://entity/payment"

activity PayOrder(Order, PaymentMethod) -> Payment
```

The fixture maps the activity to its stable owner and source-map binding without
adding registry syntax. Current grammar has no receipt, typed-path, or metric
policy declarations, so those remain docs-owned schema/fixture records until
syntax is separately designed, implemented, and tested.

Required fixture partitions are:

| Partition | Expected result |
| --- | --- |
| Display-label rename with stable IDs and equal canonical IR | `NO_SEMANTIC_DELTA` |
| Implementation-only Go refactor with equal canonical IR | `NO_SEMANTIC_DELTA` |
| Authoritative declaration/relation change with source-backed delta | `SEMANTIC_DELTA` |
| Missing, duplicate, orphan, stale, or digest-mismatched receipt | `UNKNOWN`/`FAIL_CLOSED` with the exact state code |
| Ambiguous Go observation without an accepted source assertion | Candidate path only; authority and receipt remain unresolved |

## 3. Typed paths, claims, and phases

Inference edge kinds and semantic change claims are different axes. An edge
explains how a term or relation was reached; a receipt claim explains whether
the semantic authority changed. An `OBSERVATION_CANDIDATE` edge never becomes an
accepted semantic change merely because it is visible in a code index. An
`ACCEPTED_LIFT` requires an explicit authority decision and fresh evidence.

The edge names in this document are human contract shorthand. They are not the
serialized values of the existing `semantic.InferencePathV1` wire enum. The
codec uses this total, one-to-one, versioned mapping literally:

| Human contract shorthand | `semantic.InferencePathV1` wire enum |
| --- | --- |
| `DECLARATION` | `AUTHORITATIVE_DECLARATION` |
| `DETERMINISTIC_DERIVATION` | `DETERMINISTIC_DERIVATION` |
| `PROJECTION` | `DERIVED_PROJECTION` |
| `OBSERVATION_CANDIDATE` | `OBSERVATION_CANDIDATE` |
| `ACCEPTED_LIFT` | `ACCEPTED_LIFT` |
| `INDEPENDENT_VERIFICATION` | `INDEPENDENT_VERIFICATION` |

No implementation may infer a mapping from names, aliases, similarity, or an
LLM. Alias matching and name/LLM inference are forbidden; codec mapping must
be literal and versioned. An unknown value, missing mapping, or schema-version
mismatch is `UNKNOWN`.

The document's closed edge vocabulary is exactly:

```text
DECLARATION
DETERMINISTIC_DERIVATION
PROJECTION
OBSERVATION_CANDIDATE
ACCEPTED_LIFT
INDEPENDENT_VERIFICATION
```

The document also uses the following change-claim shorthand. It is distinct
from both the edge vocabulary and the wire edge enum:

| Human contract shorthand | `semantic.SemanticChangeKind` wire enum |
| --- | --- |
| `DELTA` | `SEMANTIC_DELTA` |
| `NO_DELTA` | `NO_SEMANTIC_DELTA` |

The closed change-claim shorthand is exactly `DELTA | NO_DELTA`; it must not be
reused as an edge kind, and an edge kind must not be used as a semantic-change
claim. The change-claim codec uses the same literal, versioned discipline.

Code/semantic coupling rule: the wire enum, these mappings, and their exact
mapping fixtures form one accepted ceiling. Changing either the enum or a
mapping requires this contract and the exact mapping fixtures to change in the
same accepted ceiling. Mismatched snapshots, mapping digests, or versions are
`UNKNOWN`.

The current `semantic.InferencePathV1` logical phase enum is exactly:

| Logical inference phase | `semantic.InferencePathV1` wire phase |
| --- | --- |
| `DECLARATION` | `DECLARATION` |
| `DERIVATION` | `DERIVATION` |
| `PROJECTION` | `PROJECTION` |
| `OBSERVATION` | `OBSERVATION` |
| `LIFT` | `LIFT` |
| `VERIFICATION` | `VERIFICATION` |

Logical inference phase and execution placement are separate axes. The
following are proposed execution-placement categories for this document, not
values of the current `InferencePathV1` logical phase enum:

| Proposed execution placement | Proposed role | Current support claim |
| --- | --- | --- |
| `PRECOMPILE` | Resolve registry, authority, affected surfaces, and paths. | Design placement only. |
| `COMPILE` | Lower/normalize IR, project code, bind source maps, and form receipts. | Current compiler has related IR/projection surfaces; this contract adds no API. |
| `RUNTIME` | Observe execution occurrences/effects only when a runtime profile exists. | Not implied by a design-time `.gooo` activity. |
| `VERIFICATION` | Recompute digests, validate typed paths/receipts, and emit evidence. | Future observer placement; not a current CI promise. |

No implementation may infer execution placement from a logical inference
phase. Until an exact registered adapter exists, execution placement is
`MISSING`/`UNKNOWN`. Logical phases do not claim every phase or adapter exists
today. A design-time activity is not a runtime occurrence; runtime claims
require runtime evidence and distinct occurrence IDs.

## 4. Two inactive metric contracts

The live catalog was checked before choosing these IDs. Neither is a catalog
row: both are `DESIGN_ONLY`, `adoption=UNOBSERVED`, and
`enforcement_effect=NO_EFFECT`; adoption is a separate authority decision.

### 4.1 `gooo.metric.design.inference-path-totality.v1`

| Field | Contract |
| --- | --- |
| Question | Does every required term/code/decision relation have exactly one complete typed path to authority and independent evidence? |
| Exact inputs | Changed-surface set; term and semantic-owner registry; code/source-map bindings; origin paths; rule/evidence registry; source/IR/registry/toolchain/profile digests. |
| Normalization | Sort stable IDs; resolve each endpoint once; canonicalize the six document edge shorthands through their literal wire mapping, logical phase, proposed execution placement, rule, input/output digests, and evidence refs; reject duplicates rather than deduplicating them. |
| PASS law | Every required relation has one and only one complete path ending at authority and an evidence/verifier node; all edge kinds are registered and all digests match. |
| FAIL_CLOSED law | A known duplicate, orphan, unregistered edge kind, illegal endpoint, contradictory logical phase, or mismatched bound is `FAIL_CLOSED` with the exact path failure code. |
| UNKNOWN law | Missing/ambiguous registry, source-map, rule, evidence, snapshot, or verifier input is `UNKNOWN`; it is never an empty path or PASS. |
| N/A law | Only a future immutable catalog row may prove `NOT_APPLICABLE`; absent applicability proof is `UNKNOWN`. |
| Fixture/oracle | The billing fixture above; positives cover declaration → deterministic derivation → projection → independent verification, and negatives remove, duplicate, reverse, or relabel one edge. The oracle recomputes the required relation set and path digest. |
| Observer/layer | Future read-only `term-path-observer`, semantic/BX layer `L1`; verification evidence is retained at the catalog-selected retention level. No current LSP or CI observer is claimed. |
| Failure code/domain/owner/retry | `<MetricID>#<reason-kebab-case>`; predicate failures `FEATURE_LANE`, missing shared inputs `DEPENDENCY_LOCAL`, registry/policy corruption `REPOSITORY_INTEGRITY`; owner is the future registered semantic authority, which repins the snapshot/registry and retries. |
| Anti-gaming | Count obligations and typed paths, not edges, files, terms, or explanation text; inverse/projection copies do not add coverage; candidate paths cannot satisfy authority. |
| Dependencies | Stable-ID and relation-registry closure, source-map totality, candidate isolation, canonical IR, and independent verifier evidence. |
| Observability truth | `PARTIAL`: current repository has related IDs, IR, source spans, and projection evidence, but no adopted registry/path-totality observer. |
| Enforcement | `DESIGN_ONLY`, `UNOBSERVED`, `NO_EFFECT`; no CI gate, merge effect, promotion effect, or LSP claim. |

### 4.2 `gooo.metric.design.code-semantic-coupling-totality.v1`

| Field | Contract |
| --- | --- |
| Question | Does the changed registered code-surface set equal the set of exactly-one valid `DELTA`/`NO_DELTA` receipts, with each receipt bound to the same snapshot and semantic authority? |
| Exact inputs | Registry-resolved changed surfaces; code symbols/source maps; before/after canonical IR and authoritative-source digests; coupling receipts; registry/toolchain/profile digests; semantic delta; independent verifier result. |
| Normalization | Sort by stable `surface_id`; compute changed-surface closure from registry/source-map bindings; canonicalize receipt fields; require one current receipt per surface; compare exact digest tuples. |
| PASS law | `changed_surfaces = valid_receipt_surfaces`, cardinality is one per surface, and every `DELTA` or `NO_DELTA` claim satisfies the coupling law. |
| FAIL_CLOSED law | Established missing/orphan/duplicate/stale/mismatched/unregistered receipt state, contradictory claim, or invalid delta is `FAIL_CLOSED` with its exact receipt code. |
| UNKNOWN law | Unavailable or ambiguous snapshot, registry, toolchain/profile, source-map, IR, source, or verifier input is `UNKNOWN`; it never becomes `NO_DELTA`. |
| N/A law | Only a future catalog predicate may prove inapplicability for a snapshot; missing proof is `UNKNOWN`. |
| Fixture/oracle | Implementation-only refactor, semantic declaration change, candidate-only observation, and each negative receipt partition above. The oracle recomputes changed-surface closure, IR equality/delta, and receipt cardinality. |
| Observer/layer | Future read-only `coupling-receipt-observer`, semantic/BX `L1` plus exact CI-snapshot binding `L4`; it is not a current CI job or stable CLI command. |
| Failure code/domain/owner/retry | `<MetricID>#<reason-kebab-case>`; feature predicate failures `FEATURE_LANE`, missing shared evidence `DEPENDENCY_LOCAL`, registry/catalog corruption `REPOSITORY_INTEGRITY`; future semantic authority owner repins and reruns after the exact cause is repaired. |
| Anti-gaming | File counts, changed-line counts, receipt counts, or equal final text are insufficient; unregistered symbols are UNKNOWN, generated expansion maps to one origin, and a no-op receipt with false digests fails. |
| Dependencies | Semantic scope closure, stable IDs, source-map totality, canonical IR, provenance freshness, and the independent verifier. |
| Observability truth | `PARTIAL`: current code has semantic hashes, source spans, provenance-related structures, and generated bindings, but no adopted registry-wide receipt verifier. |
| Enforcement | `DESIGN_ONLY`, `UNOBSERVED`, `NO_EFFECT`; no adoption, CI, merge, promotion, or LSP behavior changes. |

## 5. Pressure selection and protected floors

All applicable baseline floors are evaluated on every attempt. Let `N` be the
system-owned number of baseline floor coordinates and let `M` be the
system-owned number of selected cross-pressures for this feature. The only
language minimum is `M >= 2`; neither `N` nor `M` is a hardcoded repository
constant. A system may have any cataloged candidate universe and owns its
chosen `N` and `M` through protected policy.

The selector first reserves every applicable base/floor guard. It then ranks
eligible cross-pressures by the canonical tuple:

```text
(cataloged_dependency_closure_risk descending,
 consequence_risk descending,
 observability_gap descending,
 parallel_unblocking_value descending,
 deterministic_cost ascending,
 stable_pressure_id ascending)
```

The fields are normalized from cataloged edges, risk classes, required versus
observed inputs, compatible unblocking work, and a pinned work receipt. Missing
or ambiguous fields are `UNKNOWN`, not zero or agent preference. Take the first
`M` eligible pressures; fewer than `M` records a shortfall and `UNKNOWN`.

Unselected expensive pressures retain their observed baseline floor guards and
may not regress below frozen floors. Their full evaluation state is
`UNSELECTED`, never `PASS` or `NOT_APPLICABLE`; an unavailable guard is
`UNKNOWN`. A selected pressure is not allowed to compensate for an unselected
floor regression. Before a ceiling ratchet, every applicable base and cross
pressure must requalify in the full vector, with fresh digest-bound evidence.

The immutable pressure provenance record binds:

```text
PressureSelection {
  feature_snapshot_digest, policy/catalog/dependency digests
  N, M, baseline_floor_set, candidate_set
  selected_set, unselected_set, unknown_set, selection_shortfall
  rank_inputs, attempts[], next_path_ids[], work_receipt_digests[]
  selector_version, selection_digest
}
```

Attempts and next paths are append-only children. A worker may execute a leased
path but cannot alter the set, invent an edge, close an obligation, or approve
a ceiling.

Non-normative example: a system declares three base guards and selects `M=2`
from four candidates. Coupling ranks first; inference-path ranks second; future
LSP and runtime pressures are unselected but keep floor guards. The selection
digest includes both sets, `UNKNOWN`, two attempts, and the next-path frontier;
the example sets no system default and claims no current observer.

## 6. Read-only explanation flow

The required user-facing chain is:

```text
changed code symbol
  -> stable semantic owner
  -> term definition/version/digest
  -> typed origin path
  -> DELTA or NO_DELTA receipt
  -> independent verifier result
```

An exact-snapshot explanation API is expected to return:

```text
Explain(code_position, snapshot_digest) ->
  { code_binding, semantic_owner, term, origin_path, coupling_receipt,
    verifier_result, evidence_refs, explanation_digest }
  | NO_LINK(reason: AMBIGUOUS | STALE | UNREGISTERED | MISSING)
```

The LSP projection is read-only and uses the same snapshot, registry, and
normalization digests as CI. It navigates code → owner → term → rule → evidence
and returns `NO_LINK` rather than guessing on ambiguity or staleness. This is a
future API expectation, not a claim that the current `lsp` CLI is supported.

CI consumes the same envelope and independently recomputes receipt/path
predicates. It may report `PASS`, `FAIL_CLOSED`, or `UNKNOWN`; an LSP
explanation, agent narrative, or generated projection is not verification
evidence. Users should not reconstruct proof manually from logs.

## 7. Baseline and critique

| Existing practice | Covers well | Remaining gap for this contract |
| --- | --- | --- |
| Documentation and codeowners | Human ownership and prose navigation | No exact semantic identity, typed path, or receipt cardinality. |
| BDD/conformance fixtures | Example behavior and positive/negative outcomes | Usually no code-symbol-to-authority closure or implementation-only receipt. |
| Benchmarks | Resource and performance observations | Cannot prove authority, identity, or semantic non-change. |
| Source maps | Code/output location correspondence | A map alone does not establish a semantic owner, typed inference, or current evidence. |
| PROV-O/PROV-DM/PROV-CONSTRAINTS | Provenance vocabulary, relation direction, qualification, and validation concepts | They do not choose this repository's business authority, registry, receipt law, or CI policy. |

The Gooo layer is justified only when one stable cross-view authority must be
explained through typed, digest-bound evidence and independently checked. Stable
IDs, source maps, append-only evidence, typed edges, and closed-world failure
states are established primitives; the repository-specific coupling law is the
contribution, not novelty for those primitives.

The W3C sources support the boundary but do not supply the repository contract:
PROV-O describes classes/properties/restrictions, PROV-DM says core relations
are binary, and PROV-CONSTRAINTS defines validation/inference and normal forms.
Short excerpts are deliberately bounded: “It provides a set of classes, properties, and restrictions”
([PROV-O](https://www.w3.org/TR/prov-o/)); “In the core of PROV, all relations
are binary” ([PROV-DM](https://www.w3.org/TR/prov-dm/)); and “A PROV instance is
a set of PROV statements” ([PROV-CONSTRAINTS](https://www.w3.org/TR/prov-constraints/)).
Applying those concepts to the Gooo registry, phases, and receipts is a
repository design inference, not a claim that W3C standards define this DSL.

## 8. Downstream contract expectations

Semantic lanes should expose stable IDs, canonical before/after IR digests,
source-backed accepted facts, source-map binding digests, and deterministic
delta evidence; candidate-only promotion remains rejected.

CI lanes should accept the exact changed-surface closure, registry/toolchain/
profile/catalog digests, one receipt per surface, typed paths, independent
verifier results, and owner/retry fields. They distinguish `UNKNOWN` from
`FAIL_CLOSED` and retain evidence append-only. Until those entry points exist,
these are expectations only and current CI support is not claimed.

The contract digest is defined as:

```text
contract_digest = SHA-256(canonical UTF-8 bytes of this document)
```

Consumers bind the digest together with the registry, toolchain, profile,
catalog, and snapshot digests. A document edit changes the contract digest and
requires a fresh explanation/receipt evaluation; it cannot silently reuse an
older result.
