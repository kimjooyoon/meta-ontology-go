# Track inputs across body construction and execution

Gooo first fills the two holes in `Lift`, then executes the `Normalize → Lift`
graph. The input-separation observation includes both that earlier construction
and the values actually delivered during native execution.

```sh
gooo package execute --json \
  --cases examples/workspace-input-observations/cases.json \
  examples/workspace-input-observations/gooo.workspace.json
```

Pass `--go /path/to/go1.27.2` when the Go executable on PATH differs from the
required toolchain. This example runs locally with deterministic construction.

`gooo package inspect --json <workspace.json>` reports `transport_capacity`
without constructing or executing bodies. Input JSON and its normalized suite
each have a 1 MiB aggregate limit; native composition stdout has a separate
1 MiB limit. Each decoded Text value remains limited to 1,024 UTF-8 bytes and a
request to 128 rows. JSON escapes count toward the aggregate wire limit: one
decoded NUL byte occupies six bytes as `\u0000`. These independent limits do not
promise acceptance of every combination of maximum fields, Text values and rows.
Oversized input is refused; values and rows are never shortened to fit. Recorded
origin pages retain their separate 32 KiB export limit. Provenance files and
saved receipts also retain their existing limits.

| Root input | Input delivered to Lift | Earlier observation | Classification |
| ---: | ---: | --- | --- |
| -9 | 0 | Construction case | Overlapping |
| 11 | 11 | None in the recorded construction inputs | Disjoint |
| 100 | 100 | Holdout case | Overlapping |
| 11 | 11 | Duplicate root input | Counted with the previous 11 |

Four rows specify eight activity outputs. The whole-case measurement has three
unique roots, one duplicate, two overlapping cases and one disjoint case. Its
finite fraction is 1/1 when all expectations for input 11 match. A conflicting
expectation on its duplicate changes that fraction to 0/1 and `PROGRESS`.

`result.runtime.input_separation.earlier_stages` binds each earlier body fill to
its activity, original source, plan and canonical input-set digest. Its inputs
include construction cases, holdout cases and generated behavioral probes.
The actual tuples remain in `result.body_fills`. Integer vectors and typed
record tuples retain their port order and exact integer values.

The same measurement supports source-declared fills and externally supplied
fill plans. Missing input evidence remains `UNKNOWN`. Input-only executions
keep their overlap counts but have no expected-output success count. Exposure
in a model's training data remains unknown.

## Reuse the constructed program

Save the first command's JSON as `execution.json`, then replay it with:

```sh
gooo package replay --json --receipt execution.json \
  --cases examples/workspace-input-observations/cases.json \
  examples/workspace-input-observations/gooo.workspace.json
```

Replay reconstructs the retained fill, executes the saved program twice on the
requested rows, and recomputes input separation from these fresh traces. The
earlier input-set identities stay the same. New runtime rows and expectations
determine this invocation's counts; historical accuracy counters are not reused.
Passing `--inputs` instead of `--cases` retains overlap observations and marks
expected-output evidence as unknown. Replay performs zero model calls.

Replay validates the saved construction against the current sources, then makes
new runtime observations. It does not certify the internal consistency of the
old runtime's redundant `actual` and `actual_fields` representations. A consumer
that uses the old observation needs to check that observation separately; a
successful fresh replay does not make a conflicting old display value valid.

The outer `replayed_from_sha256` identifies the exact supplied receipt bytes.
The inner `result.replay.prior_result_sha256` identifies Go's JSON serialization
of the decoded saved `result`, including its historical runtime fields. They
have different inputs and can differ. Both are content identities; neither is
an attestation of the old observation's correctness. `OBSERVED` with `--inputs`
continues to mean fresh actual values with no expected-output success count.

The [published run](../../docs/research/workspace-inputs-20261008/summary.json)
records clean compiler revision `093a8391`, the raw receipt, eight matched
activity outputs, one passing disjoint case, two native runs and zero model calls.

The [replay and explanation observation](../../docs/research/workspace-inputs-20261008/replay-summary.json)
uses clean compiler `139b932a` to read that earlier receipt. Fresh execution again
matches 8/8 activity outputs and 1/1 disjoint case, retaining the earlier
construction-input identities. The CLI takes 1.16 seconds in this single local
run; build-cache conditions were not controlled for a performance comparison.

Save replay's JSON as `replay.json` and let a Gooo tool read its construction:

```sh
gooo package execute --json --construction-receipt replay.json \
  examples/assembly-explainer/gooo.workspace.json
```

Gooo's source-defined rules return `OBSERVE_NEW_INPUTS` for the candidate
matching 3/3 construction cases and `USE_OBSERVED_CANDIDATE` for the candidate
matching 0/3. The explanation command took 0.38 seconds and ran its native
program twice. Both commands made zero model calls. The explanation is
`OBSERVED`: it returns suggestions as data, with no expected-output accuracy
denominator and no automatic repair action.
