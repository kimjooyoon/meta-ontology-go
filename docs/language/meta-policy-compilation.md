# Meaning-preserving compilation of a Gooo meta-policy

`examples/meta-policy-compilation/policy.gooo` owns the policy, state,
transition, case, evidence, and resolution nodes for the eight obligations.
The compiler lowers these typed nodes through the repository semantic IR and
checks only their fixed safety envelope. It does not maintain a second Go list
of policy meaning. The old `computes` marker program remains accepted for
existing sources, but is not used by the canonical policy.

The producer uses the public `gooo` CLI path (`check --semantic` and
`generate`) and also compiles the raw file for its receipt. The independent
consumer reads and parses the same raw file separately, derives its own local
policy view, and has a CI-checked import boundary of `0` producer imports.

The canonical matrix is three source executions, three generated executions,
and three independent reconstructions. The valid SHA-256 contradiction is a
known refutation and is ordered before UNKNOWN conditions. Empty or malformed
evidence is lower-resolution UNKNOWN and retains `stage`, `step`, `reason`,
`unknown_class`, `next_operation`, and `blocked_by`. An unrecognized upper
decision is fail-closed with reason
`FEEDBACK_COVERAGE_DECISION_UNKNOWN`; it cannot be accepted as `FIXED_POINT`.

The receipt contains eight distinct predicate observations per case and a
two-transition append-only lifecycle for each predicate. Its synthetic case
evidence is kept apart from a `CURRENT_EVIDENCE` runner observation. The CI
producer also uploads a fixed-metrics artifact with exact AST/IR node counts,
binding counts, marker before/after counts, case classifications, UNKNOWN
six-field preservation, generated artifact count, repository writes, local
test executions, and cross-project gates. Marker improvement is explicitly
`UNKNOWN` because the before/after source forms are not the same condition.
The write-set claim compares exact sorted file snapshots (path, mode, size,
and content digest) at the start and end of the producer run; it claims only a
net repository change of zero, not that no system call wrote a file.

## Typed policy decision proposals (internal compiler API)

The compiler exposes `ProposePolicyDecisionRevision` for an explicit,
source-bound revision of a first-class Gooo policy. The request names the exact
source digest, condition, previous decision and proposed decision. It does not
choose which decision would be useful or declare the revision an improvement.

The API parses and validates the source, clones its typed policy AST, and changes
exactly one transition target and the matching case resolution decision. The
existing Gooo formatter produces a candidate source string, which is compiled
again. Both compiled contracts are returned as `Original` and
`Candidate`, preserving their source/semantic identities and source-owned
rule metadata. `ChangedCoordinates` records `transition.to` and
`case.resolution.decision` for the explicitly requested condition.

The candidate is a **canonical semantic projection**, not a byte-preserving
source patch. Formatting, comments and original source layout are not promised
to survive in the new string. The original input bytes are never modified.
Baseline canonicalization must preserve the compiled semantic digest before a
proposal is emitted. Two changed semantic coordinates do not mean two changed
text lines.

Missing or stale identities, unknown conditions/decisions, no-op revisions,
ambiguous bindings and invalid candidates return an error without a partial
proposal. UNKNOWN context is neither synthesized nor deleted: a decision-only
change that requires other resolution fields must be rejected. Opaque legacy
activity programs are not rewritten by guessing at embedded strings.

The API remains internal; the public generation profile below supplies an
explicit CLI and proposal-report boundary, not an approval or conformance
receipt. Successful compilation proves only the bounded structural contract.
It does not run generated code, write a repository, establish external
utility, authorize a weaker policy, or prove independent conformance. Both
stricter and looser explicit proposals require separate acceptance against their
pinned original policy; the candidate must not certify its own promotion.

The existing source-authority test helper now uses this API instead of a
line-oriented string editor. CI checks exact compiled-coordinate changes,
source immutability, deterministic replay, an explicit reverse revision and
rejection cases. A separate native test builds one generated candidate program
and observes its changed decision. Its evidence values are synthetic test inputs,
not verified external claims, and the proposal API remains nonexecuting.


## Public policy revision profile

`gooo generate` accepts the opt-in `meta-policy-revision-v1` profile.
It exposes the typed proposal operation without requiring caller-written Go.
The caller supplies an exact input-byte digest, condition, previous decision and
proposed decision; the compiler does not infer missing values or choose a policy.

```sh
gooo generate ./project/policy.gooo \
  --profile meta-policy-revision-v1 \
  --profile-package metapolicycompilation \
  --profile-namespace metapolicycompilation \
  --profile-project-root ./project \
  --profile-source-digest "$SOURCE_DIGEST" \
  --profile-condition SEMANTIC_EQUIVALENCE \
  --profile-from-decision PASS \
  --profile-to-decision FAIL_CLOSED \
  --out ./proposal-output --json
```

Here `policy.gooo` must be inside the explicitly declared project boundary,
and `SOURCE_DIGEST` must be its exact `sha256:`-prefixed byte digest.
The existing compilation profile's source-bound manifest provides that digest;
neither a filename nor a semantic digest substitutes for it. The example paths
must be adjusted so the output is outside the input project.

