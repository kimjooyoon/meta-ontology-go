# Codegen experiment contract

This note turns the code-generation research questions into falsifiable experiments
that future AST, semantic IR, bidirectional transformation (BX), generator, LSP,
cache, provenance, and CI implementations can share. It is an independent follow-up
to the open codegen research work; it does not modify or extend that PR.

The current Go-hosted baseline is authoritative. The generator and semantic CLI are
not assumed to exist on every `integration` snapshot. An unavailable operation is
`DEFERRED`, never `PASS`. A candidate `.gooo` host may emit evidence for comparison,
but it cannot promote itself or replace the Go verifier.

## Experiment record

Every fixture uses one versioned record shape. The record is an input/output contract,
not a claim that the current CLI already reads JSON.

```json
{
  "schema": "gooo/codegen-experiment/v1",
  "fixture": "pay-order-v1",
  "operation": "generate-roundtrip",
  "input": {
    "dsl_sha256": "sha256:<digest>",
    "ir_sha256": "sha256:<digest>",
    "previous_go_sha256": null,
    "generator": {
      "contract": "gooo/generator/v1",
      "implementation": "sha256:<digest>"
    },
    "policy_sha256": "sha256:<digest>"
  },
  "expected": {
    "status": "pass",
    "failure_code": null,
    "semantic_equivalent": true,
    "allowed_region_ids": ["billing://activity/pay-order"],
    "protected_region_ids": [],
    "required_evidence_facts": [
      "ast.normalized",
      "ir.normalized",
      "codegen.regions",
      "bx.roundtrip"
    ]
  },
  "observed": null
}
```

The runner fills `observed` only after execution. Goldens must not contain a digest
that was copied from the implementation under test; the first golden is produced by
an independently reviewed reference or by a human-reviewed canonical record.

### Canonicalization rules

- All paths are repository-relative slash-separated paths.
- Stable semantic IDs, fixture IDs, schema versions, and rule IDs are canonical.
- Arrays that represent sets are sorted by stable ID; declared activity port order is
  preserved because it is part of the Go signature.
- JSON uses UTF-8, no insignificant whitespace, and a trailing newline for files.
- Digests are lowercase SHA-256 over canonical bytes and are prefixed with `sha256:`
  in human-readable records.
- Timestamps, hostnames, temporary roots, process IDs, absolute paths, and map
  iteration order are provenance only and never enter a canonical digest.
- An absent digest or a `deferred`/`not-run` result cannot compare equal to `pass`.

## Reusable stage outputs

Each implementation adapter should expose the same logical outputs even if its native
data structures differ:

| Stage | Input | Canonical output | Consumer |
| --- | --- | --- | --- |
| AST | DSL bytes and URI | normalized declarations, source spans, diagnostics | semantic lowering and LSP |
| Semantic IR | AST plus policy | stable nodes/facts, normalized semantic digest | BX, codegen, query, cache |
| BX lift/reconcile | Go symbols/source map | sourced delta, candidates, locality IDs, conflicts | semantic verifier and CI |
| Codegen | normalized IR plus previous artifact | source, region manifest, imports, source map, version | compiler, LSP, freshness gate |
| LSP | source map plus position/semantic ID | resolved URI/range and stable ID | editor and agent context |
| Cache | canonical input tuple | key, hit/miss, artifact digest, schema | incremental build and CI |
| Provenance | stage result and evidence facts | producer-independent evidence bundle | independent comparison |
| CI | evidence bundles and changed scope | verdict, rule IDs, promotion state | merge gate |

The adapter boundary must return an explicit status:

```text
PASS       required operation ran and every declared invariant passed
FAIL       required operation ran and at least one invariant failed
DEFERRED   operation is unavailable or its promotion stage is not enabled
```

`DEFERRED` is useful evidence about rollout state. It is not a softer spelling of
`PASS` and must remain visible in the final evidence bundle.

## Minimal fixture pack

The smallest useful corpus has one positive fixture per invariant family and one
negative mutation per failure class. Every fixture has `input`, `golden`, and, when
needed, `negative` records. These names should remain stable when implementations
move packages.

### `pay-order-v1`: AST → IR → Go → BX

Authoritative DSL input:

