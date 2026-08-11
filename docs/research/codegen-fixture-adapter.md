# Codegen fixture-runner and adapter contract

This note defines the smallest executable boundary for running code-generation
fixtures across independent implementations. It is a follow-up to the hypothesis
and fixture catalog, but it is intentionally self-contained enough for an AST, IR,
BX, generator, LSP, cache, provenance, or CI lane to implement one adapter without
sharing private Go types.

The runner is an oracle boundary, not a compiler. It evaluates whether an observed
result matches an immutable fixture expectation. An adapter may report `DEFERRED`
when a stage is unavailable, but it must not turn that state into `PASS`. The Go-hosted
verifier remains the current trust root; a future `.gooo` adapter is a comparison
producer until independent promotion gates succeed.

## Boundary and responsibilities

```text
fixture JSON ──> runner ──request──> adapter
     ▲              │                  │
     │              └─oracle/comparison┘──> canonical response
     │                                      │
     └──────────── immutable golden <──────┘
```

| Component | Owns | Must not own |
| --- | --- | --- |
| Fixture | authoritative input, expected status, allowed changes | implementation-specific normalization |
| Runner | canonical serialization, repetition, oracle comparison, exit status | business semantics or artifact repair |
| Adapter | AST/IR/BX/codegen/LSP/cache operation | changing fixture expectations or CI policy |
| Evidence bridge | stable facts and producer identity | certifying its own promotion |
| CI gate | branch scope, evidence comparison, promotion decision | rewriting fixtures or generated output |

The runner never patches a previous Go artifact. It passes the previous bytes to the
adapter and records whether the adapter rejected, preserved, or replaced them. A
failure must leave the previous artifact and golden untouched.

## Versioned wire protocol

The runner and adapter communicate through one UTF-8 JSON request and one UTF-8 JSON
response per invocation. A CLI may use stdin/stdout, files, or an in-process bridge,
but the canonical request and response must be identical.

### Request

```json
{
  "schema": "gooo/codegen-adapter/v1",
  "fixture": "pay-order-v1",
  "operation": "generate",
  "input": {
    "dsl": "package billing\nnamespace billing\n...\n",
    "ir": null,
    "previous_go": null,
    "source_uri": "fixtures/codegen/pay-order.gooo"
  },
  "contract": {
    "ast": "gooo/ast/v1",
    "ir": "gooo/ir/v1",
    "generator": "gooo/generator/v1",
    "marker": "gooo/marker/v1",
    "policy_sha256": "sha256:<digest>"
  },
  "options": {
    "canonical_output": true,
    "allow_migration": false
  },
  "expected": {
    "status": "pass",
    "failure_code": null
  }
}
```

Rules:

- `fixture` and `operation` are required stable names; unknown values are runner
  errors, not adapter-defined extensions.
- Exactly one of `input.dsl` and `input.ir` is authoritative for an operation. If a
  previous artifact is needed, `previous_go` is explicit and may be empty.
- Embedded source uses UTF-8 bytes. An implementation that reads from a filesystem
  must verify the supplied `source_uri` digest and must not use the host path as
  canonical identity.
- Contract versions are compared before invoking the adapter. An unknown version
  yields `DEFERRED` only when the fixture declares compatibility as optional;
  otherwise it yields `FAIL` with `ADAPTER/UNSUPPORTED_CONTRACT`.
- `allow_migration` is false by default. A marker or slot migration requires an
  explicit fixture migration record and an evidence fact naming the old and new IDs.
- `expected` belongs to the runner oracle. The adapter must not read it to change
  generation behavior; it is compared only after the observed response is normalized.

### Response

```json
{
  "schema": "gooo/codegen-adapter/v1",
  "fixture": "pay-order-v1",
  "operation": "generate",
  "status": "pass",
  "failure": null,
  "observed": {
    "semantic_digest": "sha256:<digest>",
    "source_digest": "sha256:<digest>",
    "region_digest": "sha256:<digest>",
    "import_digest": "sha256:<digest>",
    "source_map_digest": "sha256:<digest>",
    "cache_key": "sha256:<digest>",
    "regions": [],
    "slots": [],
    "imports": [],
    "source_map": [],
    "delta": null
  },
  "measurements": {
    "source_span_count": 0,
    "protected_bytes_equal": true,
    "unrelated_region_count": 0,
    "no_write": true
  },
  "evidence": {
    "producer": "go",
    "stage": 0,
    "facts": []
  }
}
```

