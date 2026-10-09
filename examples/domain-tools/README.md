# Two Gooo tools, one report interface

These tools interpret observations supplied by a caller:

| Domain | Input meaning | Example next operations |
| --- | --- | --- |
| Code assembly | Matched cases, total cases, evaluated candidates and budget | Evaluate candidates, continue, expand choices, observe new inputs |
| Documentation | Declared, implemented and documented features; stale examples | Implement declarations, document missing features, update examples |

Each domain owns its record IDs, fields, conditions, alternatives and construction
cases in a Gooo source. `report.gooo.fixture` defines a shared result contract;
`render.gooo.fixture` formats that contract. Both files are unchanged when the
manifest substitutes the policy package. The compiler checks imports, stable type
IDs and the explicit `Assess -> Render` binding.

## Run a tool

From this repository, use Go 1.27.2 to build the CLI:

```sh
go build -o /tmp/gooo-domain-tools ./cmd/gooo
/tmp/gooo-domain-tools package execute --json \
  --inputs examples/domain-tools/assembly.inputs.json \
  examples/domain-tools/assembly.workspace.json
/tmp/gooo-domain-tools package execute --json \
  --inputs examples/domain-tools/documentation.inputs.json \
  examples/domain-tools/documentation.workspace.json
```

If the native build would find another Go version on PATH, add
`--go /path/to/go1.27.2`. Outputs and inputs appear in
`result.runtime.traces`. Input-only runs are `OBSERVED`, with no scored runtime
expectations. The state inside each returned report is the Gooo tool's domain
judgment. For example, a returned `UNKNOWN` explains a missing feature roster;
it is distinct from the CLI's execution status.

The tools use caller-supplied observations. They do not collect candidate scores,
inspect a repository or establish whether its documents match its implementation.
An adapter can supply measured counts and staleness to these Gooo policies.

## Reproduce the finite checks

```sh
/tmp/gooo-domain-tools package execute --json \
  --cases examples/domain-tools/assembly.cases.json \
  examples/domain-tools/assembly.workspace.json
/tmp/gooo-domain-tools package execute --json \
  --cases examples/domain-tools/documentation.cases.json \
  examples/domain-tools/documentation.workspace.json
```

Each suite has seven input records and two expected activity outputs per record.
The inputs cover missing observations, partial results, completion, invalid counts
and integers above 2^53. They are disjoint from the three construction inputs in
each source. This separation describes the declared construction cases; model
training exposure is independently unknown.

## Replace the construction ordering

Each policy has three binary field-update choices and a budget of eight candidates.
The assembly policy needs all three alternatives; the documentation policy retains
its middle original assignment. This includes a choice where keeping existing code
satisfies the declared cases.

Run the same command with
`--assembly-model /path/to/shared-qat/model.json` to use the existing
[own compact model](https://huggingface.co/asketeddy/gooo-record-shared-field-tiny-v1/tree/5800946afb35506d66357ee3ea6f956f506795b9/models/qat_ternary).
Omit the flag for deterministic ordering. The source, candidates, case files,
compiler and attempt budget remain fixed within each comparison. Retain each
JSON result to inspect model prediction, attempted masks, scores, selected source
and native outputs. A model prediction is a suggested starting point; Gooo scores
the complete candidate and continues within the declared budget.

This demonstrates substituting two explicit domain declarations within today's
supported record/body profile. Their different input identities and rules remain
visible. Supplying assembly observations to the documentation profile fails its
input contract. The experiment is a bounded basis for expanding reusable domain
tools and their language support.

## Recorded construction

The [paired observation](../../docs/research/domain-tools-20261007/summary.json)
contains the compiler and model identities, all four raw receipts and timing scope.

| Domain | Deterministic attempts | Model attempts | Model prediction | Final native outputs |
| --- | ---: | ---: | --- | --- |
| Code assembly | 8 | 1 | Mask 7 satisfied 3/3 construction cases | 14/14 in both routes |
| Documentation | 6 | 2 | Mask 7 satisfied 2/3; mask 5 then satisfied 3/3 | 14/14 in both routes |

The documentation model assigned probability 0.99983287 to its first proposal,
which changed the required `DOCUMENT_MISSING` action. The finite checks caught
that mismatch. Model scores describe the ranking output; they are not calibrated
probabilities that the tool fulfills its task.

Model prediction took 20,834 and 21,500 ns; setup took 0.741083 and 0.196458 ms.
The loaded tensor storage was 2,096 bytes. The entire commands, including native
build/execution, reported about 86 MB maximum RSS. The summary retains exact
process measurements; host CPU utilization was not sampled. The runs were single
sequential pairs with uncontrolled caches, so this observation establishes the
candidate counts and finite outcomes rather than a general speed advantage.

## Let another Gooo tool explain the recorded attempts

The [assembly explainer](../assembly-explainer/README.md#interpret-an-actual-construction-receipt)
can consume these records directly with `package execute --construction-receipt`.
The adapter reconstructs the candidate observations, then Gooo interprets each
attempt using its source-defined rules. The saved documentation-model run yields
“continue candidates” for its first 2/3 result and “observe new inputs” for the
following 3/3 result, with zero additional model calls.
