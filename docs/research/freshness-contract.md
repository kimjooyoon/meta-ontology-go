# Freshness contract experiment

This is an executable research fixture, not a claim that the future `.gooo`
self-hosting stage is complete. The reference runner lives in
`internal/research/freshness`; its only inputs are the JSON contract and the
controlled mutations declared in that contract.

## Falsifiable hypothesis

For a canonical source → generated projection → verification evidence graph,
SHA-256 content and dependency digests are sufficient to detect source edits,
projection edits, missing evidence, and broken PROV references without using
timestamps or map iteration order. The hypothesis passes only when every
demonstrated case passes and no case fails. A deferred case is evidence of an
unimplemented boundary, never a pass.

The minimum fixture contains one authoritative `.gooo` source, one generated
projection, and one evidence entity. Each record has a stable ID, kind,
content, input IDs, recorded digests, and a PROV-shaped activity/entity
envelope. Omitted recorded digests are derived during normalization so the
fixture remains readable; adapters should persist the normalized values.

## Cases and criteria

| Case | Mutation | Expected | Criterion |
| --- | --- | --- | --- |
| `baseline` | none | projection/evidence `fresh` | pass |
| `source-change` | change source content | projection/evidence `stale` | pass |
| `projection-change` | change generated content | projection/evidence `stale` | pass |
| `missing-evidence` | remove required evidence | evidence `missing` | pass |
| `broken-provenance` | replace `used_ids` with an unknown ID | evidence `invalid` | pass |
| `future-gooo-hosted-generator` | self-hosted generator boundary | not evaluated | deferred |

The deterministic measurement output includes active record count, dependency
edge count, provenance edge count, content bytes, non-fresh count, and a
fingerprint of the sorted observations. It intentionally excludes wall-clock
duration. The current fixture asserts two non-fresh observations for source
and projection mutations and one for missing evidence and broken provenance.

## Reusable input/output contract

The JSON fixture is the interchange boundary for future implementations:

```text
AST/source spans + bytes
        │  stable ID, content
        ▼
IR graph ── input_ids ──> projection/cache record
   │                         │
   └── used_ids/provenance ──┴──> evidence record
                                  │
                                  ▼
                         observations + measurements
```

Adapters should preserve these fields:

- AST: stable source ID, source span, and source content digest;
- Semantic IR/BX: stable node IDs and deterministic input dependency order;
- codegen: projection kind, recorded input/content digests, and generated
  entity/activity IDs;
- LSP: a controlled mutation ID and source-backed replacement content, with no
  editor timestamp as a freshness input;
- cache: the same input digest and content digest, with a missing cache entry
  represented as `remove`;
- provenance: `activity_id`, `entity_id`, and declared `used_ids`;
- CI: fail on `fail`, preserve `deferred` as a visible non-success state, and
  publish the normalized contract plus measurement fingerprint as evidence.

## Self-hosting boundary

| Stage | Host | Evidence allowed today |
| --- | --- | --- |
| Initial | Go package reads JSON and evaluates cases | demonstrated pass/fail cases |
| Transition | Go generator emits the same contract from AST/IR | deferred until generator adapter exists |
| Future | `.gooo` declares and regenerates its own freshness contract | deferred until gooo-hosted bootstrap is verified |

The next implementation can replace the reference evaluator one adapter at a
time, provided it consumes the same normalized input and emits equivalent
sorted observations. A changed textual representation is acceptable only when
the stable IDs, statuses, measurements, and provenance edges remain
equivalent.