The output directory must be empty and external to the project. A successful
invocation writes exactly `candidate.gooo` and `proposal.json`, using the
existing guarded atomic writer. `--json` emits the same proposal report bytes.
Invalid revision requests are rejected before output preparation; nonempty,
overlapping or disallowed output paths are not an overwrite/repair instruction.

The `gooo/meta-policy-decision-proposal/v1` report binds the requested decision
change, both compiled policies and their source/semantic identities, two changed
semantic coordinate kinds, and the two artifact names. The candidate is a
canonical semantic projection, not a byte-preserving patch. The original source
remains unchanged. Execution is false, current conformance is UNKNOWN, and
repository writes, mutation authority and promotion authority are zero.

The emitted candidate can be checked with `gooo check`, then supplied to the
unchanged `meta-policy-compilation-v3` profile with the proposal directory as
its project boundary and a different empty external generation directory.
That profile still produces its original four artifacts. Applying a revision,
executing generated code and deciding whether a policy change is acceptable are
separate operations, not effects of proposal generation.

CI exercises one public route from original Gooo through deterministic proposal
replay, candidate checking, ordinary public generation and generated decision
observation. It also exercises eight invalid request/boundary cases, duplicate
options and occupied output. The parent builds one Gooo CLI and one candidate
judge in temporary directories; these are CI-only test costs, not a language
runtime benchmark. Synthetic digest declarations remain unverified external
claims, even when the generated decision matches the source.

## Source-owned native revision roles

The proposal primitive now consumes an embedded Gooo operation contract before
it compiles or transforms the policy source. The contract declares three
opaque entities and one activity:

```gooo
entity PolicySource id "gooo://meta-policy-revision/source"
entity PolicyDecisionRevision id "gooo://meta-policy-revision/request"
entity PolicyDecisionProposal id "gooo://meta-policy-revision/proposal"
activity ProposePolicyDecisionRevision(PolicySource, PolicyDecisionRevision) -> PolicyDecisionProposal
```

The compiler checks the exact native signature, lowers this declaration and
requires both Used input facts and the WasGeneratedBy proposal fact. The
request/proposal entity names are bound to the compiled Go API types; their
opaque payload representation and the transformation implementation remain a
native foundation, not Gooo-generated function bodies. Extra declarations,
runtime bindings, entity fields, or an unadmitted value program cannot silently
widen this role contract. Stable entity IDs come from Gooo rather than a second
native ID table.

A successful proposal retains the binding in `operation_binding`: exact
contract source/semantic digests, activity/entity identities and three observed
relation facts. It also carries `request_digest`, covering the canonical typed
request fields `expected_source_digest`, `condition`, `from_decision` and
`to_decision`. The original policy digest, this request digest and the candidate
digest therefore describe separate inputs and output; none substitutes for an
execution observation.

The public revision report exposes these bindings without adding an artifact.
The revision profile still writes two files, and the ordinary compilation
profile still writes four. Input source remains immutable and the existing
non-executing, UNKNOWN-conformance, zero-authority boundaries remain intact.

CI adds three accepted role/request cases and twelve rejected contract cases,
and checks the new bindings along the existing public generation/execution
route without adding another CLI or judge build. These are declared test
requirements until that candidate's native CI reports their outcomes.

This binding does not add a sixth operation to the common selector. The
existing four legacy operations and separately bound syntax-registration
operation remain unchanged. Worker/verifier integration, common action
selection, policy adoption and external utility are separate, still-open work.


## Discovering the generated input contract

The generated executable can describe the input accepted by its
`--declared-input` mode without requiring an input case:

```sh
# Query in caller-owned CI/runner output after generation.
go run "$OUT/judge.go" --input-schema > "$OUT/input-schema.json"
```

The `gooo/generated-policy-input-schema/v1` document contains the source and
semantic identities, the target evaluation mode, the fields and a complete
`default_input` object. Each field carries its exact `json_field`, `go_type`,
`default`, `required` and `nullable` values. Names, types and zero values come
from the generated input type; there is no separately maintained schema field
list. The current eight scalar fields are optional and nullable because the
runtime decoder defaults missing and null values to their zero values.

An agent can obtain field names and a well-typed starting object from the
executable instead of guessing them from prose. This does not supply evidence:
executing the returned default input against the canonical policy yields
UNKNOWN, not PASS. Selecting real values still requires the caller's actual
observations. The ABI is compiler-owned; Gooo supplies the policy reductions.
This mode does not claim that arbitrary Gooo declarations define a new ABI.

Schema discovery is handled before stdin is read and performs no policy
evaluation. `policy_evaluation_observed` is false, and mutation and promotion
authority remain zero. Embedded identities are a self-description, not a
signature or CI attestation. The schema output is produced only by this
separate caller invocation; the public generation profile still emits exactly
four artifacts and does not execute the generated program.