`observed` is normalized before comparison. The runner, not the adapter, sorts set-like
arrays, rejects duplicate IDs, computes canonical digests, and checks that all ranges
are within the supplied bytes. The response itself ends with exactly one newline;
diagnostics may go to stderr and are not part of the canonical response digest.

## Operation contract

The initial operation set is deliberately small:

| Operation | Authoritative input | Required observed output | Negative oracle |
| --- | --- | --- | --- |
| `parse-ast` | DSL bytes | declarations, IDs, spans, diagnostics | invalid span/duplicate ID |
| `lower-ir` | AST or DSL | normalized facts and semantic digest | namespace/identity collision |
| `generate` | IR plus optional previous Go | source, regions, slots, imports, source map | malformed/tampered previous artifact |
| `lift-bx` | Go plus source map | sourced delta, candidates, conflicts, locality IDs | deterministic fact without source |
| `resolve-lsp` | source map plus query | semantic ID and bidirectional range | stale or out-of-bounds mapping |
| `cache-key` | canonical input tuple | key and component digests | temp-path/tool-order key drift |
| `emit-evidence` | observed result | producer-independent stable facts | missing/deferred fact reported as pass |
| `compare-evidence` | two evidence artifacts | equal/mismatch plus rule ID | producer label changes comparison |

An adapter may implement fewer operations, but the response must be `DEFERRED` with a
missing capability. Returning an empty successful object is invalid because it hides
partial implementation.

## Canonical observed data

### AST snapshot

The AST adapter returns declaration records with:

```text
kind, display_name, semantic_id, namespace, source_uri,
start_offset, end_offset, start_line, start_column, end_line, end_column
```

Offsets are bytes into the exact supplied source. A file-level diagnostic may have no
span; a declaration-level result may not. Diagnostics are sorted by offset, rule ID,
and message code. Display names are never used as identity.

### Semantic IR snapshot

Facts are represented as:

```text
subject_id, predicate, object_id,
subject_kind, object_kind, fact_class,
source_span, candidate_reason
```

`fact_class` is one of `authoritative`, `deterministic`, `candidate`, or `rejected`.
The normalized semantic digest includes IDs, predicates, kinds, and the declared
source relation; it excludes display-only whitespace and host paths.

### Generator projection snapshot

The generator result is not only source bytes:

```text
generator_contract, implementation_digest, package
regions: kind, semantic_id, owner_id, start_offset, end_offset, body_digest
slots: slot_id, owner_id, start_offset, end_offset, body_digest
imports: path, alias, used_by_ids
source_map: semantic_id, kind, source_range, generated_range
source_digest, semantic_digest
```

A region ID or slot owner change is a semantic change even if the output still parses.
An output that parses but does not compile is `CODEGEN/COMPILE_FAILURE`. A previous
artifact with malformed markers is rejected before any replacement is written.

### BX delta snapshot

`lift-bx` returns:

```text
added_facts, removed_facts, candidate_facts,
locality_ids, conflicts, source_spans
```

Every deterministic added or removed fact must have a source span. A generic helper
call without a registered semantic ID remains implementation detail. A candidate fact
is retained as a candidate and does not alter the authoritative semantic digest.

### LSP/source-map snapshot

For each region and slot, both directions are required:

```text
semantic ID -> generated range -> source/IR range
source/IR range -> semantic ID -> generated range
```

The runner checks offsets against the final generated source, not a pre-format buffer.
Missing or stale mappings return `LSP/STALE_MAPPING`; they cannot silently resolve to a
neighboring region.

### Cache snapshot

The cache adapter exposes the canonical key inputs and key:

```text
cache_schema
ast_contract
ir_contract
generator_contract
generator_implementation_digest
normalized_ir_digest
previous_artifact_digest
policy_digest
toolchain_contract
```

Absolute roots, timestamps, environment variable order, and producer labels are not
key inputs. A display-name-only change may share a semantic key but must use a distinct
text artifact key if source bytes change.

## Fixture runner behavior

