# Source-bound syntax registration candidate

This is the candidate-generation backend for issue #742, not admission of a
fifth operation into the existing automatic meta-operation planner.

The embedded Gooo contract declares an execution-identity pinning activity,
RegisterSyntaxCapability and nine artifact activities. Native lowering binds their Used/WasGeneratedBy relations and carries
the source digest, semantic digest, activity IDs, exact input digest and explicit
typed request into a complete candidate. No request parameter is hidden in a
subject name or inferred from prose.

## Exact first operation

The input source already exists in the caller snapshot. A draft request declares a
VALID LANGUAGE_CAPABILITY case, its source path, entity-fields or implicit-port
support, proof choice and current denominator version. A compact `--intent`
request can instead ask the fixed Gooo `ComposeCase` activity to produce those
case values from an ID, path, proof choice and profile selector. The caller may
omit `base_version`: repository observation selects it only after checking the
contiguous v22–current denominator history and matching the latest syntax target
to the corpus. Inspection adds the
input/source digests, Go runtime version and execution_identity to the pinned
request. The identity includes SHA-256 digests of
the running generator, the selected Go driver and its compiler, plus Go version,
OS and architecture. The source must satisfy the existing five syntax/BX obligations.
A missing input is UNKNOWN/DIRECT_MISSING; changed content or runtime identity is
UNKNOWN/STALE. Both retain stage, step, reason, unknown_class, next_operation and
blocked_by. Duplicate case IDs or paths are REFUTED.

Generation coordinates exactly nine semantic artifact roles: corpus JSON, native registry,
syntax counters, syntax conformance expectations, denominator admission,
denominator selection, digest pin, one new denominator JSON, and migration tests.
The canonical ordinary registration writes exactly nine physical member files.
A canonical meta-source promotion writes ten: its case literal and MetaSources
list have distinct source owners. The nine semantic artifact roles are unchanged.
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

### Gooo-owned append record

The Gooo `ComposeRegistrationAppend` activity receives an observed case ID,
source path and digest, the exact prior denominator digest and version, and
independently counted syntax targets from the denominator and corpus. It emits
a typed append record with the next version and syntax target. The candidate
exposes this record and the activity, source and projection digests so an agent
can see which source decision produced the proposed boundary.

The syntax counter projection takes its next capability target from the record
and checks it against the independently counted corpus. The native denominator
reader checks the prior JSON bytes and schema, repeats the target arithmetic,
and preserves the other boundaries and link obligations.
The source replay and candidate reconstruction remain separate checks. Changed
Gooo append logic, a contradictory corpus count and changed historical bytes
are rejected. The record is a proposal; it does not grant write or merge
authority.

This step transfers the append decision into Gooo source. It adds no historical
rewrite and does not reduce the nine semantic roles or nine physical members of
an ordinary registration. Consolidating the registry and counter *files* still
requires a separate source-backed change and measured before/after cost.

### Append-only history projection

The Gooo `GenerateDenominatorAdmission` and `GenerateDenominatorSelection`
activities declare `syntax.history.append:v1` and
`syntax.selection.append:v1`. Their native projections read every existing
version-to-syntax-target and capability-to-denominator case before proposing a
new one. An exact historical version is unchanged; a contradictory target or
an insertion into an older version gap is rejected. For the bounded helper
form, the history projection edits only the terminal helper. When it reaches
24 version cases, it links one new helper and leaves all earlier cases in
place. Each of the two active-denominator selectors similarly appends to its
last 16-case chunk. A full chunk gains one new chunk and a new wrapper call;
earlier selection cases remain byte-identical. Legacy inline switches get one
normalization into bounded helpers first.

This limits history-source churn; it does not reduce the nine required artifact
roles or waive corpus, denominator, and conformance checks. In the synthetic
five-owner history fixture, a one-version append changes one Go source unit;
the old full normalization changed all five. The current repository keeps its
history helpers in one physical file, so this is a smaller edit within that
file rather than a reduction in the nine-file registration proposal. No CI
runtime improvement is inferred from the source edit count.

## Caller path

An agent can inspect a source before choosing the intent's parser profile:

~~~sh
go run ./cmd/syntax-registration-candidate -root PROJECT -profile-hint examples/my-new-capability/main.gooo
~~~

This read-only mode pins the source bytes, replays them under the ordinary,
implicit-port and exact V1–V6 profiles, and reports every observed decision.
The Gooo `SuggestRegistrationProfile` activity chooses one passing profile in
that order. The result includes the source digest and activity/projection
origins. Its `ADVISORY_ONLY` authority never fills in `CaseIntent.profile` or
changes registration: the agent must still supply that field, and generation
independently replays the selected profile. An unreadable source reports
`UNKNOWN`; a source rejected by all profiles reports `NO_MATCH`. Neither
result writes files or grants admission.

