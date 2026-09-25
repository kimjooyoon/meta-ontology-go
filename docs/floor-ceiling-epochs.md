# Floor/Ceiling Epochs: deterministic optimization and proof

Status: design-only. This is a future architecture contract, not a verifier,
workflow, merge rule, blocking check, MetricID, catalog row, or adoption change.
All 54 catalog rows remain `UNOBSERVED`; the 171 conversation candidates remain
frozen hypotheses. Existing tests and CI are not formal observations of this
model.

## 1. Boundary and views

The model keeps four layers distinct:

| Layer | Authority |
| --- | --- |
| A — semantic decisions | declarations, obligations, accepted decisions, stable IDs |
| B — decision metrics | normalized observations bound to an epoch snapshot |
| C — operational snapshots | refs, jobs, hosts, leases, attempts, current run state |
| D — resource guards | deterministic work, memory, artifact, and other limits |

B, C, and D may inform an A decision, but cannot write A. C cannot prove
semantic correctness; D cannot compensate for an A/B failure; B is a record,
not a score. This preserves `.gooo` business authority, normalized IR,
classified Go observations, and append-only provenance.

| User view | Critical assistant/system view |
| --- | --- |
| “Improve EntityFields safely and quickly.” | An immutable epoch with floors, targets, obligations, finite paths, registered evaluators, leases, and a proof snapshot. |
| “Try implementations in parallel.” | The deterministic scheduler emitted compatible leased tasks; workers cannot choose extra paths or approve results. |
| “The best attempt won.” | The system checked independent predicates. `SUCCEEDED` is not ceiling promotion. |
| “All paths were considered.” | A finite declared graph was covered or exhausted relative to its digest. |

Assistant explanations and hypotheses are useful views, never authority,
selection, path completeness, or proof.

## 2. Generic epoch model

Let `B` be base metrics and `X` expensive cross pressures. Protected policy/SPI
declares their cardinalities and active set:

```text
|B| = N, |X| = M, requested_K = protected-policy value
optimization epoch: normally 1 <= requested_K <= M
maintenance/no-cross epoch: separately cataloged policy case, if allowed
```

`N`, `M`, and `requested_K` are entirely protected-policy-defined. `20/20/K=10`
is illustrative only, never a language constant or default. The language
invariant is instead an acceptance-rank predicate: every authorization or
promotion must evaluate at least two independent, non-compensating acceptance
dimensions. Those dimensions may come from floor guards, selected pressures,
and, at a ceiling, the full required vector. A K shortfall cannot be counted as
one of those dimensions or authorized as if the requested selection occurred.

Every attempt observes all declared base metrics. Every declared metric is also
considered for freshness, slack, and dependency invalidation. That does not
require every expensive cross observer to execute: source/policy/dependency
digests can mark a prior observation fresh, stale, or invalidated before the
selector chooses the full observers to run.

A cross pressure can have:

- a cheap floor guard run on every attempt after promotion into the floor; and
- an expensive full evaluator selected for `K` pressures, or required during
  full-vector ceiling requalification.

A floor coordinate without an every-attempt guard is `UNKNOWN`, not ignored;
its path cannot be authoritative or promote. A cross pressure without a cheap
guard remains ceiling-only. An unselected expensive pressure may drift, but
absence is `UNSELECTED`, never `PASS` or `NOT_APPLICABLE`.

The exact immutable epoch record is:

```text
Epoch {
  epoch_id, parent_epoch_id
  authority_snapshot_digest       # source, IR, accepted decisions
  baseline_vector                 # every declared coordinate
  floor_mask                      # all base + previously promoted cross coordinates
  hard_invariants                 # authority, identity, safety, no-write
  slack_tolerance                 # direction, amount, observation rule
  requested_K, selected_set, target_vector
                                   # deterministic requested-K selection
  selection_shortfall              # explicit shortfall/UNKNOWN record
  obligation_graph, path_graph    # finite, declared, digest-bound graphs
  evaluator_registry              # phase, inputs, outputs, trust, version
  cost_receipts, leases           # deterministic work and conflict state
  feature_priority, snapshot_digest
}
```

