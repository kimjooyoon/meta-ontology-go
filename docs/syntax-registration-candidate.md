# Source-bound syntax registration candidate

This is the candidate-generation backend for issue #742, not admission of a
fifth operation into the existing automatic meta-operation planner.

The embedded Gooo contract declares RegisterSyntaxCapability and nine artifact
activities. Native lowering binds their Used/WasGeneratedBy relations and carries
the source digest, semantic digest, activity IDs, exact input digest and explicit
typed request into a complete candidate. No request parameter is hidden in a
subject name or inferred from prose.

## Exact first operation

The input source already exists in the caller snapshot. The request declares a
VALID LANGUAGE_CAPABILITY case, its source path, entity-fields or implicit-port
support, proof choice, current denominator version, input/source digests and Go
runtime version. The source must satisfy the existing five syntax/BX obligations.
A missing input is UNKNOWN/DIRECT_MISSING; changed content or runtime identity is
UNKNOWN/STALE. Both retain stage, step, reason, unknown_class, next_operation and
blocked_by. Duplicate case IDs or paths are REFUTED.

Generation coordinates exactly nine members: corpus JSON, native registry,
syntax counters, syntax conformance expectations, denominator admission,
denominator selection, digest pin, one new denominator JSON, and migration tests.
Go changes are bounded AST-position edits. Existing denominator files are
immutable inputs; only the new version is emitted. Current obligations increase
by one and unrelated boundary/link targets are preserved.

The operation emits PROPOSAL_ONLY / UNASSESSED, never apply or promotion
permission. Missing, duplicated, altered or historical-path candidate members
fail candidate validation. Replay equality alone is not semantic acceptance.

## Caller path

Use the implemented experimental entry point:

~~~sh
go run ./cmd/syntax-registration-candidate -root PROJECT -request request.json -inspect
go run ./cmd/syntax-registration-candidate -root PROJECT -request request.json -output NEW_EXTERNAL_DIRECTORY
~~~

The first command reports observed snapshot/source digests and Go runtime
version. Pin them explicitly in the request before generation. The second writes
nine member files, candidate.json and execution.json outside the input project.
Existing output directories and parents resolving inside the input are rejected.
It does not apply the candidate to the project.

Request example, with observed values filled by the caller:

~~~json
{
  "case": {
    "id": "my-new-capability",
    "path": "examples/my-new-capability/main.gooo",
    "kind": "VALID",
    "expected_decision": "PASS",
    "proof_choice": "COHERENCE",
    "meta_operation": "replay-language-syntax",
    "scope": "LANGUAGE_CAPABILITY",
    "entity_fields": true
  },
  "base_version": 30,
  "snapshot_digest": "sha256:<observed-snapshot>",
  "source_digest": "sha256:<observed-source>",
  "toolchain": "go1.27.0"
}
~~~

## Native evidence and remaining integration

The dedicated Actions job runs the real generator twice, applies all nine
members only to a CI temporary repository, and runs the existing syntax and
vertical-slice conformance tests. It reports the actual emitted/required members,
replay comparison, native command result and elapsed milliseconds. Zero manual
follow-up edits is reported only after those unchanged conformance commands pass.
No local test, build or formatter execution is part of this development workflow.

The existing automatic generation.Action input schema, DefaultRegistry,
meta-execution dispatch and shared operation receipts are NOT connected yet.
Go runtime version binding is not an executable/toolchain-binary digest proof.
Issue #742 remains open until those input/identity/executor integrations and all
six original acceptance obligations are supported by exact-head native evidence.
Neither candidate byte equality nor the backend tests close product utility or
the entire language's self-improvement goal.