```gooo
package billing
namespace billing

entity Order id "billing://entity/order"
entity Payment id "billing://entity/payment"

activity PayOrder(Order) -> Payment
```

Required normalized IR facts:

```text
entity billing://entity/order name=Order
entity billing://entity/payment name=Payment
activity billing://activity/pay-order name=PayOrder
used billing://activity/pay-order billing://entity/order
generated billing://entity/payment billing://activity/pay-order
```

Required projection facts:

```text
region activity billing://activity/pay-order
slot billing://activity/pay-order/implementation owner=billing://activity/pay-order
port input order type=Order
port output payment type=Payment
```

Pass requires that normalized IR after `lift(project(IR))` is semantically equivalent
to the input and that every lifted fact has a source span or is recorded as a
candidate. A helper such as `strings.TrimSpace` must not become a semantic relation
merely because it appears in the slot body.

### `imports-permute-v1`: deterministic import projection

Use the same IR with an explicit import set:

```text
imports:
  encoding/json
  time
  clock time/tzdata
```

Generate all six permutations of the three import inputs, plus two permutations of
the entity and activity arrays. The golden records the canonical import list and
source digest. A counterexample adds `timer time` beside the existing `time` import
or aliases a package to a generated top-level name. Both must fail with a stable
import rule, not choose an alias based on arrival order.

### `marker-tamper-v1`: protected marker negative corpus

Start with one generated activity, one slot, one protected handwritten block, and
one marker-outside helper. Apply exactly one mutation per case:

```text
missing-end
nested-generated
mismatched-close-id
duplicate-slot-id
slot-outside-generated
retagged-region
generated-signature-edit
protected-body-edit
slot-body-edit
```

Only `slot-body-edit` is potentially allowed. Every other mutation must return
`FAIL`, write no replacement artifact, and name the marker kind, stable ID, and line
when available. `retagged-region` is the important counterexample: a parser that only
checks balanced syntax may incorrectly transfer handwritten code to another
activity.

### `locality-pay-order-v1`: handwritten locality

The previous Go artifact contains two generated regions, one slot, one protected
handwritten block, and a helper outside markers. Candidate outputs are:

```text
slot-edit
same-id-display-rename
add-unrelated-activity
remove-pay-order
outside-comment-edit
protected-helper-edit
```

The golden stores the allowed semantic region closure and exact protected bytes. The
negative cases are `outside-comment-edit` and `protected-helper-edit`; a generator
must not silently treat them as owned output. `remove-pay-order` must not leave stale
executable code behind.

### `version-skew-v1-v2`: compatibility boundary

Keep the semantic IDs and slot owner constant while changing only a display name and
formatter output between generator contract versions. Then mutate the marker schema,
slot ID, and removed-region policy one at a time. The golden distinguishes:

```text
same meaning + allowed formatting change -> PASS or policy-approved TEXTUAL_DRIFT
unknown marker/schema without migration -> FAIL VERSION_SKEW
changed semantic ID or slot owner -> FAIL OWNERSHIP_MIGRATION_REQUIRED
old reader silently accepts new ownership semantics -> FAIL UNDECLARED_COMPATIBILITY
```

### `bootstrap-stage-0-v1`: explicit deferred state

Use `examples/bootstrap/main.gooo` and the existing bootstrap evidence shape. The
Go-hosted record may have ordinary Go checks as `pass`, while semantic CLI and
bootstrap comparison remain `deferred`/`not-run`. The future gooo-hosted record is
`not-run` and `promotion_eligible: false` until an independent verifier compares it.

This fixture's counterexample is a record that sets `promotion_eligible: true` while
`semantic_cli` is `deferred`. CI must reject it as `UNIMPLEMENTED_STAGE_CLAIM`.

## Falsifiable hypotheses

Correctness hypotheses use exact counts rather than subjective review. A hypothesis
is `PASS` only when all required cases meet its threshold; one false acceptance is a
failure.