The baseline records every coordinate so full requalification can expose drift;
the mask identifies currently enforced floors. It starts with all base metrics
and may expand only after ceiling promotion. Epoch state is immutable; retries,
observations, and operational facts are append-only children.

## 3. Non-compensating progress and promotion

Orient ordered values so larger is better, or use an exact predicate for a
non-ordered invariant. Progress is a product, never an aggregate:

```text
Progress(e, a) = Hard(e,a)
              AND Floor(e,a)
              AND SelectedTarget(e,a)
              AND Provenance(e,a)
              AND Resources(e,a)

AcceptanceDimensions(e, a) = independently declared floor-guard predicates,
                              selected-pressure predicates, and, at a ceiling,
                              full-required-vector predicates
AcceptanceRank(e, a) = rank(AcceptanceDimensions(e,a))
Authorize(e, a) iff Progress(e,a)
              AND AcceptanceRank(e,a) >= 2
AuthorizeOptimization(e, a) iff Authorize(e,a)
                         AND RequestedKShortfall(e,a) = false
```

For every enforced floor coordinate `i`, the candidate must satisfy
`q_i(candidate) >= q_i(floor) - tolerance_i`. Every selected pressure `j` must
meet its target. Hard invariants cover stable identity, allowed semantic scope,
accepted source-backed deltas, protected-policy integrity, leases, and no-write
rules. Provenance covers source spans, provider/evaluator identities, output
digests, and graph witnesses. D cannot make another conjunct true.

Unselected coordinates are explicitly `GUARDED`, `UNSELECTED`, or `UNKNOWN`.
`UNKNOWN` blocks the relevant authority path. It is never made inapplicable
because the observer is costly.

The rank counts only independently declared acceptance dimensions; duplicate
views of one predicate do not increase it. Ceiling promotion requires:

```text
full vector observed
AND all hard invariants true
AND every required floor preserved
AND explicit per-coordinate thresholds OR a Pareto rule
AND complete obligation and path proofs
AND current snapshot and receipts match
AND AcceptanceRank(e, candidate) >= 2
AND (maintenance/no-cross case OR RequestedKShortfall(e,candidate) = false)
```

Strictly increasing every metric is neither required nor generally meaningful.
A weighted score is unsafe because improvement in speed, size, or one business
dimension can mask a safety, authority, or correctness regression. After
promotion, the accepted full vector, invariants, proof bundle, and snapshot are
frozen as the next floor.

## 4. Deterministic selector and scheduler

Inputs must be source-backed:

- semantic impact closure of changed stable IDs;
- dependency and obligation DAGs;
- current slack, target deficit, freshness, and invalidation state;
- evaluator/path-provider availability;
- retained deterministic cost/work receipts;
- leases, semantic conflicts, resource claims;
- immutable feature priority; and
- stable IDs for metrics, tasks, providers, and conflict groups.

The system never infers priority from names, prose, agent confidence, recent
success, or an LLM. Missing or ambiguous values are `UNKNOWN`, not zero.

For each eligible cross pressure, compute canonical `deficit`, `slack`, and
freshness/invalidation status. Sort lexicographically by protected policy, for
example:

```text
(floor-risk descending,
 feature-priority ascending,
 invalidated/stale first,
 deficit descending,
 impact rank ascending,
 retained work units ascending,
 stable metric ID ascending)
```

Take the first `requested_K` eligible pressures after reserving all base and
floor guards. If fewer than `requested_K` are eligible, record a deterministic
selection shortfall and `UNKNOWN`; do not authorize the attempt as though
`requested_K` were satisfied, reuse stale results, or relabel absence
`NOT_APPLICABLE`. A separately cataloged maintenance/no-cross case may use a
different policy path, but cannot be reported as optimization progress. All
metrics still receive freshness/slack/invalidation updates.

Each selected pressure expands to declared tasks for guards, full observation,
obligations, path providers, proof, and resource receipts. A task declares
phase, dependencies, semantic read/write closure, provider/evaluator binding,
lease requirements, resource claims, and retained cost.

