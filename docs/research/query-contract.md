# Query conformance contract

Status: research/prototype contract. This document adds a fixture oracle and
adapter contract; it does not claim that the query engine, query syntax, or a
gooo-hosted verifier is implemented.

## Falsifiable hypotheses

The fixture at [query-contract.fixture](../../examples/conformance/query-contract.fixture)
defines a small PROV graph with stable IDs and two fact layers. The hypotheses
are deliberately observable by any future AST, IR, BX, or query adapter:

1. **Normalization invariance.** Reordering declarations, facts, or map insertion
   order, and renaming a display label while preserving its stable ID, produces
   the same canonical fact rows and digest.
2. **Candidate non-interference.** Adding a candidate fact changes only candidate
   results and evidence. It cannot add a deterministic exact match, deterministic
   path, semantic scope edge, or promotion decision.
3. **Bounded traversal.** A traversal with `max_depth=N` emits no path deeper than
   `N`. A cycle is finite under simple-path semantics; `max_depth=0` is an error,
   not an implicit unbounded walk.
4. **Namespace safety.** Equal display names with different stable IDs do not
   match. The settlement observation must remain candidate and must not satisfy a
   billing query.

Each hypothesis is falsified by a changed digest, an unexpected row, a candidate
appearing in the deterministic layer, a repeated ID in a simple path, or an
accepted unbounded query.

## Minimal input and expected output

The fixture contains five deterministic facts, two candidates, six nodes, and
one three-edge cycle used only as a negative case. It specifies these queries:

| Query | Expected result |
| --- | --- |
| `exact_input` | one deterministic `used` fact, no candidate |
| `exact_candidate` | no deterministic fact, one candidate `wasDerivedFrom` fact |
| `bounded_neighborhood` | four deterministic paths and one candidate path, depth at most 3 |
| `negative_namespace` | the settlement observation is visible only as one candidate |
| `negative_unbounded` | `invalid-max-depth` error |
| `counterexample.cycle` | no repeated IDs and no emitted path deeper than 3 |

The canonical oracle is seven rows sorted by status, subject, predicate, and
object. The fixture checker records the measured oracle digest so implementations
can compare bytes rather than rely on a prose assertion.

## Measurement protocol

Run:

```sh
./scripts/query-contract-check.sh
```

The checker validates the fixture's row counts, query oracle, and SHA-256 digest.
The recorded baseline is:

| Measurement | Value |
| --- | ---: |
| fixture lines | 58 |
| deterministic facts | 5 |
| candidate facts | 2 |
| canonical oracle rows | 7 |
| canonical oracle digest | recorded in the fixture's `measurements` section |

The checker itself may report `fixture oracle PASS`; that is only a pass for the
static contract artifact. `query_engine_conformance` remains `DEFERRED` until an
adapter evaluates the same queries against a semantic graph. `gooo_hosted_stage`
is `NOT_RUN` and is never promoted by this fixture.

## Pass, fail, and deferred rules

**Pass** when the fixture oracle is internally consistent, canonical output is
byte-stable across repeated runs, and a future adapter matches every expected
query result and negative case.

**Fail** when any count or digest changes without a contract revision, when
candidate facts affect deterministic output, when namespace boundaries are
ignored, when a cycle escapes its bound, or when a deferred/unimplemented host
is reported as authoritative.

**Deferred** for query syntax parsing, semantic IR integration, BX round trips,
Go projection, LSP presentation, cache reuse, provenance publication, and CI
promotion evidence. A deferred item is missing evidence, not a pass.

## Reusable implementation contracts

The following contracts are the handoff between this fixture and future lanes.
Names are proposed stable fields, not an assertion that every package already
exports them.

| Boundary | Required input | Required output | Negative/evidence rule |
| --- | --- | --- | --- |
| AST/query syntax | query text plus source span | canonical `QueryPlan` with IDs, predicate, direction, and depth | unqualified IDs and zero depth are diagnostics; preserve span |
| Semantic IR | normalized stable IDs and PROV facts with status | deterministic/candidate snapshots and canonical fingerprint | candidate cannot mutate deterministic snapshot |
| BX | `Get`/`Put` model plus query view | semantically equivalent model; query view is reconstructable | query results never become business intent or implicit deletion |
| Codegen | IR/query projection options | stable generated region and source map | query projection must not rewrite unrelated handwritten slots |
| LSP | plan/result plus source/evidence spans | stable ordered diagnostics, matches, and path edges | candidate is visibly marked; missing evidence is not hidden |
| Cache | semantic digest, canonical plan, policy/rule version | reconstructable result keyed by all inputs | stale or partial cache is a miss, never authority |
| Provenance | query, input digest, result digest, status, producer ID | append-only evidence envelope | candidate/deferred/not-run blocks promotion |
| CI | fixture, toolchain, policy revision, branch ownership | deterministic pass/fail/deferred decision | Stage 0 Go verifier remains authority; future stages fail closed |

For AST and LSP adapters, a path result should carry both the ordered stable IDs
and the canonical relation facts. For cache and provenance adapters, the seven-row
oracle digest is a result digest, not a semantic identity. For BX and codegen,
changing a display label while preserving IDs is the required locality case.

## Self-hosting evidence boundary

The current Go-hosted baseline remains authoritative. This contract can be used
by a future gooo-hosted candidate as shadow evidence, but it does not grant the
candidate authority. A comparison envelope should identify:

```text
source_digest + semantic_digest + query_plan_digest + result_digest
producer + verifier + decision + evidence_status + promotion_eligible
```

`deferred`, `not-run`, `candidate`, or `mismatch` is not equivalent to `pass` for
promotion. The fixture therefore supports the staged bootstrap evidence shape in
`docs/bootstrap-evidence.md` without claiming a new self-hosting stage.

## Ownership and follow-up

This is an independent research/contract artifact. The current protected CI
ownership map has no `agent/query-contract` alias. CI ownership registration is
delegated to the CI owner; this change does not edit `internal/verify` or weaken
the gate. After registration, a query implementation lane can consume this
fixture, add the adapter, and replace `query_engine_conformance: DEFERRED` with
measured evidence only when the implementation actually exists.