### Deterministic execution

For each fixture, the runner performs these steps:

1. Validate the fixture schema, operation, expected status, and stable IDs.
2. Compute the request digest from canonical bytes.
3. Invoke the adapter in a clean process or isolated in-process context.
4. Normalize observed arrays, ranges, diagnostics, and evidence facts.
5. Compare observed output with the golden oracle and expected negative behavior.
6. Repeat the request when the fixture declares a repeat count.
7. Emit one canonical response and preserve all failure evidence.

The runner must not call a formatter, repair malformed markers, update a golden, or
rewrite a previous artifact while evaluating a case.

### Exit status

The JSON response carries the semantic status; the process exit code carries evaluation
status. This prevents an expected negative case from looking like a runner crash:

| Exit | Meaning | Example |
| --- | --- | --- |
| `0` | observed result matches the fixture expectation | expected `FAIL` was correctly rejected |
| `10` | unexpected result or false acceptance | malformed marker was accepted |
| `20` | adapter capability/stage is deferred | generator package is unavailable |
| `30` | fixture/runner/transport error | invalid fixture schema or duplicate golden |
| `40` | evidence or canonicalization error | nondeterministic response or range overflow |

CI may allow exit `20` only in an explicitly Stage 0/deferred job. It must never count
exit `20` as promotion evidence. Exit `0` for an expected negative case means the
oracle was satisfied, not that the negative input was accepted by the compiler.

### Repetition and equality measurements

The runner records:

```text
repeat_count
canonical_equal_count
source_equal_count
semantic_equal_count
region_equal_count
source_map_resolved_count
no_write_count
false_acceptance_count
environment_leak_count
```

Correctness fixtures require exact counts. For example, a 32-run deterministic test
passes only at `canonical_equal_count=32`; a 9-case marker mutation suite passes only
at `false_acceptance_count=0` and `no_write_count=9`.

Timing is optional provenance. It is never part of a golden digest and cannot convert
a correctness failure into a performance result.

## Minimal fixture examples

### Positive generation request

```json
{
  "schema": "gooo/codegen-adapter/v1",
  "fixture": "pay-order-v1",
  "operation": "generate",
  "input": {
    "dsl": "package billing\nnamespace billing\nentity Order id \"billing://entity/order\"\nentity Payment id \"billing://entity/payment\"\nactivity PayOrder(Order) -> Payment\n",
    "ir": null,
    "previous_go": null,
    "source_uri": "fixtures/codegen/pay-order.gooo"
  },
  "contract": {
    "ast": "gooo/ast/v1",
    "ir": "gooo/ir/v1",
    "generator": "gooo/generator/v1",
    "marker": "gooo/marker/v1",
    "policy_sha256": "sha256:<fixture-policy>"
  },
  "options": {"canonical_output": true, "allow_migration": false}
}
```

The positive golden requires one activity region, one implementation slot owned by
that activity, two entity IDs, one input port, one output port, valid Go formatting,
and reversible source-map ranges. Exact digests are generated by the independent
golden process, not hand-waved as constants in this contract.

### Expected-negative marker request

The input is the same as `pay-order-v1`, but `previous_go` contains:

```go
//gooo:generated:start id="billing://activity/other" kind="activity"
func PayOrder(order Order) Payment {
	//gooo:slot:start id="billing://activity/pay-order/implementation"
	return Payment{}
	//gooo:slot:end id="billing://activity/pay-order/implementation"
}
//gooo:generated:end id="billing://activity/other"
```

Expected response:

```json
{
  "status": "fail",
  "failure": {
    "code": "MARKER/OWNERSHIP_UNPROVEN",
    "semantic_id": "billing://activity/other",
    "no_write": true
  }
}
```

The fixture evaluation exits `0` because the expected negative was correctly
rejected. The adapter did not succeed; the observed compiler action is still a
fail-closed rejection.

### Deferred bootstrap request

```json
{
  "schema": "gooo/codegen-adapter/v1",
  "fixture": "bootstrap-stage-0-v1",
  "operation": "compare-evidence",
  "input": {
    "go_evidence": "examples/bootstrap/go-hosted-baseline.json",
    "gooo_evidence": "examples/bootstrap/gooo-hosted-proposed.json"
  },
  "expected": {"status": "deferred", "promotion_eligible": false}
}
```