```text
ready := tasks with satisfied dependencies
ordered := sort(ready, phase, pressure rank, criticality,
                retained work units, stable task ID)
scheduled := {}
for task in ordered:
  if compatible(task, scheduled, leases, resource claims):
    scheduled += task
repeat until no ready task can be added
```

The result is a maximal compatible set: no additional ready task can be added
without violating a declared dependency, lease, conflict, or resource claim.
It is deterministic, not promised globally optimal. Workers execute leased
tasks and return attempts; they cannot choose pressure sets, invent paths,
close obligations, write proof, or self-approve.

## 5. Finite path proof and undecidability boundary

The system checks two finite graphs:

```text
G_obligation = declared obligations + typed dependencies
G_path       = providers + phase edges + alternative paths
```

An obligation is covered only by a declared path reaching its verifier with a
source/snapshot-bound witness. Each edge records provider, input/output digest,
phase, and next obligation. An alternative is exhausted only with a retained
terminal result and no untried declared continuation.

The proof verifier computes:

```text
required = obligations reachable from declared epoch roots
covered  = obligations with valid path witnesses
complete = covered == required
        AND every reachable alternative is witnessed or exhausted
        AND every edge has a registered provider/evaluator
        AND graph digest equals the epoch graph digest
```

Nodes/edges are sorted by stable ID; duplicate, ambiguous, unbound, and
scope-crossing edges fail. Cycles require a finite SCC rule or explicit finite
iteration bound. An unbounded cycle is not complete.

This proves coverage/exhaustion only relative to the declared finite graph. It
cannot decide arbitrary program solvability, termination, semantic equivalence,
or whether an omitted open-world implementation exists. Missing provider,
evaluator, binding, or input is `UNKNOWN`; user reasoning and assistant
confidence cannot fill it.

## 6. Attempt provenance and failure separation

Each immutable attempt records:

```text
attempt/parent IDs, epoch/task/path IDs, source/IR/target digests,
provider/evaluator/phase, lease, input/output digests, state/reason,
work receipt, obligation witnesses, conflicts, and snapshot digest
```

```text
PROPOSED -> RUNNING -> SUCCEEDED
                    -> FAILED_EXPECTED
                    -> FAILED_INVALID
                    -> DEFERRED
                    -> EXHAUSTED
FAILED_EXPECTED -> PROPOSED   # retry or declared alternative
DEFERRED        -> PROPOSED   # new snapshot, lease, or available SPI
FAILED_INVALID  -> PROPOSED   # new repair/alternative attempt only
```

`SUCCEEDED` proves only the leased task, never promotion. `FAILED_EXPECTED`
retains acceptable exploration that missed a selected target while preserving
authority and safety. `FAILED_INVALID` records an authority/safety violation:
unbound write, forged receipt, unauthorized path, lease conflict, protected
policy mutation, or invalid source span. It may stop its dependency lane or,
for kernel corruption, the repository gate. `DEFERRED` records unavailable
SPI/evaluator, lease, resource, or snapshot conditions; it is not success or
`NOT_APPLICABLE`. `EXHAUSTED` means no declared path/retry remains for this
epoch. No transition rewrites prior evidence.

## 7. SPI and phase contract

Progress requires these registered interfaces, analogous to requiring a JWT
verifier before trusting a JWT:

| SPI | Required result |
| --- | --- |
| Semantic oracle | canonical answer, oracle digest, source-backed witness, or `UNKNOWN` |
| Metric observer | normalized value/predicate, inputs, evaluator digest, evidence reference |
| Proof verifier | verified invariants, path/obligation coverage, receipts, or exact failure/unknown |
| Path provider | finite graph fragment, provider digest, edge witnesses, exhaustion frontier |
| Cost/work receipt | canonical work units for fixture/options/toolchain/provider and receipt digest |
| Phase adapter | phase-bound input/output envelope and adapter identity |

The registry owns implementation, version, trust, inputs, and phase. Missing,
ambiguous, stale, or mismatched SPI/input is `UNKNOWN` and blocks its path;
absence is never catalog-proven inapplicability.

