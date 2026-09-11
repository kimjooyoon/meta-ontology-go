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
