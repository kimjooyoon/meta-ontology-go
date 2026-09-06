# Source-bound syntax registration candidate

This is the candidate-generation backend for issue #742, not admission of a
fifth operation into the existing automatic meta-operation planner.

The embedded Gooo contract declares an execution-identity pinning activity,
RegisterSyntaxCapability and nine artifact activities. Native lowering binds their Used/WasGeneratedBy relations and carries
the source digest, semantic digest, activity IDs, exact input digest and explicit
typed request into a complete candidate. No request parameter is hidden in a
subject name or inferred from prose.

## Exact first operation

The input source already exists in the caller snapshot. The request declares a
VALID LANGUAGE_CAPABILITY case, its source path, entity-fields or implicit-port
support, proof choice, current denominator version, input/source digests and Go
runtime version and execution_identity. The identity includes SHA-256 digests of
the running generator, the selected Go driver and its compiler, plus Go version,
OS and architecture. The source must satisfy the existing five syntax/BX obligations.
A missing input is UNKNOWN/DIRECT_MISSING; changed content or runtime identity is
UNKNOWN/STALE. Both retain stage, step, reason, unknown_class, next_operation and
blocked_by. Duplicate case IDs or paths are REFUTED.

Generation coordinates exactly nine semantic artifact roles: corpus JSON, native registry,
syntax counters, syntax conformance expectations, denominator admission,
denominator selection, digest pin, one new denominator JSON, and migration tests.
The canonical acceptance case still requires exactly nine physical member files.
After repository projection, one role can own several physical files and a file
can participate in several roles. All nine activity/output bindings remain
mandatory; emitted/required physical members are the complete distinct write set,
not a substituted language-completeness score.

Native symbols are resolved across the declared package namespaces. Every Go
source unit in those namespaces is included in the input snapshot; missing,
duplicated or newly added units cannot silently inherit prior evidence.
AST token-file ownership binds edits to actual paths, ignoring misleading
line-directive aliases. `artifacts` carries nine lowered activity/output IDs and
their member paths; each member carries all generating `activity_ids`.
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

The first command reports observed snapshot/source digests, Go runtime version
and execution_identity. Pin all of them explicitly in the request before generation.
The Go driver is observed with GOTOOLCHAIN=local and GOWORK=off, as in native
evaluation. Observation invokes only go env and reads executable bytes; it does
not build, test, install, repair or grant authority. No absolute tool path is
serialized into the identity. The second writes
the complete member files, candidate.json and execution.json outside the input project.
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
  "toolchain": "go1.27.0",
  "execution_identity": {
    "go_version": "go1.27.0",
    "goos": "linux",
    "goarch": "amd64",
    "executable_sha256": "sha256:<observed-generator>",
    "go_command_sha256": "sha256:<observed-go-driver>",
    "compiler_sha256": "sha256:<observed-compiler>"
  }
}
~~~

## Native evidence and remaining integration

The dedicated Actions job retains the canonical nine-file end-to-end case.
The repository-projection job additionally uses its actual authorized transformed
workspace, copies it with source links dereferenced, generates the complete
candidate twice, applies every member only to that CI temporary copy, and runs
the existing syntax and vertical-slice conformance tests. It neither substitutes
a historical checkout for the projected case nor skips the original full-module
test. Arbitrary source-file relocation, ambiguous symbol owners, added source
units and forged source locations have separate regression cases. It reports the actual emitted/required members,
replay comparison, native command result and elapsed milliseconds. Zero manual
follow-up edits is reported only after those unchanged conformance commands pass.
No local test, build or formatter execution is part of this development workflow.

The existing automatic generation.Action input schema, DefaultRegistry,
meta-execution dispatch and shared operation receipts are NOT connected yet.
The syntax.register:v2 contract routes every generation activity through
PinnedRegistrationInput, produced by PinRegistrationExecutionIdentity. Compilation
and generation both reobserve and compare execution identity. Missing identity
is UNKNOWN/DIRECT_MISSING, changed bytes are UNKNOWN/STALE, and malformed digests
are REFUTED. All six UNKNOWN fields remain present. Candidate execution_binding
carries the lowered pin activity and input/output IDs together with the observed
identity. Candidate validation rejects a substituted identity.

This pins three executable identities, not the entire environment, toolchain
publisher authenticity or semantic correctness. The new identity requirement is
an explicit experimental request-ABI change: version-only requests are no longer
accepted. Native reports carry the actual binding; existing nine-role and
canonical/projected conformance obligations are unchanged.
Issue #742 remains open until those input/identity/executor integrations and all
six original acceptance obligations are supported by exact-head native evidence.
Neither candidate byte equality nor the backend tests close product utility or
the entire language's self-improvement goal.