| ID | Hypothesis | Minimum measurement | PASS threshold |
| --- | --- | --- | --- |
| `H-IMP-001` | unordered IR input does not affect canonical imports | 6 import × 4 declaration permutations; source/import digests | 24/24 equal |
| `H-MARK-001` | protected marker mutations fail closed | 8 structural mutations plus retagging | 9/9 reject; 0 writes |
| `H-LOC-001` | slot-only edits remain local | 3 positive and 3 negative locality cases | positives have exact allowed closure; negatives reject |
| `H-BX-001` | projection/lift preserves semantic meaning | 10 round-trips over two namespaces | 10/10 normalized IR equivalent |
| `H-SKEW-001` | version changes require declared compatibility | v1/v1, v2/v2, v2/v1, v1/v2 | no undeclared acceptance; owner changes reject |
| `H-CACHE-001` | cache keys are path/environment independent | 8 temp roots × 8 input permutations | equal keys for equal canonical input |
| `H-LSP-001` | source maps resolve stable IDs after formatting | all region and slot mappings in golden | 100% in-bounds, reversible mappings |
| `H-BOOT-001` | repeated Stage 0 builds reproduce evidence | 2 clean roots × 2 runs | canonical payloads equal; builder metadata may differ |
| `H-CI-001` | unknown branch ownership cannot pass | known and unknown scope branches | unknown branch always rejects |

The runner should also report counts that help diagnose partial progress:

```text
cases_total, cases_pass, cases_fail, cases_deferred
false_acceptances, false_rejections, no_write_violations
semantic_equal, source_equal, region_equal, source_map_resolved
cache_key_equal, evidence_equal, environment_leaks
```

Timing may be recorded for performance research, but a timing improvement never
overrides a correctness failure and timing is excluded from canonical evidence.

## Cross-stage output contracts

### AST and source spans

The AST adapter must emit a normalized declaration record with:

```text
kind, display_name, semantic_id, namespace, source_uri,
start_offset, end_offset, start_line, start_column, end_line, end_column
```

Offsets are byte offsets in the original UTF-8 source. A diagnostic without a source
span is allowed only for file-level failures. Re-parsing equivalent source with line
ending normalization must either preserve byte offsets by policy or declare a new
source digest; it must not silently reuse stale LSP ranges.

### Semantic IR and BX

The normalized IR is the comparison form. Its canonical facts contain:

```text
subject_id, predicate, object_id, subject_kind, object_kind,
fact_class, source_span, candidate_reason
```

`fact_class` distinguishes authoritative, deterministic, candidate, and rejected
facts. A BX adapter returns:

```text
added_facts, removed_facts, candidate_facts,
locality_ids, conflicts, source_spans
```

The negative contract is explicit: a deterministic fact without source evidence must
be rejected as `BX/MISSING_SOURCE`, not silently promoted from implementation text.

### Codegen and region manifest

The generator result must expose these records independently of the rendered bytes:

```text
generator_contract, implementation_digest, package
regions: kind, semantic_id, owner_id, start_offset, end_offset, body_digest
slots: slot_id, owner_id, start_offset, end_offset, body_digest
imports: path, alias, used_by_ids
source_map: semantic_id, kind, source_range, generated_range
source_digest, semantic_digest
```

Region and slot manifests are the locality comparison surface. A source digest may
change under an allowed formatter revision, but an owner or semantic ID must not
change without a migration record.

### LSP and source maps

For every generated region and slot, the LSP adapter must resolve both directions:

```text
semantic ID -> generated range -> source/IR range
source/IR range -> semantic ID -> generated range
```

The negative fixture deletes a region while retaining an old source-map entry. The
resolver must return `STALE_MAPPING`, not a range into an unrelated region. Mapping
offsets are checked against the final generated bytes, after formatting.

### Cache identity

The proposed cache key is the SHA-256 of a canonical tuple:

```text
cache_schema
generator_contract
generator_implementation_digest
normalized_ir_digest
previous_artifact_digest
policy_digest
toolchain_contract
```

Absolute worktree paths, timestamps, producer labels, and temporary directory names
are excluded. A changed semantic IR, generator contract, policy, or previous artifact
must change the key. A display-name-only change with the same stable ID may change the
source digest while retaining the semantic cache identity only if the cache contract
explicitly separates semantic and textual artifacts.

### Provenance and CI evidence

The current `internal/verify` evidence contract uses `gooo/evidence/v1`, a stage,
fixture, decision, and sorted stable-ID facts. Codegen experiments should map their
results into facts such as:

```text
codegen/semantic-digest
codegen/source-digest
codegen/region-digest
codegen/import-digest
codegen/source-map-digest
codegen/failure-code
codegen/status
```

Producer identity remains outside the canonical payload comparison. A Go host and a
future gooo host may have different producer/attestation identity while agreeing on
the normalized facts. A candidate must not emit a `pass` decision for a fixture whose
required operation returned `DEFERRED`.

## Pass, fail, and deferred rules

The fixture runner applies these rules in order:

1. Invalid fixture schema, missing stable IDs, malformed markers, or unsafe paths are
   `FAIL` before generation.
2. Expected negative mutations must fail with the declared primary rule class. A
   successful generation is a false acceptance and fails the experiment.
3. Positive fixtures compare semantic, ownership, locality, source-map, cache, or
   evidence fields according to their contract. Unspecified fields do not become
   accidental equality requirements.
4. If the adapter, generator, CLI, or stage is unavailable, return `DEFERRED` with a
   missing capability and no promotion eligibility.
5. Any `DEFERRED`, missing evidence fact, or comparison mismatch blocks promotion;
   none is coerced to `PASS`.

Stable failure codes should use the following reusable categories:

| Code | Meaning | Required action |
| --- | --- | --- |
| `AST/INVALID_SPAN` | source span is missing or out of bounds | reject AST result |
| `IR/IDENTITY_COLLISION` | stable IDs collide across kinds/namespaces | reject lowering |
| `BX/MISSING_SOURCE` | deterministic lifted fact has no source evidence | keep candidate or reject |
| `CODEGEN/IMPORT_DRIFT` | canonical import projection changed unexpectedly | reject projection |
| `MARKER/OWNERSHIP_UNPROVEN` | marker retagging or owner mismatch | reject without write |
| `LOCALITY/ESCAPE` | unowned/protected/unrelated bytes changed | reject refresh |
| `LSP/STALE_MAPPING` | source map resolves to old or unrelated bytes | reject mapping |
| `CACHE/KEY_DRIFT` | equal canonical input produced a different key | invalidate contract |
| `SKEW/UNDECLARED` | reader/writer accepted an undeclared version | reject compatibility |
| `EVIDENCE/MISMATCH` | producer-independent facts differ | block promotion |
| `CI/UNKNOWN_SCOPE` | branch has no ownership alias | fail closed; CI owner registers alias |
| `STAGE/DEFERRED` | requested implementation/stage is unavailable | report deferred, never pass |

## Follow-up implementation contract

The first implementation lane may provide only the generator adapter and its tests.
Later lanes must consume the same fixture record without changing fixture IDs or
failure meanings:

```text
RunFixture(record) -> observed record
NormalizeAST(source) -> AST snapshot
LowerIR(AST) -> normalized IR + semantic digest
Generate(IR, previous) -> projection + region/source-map manifest
Lift(Go, manifest) -> sourced BX delta
ResolveLSP(manifest, query) -> stable ID/range or STALE_MAPPING
CacheKey(canonical inputs) -> digest
EmitEvidence(observed) -> EvidenceArtifact
CompareEvidence(go, gooo) -> equal or EVIDENCE/MISMATCH
CheckScope(changed IDs, branch) -> pass or CI/UNKNOWN_SCOPE
```

The adapter must be pure over its canonical inputs where possible. Filesystem writes
are atomic and occur only after validation; failed experiments retain their evidence
and do not overwrite goldens, prior artifacts, or rollback records. CI remains the
guardian of ownership and promotion. The generator, candidate host, or fixture runner
must not weaken that boundary to make an experiment pass.

## Current baseline status

At the time this contract is authored:

- Go 1.26.5 and the ordinary Go format/vet/test/race gates are available;
- bootstrap evidence fixtures and producer-independent evidence comparison exist;
- the semantic CLI and generated-freshness command remain explicitly deferred on the
  baseline;
- `agent/codegen-hypotheses` requires a CI ownership alias before its path-scope check
  can pass; the alias request is delegated to the CI-owned lane;
- no hypothesis in this document is marked `PASS` until an implementation produces
  the measured record. The document defines the experiment, not its result.