| Phase | Placement |
| --- | --- |
| Precompile | authority, impact closure, paths, selection, evaluator availability, leases |
| Compile | BX Get/Put, projection, generated/source maps, compile work receipts |
| Runtime | effects, latency, host observations, output and resource receipts |
| CI aggregation | exact snapshot, floor/result aggregation, append-only provenance, full requalification, promotion |

These are future bindings, not claims that current adapters exist.

## 8. Meta-level `.gooo` shape

Application DSL remains business intent. A future protected meta surface can
name capabilities, obligations, and phases without workflow commands, mutable
budgets, `K`, adoption, or promotion in business declarations:

```gooo
meta capability EntityFieldCarrier {
  id "gooo://capability/entity-field-carrier"
  obligation EntityFieldRoundTrip
  phase compile
}
meta obligation EntityFieldRoundTrip {
  id "gooo://obligation/entity-field-round-trip"
  subject "gooo://entity/entity-field"
  oracle "gooo://oracle/entity-field-bx"
  path "gooo://path/entity-field-bx"
  phase compile
}
```

Pseudocode only; it is not supported syntax. Protected policy owns thresholds,
floor mask, `N/M/K`, evaluators, retries, adoption, and promotion. The meta
declaration cannot say `run`, `budget=`, `select=`, `adopt=`, `promote=`, or
`PASS`.

## 9. Shared BDD/performance fixture

Both dimensions bind one immutable envelope:

```text
FixtureEnvelope {
  source_digest, semantic_input_digest, oracle_digest,
  obligation_partition, output_digest, work_receipt_digest,
  toolchain/options digest
}
```

BDD/conformance checks obligations and oracle output; D checks deterministic
work for the same source, oracle, output, options, and toolchain. A fast
omission fails the semantic/output dimension; a complete but slow result may
fail the resource guard. Neither is averaged into a score. Unpinned wall time
is shadow evidence; deterministic work units are the default.

## 10. Adversarial cases

- **Weighted masking:** latency rises while an authority invariant falls;
  `Hard=false`, so the attempt is `FAILED_INVALID` despite any score.
- **Unselected laundering:** an omitted expensive observer is `GUARDED`,
  `UNSELECTED`, or `UNKNOWN`; omission never becomes success or N/A.
- **Invented path:** an agent names an unregistered provider; graph proof
  rejects the unbound edge.
- **Forged speed:** a work receipt lacks canonical fixture/options/toolchain
  binding; D is `UNKNOWN`, so the path cannot promote.
- **Entity collision:** duplicate field ID, moved parent, or ambiguous type
  registry rejects transactionally and preserves the prior graph/write set.
- **Lost exploration:** expected failures remain append-only; all exhausted
  alternatives produce `EXHAUSTED`, not a rewritten history.

## 11. EntityFields worked example

PR #197 merged the semantic field model: stable field IDs, immutable Entity
parents, order, names/aliases, spans, presence/cardinality, stable-ID type
references, and transactional validation. PR #198 merged a parser-neutral
latent field carrier while leaving the public grammar unchanged. Draft PR #199
proposes bidir lowering/write-back, typed registry resolution, and rejection
fixtures; it is not merged or authoritative.

```text
field_id    = billing://field/order/customer-id
parent_id   = billing://entity/order
type_ref    = billing://type/string
presence    = required
cardinality = one
order       = 0
```

```text
latent FieldDecl
  -> syntax carrier (#198)
  -> bidir Get/Put (#199 draft)
  -> TypeRegistry
  -> semantic Field/Entity graph (#197)
  -> BX + source-span proof
```

Base guards cover stable identity, parent immutability, transactional no-write,
and fieldless readback. Selected cross pressures may include typed round-trip,
presentation order, and source-map coverage. Duplicate ID/parent move is
`FAILED_INVALID`; missing TypeRegistry SPI is `DEFERRED`/`UNKNOWN`; a valid
task is `SUCCEEDED` but does not promote the epoch. Full field-vector and path
requalification are still required for a ceiling. Documenting this example
changes no metric or adoption state.