If the candidate host is not implemented, the response is `DEFERRED` with exit `20`.
It is not a passing equality result and cannot promote the candidate verifier.

## Negative oracle catalog

The following cases must remain reusable across adapters:

| Fixture mutation | Required response | Failure code |
| --- | --- | --- |
| unknown protocol schema | no adapter invocation | `RUNNER/UNSUPPORTED_SCHEMA` |
| unknown operation | no adapter invocation | `RUNNER/UNSUPPORTED_OPERATION` |
| duplicate fixture ID or duplicate fact ID | no golden comparison | `RUNNER/DUPLICATE_ID` |
| missing source span on deterministic fact | reject or candidate-only | `BX/MISSING_SOURCE` |
| marker retagging or mismatched close | no write | `MARKER/OWNERSHIP_UNPROVEN` |
| import alias collision | no generated artifact | `CODEGEN/IMPORT_DRIFT` |
| unrelated region bytes changed | reject refresh | `LOCALITY/ESCAPE` |
| stale source-map range | no LSP answer | `LSP/STALE_MAPPING` |
| equal canonical input, different cache key | invalidate cache result | `CACHE/KEY_DRIFT` |
| producer label changes only | canonical payload remains equal | no failure |
| producer-independent fact changes | comparison rejects | `EVIDENCE/MISMATCH` |
| unknown branch ownership | CI rejects | `CI/UNKNOWN_SCOPE` |
| unavailable implementation stage | explicit deferred response | `STAGE/DEFERRED` |

Negative fixtures must assert both the primary code and the safety action. A message
that says “invalid” without `no_write`, candidate retention, or promotion state is not
enough evidence for a semantic compiler.

## Evidence bridge

The current verifier evidence model uses `gooo/evidence/v1`, a stage, fixture,
decision, and producer-independent sorted facts. The adapter maps observed results to
facts without changing their meaning:

```text
codegen/status              pass|fail|deferred
codegen/semantic-digest     sha256:<digest>
codegen/source-digest       sha256:<digest>
codegen/region-digest       sha256:<digest>
codegen/import-digest       sha256:<digest>
codegen/source-map-digest   sha256:<digest>
codegen/failure-code        stable rule ID or empty on pass
codegen/no-write             true|false
```

`EvidenceArtifact.Producer` may be `go` or `gooo`, while the canonical payload must
remain equal when the normalized result is equal. Producer identity and builder
metadata are attestation context, not semantic facts. A candidate may emit a mismatch
fact, but only the independent verifier decides whether a stage can be promoted.

## Adapter interface

An implementation can use native Go interfaces, subprocesses, or a future `.gooo`
adapter. The logical interface remains:

```text
Adapter.Describe() -> adapter name, contract versions, supported operations
Adapter.Execute(request) -> observed response or typed deferred/error result
Adapter.Normalize(response) -> canonical response fields
Adapter.Evidence(response) -> producer-tagged evidence artifact
```

The runner owns canonicalization and oracle comparison. `Normalize` therefore cannot
change the fixture expectation or suppress a conflict; it may only produce the
adapter's declared normalized representation. A typed error must preserve:

```text
code, operation, fixture, semantic_id, source_span, no_write, promotion_eligible
```

This makes the same negative case usable by a Go generator, a `.gooo` generator, a
standalone LSP server, and a CI-only evidence comparator.

## Current rollout contract

- Go 1.26.5, `gofmt`, `go vet`, `go test`, and race tests remain mandatory.
- The current baseline may report semantic CLI/generator operations as `DEFERRED`.
- Unknown branch ownership is a CI policy failure, not a fixture pass condition.
- The fixture adapter lane must not edit protected policy files to make its branch
  pass; its ownership alias is registered by the CI-owned lane.
- No generated artifact, golden, cache entry, or bootstrap evidence is overwritten by
  a failed experiment.
- A future implementation may add operations only with a new schema or a declared
  backward-compatible extension; it must preserve existing fixture IDs and failure
  codes.

The contract is complete when an independent runner can execute the positive,
expected-negative, and deferred examples above and produce deterministic responses
that the existing verifier can compare without importing generator internals.