CI queries the same already-built generated binary with absent and invalid
stdin, checks the fields against the typed ABI, and evaluates the returned
default input. Conflicting mode arguments fail rather than silently selecting
a different mode. These are interface conformance checks, not external utility
evidence or a language-completeness percentage.

## Declared inputs in the generated executable

The generated standalone judge accepts an explicit `--declared-input` mode:

```sh
# Execute in caller-owned CI/runner output, not during gooo generate.
go run "$OUT/judge.go" --declared-input < "$CASE_JSON"
```

No handwritten Go wrapper or private witness is needed for this mode. Gooo
still supplies the reduction rows. Generation still emits four artifacts and
does not execute them; the caller separately chooses to run the generated
program. This is a generated-executable mode, not a new Gooo CLI subcommand.

The `gooo/generated-policy-declared-input/v1` envelope records the generated
decision, matched rule and UNKNOWN context, source and semantic identities,
raw-input and effective-input digests, and per-field supplied/effective values.
Bindings are derived from the generated program's actual input type and JSON
tags, rather than another field registry. Missing, null and explicit false
remain distinguishable even when their effective value and decision agree.

The emitted input declaration is itself rendered from the compiler's typed
runtime-input ABI. Field names, types, order and JSON tags are not repeated as
a second hand-maintained list in the code template. CI parses the emitted Go
declaration and exercises renamed, reordered and retyped renderer fixtures.
These fixtures check schema derivation, not arbitrary extensibility of the
policy's runtime input domain.

The generated execution input has eight fields. It is not the eleven-field
compiler Case: validator expectation, evidence classification and provenance
metadata are not generated-judge input fields. Exact JSON names and a single
object are required in declared mode; aliases, duplicate keys, unknown fields,
wrong types and trailing documents do not produce a report.

An emitted report describes conditional generated-policy evaluation. External
evidence remains UNKNOWN/not verified and full conformance remains UNKNOWN/not
executed. The program reports that it ran, but this is
`SELF_REPORTED_NOT_ATTESTED`, not authenticated CI execution evidence.
Mutation and promotion authority are both zero. A conditional PASS does not
verify the truth of supplied availability or digest claims.

The no-argument mode preserves the existing decision-only output schema.
Declared-mode replay equality covers this report, not external utility or
permission to mutate a repository. CI exercises the input corpus using one
compiled generated binary rather than rebuilding it for each document.

## Declared-case evaluation boundary (internal compiler API)

`policycompilation.EvaluateDeclaredCase` compiles caller-supplied Gooo source
with an explicit package and namespace, then evaluates an explicit JSON Case.
This is an internal compiler API, not a new public CLI command.

The `gooo/meta-policy-declared-case-evaluation/v1` result preserves the existing
source decision, including its matched rule and UNKNOWN context. It separately
records source and semantic digests, the original input-byte digest, and the
effective typed-input digest. Equal effective-input digests do not prove equal
field presence or the truth of a caller's claims.

Field bindings are derived from the existing Go Case type and its JSON tags,
not a second hand-maintained field list. Each records its supplied and effective
value as `DECLARED`, `MISSING_DEFAULTED`, or `NULL_DEFAULTED`. Thus an omitted
boolean, explicit null, and explicit false can retain the same legacy decision
while remaining distinguishable. The JSON `omitempty` tag is not interpreted
as an optional-field or evidence rule.

This API requires a JSON object and exact field spellings. Case-folded aliases
are rejected rather than attributed to a different or missing field. Existing
Case decoding and the default generated-judge behavior are unchanged.

Availability, provenance, evidence classifications, and observed digests remain
caller declarations. The result is conditional source-policy evaluation:
external evidence is UNKNOWN/not verified, full conformance is UNKNOWN/not
executed, generated execution is false, and mutation/promotion authority is zero.
It does not replace the existing source/generated/independent conformance receipt.
The public command admission and end-user route remain separate unfinished work.

The declared-case regression uses a bounded source intervention: exactly one
`SEMANTIC_EQUIVALENCE` transition and its matching case decision change from
`PASS` to `FAIL_CLOSED`. Missing or ambiguous target bindings reject the candidate;
only those two source lines may change. Native CI checks both the explicit decision
and source/semantic identity changes. This fixture transformation is not a public
repair API, external evidence, or permission to weaken a failing acceptance rule.

## Public generation profile

Callers that need a portable compilation boundary can use
`gooo generate --profile meta-policy-compilation-v3` with explicit
`--profile-package`, `--profile-namespace`, and `--profile-project-root`
arguments. The output directory must be an empty caller-owned location outside
the project; the profile writes `policy.json`, `artifact.json`, `judge.go`,
and `generation-manifest.json` atomically. The manifest binds the actual byte
digests, records `execution_observed=false`, and reports conformance as
`UNKNOWN`; it is generation metadata, not execution or promotion evidence.

The private witness consumes those public artifacts and executes the generated
judge independently. Its existing six-file runner-temp write set remains the
execution/conformance boundary, while the public profile directory is carried
as a separate caller-owned artifact boundary.
