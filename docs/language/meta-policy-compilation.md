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
Case decoding and generated-judge behavior are unchanged.

Availability, provenance, evidence classifications, and observed digests remain
caller declarations. The result is conditional source-policy evaluation:
external evidence is UNKNOWN/not verified, full conformance is UNKNOWN/not
executed, generated execution is false, and mutation/promotion authority is zero.
It does not replace the existing source/generated/independent conformance receipt.
The public command admission and end-user route remain separate unfinished work.

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