## 12. Critical comparison

| Adjacent system | Reused | Added here | Not promised |
| --- | --- | --- | --- |
| BDD/performance | fixtures, oracles, budgets | one digest-bound semantic/output/work envelope; non-compensating results | counts prove coverage or wall time is stable |
| Bazel/DAG/cache | closures, actions, replay | semantic obligations, finite path proof, leases, epochs | optimal scheduling or cache soundness without a key |
| Compiler passes | phases, IR, lowering, maps | capability/obligation-bound observers and proof across phases | replacement for a compiler or equivalence solver |
| Provenance | IDs, digests, append-only lineage | computed coverage/exhaustion bound to an epoch | truth outside the declared graph |
| Proof assistants | checked witnesses and boundaries | operational evaluator/path/work receipts as finite build proof | theorem proving or arbitrary solvability |

The useful novelty is the composition: non-compensating progress, finite path
exhaustion, deterministic selection, leased parallel attempts, and append-only
provenance share one immutable snapshot. The pieces should reuse current
compiler, conformance, CI-evidence, and provenance surfaces.

## 13. Proposal, critique, resolved contract

| Proposal | Critique | Resolution |
| --- | --- | --- |
| 20/20/K=10 and selected expensive pressures | Counts could become policy; skipped observers hide drift. | Protected `N/M/requested_K`; at least two independent acceptance dimensions by rank; all metrics enter freshness/slack/invalidation; only the requested selection runs full observers; a shortfall is `UNKNOWN`; ceiling requalifies the vector. |
| Base floor and rising ceiling | A floor without an observer is unsafe; strict rise is not a valid multidimensional rule. | Immutable vector/mask, hard invariants, tolerances, selected targets, Pareto or explicit thresholds. |
| Parallel agent heuristics | Agents cannot choose paths or classify safety failures. | System scheduler/registries decide; leased workers emit closed-state attempts. |
| Complete declared paths | Arbitrary completeness is undecidable. | Finite digest-bound graph coverage/exhaustion, with explicit boundary. |
| User-supplied SPIs | Missing input cannot mean harmless N/A. | Missing SPI/input is `UNKNOWN` and blocks its path. |
| BDD plus performance | Separate fixtures permit fast omission or slow completeness to be misread. | Shared fixture/oracle/output/work receipt with independent predicates. |

## 14. MVP and constraints

1. Define canonical epoch, fixture, task, attempt, path, and receipt records;
   add no catalog rows or MetricIDs.
2. Implement a pure selector with all base guards and a policy-declared
   requested `K` (normally `1 <= K <= M`), lexicographic ties,
   freshness/slack/invalidation, shortfall `UNKNOWN`, and no model inference;
   require an acceptance rank of at least two for authorization.
3. Verify finite graphs, stable digests, alternate exhaustion, and `UNKNOWN`.
4. Add leases/conflicts and append-only expected/invalid/retry/exhausted cases
   on one EntityFields fixture.
5. Bind BDD and work receipts in shadow evidence; adoption/blocking remains a
   separate future authority change.
6. Add phase adapters/CI aggregation only after replay is exact; promotion
   remains disabled until full-vector evidence is exact.

Keep the algebra `coordinate + obligation + path + evaluator + receipt`.
Avoid a distributed planner, graph database, theorem prover, LLM scheduler,
unbounded path search, mutable business-DSL budgets, weighted score, or
automatic adoption.

## 15. Open questions

- Which work units and receipt comparisons survive evaluator/toolchain changes?
- Which cross pressures get cheap floor guards versus ceiling-only treatment?
- What threshold/Pareto rule, tolerance decay, and floor expansion are safe?
- Which cycle forms and finite loop bounds are admitted?
- How are provider trust roots rotated without worker control?
- How do leases recover after host loss without duplicate acceptance?
- Which phase adapters are implementable without implying unsupported APIs?
- How can future adoption relate to the design-only catalog without premature
  MetricID versioning or state change?
