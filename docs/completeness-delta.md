# Comparing completeness observations

`gooo completeness-delta --before before.json --after after.json` records what
changed between two shared completeness receipts. Inputs may be standalone
receipts, `body-codegen --json` results, or `body-execute` results. The operation
reads bounded local JSON and emits one deterministic JSON report to stdout.
It performs zero model calls, external requests, generated-program executions,
or repository writes.

The comparison structure is declared in
[`delta.gooo`](../internal/completenessdelta/delta.gooo), which generates its Go
types and JSON Schema. The existing completeness receipt declaration retains
its original digest, so previous published observations remain readable.

## What the result means

Both original receipts are preserved, including their first unresolved claim,
evidence, exclusions, and failure reasons. Exact input and selected receipt byte
digests are recorded separately. Scope changes have sorted JSON Pointer paths.
Axes retain their original order, followed by newly added axes in after-order.

| Observation | Recorded behavior |
| --- | --- |
| Same registered profile, source, plan, suite, evaluator and investment | `SAME_MEASUREMENT_SCOPE` |
| Same nonzero denominator and unit; both states PASS or PROGRESS | Exact numerator direction and integer magnitude |
| UNKNOWN becomes observed | `BECAME_OBSERVED`, numeric delta stays null |
| An observed axis becomes UNKNOWN | `OBSERVATION_LOST` |
| A failure appears | Failure transition and original cause |
| An axis is added or removed | Explicit `ADDED` / `REMOVED`; removed obligations are flagged |
| Denominator, unit, profile or measurement binding changes | Both values retained; numeric comparison withheld |
| Generation receipt is the exact parent of a runtime receipt | `PARENT_RUNTIME_CONTINUATION`, new observations retained |

For a before/after counter decrease under comparable conditions,
`regression=COUNT_REGRESSION`. Other regression labels describe observation loss,
new failure, and removed obligations. These are categories with separate counts.
The report carries no aggregate completeness percentage.

Report `status=PASS` means the registered measurement scope matched;
`PROGRESS` means a continuation or incomparable scope was recorded. A successful
CLI exit means a comparison was emitted. Inspect the original product decisions
and dimension changes to understand the outcome.

## Run generation, execution, and comparison

With Go 1.27.1 available as `go`, from the compiler checkout:

```sh
go run ./cmd/gooo body-codegen --json --activity Combined \
  --path-plan examples/body-codegen/typed-path-compound-plan.json \
  examples/body-codegen/typed-path-compound.gooo.fixture > /tmp/gooo-generation.json
go run ./cmd/gooo body-execute \
  --source examples/body-codegen/typed-path-compound.gooo.fixture \
  --path-plan examples/body-codegen/typed-path-compound-plan.json \
  --generation /tmp/gooo-generation.json \
  --cases examples/body-codegen/typed-path-runtime-cases.json \
  --go-bin "$(command -v go)" > /tmp/gooo-runtime.json
go run ./cmd/gooo completeness-delta \
  --before /tmp/gooo-generation.json --after /tmp/gooo-runtime.json
```

The typed-path generation and runtime profiles are the initial supported scope
contracts. Different compiler/model/context identities are retained as scope
changes, allowing descriptive comparisons of observations under the same task.
Attributing a change to a model requires a separately controlled experiment.
Runtime-to-runtime comparisons additionally require the same runtime suite,
activity identity and suite authority.

The registered runtime profiles are `gooo/typed-path-runtime-v1` (fresh execution)
and `gooo/typed-path-runtime-v2` (one owned executable). The latter records the
original source-bound build separately from work performed by the current call.
Both envelopes preserve the exact parent receipt. A v2 envelope also requires a
matching registered owned-artifact record, observation digest and original build
binding. Failed observations remain readable when no successful build exists.
The v2 bare receipt must retain its ownership contract. Future profiles are
unregistered and do not become comparable by sharing a name prefix.

File-based execution saves comparison inputs directly:

```sh
gooo body-path-run \
  --source examples/body-codegen/typed-path-compound.gooo.fixture \
  --activity Combined --path-plan examples/body-codegen/typed-path-compound-plan.json \
  --cases examples/body-codegen/typed-path-runtime-cases.json \
  --repeat 2 --out body-comparison-results
gooo completeness-delta --before body-comparison-results/run-1-generation.json \
  --after body-comparison-results/run-1-runtime.json
gooo completeness-delta --before body-comparison-results/run-1-runtime.json \
  --after body-comparison-results/run-2-runtime.json
```

Generation-to-runtime is a parent continuation with no numeric delta across
profiles. First-build-to-reuse v2 observations can have the same finite expectation
scope. Their current-child resource axes have different units and denominators
(four children versus three with the earlier build excluded), so those axes retain
both observations without a numeric improvement. Changed runtime expectations
change the suite digest and are incomparable. The reader performs no new runtime
execution or model operation, and JSON consistency does not attest past execution.

## Input and evidence boundaries

Inputs are limited to 2 MiB each; each shared receipt remains limited to 1 MiB.
Invalid UTF-8, duplicate keys, trailing JSON, excessive nesting, ambiguous producer
paths and invalid shared accounting are rejected before output. Runtime envelopes
retain their base64 parent receipt and must agree on its exact-byte hash.

The command compares caller-provided observations. Schema validation and hash
consistency establish what was compared. Actual execution evidence comes from
the source replay, tool and runtime captures supplied by producers. The command
grants no host permission and makes no new claim about earlier model execution.

Natural-language discovery, other profiles, independently observed workflows,
and causal quality attribution remain development work under
[issue #1023](https://github.com/kimjooyoon/meta-ontology-go/issues/1023).
