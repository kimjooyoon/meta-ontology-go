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

## Source-derived receipt reconstruction

A receipt digest and a valid claim-event hash chain establish internal byte
integrity, not that the recorded claims follow from the policy and evidence.
Receipt verification therefore reconstructs the existing receipt from the
compiled Gooo policy, canonical input cases, and the already checked generated
and independent decision observations. It compares the complete canonical
content, not only self-reported success flags and denominator totals.

The reconstruction reuses the existing source-driven receipt builder. It adds
no parallel rule registry, policy semantics, receipt schema, artifact, or
external dependency. Source-owned meta-operation/proof bindings, exact decision
and predicate counts, predicate outcomes, event provenance, initial chain
boundary and verification claims must all agree. UNKNOWN cannot be discharged,
and a refuted predicate cannot become discharged, merely by rewriting the
ledger and recomputing its hashes.

Caller case order remains immaterial; canonical receipt order is reconstructed
by case ID. The explicitly supplied current-evidence provenance description is
retained, not replaced with a compiler-owned string. Existing write-boundary
and public-CLI evidence checks still apply before reconstruction.

The native regression suite declares exactly three accepted reconstruction
cases and sixteen resealed counterexamples. Counterexamples retain valid
receipt digests and internal event chains, including count-preserving summary
changes and unsupported UNKNOWN/refutation discharge. Those are test
requirements, not a completion or utility score; their pass counts come from
the corresponding CI run.

This is source-derived consistency, not a second independent implementation or
fresh execution attestation. The unit fixtures are explicitly synthetic and do
not execute a generated judge or Gooo CLI. External utility, actual execution
and policy adoption still require their own observations. No policy candidate
is applied, and no mutation or promotion authority is added.


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


## Independent raw-source observation

After CI builds the consumer binary, the experimental read-only mode can
observe a policy without the three synthetic cases or producer artifacts:

~~~sh
meta-policy-compilation-consumer -observe-source \
  -policy examples/meta-policy-compilation/policy.gooo
~~~

The mode writes one `gooo/meta-policy-source-observation/v1` JSON document to
stdout. It reparses raw Gooo with the consumer's policy reader and preserves
source digest, semantic digest, expected package/namespace, declared rule
denominator and source-derived rule/meta-operation bindings. It does not
read producer JSON, bind placeholder case digests, execute a generated judge,
write a report file or mutate the source.

`-profile-package` and `-profile-namespace` retain their existing defaults
and may be supplied explicitly. Mixing this mode with `-cases`, `-artifact`,
`-manifest`, `-output`, unknown mode flags or positional arguments is rejected
before reading a policy. The existing three-case consumer path is unchanged.

Raw-source reconstruction is not policy conformance or admission.
`current_conformance` remains `UNKNOWN` with stage, step, reason,
unknown_class, next_operation and an explicit empty blocked_by frontier.
Policy execution and producer-artifact observation remain false; repository
writes, mutation authority and promotion authority remain zero.

The consumer shares the language parser and IR lowering with other language
tools. This is separate policy reconstruction, not an independently
implemented compiler. Its digests are source observations for a later
source-bound execution/producer comparison, not permission to adopt a
revision or relabel synthetic fixtures as current execution evidence.
Read, parse and output failures remain errors and must not become success
or silently supplied digests in a caller.

## Read-only revision receipt reconstruction

The consumer can inspect a revision execution report without executing its
generated program or adopting the candidate:

```sh
go run ./cmd/meta-policy-compilation-consumer \
  -policy policy.gooo \
  -revision-request request.json \
  -observe-revision-receipt revision-report.json
```

The three inputs are read-only. The observer writes JSON to stdout and accepts
neither an output directory nor mixed legacy/source-observation modes.
Each input is bounded to 16 MiB. Raw source, request and report bytes retain
separate digests. The canonical typed request digest does not replace the raw
request identity. Source filenames are parsing context, not filesystem
provenance or proof of where the producer read its bytes.

Six bounded checks connect the report back to the Gooo policy's source rules
and meta-operation bindings:

| Check | Reconstruction |
| --- | --- |
| SOURCE_BINDING | Consumer-parsed baseline/candidate identities, rules and semantic digests |
| REQUEST_BINDING | Original raw/canonical request and paired case identities |
| REVISION_SCOPE | Only the requested transition/resolution decision change |
| DECLARED_INPUT_PRESERVATION | Caller snapshots, order, classes and provenance, without digest repair |
| RESULT_RECONSTRUCTION | Every source/first/replay result field, including UNKNOWN cause/frontier |
| ACCOUNTING_RECONSTRUCTION | Counts, transitions, incomplete attempts, pending claims and zero authority |

`RECEIPT_CONSISTENT_ONLY` means these supplied records agree with the consumer's
source interpretation. It does not prove the producer executed those records.
A contradictory result or advertised total is REFUTED. An honestly incomplete
attempt remains UNKNOWN with a concrete stage, step, reason, class, next
operation and blocked-by frontier. Missing source reconstruction blocks later
checks explicitly; it never supplies a fabricated interpretation.

The consumer uses its existing raw-source parser and its own condition
evaluation, not the producer's compiler, interpreter, generator or executor.
The Gooo syntax frontend and the versioned JSON wire schema remain shared
assumptions, declared in the report. Wire declarations are consumer-owned;
neither the observer nor its tests import the producer implementation. Structural metrics, actual generated-program semantics, process
execution, wall-time accuracy, repository-wide writes, external utility and
causal improvement are not established by this mode. Its process-execution
claim remains UNKNOWN even when all six receipt checks close. Mutation and
promotion authority are zero.

The native test corpus includes a real generated baseline/candidate report,
13 named counterexamples, an actual failed-toolchain attempt, mode rejections, and
an actual consumer CLI path. The positive cohort has three paired inputs;
18 comparisons include each side's source interpretation, first result and
replay result. These are bounded observations, not language-completeness or
external-utility percentages. CI results, not this description, determine
whether the authored cases pass. Test inputs come from a separately built witness
process, reused within each fixture, not direct calls into the producer package.
A bootstrap execution supplies actual judge bytes before caller snapshots are
declared; the completed request is then executed through the same public CLI.
The missing-toolchain case changes only the fixture child process environment.
That counterexample does not change repository settings or installed tools.

The full-repository CI suite executes the consumer tests. Its non-verbose
successful output does not retain individual receipt logs. The source-owned
test logs emit `REVISION_RECEIPT_FIXTURE_PROCESS` and
`POLICY_REVISION_RECEIPT_OBSERVATION` for a JSON-enabled CI invocation.

