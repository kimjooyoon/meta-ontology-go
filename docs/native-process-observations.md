# Native process observations

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
