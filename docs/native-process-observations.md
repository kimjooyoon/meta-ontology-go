# Native process observations

## Optional package wall diagnostics

`gooo package execute`, `replay` and `resume` accept `--timing`. The option writes
one `gooo/package-wall-diagnostic/v1` JSON line to stderr after the ordinary stdout
result is published. It reuses the bounded wall recorder used by body-path files.
The default path allocates no recorder and the phase hooks read no clock. Timing
does not enter the execution receipt, source identity, selection or admission.
Diagnostic writer errors leave the original execution outcome unchanged.

```sh
gooo package execute --timing --json --receipt-compression gzip \
  --cases examples/assembly-explainer/cases.json \
  examples/assembly-explainer/gooo.workspace.json \
  > /tmp/execution.json.gz 2> /tmp/execution-diagnostics.jsonl
```

The line binds the current execution source, manifest and case/input digests
when those are available. Read its `wall.phases`, `wall.capture_ns` and
`wall.unassigned_ns` separately from the primary receipt's execution decision.
stderr can also contain ordinary errors; only the line with this schema is the
timing observation. A later replay records fresh intervals rather than accepting
the saved execution's timings.

| Phase | Observed work |
| --- | --- |
| `composition_receipt` | Initial runtime identities and source/suite digests |
| `source_replay` | Reconstructing and verifying the saved composition |
| `composition_call_context` | Binding an explicit joint-call context, when present |
| `runtime_suite_prepare` | Resolving current native input rows |
| `go_tool_hash`, `toolchain_bind` | Binding the selected Go executable and toolchain |
| `native_projection` | Constructing source-backed native observation code |
| `executable_prepare` | Preparing or reusing the owned executable, including a build when needed |
| `native_prepare_1`, `native_prepare_2` | Resetting each execution's observation files |
| `native_run_1`, `native_run_2` | Waiting for each actual native execution |
| `runtime_decode_1`, `runtime_decode_2` | Reading, decoding, binding and comparing current native observations |
| `input_separation` | Comparing runtime inputs with previously observed selection inputs |
| `executor_gate` | Waiting for an owned executor, when that API is used |

Intervals are sequential wall observations, including scheduling and contention.
Existing process start/wait/build intervals are contained within these phases;
adding both would double-count time. Unassigned time includes workspace generation,
workspace-origin binding, receipt encoding/publication and bookkeeping. It is
measured remainder, with no claim that all of it can be removed. CPU and memory
retain their existing independent observations.

The recorder retains at most 32 phases. Multiple sequential compositions can
repeat phase names. Overflow, overlapping intervals or invalid measurements yield
`INCOMPLETE`; absent phases are unobserved. This diagnostic status does not change
program behavior or lower a required check. Phase outcome `completed` means the
measured operation returned normally, including a classification of `UNKNOWN`;
it does not establish correctness of all inputs. CLI argument rejection happens
before an execution diagnostic is produced.

## Individual native processes

Development source records process creation and termination waiting separately.
Each newly executed toolchain/build/native process has an optional `timing`
object alongside its existing total, CPU, output-digest and failure fields:

| Field | Meaning |
| --- | --- |
| `timing.start_ns` | Wall time inside the host's `exec.Cmd.Start` call |
| `timing.wait_ns` | Wall time after Start returns, including Wait and pipe I/O |
| `timing.wait_limit_ns` | Budget armed after a successful Start for a native run; omitted for ordinary toolchain/build calls and old records |
| `timing.deadline_remaining_ns` | Remaining caller deadline immediately before Start; omitted when no deadline exists |
| `timing.context_at_start_return` | `ACTIVE`, `CANCELED` or `DEADLINE_EXCEEDED` observed when Start returns |
| `failure_phase` | `START`, `WAIT` or `OUTPUT` for the host operation reporting the failure; omitted on success |

Start and wait times sum exactly to `wall_ns`. A failed Start has zero wait time.
A successful Start always calls Wait to release the child and its pipes.
Native execution arms a two-second wait budget after a successful Start. The
original caller deadline still covers queueing, toolchain checks, compilation,
host creation and execution; starting the wait budget never extends that
deadline. Native executor calls already cap this enclosing work at 60 seconds.
The one-second pipe wait delay and process-group cleanup continue to apply.
Total wall time includes both phases. Cancellation still calls Wait to join a
successfully created child, and a completed wait stops its timer.
Cancellation and pipe cleanup can make observed Wait time exceed its requested
budget; the observation retains that elapsed time.

This separates host creation from the native execution interval. Go's Start
call can itself return after a caller deadline; the context at return records
that fact and the child is joined without restoring the expired caller budget.
The two-second wait interval includes runtime startup, generated code, pipe I/O
and termination. It is not a measurement of CPU time or solely the user body.

A long Start interval locates a delay in host process creation. A long Wait
interval can include runtime startup, program execution, scheduling and pipe
handling. Finding when generated user code began requires further observation.
The context snapshot describes the host deadline at Start's return, while CPU
time continues to come from the child's process state.

Older receipts can have no `timing` object or failure phase. That means these
details were unobserved; their original totals and results remain meaningful.
Saved replay creates fresh process observations and preserves prior history.
Executor cache copies keep timing objects independent of the cached original.

## Release example failures

The native language release examples save successful output under the existing
`<platform>-<example>-construct.json` and `-replay.json` paths. A failed command
saves its exact combined stdout/stderr bytes under `.failed-output`, including
an empty file if the command produced no bytes. This allows partial JSON and
arbitrary diagnostic bytes to survive the failing CI step. Both command and
file-write errors remain available if persistence also fails. The platform
workflow already uploads its output directory on failure.

On 2026-10-09, two Windows readiness failures at compiler candidate `8951c5f8`
had empty output, zero recorded CPU time and an expired native deadline. Those
older records do not identify the Start/Wait boundary. These fields provide
evidence for that distinction in subsequent executions.

The [later Windows ledger run](https://github.com/kimjooyoon/meta-ontology-go/actions/runs/37883255912/job/113667493818)
at `aa79dc94` recorded Start at 2,854,721,800 ns, Wait at 510,000 ns, and
`DEADLINE_EXCEEDED` when Start returned. Its original two-second deadline had
already expired inside host creation. The same executable digest completed in
the separate readiness run. This locates the consumed budget in Start; it does
not identify the Windows component responsible for the delay. The subsequent
budget split addresses that boundary without repeating an execution or changing
the generated program. Original observations retain their original timings.