Use the implemented experimental entry point:

~~~sh
go run ./cmd/syntax-registration-candidate -root PROJECT -intent case-intent.json -output NEW_EXTERNAL_DIRECTORY
~~~

The direct form composes the case in Gooo, observes and pins the source and
execution identity, then compiles, rechecks and exports the proposal in one
process. No intermediate request file is needed. For inspection before
generation, the same intent supports the two-step form:

~~~sh
go run ./cmd/syntax-registration-candidate -root PROJECT -intent case-intent.json -inspect > inspection.json
jq '.pinned_request' inspection.json > pinned-request.json
go run ./cmd/syntax-registration-candidate -root PROJECT -request pinned-request.json -output NEW_EXTERNAL_DIRECTORY
~~~

Compact case intent example:

~~~json
{
  "id": "my-new-capability",
  "path": "examples/my-new-capability/main.gooo",
  "proof_choice": "COHERENCE",
  "profile": "gooo.entityfields.go-projection.v6"
}
~~~

`ComposeCase` supplies `VALID`, `PASS`, `LANGUAGE_CAPABILITY` and
`replay-language-syntax`, and selects one parser profile. The resulting
`pinned_request` retains the exact case intent and Gooo contract digest.
Compilation re-evaluates the same activity and refuses changed case values,
version, promotion flag or source identity. The host independently checks case
shape and source replay; composition does not admit the candidate. Both forms
export only `PROPOSAL_ONLY` / `UNASSESSED` artifacts outside the input project.

Selectors are `ordinary`, `text-fields`, `implicit-ports`,
`ordered-collections`, or an exact
`gooo.entityfields.go-projection.v1` through `.v6` profile ID. Use the exact
profile supported by the source; the compact intent does not infer one from
the file even when a profile hint was requested. An explicit `base_version`
remains available and must equal the
observed current version. Missing, altered or contradictory history refuses
baseline selection; a changed input after inspection fails the later snapshot
recheck. Baseline observation selects repository state, while Gooo source owns
the case profile decision.

The full draft request remains available when the caller already has exact
case fields:

~~~sh
go run ./cmd/syntax-registration-candidate -root PROJECT -request draft-request.json -inspect > inspection.json
jq '.pinned_request' inspection.json > pinned-request.json
go run ./cmd/syntax-registration-candidate -root PROJECT -request pinned-request.json -output NEW_EXTERNAL_DIRECTORY
~~~

The first command reports observed snapshot/source digests, Go runtime version
and execution_identity, and includes `pinned_request` with those values filled
into the original case and version. This avoids copying six identity fields by
hand. The inspection is still `UNASSESSED`: generation reobserves the source,
snapshot and executable identities and checks the case before proposing files.
If any input changes between commands, the pinned request becomes stale.
The Go driver is observed with GOTOOLCHAIN=local and GOWORK=off, as in native
evaluation. Observation invokes only go env and reads executable bytes; it does
not build, test, install, repair or grant authority. No absolute tool path is
serialized into the identity. The second writes
the complete member files, candidate.json and execution.json outside the input project.
Existing output directories and parents resolving inside the input are rejected.
It does not apply the candidate to the project.

Draft request example; select the latest existing denominator version in the
input snapshot (v99 at this revision). `--inspect` fills the observation fields
before generation:

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
  "base_version": 99
}
~~~

## Internal meta source registration

### Canonical capability profiles

The request selects one canonical profile. `ordered_collections: true` already
includes record fields; combining it with `entity_fields: true` is rejected by
the existing case contract before any candidate member is written.

| Required surface | Case flags |
| --- | --- |
| Ordinary body | Neither flag |
| V1 required single Text record fields | `entity_fields: true` |
| V5 Text/Boolean/Integer, optional fields or ordered collections | `ordered_collections: true` only |
| V6 Text fields with `max_utf8_bytes`, including ordered collections | `entity_fields_profile: "gooo.entityfields.go-projection.v6"` only |

The explicit profile replaces the legacy Boolean flags. For example, the
[scope-fragment program](../examples/scope-fragments/README.md) uses both a list
and bounded Text outputs, so its registration selects V6. Its first V5 request
was rejected with `REGISTRATION_SOURCE_CONFORMANCE_FAILED`; selecting the source's
actual V6 contract allowed generation. Runtime execution had already succeeded.

The local-version reader's first request recorded this combination error. Its
corrected source-bound request generated the complete nine-member proposal.
Both attempts and their separate authority boundaries are retained in the
[observation record](observations/local-version-origins/observations.json).

The flag names are historical profile selectors. V1 covers required single Text
fields; an Integer field needs the V5 selector even if the program uses no
collections. The Text-origin example retained its original V1 rejection and
subsequent V5 request. Source fields were preserved throughout. Use the exact
profile table before pinning an execution request.