Native run [34682470696](https://github.com/kimjooyoon/meta-ontology-go/actions/runs/34682470696)
retained those records at commit `71cf19fc2a2b1f4ea9b3eadc5726335ffac83dfd`.
That candidate's one-line CI capture change was withdrawn after the default
Guardian rejected its protected-workflow authorization. Its archived evidence
is historical, not acceptance of a later head or a permanently enabled capture
route. The original CI configuration and all consumer tests are preserved;
no Guardian exception, permission change or adoption claim is introduced.

## Source-bound revision execution observation

The witness has a separate experimental mode for executing an exact policy
revision without changing the existing three-fixture producer path:

~~~sh
meta-policy-compilation-witness \
  -policy policy.gooo \
  -observe-revision revision-observation-request.json
~~~

The request has `expected_source_digest`, `condition`, `from_decision`,
`to_decision` and a nonempty `cases` array. Each array entry contains a
`baseline` and `candidate` Case with the same unique ID and explicit evidence
class/provenance. These are caller declarations, not authenticated provenance.
Duplicate JSON keys, unknown fields, trailing documents, stale source identity,
duplicate/mismatched case IDs and mixed legacy-mode flags are rejected.

The result is one `gooo/meta-policy-revision-observation/v1` JSON document on
stdout. It retains the exact typed request, its canonical digest, the raw
request artifact digest, both compiled source contracts and their source-owned
rule/meta-operation bindings, the candidate Gooo source, and both generated Go
judges with their digests. The existing revision primitive edits the detached
first-class policy AST; this mode actually executes its generated consequence.

Each source version uses one generated-judge build invocation. Its declared
cases are then executed twice as fresh processes using that binary. First
results, replay results and producer-side source interpretations remain separate.
The record counts actual decoded pairs, source comparisons, replay comparisons,
mismatches and requested decision transitions. Process/build failure retains
the successful prefix, identifies its exact batch boundary and returns a
nonzero command exit. Missing outputs are not invented.

No case digest or validator expectation is rebound. A predecessor snapshot
supplied unchanged to the candidate remains stale. Separate caller-supplied
before/candidate snapshots remain different inputs, not a same-input causal or
performance comparison. `requested_transition_observed` reports only the
observed condition and decisions; `causal_attribution` remains `UNASSESSED`.

`execution_conformance=PASS` means generated/source agreement and replay
agreement for the supplied cases, not independent validation or admission.
A source/replay contradiction is `REFUTED`; failed execution is explicitly
`execution_status=FAILED` with its error and incomplete evidence. A request
whose target condition was not exercised retains a REQUEST_COVERAGE UNKNOWN.
Independent revision admission always remains UNKNOWN with stage, step, reason,
unknown_class, next_operation and an explicit blocked_by frontier. Neither
self-authored expectations nor successful replay closes that claim.

Wall time is an integer observation of each generated batch, not an improvement
claim. Improvement remains UNKNOWN. Repository-wide write observation and peak
RSS are not collected by this mode; it does not claim a measured zero write set
or zero memory. Mutation/promotion authority remains zero. Generated execution
uses the existing temporary judge workspaces, while the report is stdout-only.
This is not yet registration in the common meta-operation selector or automatic
adoption of the generated candidate.

## Gooo-bound policy revision observation

The explicit revision path can bind a caller-supplied Gooo operation contract
before invoking the existing bounded native worker:

```sh
go run ./cmd/meta-policy-compilation-witness \
  -policy /caller/input/policy.gooo \
  -observe-revision /caller/input/request.json \
  -revision-operation internal/meta/policycompilation/revision-operation.gooo \
  -profile-package metapolicycompilation \
  -profile-namespace metapolicycompilation
```

The contract declares PolicySource, RevisionRequest and RevisionObservation.
ObservePolicyDecisionRevision must use both inputs and generate the observation
in lowered semantic IR. Its computes program is the pinned
`policy.revision.observe:v1` native ABI, not an arbitrary program selected by
a Gooo string. Contract source bytes, semantic identity, embedded native contract,
policy source bytes and exact request bytes retain separate identities.

The binding rejects absent relations, unknown programs and incompatible semantic
contracts before calling the worker. Strict request decoding and exact policy-byte
binding also precede the call. A native invocation or cancelled attempt is not a
successful generated execution; the nested observation retains actual counts,
failure causes and the six-field UNKNOWN records.

With this explicit flag, stdout uses `gooo/meta-policy-revision-operation/v1`.
Its `observation` member remains the unchanged v1 producer receipt. That member
can be passed to the existing independent consumer together with the original
policy and request bytes. The new native CLI test exercises those two separate
processes and checks that all three caller inputs remain unchanged. Consumer
receipt consistency does not attest process execution or grant policy adoption.

Without the flag, the existing revision-observation output remains unchanged.
The four-artifact compilation profile, two-artifact revision profile, legacy
inventory selector and common registry are unchanged. This is an explicitly
requested single operation, not a fabricated second inventory action.

`native_worker_invocations` counts actual calls to the bounded native API.
Per-case/source/replay counts come from its existing observation, not a new score.
Synthetic case pairs remain synthetic, including separately declared candidate
expectations. Improvement and admission remain UNKNOWN; mutation and promotion
authority remain zero. This connects Gooo relations to execution but does not
close the independent adoption and next-run-use requirements in #804.