### Register an internal source

An explicit request may also name an existing `internal/meta/**/*.gooo`
source. This is the same VALID LANGUAGE_CAPABILITY registration operation,
not a meta-source exemption or a smaller denominator. Lexical traversal,
absolute paths, look-alike prefixes, other internal namespaces and non-Gooo
extensions remain rejected before input observation. This path predicate is
not a filesystem sandbox or write permission.

The source is read-only input. All nine generated artifact roles remain
mandatory, the next denominator version is appended, historical versions
remain unchanged, and candidate application/promotion authority stays false.
Source, snapshot, typed request and executable identities remain pinned.

The existing dedicated Actions workflow now exercises both the original
example source and a policy-revision-shaped internal meta source. Each uses
an isolated CI copy, exact candidate reconstruction, deterministic replay
and the unchanged syntax/vertical-slice conformance packages. Internal-source
evidence is kept under `internal-meta-source/` alongside the original
case's artifacts. Both retain the exact request and Gooo source in addition
to the candidate, native report and command output.

The internal case is a declared synthetic compiler regression, not external
adoption or proof of policy revision execution. Case-specific wall time is
an observation only; no speed or memory improvement is claimed without a
matched before/after pair. No extra local build, test, formatter or generator
execution is part of this workflow.

## Explicit promotion of an existing meta source

Set `"promote_meta_source": true` in the typed request to move exactly one
already registered meta source into the executed LANGUAGE_CAPABILITY corpus.
The request must name a source present exactly once in `meta_sources`, not
an existing executed case or package member. Omission or false preserves the
original unregistered-source behavior and cannot authorize this promotion.

The native operation removes that one membership from both corpus and native
registry while adding the exact requested case. The same nine artifact roles
also increase syntax counters, conformance expectations and the next
denominator version. Other cases, other meta sources, package units, source
bytes and denominator history remain unchanged. An incomplete or resealed
candidate is rejected by full reconstruction; matching file counts alone is
not acceptance.

The explicit flag is included in the canonical request digest. It is not
inferred from a path, title, score or failed check and does not grant apply,
merge, release or planner-admission authority.

The Actions-only `registered-meta-source/` case promotes the compiler's
existing `internal/meta/syntaxregistration/contract.gooo` without overwriting
that source even in the temporary snapshot. It retains the exact request,
source, complete generated candidate and unchanged conformance output.
The original example and previously unregistered internal-source cases remain
separate. These are synthetic compiler regressions, not external utility.

## Native evidence and remaining integration

The dedicated Actions job retains the canonical nine-file ordinary registration
and ten-file meta-source promotion end-to-end cases.
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

The dedicated candidate witness does not assess common-planner or dispatch
admission. Its `global_planner_admission` observation is `UNASSESSED`; it does
not claim those integrations are absent or certified. The separate native
common-execution checks own that evidence.
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
canonical/projected conformance obligations remain bound to the complete write set.
The original six acceptance obligations of issue #742 were closed in its
[bounded acceptance record](https://github.com/kimjooyoon/meta-ontology-go/issues/742#issuecomment-5563071729).
That closure covers the coordinated registration operation and its stated native
checks; broader language capability and user utility keep their own scope.
Neither candidate byte equality nor the backend tests close product utility or
the entire language's self-improvement goal.

## Registry helpers and declaration origins

The native `Cases` field may hold a direct case-slice literal or a local
zero-argument helper. A helper may return another such helper or assemble a
sequence using one local declaration, ordered `append(sequence, helper()...)`
assignments, and a final return. Each reachable helper and literal has one use.
The final literal owns the insertion; earlier groups and unrelated literals
retain their bytes. Inline literal punctuation and real token-file paths are
preserved, including source units carrying line directives.

Branches, cycles, repeated helper paths, external calls, arguments, named return
bindings and shadowed append/helper bindings remain unresolved. Generation
returns no candidate for those shapes. This is a finite source observation used
by the Gooo `GenerateNativeRegistry` activity. It does not evaluate arbitrary
Go or use a model to invent registry members.

The [recorded source cohorts](observations/registry-source-origins/README.md)
include the failed inline punctuation case and its changed-source successor.
Full native canonical/projected acceptance belongs to the submitted head's own
original Actions runs.

## Historical targets during continued generation

Generating the next corpus version pins the previous syntax target as an
explicit source case before the current capability count grows. Historical
migration checks compare their own immutable evidence, while current checks
follow the selected source version. This keeps a v72 observation at108 when
v73/v74 use109/110. Original source-case membership, the other five boundaries
and all thirteen link obligations are preserved.

[Recorded continuation](observations/refactor-type-origins/README.md) contains
actual compiled two-step decoder observations and the standard-type reader
lifetime used by Gooo function decomposition.
